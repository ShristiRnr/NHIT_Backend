package main

import (
	"log"
	"net"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/services"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("department-service")
	log.Printf("🚀 Starting %s on port %s", cfg.ServiceName, cfg.ServerPort)

	// Connect to database
	conn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize sqlc queries
	queries := db.New(conn)

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
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Department Service listening on %s", cfg.ServerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
