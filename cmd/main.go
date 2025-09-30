package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/email"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/http_server"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	port := getEnv("SERVER_PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://user:pass@localhost:5432/nhit?sslmode=disable")

	// Connect to Postgres
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}
	defer conn.Close()

	// Ping DB to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	queries := db.New(conn)

	// ---------- Token TTL ----------
	tokenTTL := 24 * time.Hour

	// ---------- Repositories ----------
	userRepo := repository.NewUserRepo(queries)
	roleRepo := repository.NewRoleRepo(queries)
	sessionRepo := repository.NewSessionRepo(queries)
	refreshRepo := repository.NewRefreshRepo(queries)
	orgRepo := repository.NewOrganizationRepo(queries)
	userOrgRepo := repository.NewUserOrganizationRepo(queries)
	pagRepo := repository.NewPaginationRepo(queries)
	resetRepo := repository.NewPasswordResetRepo(queries)
	emailRepo := repository.NewEmailVerificationRepo(queries)
	tenantRepo := repository.NewTenantRepo(queries)
	userLoginRepo := repository.NewUserLoginRepo(queries)
	deptRepo := repository.NewDepartmentRepository(queries)
	designationRepo := repository.NewDesignationRepository(queries)
	vendorRepo := repository.NewVendorRepositoryPG(queries)

	// ---------- Services ----------
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, roleRepo, sessionRepo, refreshRepo)
	orgService := services.NewOrganizationService(orgRepo)
	userOrgService := services.NewUserOrganizationService(userOrgRepo)
	pagService := services.NewPaginationService(pagRepo)
	deptService := services.NewDepartmentService(deptRepo)
	designationService := services.NewDesignationService(designationRepo)
	vendorService := services.NewVendorService(vendorRepo)

	// TLS SMTP Sender for Gmail (port 465)
	tlsSender := &email.SMTPTLSender{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnvInt("SMTP_PORT", 465),
		Username: getEnv("SMTP_USER", "your_email@gmail.com"),
		Password: getEnv("SMTP_PASS", "app_password"),
		From:     getEnv("SMTP_FROM", "your_email@gmail.com"),
		AppName:  getEnv("APP_NAME", "NHIT"),
	}

	resetService := services.NewPasswordResetService(resetRepo, userRepo, tokenTTL, tlsSender)
	refreshService := services.NewRefreshTokenService(refreshRepo)
	roleService := services.NewRoleService(roleRepo)
	sessionService := services.NewSessionService(sessionRepo)
	tenantService := services.NewTenantService(tenantRepo)
	userLoginService := services.NewUserLoginService(userLoginRepo)

	emailService := services.NewEmailVerificationService(
		emailRepo,
		userRepo,
		tlsSender,
		getEnv("BASE_URL", "http://localhost:"+port),
	)

	// ---------- Adapters / Middleware ----------
	userAdapter := adapters.NewUserServiceAdapter(authService)
	authMiddleware := http_server.NewAuthMiddleware(userAdapter) // <- pointer, fixed

	// ---------- Handlers ----------
	authHandler := http_server.NewAuthHandler(authService)
	orgHandler := http_server.NewOrganizationHandler(orgService)
	userOrgHandler := http_server.NewUserOrganizationHandler(userOrgService)
	pagHandler := http_server.NewPaginationHandler(pagService)
	resetHandler := http_server.NewPasswordResetHandler(resetService)
	refreshHandler := http_server.NewRefreshTokenHandler(refreshService)
	roleHandler := http_server.NewRoleHandler(roleService, authMiddleware)
	sessionHandler := http_server.NewSessionHandler(sessionService)
	tenantHandler := http_server.NewTenantHandler(tenantService)
	userLoginHandler := http_server.NewUserLoginHandler(userLoginService)
	userHandler := http_server.NewUserHandler(userService)
	emailHandler := http_server.NewEmailVerificationHandler(emailService, tlsSender)
	deptHandler := http_server.NewDepartmentHandler(deptService)
	designationHandler := http_server.NewDesignationHandler(designationService)
	vendorHandler := http_server.NewVendorHandler(vendorService, authMiddleware)

	// ---------- Router ----------
	r := chi.NewRouter()

	// ---------- Department Routes ----------
	r.Route("/departments", deptHandler.Register)         // make sure Register exists
	r.Route("/designations", designationHandler.Register) // make sure Register exists
	r.Route("/vendors", vendorHandler.Routes)             // <- fixed, use Routes

	// ---------- Auth ----------
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)

	// ---------- Organization ----------
	orgHandler.RegisterRoutes(r)

	// ---------- UserOrganization ----------
	userOrgHandler.RegisterRoutes(r)

	// ---------- Pagination ----------
	pagHandler.RegisterRoutes(r)

	// ---------- Password Reset ----------
	resetHandler.RegisterRoutes(r)

	// ---------- Refresh Token ----------
	refreshHandler.RegisterRoutes(r)

	// ---------- Roles ----------
	roleHandler.RegisterRoutes(r)

	// ---------- Sessions ----------
	sessionHandler.RegisterRoutes(r)

	// ---------- Tenant ----------
	tenantHandler.RegisterRoutes(r)

	// ---------- User Login ----------
	userLoginHandler.RegisterRoutes(r)

	// ---------- User ----------
	userHandler.RegisterRoutes(r)

	// ---------- Email Verification ----------
	emailHandler.Routes(r)

	// ---------- Designation ----------
	designationHandler.Register(r)

	// ---------- Start HTTP Server ----------
	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// getEnv reads environment variables with a fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt reads integer env variables with a fallback
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var val int
		if _, err := fmt.Sscanf(v, "%d", &val); err == nil {
			return val
		}
	}
	return fallback
}