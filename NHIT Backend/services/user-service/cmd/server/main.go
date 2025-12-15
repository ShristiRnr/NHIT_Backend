package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"


	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/grpc"
	httpHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/http"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository"
	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/services"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Get configuration from environment
	grpcPort := os.Getenv("USER_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("USER_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	}

	log.Printf("🚀 Starting User Service - gRPC:%s, HTTP:%s", grpcPort, httpPort)

	// Connect to database using pgxpool with optimized configuration
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	// ✅ Optimized connection pool limits for horizontal scaling
	poolConfig.MaxConns = 5                       // Increased for higher load
	poolConfig.MinConns = 2                       // Maintain warm connections
	poolConfig.MaxConnLifetime = time.Hour        // Recycle connections hourly
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Close idle connections
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Health check frequency

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established (Pool: Max=30, Min=10)")

	// Initialize sqlc queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	tenantRepo := repository.NewTenantRepository(queries)
	userRoleRepo := repository.NewUserRoleRepository(queries)
	roleRepo := repository.NewRoleRepository(queries)
	permissionRepo := repository.NewPermissionRepository(queries)
	loginHistoryRepo := repository.NewLoginHistoryRepository(queries)
	activityLogRepo := repository.NewActivityLogRepository(queries)

	// Initialize services
	userService := services.NewUserService(userRepo, tenantRepo, userRoleRepo, roleRepo, permissionRepo, loginHistoryRepo, activityLogRepo)

	// Connect to Auth Service for token validation / context
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "localhost:50052"
	}
	authConn, err := grpc.Dial(authServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth service at %s: %v", authServiceURL, err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	// Connect to Department Service
	departmentServiceURL := os.Getenv("DEPARTMENT_SERVICE_URL")
	if departmentServiceURL == "" {
		departmentServiceURL = "localhost:50054"
	}
	deptConn, err := grpc.Dial(departmentServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to department service at %s: %v", departmentServiceURL, err)
	}
	defer deptConn.Close()
	log.Printf("✅ Connected to Department Service at %s", departmentServiceURL)

	// Connect to Designation Service
	designationServiceURL := os.Getenv("DESIGNATION_SERVICE_URL")
	if designationServiceURL == "" {
		designationServiceURL = "localhost:50055"
	}
	desigConn, err := grpc.Dial(designationServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to designation service at %s: %v", designationServiceURL, err)
	}
	defer desigConn.Close()
	log.Printf("✅ Connected to Designation Service at %s", designationServiceURL)

	// Initialize handlers (pass DB pool so handler can query user_organizations/organizations)
	userGrpcHandler := grpcHandler.NewUserHandler(userService, pool, authClient, deptConn, desigConn)
	tenantHttpHandler := httpHandler.NewTenantHTTPHandler(userService)

	// Initialize RBAC interceptor for gRPC
	rbacInterceptor := middleware.NewRBACInterceptor(authClient)
	
	// Register public methods (methods that don't require authentication)
	// These are needed for auth-service to fetch user data during login
	rbacInterceptor.RegisterPublicMethod("/UserManagement/ListUserOrganizations")
	rbacInterceptor.RegisterPublicMethod("/UserManagement/ListRolesOfUser")
	rbacInterceptor.RegisterPublicMethod("/UserManagement/CreateUserLoginHistory")
	rbacInterceptor.RegisterPublicMethod("/UserManagement/CreateActivityLog")
	rbacInterceptor.RegisterPublicMethod("/UserManagement/CreateTenant")
	
	// Start gRPC server in goroutine
	go func() {
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(rbacInterceptor.UnaryServerInterceptor()),
		)
		userpb.RegisterUserManagementServer(grpcServer, userGrpcHandler)

		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}

		log.Printf("✅ gRPC Server listening on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Setup HTTP server with tenant endpoints
	router := mux.NewRouter()
	tenantHttpHandler.SetupRoutes(router)

	userHttpHandler := httpHandler.NewUserHTTPHandler(userService)
	userHttpHandler.RegisterRoutes(router)

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	log.Printf("✅ HTTP Server listening on port %s", httpPort)
	log.Printf("📡 HTTP endpoints available:")
	log.Printf("   POST http://localhost:%s/api/v1/tenants", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/tenants/{tenant_id}", httpPort)
	log.Printf("   POST http://localhost:%s/api/v1/users", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/users?tenant_id={tenant_id}", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/users/{user_id}", httpPort)
	log.Printf("   PUT  http://localhost:%s/api/v1/users/{user_id}", httpPort)
	log.Printf("   DELETE http://localhost:%s/api/v1/users/{user_id}", httpPort)
	log.Printf("   POST http://localhost:%s/api/v1/users/{user_id}/roles", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/users/{user_id}/roles", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/permissions", httpPort)
	log.Printf("   POST http://localhost:%s/api/v1/roles", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/roles", httpPort)
	log.Printf("   GET  http://localhost:%s/api/v1/roles/{role_id}", httpPort)
	log.Printf("   PUT  http://localhost:%s/api/v1/roles/{role_id}", httpPort)
	log.Printf("   DELETE http://localhost:%s/api/v1/roles/{role_id}", httpPort)

	if err := http.ListenAndServe(":"+httpPort, router); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
