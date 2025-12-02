package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Configuration
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	port := getEnv("PORT", "50055")

	log.Printf("🚀 Starting Designation Service on port %s", port)

	// Connect to database using pgxpool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize SQLC queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repository
	designationRepo := repository.NewDesignationRepository(queries)

	// Initialize service
	designationService := services.NewDesignationService(designationRepo)

	// Initialize gRPC handler
	designationHandler := grpcHandler.NewDesignationHandler(designationService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	designationpb.RegisterDesignationServiceServer(grpcServer, designationHandler)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Designation Service listening on port %s", port)
	log.Printf("📡 gRPC server ready")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	// Start serving
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
