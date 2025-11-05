package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Connect establishes a connection to the PostgreSQL database
func Connect(dbURL string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping db: %w", err)
	}

	log.Println("✅ Database connection established")
	return conn, nil
}
