package main

import (
	"log"
	"net"
	"os"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/kafka"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/repository"
	orgConfig "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	log.Println("✅ Repositories initialized")

	// Connect to Auth Service for token validation
	authConn, err := grpc.Dial(cfg.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to auth service at %s: %v", cfg.AuthServiceURL, err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	log.Println("✅ Connected to Auth Service", cfg.AuthServiceURL)

	// Initialize Kafka publisher (real implementation)
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "organization.events"

	kafkaPublisher, err := kafka.NewRealKafkaPublisher(kafkaBrokers, kafkaTopic, nil)
	if err != nil {
		log.Printf("⚠️ Failed to initialize real Kafka publisher, falling back to mock: %v", err)
		kafkaPublisher = kafka.NewMockKafkaPublisher(nil)
	} else {
		log.Println("✅ Real Kafka publisher initialized")
	}
	defer kafkaPublisher.Close()

	// Initialize MinIO client for logos
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}
	minioBucket := os.Getenv("MINIO_BUCKET_LOGOS")
	if minioBucket == "" {
		minioBucket = "logos"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, useSSL)
	if err != nil {
		log.Printf("⚠️ Failed to initialize MinIO client for organizations: %v", err)
	}

	// Initialize gRPC handlers (Adapters Layer) and pass DB pool, auth client, kafka publisher and minio client
	orgHandler := grpcHandler.NewOrganizationHandler(orgRepo, conn, authClient, kafkaPublisher, minioClient)
	log.Println("✅ gRPC handlers initialized")

	// Initialize RBAC interceptor
	rbac := middleware.NewRBACInterceptor(authClient)
	
	// Register permissions
	for method, perms := range orgConfig.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	
	// Register public methods
	for _, method := range orgConfig.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	log.Println("✅ RBAC interceptor initialized with permissions")

	// Create gRPC server with RBAC interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rbac.UnaryServerInterceptor()),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
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
