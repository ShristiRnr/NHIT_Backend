package main

import (
	"log"
	"net"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("auth-service")
	log.Printf("🚀 Starting %s on port %s", cfg.ServiceName, cfg.ServerPort)

	// Connect to database
	conn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize sqlc queries
	queries := db.New(conn)

	// TODO: Initialize repositories and services
	_ = queries

	// Create gRPC server
	grpcServer := grpc.NewServer()
	
	// TODO: Register auth service
	_ = authpb.AuthServiceServer(nil)

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Auth Service listening on %s", cfg.ServerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
