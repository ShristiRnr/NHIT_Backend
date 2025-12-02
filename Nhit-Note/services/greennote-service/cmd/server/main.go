package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	greennotepb "nhit-note/api/pb/greennotepb"
	eventsadapter "nhit-note/services/greennote-service/internal/adapters/events"
	grpcadapter "nhit-note/services/greennote-service/internal/adapters/grpcadapter"
	memoryrepo "nhit-note/services/greennote-service/internal/adapters/repository/memory"
	sqlcrepo "nhit-note/services/greennote-service/internal/adapters/repository/sqlc"
	storageadapter "nhit-note/services/greennote-service/internal/adapters/storage"
	"nhit-note/services/greennote-service/internal/core/ports"
	"nhit-note/services/greennote-service/internal/core/services"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

// Config holds runtime configuration for the GreenNote service.
type Config struct {
	GRPCPort string
	HTTPPort string

	KafkaBrokers       []string
	KafkaTopicApproved string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	DatabaseURL         string
	AuthServiceEndpoint string
}

func loadConfigFromEnv() Config {
	cfg := Config{
		GRPCPort:            getenvWithDefault("GREENNOTE_GRPC_PORT", "50059"),
		HTTPPort:            getenvWithDefault("GREENNOTE_HTTP_PORT", "8080"),
		KafkaTopicApproved:  getenvWithDefault("GREENNOTE_KAFKA_TOPIC_APPROVED", "greennote.approved"),
		MinIOEndpoint:       os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey:      os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:      os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:         getenvWithDefault("MINIO_BUCKET", "greennote-docs"),
		DatabaseURL:         strings.TrimSpace(getenvWithDefault("GREENNOTE_DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")),
		AuthServiceEndpoint: getenvWithDefault("AUTH_SERVICE_ENDPOINT", "localhost:50052"),
	}

	if brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if useSSL := strings.TrimSpace(os.Getenv("MINIO_USE_SSL")); useSSL != "" {
		b, err := strconv.ParseBool(useSSL)
		if err == nil {
			cfg.MinIOUseSSL = b
		}
	}

	return cfg
}

func getenvWithDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := loadConfigFromEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Set up document storage (MinIO or in-memory fallback).
	docStorage, err := storageadapter.NewDocumentStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatalf("failed to initialize document storage: %v", err)
	}

	// Choose repository implementation: Postgres/sqlc when GREENNOTE_DATABASE_URL is set,
	// otherwise fall back to in-memory repository (useful for local development/tests).
	var repo ports.GreenNoteRepository
	if cfg.DatabaseURL != "" {
		db, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to open Postgres connection: %v", err)
		}
		db.SetMaxIdleConns(5)
		db.SetMaxOpenConns(20)
		db.SetConnMaxLifetime(time.Hour)

		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("failed to ping Postgres: %v", err)
		}

		repo = sqlcrepo.NewPostgresGreenNoteRepository(db, docStorage)
		log.Println("using Postgres/sqlc GreenNote repository")
	} else {
		repo = memoryrepo.NewInMemoryGreenNoteRepository(docStorage)
		log.Println("using in-memory GreenNote repository")
	}

	// Set up Kafka event publisher (or noop if brokers not configured).
	events := eventsadapter.NewEventPublisher(cfg.KafkaBrokers, cfg.KafkaTopicApproved)

	// Core application service.
	appService := services.NewGreenNoteService(repo, events)

	// gRPC service adapter.
	grpcSvc := grpcadapter.NewGreenNoteGRPCServer(appService)

	grpcErrCh := make(chan error, 1)
	httpErrCh := make(chan error, 1)

	go func() {
		grpcErrCh <- runGRPCServer(ctx, cfg, grpcSvc)
	}()

	go func() {
		httpErrCh <- runHTTPGatewayServer(ctx, cfg, grpcSvc, docStorage)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received, waiting for servers to stop...")
		select {
		case err := <-grpcErrCh:
			if err != nil {
				log.Printf("gRPC server error during shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			log.Println("gRPC server shutdown timeout")
		}
		select {
		case err := <-httpErrCh:
			if err != nil {
				log.Printf("HTTP gateway error during shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			log.Println("HTTP gateway shutdown timeout")
		}
	case err := <-grpcErrCh:
		log.Fatalf("gRPC server exited with error: %v", err)
	case err := <-httpErrCh:
		log.Fatalf("HTTP gateway exited with error: %v", err)
	}
}

func runGRPCServer(ctx context.Context, cfg Config, svc greennotepb.GreenNoteServiceServer) error {
	addr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	greennotepb.RegisterGreenNoteServiceServer(server, svc)

	go func() {
		<-ctx.Done()
		log.Println("stopping gRPC server...")
		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			log.Println("force-stopping gRPC server")
			server.Stop()
		}
	}()

	log.Printf("gRPC server listening on %s", addr)
	return server.Serve(lis)
}

func runHTTPGatewayServer(ctx context.Context, cfg Config, svc greennotepb.GreenNoteServiceServer, docs ports.DocumentStorage) error {
	gatewayMux := runtime.NewServeMux()

	if err := greennotepb.RegisterGreenNoteServiceHandlerServer(ctx, gatewayMux, svc); err != nil {
		return err
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/", gatewayMux)
	rootMux.Handle("/api/v1/green-notes/documents/upload", makeDocumentUploadHandler(docs))
	rootMux.Handle("/additional-document/upload", makeFixedSectionUploadHandler(docs, "additional", 5<<20))
	rootMux.Handle("/extension-document/upload", makeFixedSectionUploadHandler(docs, "extension", 10<<20))
	rootMux.Handle("/deviation-document/upload", makeFixedSectionUploadHandler(docs, "deviation", 10<<20))

	addr := ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:    addr,
		Handler: rootMux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP gateway shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP gateway listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func makeDocumentUploadHandler(docs ports.DocumentStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form; limit total memory usage.
		if err := r.ParseMultipartForm(12 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		noteIDStr := strings.TrimSpace(r.FormValue("note_id"))
		section := strings.ToLower(strings.TrimSpace(r.FormValue("section")))
		if noteIDStr == "" || section == "" {
			http.Error(w, "note_id and section are required", http.StatusBadRequest)
			return
		}
		noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
		if err != nil || noteID <= 0 {
			http.Error(w, "note_id must be a positive integer", http.StatusBadRequest)
			return
		}

		var maxBytes int64
		switch section {
		case "additional":
			maxBytes = 5 << 20 // 5 MB
		case "extension", "deviation":
			maxBytes = 10 << 20 // 10 MB
		default:
			http.Error(w, "invalid section", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limited := io.LimitReader(file, maxBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		if int64(len(data)) > maxBytes {
			http.Error(w, fmt.Sprintf("file too large; limit is %d bytes", maxBytes), http.StatusRequestEntityTooLarge)
			return
		}

		filename := header.Filename
		if filename == "" {
			filename = "upload.bin"
		}
		filename = sanitizeFilename(filename)
		objectName := fmt.Sprintf("greennotes/%d/%s/%d-%s", noteID, section, time.Now().UnixNano(), filename)
		contentType := header.Header.Get("Content-Type")

		if err := docs.Save(r.Context(), objectName, data, contentType); err != nil {
			http.Error(w, "failed to store document", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"success":     true,
			"message":     "document uploaded",
			"object_name": objectName,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func makeFixedSectionUploadHandler(docs ports.DocumentStorage, section string, maxBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(12 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		noteIDStr := strings.TrimSpace(r.FormValue("note_id"))
		if noteIDStr == "" {
			http.Error(w, "note_id is required", http.StatusBadRequest)
			return
		}
		noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
		if err != nil || noteID <= 0 {
			http.Error(w, "note_id must be a positive integer", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limited := io.LimitReader(file, maxBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		if int64(len(data)) > maxBytes {
			http.Error(w, fmt.Sprintf("file too large; limit is %d bytes", maxBytes), http.StatusRequestEntityTooLarge)
			return
		}

		filename := header.Filename
		if filename == "" {
			filename = "upload.bin"
		}
		filename = sanitizeFilename(filename)
		objectName := fmt.Sprintf("greennotes/%d/%s/%d-%s", noteID, section, time.Now().UnixNano(), filename)
		contentType := header.Header.Get("Content-Type")

		if err := docs.Save(r.Context(), objectName, data, contentType); err != nil {
			http.Error(w, "failed to store document", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"success":     true,
			"message":     "document uploaded",
			"object_name": objectName,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}
