package main

import (
	"log"
	"net"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("organization-service")
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

	// TODO: Register organization service

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Organization Service listening on %s", cfg.ServerPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
