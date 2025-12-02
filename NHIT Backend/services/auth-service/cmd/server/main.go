package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/notifier"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/organization"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"

	_ "github.com/lib/pq"
	grpcMiddleware "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load configuration from environment
	serviceName := "auth-service"
	serverPort := os.Getenv("AUTH_SERVICE_PORT")
	if serverPort == "" {
		serverPort = "50052"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	}

	log.Printf("🚀 Starting %s on port %s", serviceName, serverPort)

	// Connect to database
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Test connection
	if err := conn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connected")

	// Connect to User Service
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051" // Default User Service address
	}

	userConn, err := grpcMiddleware.Dial(userServiceAddr, grpcMiddleware.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()

	userServiceClient := userpb.NewUserManagementClient(userConn)
	log.Printf("✅ Connected to User Service at %s", userServiceAddr)

	// Connect to Organization Service
	orgServiceAddr := os.Getenv("ORGANIZATION_SERVICE_ADDR")
	if orgServiceAddr == "" {
		orgServiceAddr = "localhost:8082" // Default Organization Service address
	}

	orgConn, err := grpcMiddleware.Dial(orgServiceAddr, grpcMiddleware.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Organization Service: %v", err)
	}
	defer orgConn.Close()

	orgServiceClient := organizationpb.NewOrganizationServiceClient(orgConn)
	orgClientAdapter := organization.NewOrganizationClient(orgServiceClient)
	log.Printf("✅ Connected to Organization Service at %s", orgServiceAddr)

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository(conn)
	refreshTokenRepo := repository.NewRefreshTokenRepository(conn)
	passwordResetRepo := repository.NewPasswordResetRepository(conn)
	emailVerificationRepo := repository.NewEmailVerificationRepository(conn)
	userRepo := repository.NewUserRepository(conn)

	// Initialize JWT manager
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default for development
		log.Println("⚠️  Warning: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	accessTokenDuration := 2 * time.Hour       // 2 hours as requested
	refreshTokenDuration := 7 * 24 * time.Hour // 7 days

	jwtManager := utils.NewJWTManager(jwtSecret, accessTokenDuration, refreshTokenDuration)

	// Initialize email service (mock for now)
	emailService := utils.NewMockEmailService()

	// Initialize notification client
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:50060"
	}
	notificationClient := notifier.NewRealNotificationClient(notificationServiceURL)
	log.Printf("✅ Initialized notification client with URL: %s", notificationServiceURL)

	// Initialize auth service with User Service client
	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		refreshTokenRepo,
		passwordResetRepo,
		emailVerificationRepo,
		jwtManager,
		emailService,
		nil,                // TODO: inject Kafka publisher implementation
		notificationClient, // Pass notification client
		userServiceClient,  // Pass User Service gRPC client
		orgClientAdapter,   // Pass Organization Service client
	)

	// Initialize auth interceptor (middleware)
	authInterceptor := middleware.NewAuthInterceptor(authService)

	// Create gRPC server with interceptors
	grpcServer := grpcMiddleware.NewServer(
		grpcMiddleware.UnaryInterceptor(authInterceptor.Unary()),
		grpcMiddleware.StreamInterceptor(authInterceptor.Stream()),
	)

	// Initialize and register auth handler
	authHandler := grpc.NewAuthHandler(authService, orgClientAdapter)
	authpb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Auth Service listening on port %s", serverPort)
	log.Printf("📧 Email service: Mock (for development)")
	log.Printf("🔐 JWT: Access token expires in %v, Refresh token expires in %v", accessTokenDuration, refreshTokenDuration)
	log.Println("🎉 Auth Service is ready!")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
