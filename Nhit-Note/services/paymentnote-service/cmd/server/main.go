package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"nhit-note/services/paymentnote-service/internal/adapters/grpc/handler"
	"nhit-note/services/paymentnote-service/internal/adapters/repository"
	"nhit-note/services/paymentnote-service/internal/core/services"
	"nhit-note/services/paymentnote-service/internal/storage"
	paymentnotepb "nhit-note/api/pb/paymentnotepb"
)

func main() {
	// Load environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "nhit_payment_notes")
	grpcPort := getEnv("GRPC_PORT", "50053")

	// MinIO configuration
	minioEndpoint := getEnv("MINIO_ENDPOINT", "play.min.io:9443")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "")
	minioBucket := getEnv("MINIO_BUCKET", "payment-documents")
	minioUseSSL := getEnv("MINIO_USE_SSL", "true") == "true"

	// Database connection
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// ✅ Optimized connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)
	
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize MinIO client (optional - can be nil if not configured)
	var minioClient *storage.MinIOClient
	if minioAccessKey != "" && minioSecretKey != "" {
		minioClient, err = storage.NewMinIOClient(
			minioEndpoint,
			minioAccessKey,
			minioSecretKey,
			minioBucket,
			minioUseSSL,
		)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to initialize MinIO client: %v", err)
			log.Println("⚠️  Document upload will not be available")
		} else {
			log.Printf("✅ Connected to MinIO at %s (bucket: %s)", minioEndpoint, minioBucket)
		}
	} else {
		log.Println("⚠️  MinIO credentials not provided - document upload disabled")
	}

	// Initialize repository with MinIO client
	paymentNoteRepo := repository.NewPaymentNoteRepository(db, minioClient)
	log.Println("✅ Repository initialized")

	// Initialize service
	paymentNoteService := services.NewPaymentNoteService(paymentNoteRepo)
	log.Println("✅ Service layer initialized")

	// Initialize gRPC handler
	paymentNoteHandler := handler.NewPaymentNoteHandler(paymentNoteService)
	log.Println("✅ gRPC handler initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register payment note service
	paymentnotepb.RegisterPaymentNoteServiceServer(grpcServer, paymentNoteHandler)
	log.Println("✅ Payment Note Service registered")

	// Register reflection service for grpcurl
	reflection.Register(grpcServer)
	log.Println("✅ gRPC reflection registered")

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("\n🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
		db.Close()
		log.Println("✅ Server stopped")
		os.Exit(0)
	}()

	log.Printf("🚀 Payment Note Service listening on port %s", grpcPort)
	log.Println("=====================================")
	log.Println("Service Features:")
	log.Println("- Payment Note CRUD operations")
	log.Println("- Financial calculations (TDS, Net Payable, Number-to-Words)")
	if minioClient != nil {
		log.Println("- Document upload/download (MinIO)")
	}
	log.Println("- Draft workflow management")
	log.Println("- Hold/Unhold operations")
	log.Println("- UTR tracking")
	log.Println("- Approval logs & comments")
	log.Println("=====================================")
	log.Printf("Test with: grpcurl -plaintext localhost:%s paymentnote.PaymentNoteService/GeneratePaymentNoteOrderNumber", grpcPort)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
