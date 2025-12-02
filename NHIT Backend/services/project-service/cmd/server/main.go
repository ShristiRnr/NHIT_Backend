package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/grpc/handler"
	kafkaAdapter "github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/kafka"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/services"
)

func main() {
	// Load configuration from environment variables
	grpcPort := getEnv("GRPC_PORT", "50057")
	httpPort := getEnv("HTTP_PORT", "8057")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize repository
	projectRepo := repository.NewProjectRepository(db)

	// Initialize Kafka consumer (real implementation)
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "organization.events"
	kafkaGroupID := "project-service-group"

	kafkaConsumer, err := kafkaAdapter.NewRealKafkaConsumer(kafkaBrokers, kafkaTopic, kafkaGroupID, nil)
	if err != nil {
		log.Printf("⚠️ Failed to initialize real Kafka consumer, falling back to mock: %v", err)
		kafkaConsumer = kafkaAdapter.NewMockKafkaConsumer(nil)
	} else {
		log.Println("✅ Real Kafka consumer initialized")
	}
	defer kafkaConsumer.Close()

	// Initialize service
	projectService := services.NewProjectService(projectRepo, kafkaConsumer, nil)

	// Initialize gRPC handler
	projectHandler := handler.NewProjectHandler(projectService)

	// Start event consumer in background
	go func() {
		if err := projectService.StartEventConsumer(context.Background()); err != nil {
			log.Printf("⚠️ Event consumer error: %v", err)
		}
	}()

	log.Println("✅ Service components initialized")

	// Start gRPC server
	go func() {
		if err := startGRPCServer(grpcPort, projectHandler); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP gateway
	go func() {
		if err := startHTTPGateway(httpPort, grpcPort); err != nil {
			log.Fatalf("Failed to start HTTP gateway: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Project Service...")
}

func startGRPCServer(port string, handler pb.ProjectServiceServer) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProjectServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	log.Printf("🚀 gRPC server listening on port %s", port)
	return grpcServer.Serve(lis)
}

func startHTTPGateway(httpPort, grpcPort string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb.RegisterProjectServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:"+grpcPort,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	server := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 HTTP gateway listening on port %s", httpPort)
	return server.ListenAndServe()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
