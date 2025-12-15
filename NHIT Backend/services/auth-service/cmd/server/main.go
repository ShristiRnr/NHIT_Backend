package main

import (
	"context"
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

	"github.com/jackc/pgx/v5/pgxpool"
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

	// Connect to database with connection pooling
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 5               // Max connections for auth-service
	poolConfig.MinConns = 2               // Warm connections
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connected with pgxpool (Max: 25, Min: 5)")

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

	// Initialize repositories with pgxpool
	sessionRepo := repository.NewSessionRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	passwordResetRepo := repository.NewPasswordResetRepository(pool)
	emailVerificationRepo := repository.NewEmailVerificationRepository(pool)
	userRepo := repository.NewUserRepository(pool)

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
