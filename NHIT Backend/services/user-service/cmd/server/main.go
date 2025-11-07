package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/services"
	"google.golang.org/grpc"
)

func main() {
	// Get configuration from environment
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "50051"
	}
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5432/nhit_db?sslmode=disable"
	}

	log.Printf("🚀 Starting User Service on port %s", port)

	// Connect to database using pgxpool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize sqlc queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	userRoleRepo := repository.NewUserRoleRepository(queries)

	// Initialize services
	userService := services.NewUserService(userRepo, userRoleRepo)

	// Initialize gRPC handlers
	userHandler := grpcHandler.NewUserHandler(userService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userpb.RegisterUserManagementServer(grpcServer, userHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ User Service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
