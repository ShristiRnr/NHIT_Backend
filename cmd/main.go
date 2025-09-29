package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/email"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/http_server"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

func main() {
	// Load environment variables
	port := getEnv("SERVER_PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://user:pass@localhost:5432/nhit?sslmode=disable")

	// Connect to Postgres
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}
	defer conn.Close()

	// Init sqlc queries
	queries := db.New(conn)

	// ---------- Token TTL ----------
	tokenTTL := 24 * time.Hour

	// ---------- User repository & service ----------
	userRepo := repository.NewUserRepo(queries)
	userService := services.NewUserService(userRepo)

	// ---------- Auth middleware ----------
	userAdapter := adapters.NewUserServiceAdapter(userService)
	authMiddleware := http_server.NewAuthMiddleware(userAdapter)
	_ =authMiddleware

	// ---------- Organization ----------
	orgRepo := repository.NewOrganizationRepo(queries)
	orgService := services.NewOrganizationService(orgRepo)
	orgHandler := http_server.NewOrganizationHandler(orgService)

	// ---------- UserOrganization ----------
	userOrgRepo := repository.NewUserOrganizationRepo(queries)
	userOrgService := services.NewUserOrganizationService(userOrgRepo)
	userOrgHandler := http_server.NewUserOrganizationHandler(userOrgService)

	// ---------- Pagination ----------
	pagRepo := repository.NewPaginationRepo(queries)
	pagService := services.NewPaginationService(pagRepo)
	pagHandler := http_server.NewPaginationHandler(pagService)

	// ---------- Password Reset ----------
	resetRepo := repository.NewPasswordResetRepo(queries)
	smtpSender := email.NewSMTPSender(
		"smtp.gmail.com", // host
		587,              // port
		"your_email@gmail.com", // username
		"your_app_password",     // password (App Password from Gmail)
		"your_email@gmail.com",  // from
		"NHIT",                  // appName (use your app’s name)
	)

	resetService := services.NewPasswordResetService(resetRepo, userRepo, tokenTTL, smtpSender)
	resetHandler := http_server.NewPasswordResetHandler(resetService)
	
	// ---------- Refresh Token ----------
	refreshRepo := repository.NewRefreshRepo(queries)
	refreshService := services.NewRefreshTokenService(refreshRepo)
	refreshHandler := http_server.NewRefreshTokenHandler(refreshService)

	// ---------- Roles ----------
	roleRepo := repository.NewRoleRepo(queries)
	roleService := services.NewRoleService(roleRepo)
	roleHandler := http_server.NewRoleHandler(roleService, authMiddleware)

	// ---------- Sessions ----------
	sessionRepo := repository.NewSessionRepo(queries)
	sessionService := services.NewSessionService(sessionRepo)
	sessionHandler := http_server.NewSessionHandler(sessionService)

	// ---------- Tenant ----------
	tenantRepo := repository.NewTenantRepo(queries)
	tenantService := services.NewTenantService(tenantRepo)
	tenantHandler := http_server.NewTenantHandler(tenantService)

	// ---------- User Login ----------
	userLoginRepo := repository.NewUserLoginRepo(queries)
	userLoginService := services.NewUserLoginService(userLoginRepo)
	userLoginHandler := http_server.NewUserLoginHandler(userLoginService)

	// ---------- User ----------
	userHandler := http_server.NewUserHandler(userService)

	// ---------- Router ----------
	r := chi.NewRouter()
	orgHandler.RegisterRoutes(r)
	userOrgHandler.RegisterRoutes(r)
	pagHandler.RegisterRoutes(r)
	resetHandler.RegisterRoutes(r)
	refreshHandler.RegisterRoutes(r)
	roleHandler.RegisterRoutes(r)
	sessionHandler.RegisterRoutes(r)
	tenantHandler.RegisterRoutes(r)
	userLoginHandler.RegisterRoutes(r)
	userHandler.RegisterRoutes(r)

	// ---------- Email Verification ----------
	// SMTP Config (load from env)
	smtpHost := getEnv("SMTP_HOST", "smtp.gmail.com")
	smtpPort := getEnvInt("SMTP_PORT", 587)
	smtpUser := getEnv("SMTP_USER", "your-email@gmail.com")
	smtpPass := getEnv("SMTP_PASS", "app-password")
	smtpFrom := getEnv("SMTP_FROM", "your-email@gmail.com")
	appName := getEnv("APP_NAME", "NHIT")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)

	emailSender := email.NewSMTPSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom, appName)

	emailRepo := repository.NewEmailVerificationRepo(queries)
	emailService := services.NewEmailVerificationService(emailRepo, userRepo, emailSender, baseURL)
	emailHandler := http_server.NewEmailVerificationHandler(emailService)

	// Email Verification Routes
	r.Post("/users/{userID}/send-verification", emailHandler.SendVerification)
	r.Get("/verify-email", emailHandler.VerifyEmail)

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
