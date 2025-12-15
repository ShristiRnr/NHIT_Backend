package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Configuration
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	port := getEnv("PORT", "50055")

	log.Printf("🚀 Starting Designation Service on port %s", port)

	// Connect to database with optimized pool configuration
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	// ✅ Optimized connection pool limits
	poolConfig.MaxConns = 5                       // Low traffic service (mostly reads)
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established (Pool: Max=20, Min=5)")

	// Initialize SQLC queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repository
	designationRepo := repository.NewDesignationRepository(queries)

	// Initialize service
	designationService := services.NewDesignationService(designationRepo)

	// Initialize gRPC handler
	designationHandler := grpcHandler.NewDesignationHandler(designationService)

	// Connect to auth-service for RBAC
	authURL := getEnv("AUTH_SERVICE_URL", "localhost:50052")
	authConn, err := grpc.Dial(authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	log.Println("✅ Connected to Auth Service")

	// Initialize RBAC interceptor
	rbac := middleware.NewRBACInterceptor(authClient)
	for method, perms := range config.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	for _, method := range config.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	log.Println("✅ RBAC interceptor initialized")

	// Create gRPC server with RBAC
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rbac.UnaryServerInterceptor()),
	)
	designationpb.RegisterDesignationServiceServer(grpcServer, designationHandler)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("✅ Designation Service listening on port %s", port)
	log.Printf("📡 gRPC server ready")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
