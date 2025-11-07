package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/services"
	"google.golang.org/grpc"
)

func main() {
	// Get configuration from environment
	port := os.Getenv("DEPARTMENT_SERVICE_PORT")
	if port == "" {
		port = "50054"
	}
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5432/nhit_db?sslmode=disable"
	}

	log.Printf("🚀 Starting Department Service on port %s", port)

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
	departmentRepo := repository.NewDepartmentRepository(queries)

	// Initialize services
	departmentService := services.NewDepartmentService(departmentRepo)

	// Initialize gRPC handlers
	departmentHandler := grpcHandler.NewDepartmentHandler(departmentService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	departmentpb.RegisterDepartmentServiceServer(grpcServer, departmentHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Department Service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
