package main

import (
	"log"
	"net"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/services"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("user-service")
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
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ User Service listening on %s", cfg.ServerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
