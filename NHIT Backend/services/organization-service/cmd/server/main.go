package main

import (
	"log"
	"net"

	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("organization-service")
	log.Printf("🚀 Starting %s on port %s", cfg.ServiceName, cfg.ServerPort)

	// Connect to database
	conn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer conn.Close()
	log.Println("✅ Database connection established")

	// Initialize repositories (Adapters Layer)
	orgRepo := repository.NewOrganizationRepository(conn)
	userOrgRepo := repository.NewUserOrganizationRepository(conn)
	log.Println("✅ Repositories initialized")

	// Initialize services (Domain/Core Layer)
	orgService := services.NewOrganizationService(orgRepo, userOrgRepo)
	userOrgService := services.NewUserOrganizationService(orgRepo, userOrgRepo)
	log.Println("✅ Business services initialized")

	// Initialize gRPC handlers (Adapters Layer)
	orgHandler := grpcHandler.NewOrganizationHandler(orgService, userOrgService)
	log.Println("✅ gRPC handlers initialized")

	// Create gRPC server with options
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	)

	// Register organization service
	organizationpb.RegisterOrganizationServiceServer(grpcServer, orgHandler)
	log.Println("✅ Organization service registered")

	// Register reflection service (for tools like grpcurl)
	reflection.Register(grpcServer)
	log.Println("✅ gRPC reflection registered")

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", cfg.ServerPort, err)
	}

	log.Printf("✅ Organization Service listening on port %s", cfg.ServerPort)
	log.Println("=====================================")
	log.Println("Service Architecture: Hexagonal (Ports & Adapters)")
	log.Println("- Domain Layer: Business logic and domain models")
	log.Println("- Ports Layer: Interfaces for services and repositories")
	log.Println("- Adapters Layer: gRPC handlers and PostgreSQL repositories")
	log.Println("=====================================")
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}
