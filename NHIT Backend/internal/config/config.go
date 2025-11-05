package config

import (
	"log"
	"os"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort       string
	DatabaseURL      string
	JWTSecret        string
	JWTExpiry        time.Duration
	RefreshTokenExpiry time.Duration
}

// LoadConfig reads environment variables and returns a Config struct
func LoadConfig() *Config {
	serverPort := getEnv("SERVER_PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "supersecretkey")
	jwtExpiry := getEnvAsDuration("JWT_EXPIRY", 15*time.Minute)
	refreshExpiry := getEnvAsDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour) // 7 days

	return &Config{
		ServerPort:       serverPort,
		DatabaseURL:      dbURL,
		JWTSecret:        jwtSecret,
		JWTExpiry:        jwtExpiry,
		RefreshTokenExpiry: refreshExpiry,
	}
}

// getEnv reads an environment variable or returns a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}


// getEnvAsDuration reads an env variable as time.Duration or returns fallback
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("invalid duration for %s: %v, using fallback %v", key, err, fallback)
		return fallback
	}
	return value
}
