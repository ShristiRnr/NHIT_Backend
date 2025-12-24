package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orgID := "379d8ece-e140-41a9-ad3b-577ea64c2b27"
	
	fmt.Printf("--- Status Distribution for Org: %s ---\n", orgID)
	rows, err := db.Query("SELECT status, COUNT(*) FROM green_notes WHERE org_id = $1 GROUP BY status ORDER BY COUNT(*) DESC", orgID)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		fmt.Printf("Status: %-20s Count: %d\n", status, count)
	}

	fmt.Println("\n--- Sample Records ---")
	rows, err = db.Query("SELECT id, project_name, supplier_name, status FROM green_notes WHERE org_id = $1 LIMIT 5", orgID)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id, proj, supp, status string
		rows.Scan(&id, &proj, &supp, &status)
		fmt.Printf("Project: %-30s Status: %s\n", proj, status)
	}
}
