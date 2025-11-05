package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/http_server"
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
	resetService := services.NewPasswordResetService(resetRepo)
	resetHandler := http_server.NewPasswordResetHandler(resetService)

	// ---------- Refresh Token ----------
	refreshRepo := repository.NewRefreshRepo(queries)
	refreshService := services.NewRefreshTokenService(refreshRepo)
	refreshHandler := http_server.NewRefreshTokenHandler(refreshService)

	// ---------- Roles ----------
	roleRepo := repository.NewRoleRepo(queries)
	roleService := services.NewRoleService(roleRepo)
	roleHandler := http_server.NewRoleHandler(roleService)

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
	userRepo := repository.NewUserRepo(queries)
	userService := services.NewUserService(userRepo)
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

	// ---------- Start HTTP Server ----------
	log.Printf("🚀 Server listening on %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// helper for env vars with defaults
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
