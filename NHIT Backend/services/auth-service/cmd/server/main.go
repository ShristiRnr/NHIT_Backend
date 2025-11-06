package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/proto"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	
	_ "github.com/lib/pq"
	grpcMiddleware "google.golang.org/grpc"
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
		databaseURL = "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
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
	
	accessTokenDuration := 15 * time.Minute  // 15 minutes
	refreshTokenDuration := 7 * 24 * time.Hour // 7 days
	
	jwtManager := utils.NewJWTManager(jwtSecret, accessTokenDuration, refreshTokenDuration)

	// Initialize email service (mock for now)
	emailService := utils.NewMockEmailService()

	// Initialize auth service
	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		refreshTokenRepo,
		passwordResetRepo,
		emailVerificationRepo,
		jwtManager,
		emailService,
	)

	// Initialize auth interceptor (middleware)
	authInterceptor := middleware.NewAuthInterceptor(authService)

	// Create gRPC server with interceptors
	grpcServer := grpcMiddleware.NewServer(
		grpcMiddleware.UnaryInterceptor(authInterceptor.Unary()),
		grpcMiddleware.StreamInterceptor(authInterceptor.Stream()),
	)

	// Initialize and register auth handler
	authHandler := grpc.NewAuthHandler(authService)
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
