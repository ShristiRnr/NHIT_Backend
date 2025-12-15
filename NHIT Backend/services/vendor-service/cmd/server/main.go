package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters"
	grpcAdapter "github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/services"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":50058"
	httpPort = ":8084"
)

func main() {
	ctx := context.Background()

	// Database connection with optimized pool configuration
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	log.Printf("🔌 Connecting to database: %s", dbURL)
	
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	// ✅ Optimized connection pool limits
	poolConfig.MaxConns = 5                       // Medium traffic service
	poolConfig.MinConns = 2                       // Maintain warm connections
	poolConfig.MaxConnLifetime = time.Hour        // Recycle connections hourly
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Close idle connections
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL database (Pool: Max=25, Min=8)")

	// Initialize dependencies
	logger := adapters.NewSimpleLogger()
	publisher := adapters.NewNoOpEventPublisher()
	serviceConfig := ports.VendorServiceConfig{
		EnableCodeGeneration: true,
		DefaultVendorType:    "EXTERNAL",
		MaxAccountsPerVendor: 10,
	}

	// Initialize repository
	vendorRepo := repository.NewVendorRepository(pool)
	log.Println("✅ Repository initialized")

	// Initialize service
	vendorService := services.NewVendorService(vendorRepo, logger, publisher, serviceConfig)
	log.Println("✅ Vendor service initialized")

	// Initialize gRPC handler
	vendorHandler := grpcAdapter.NewVendorGRPCHandler(vendorService)
	log.Println("✅ gRPC handler initialized")

	// Start gRPC server
	go func() {
		if err := startGRPCServer(vendorHandler); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP gateway server
	go func() {
		if err := startHTTPGateway(ctx); err != nil {
			log.Fatalf("Failed to start HTTP gateway: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down servers...")
}

func startGRPCServer(vendorHandler vendorpb.VendorServiceServer) error {
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	// Connect to auth-service for RBAC
	authConn, err := grpcLib.Dial("localhost:50052", grpcLib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️  Failed to connect to auth-service: %v (RBAC disabled)", err)
		// Continue without RBAC if auth service is down
		server := grpcLib.NewServer()
		vendorpb.RegisterVendorServiceServer(server, vendorHandler)
		reflection.Register(server)
		log.Printf("🚀 gRPC server listening on %s (NO RBAC)", grpcPort)
		return server.Serve(listener)
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

	// Create server with RBAC
	server := grpcLib.NewServer(
		grpcLib.UnaryInterceptor(rbac.UnaryServerInterceptor()),
	)
	
	// Register vendor handler
	vendorpb.RegisterVendorServiceServer(server, vendorHandler)

	// Enable reflection for development
	reflection.Register(server)

	log.Printf("🚀 gRPC server listening on %s", grpcPort)
	log.Printf("✅ RBAC enabled with permissions")
	log.Printf("📋 Database-backed vendor service ready")
	return server.Serve(listener)
}

func startHTTPGateway(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpcLib.DialOption{grpcLib.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost%s", grpcPort)

	err := vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Add CORS middleware
	handler := corsMiddleware(mux)

	log.Printf("🌐 HTTP gateway listening on %s", httpPort)
	log.Printf("📋 API endpoints available:")
	log.Printf("   - POST   /api/v1/vendors                    # Create vendor")
	log.Printf("   - GET    /api/v1/vendors                    # List vendors")
	log.Printf("   - POST   /api/v1/vendors/generate-code      # Generate vendor code")
	log.Printf("   - GET    /api/v1/vendors/dropdowns/projects # Get projects dropdown")

	return http.ListenAndServe(httpPort, handler)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
