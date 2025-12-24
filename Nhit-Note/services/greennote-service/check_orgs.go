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

	fmt.Println("--- Organization Summary ---")
	rows, err := db.Query(`
		SELECT o.org_id, o.name, COUNT(gn.id) as note_count 
		FROM organizations o 
		LEFT JOIN green_notes gn ON o.org_id = gn.org_id 
		GROUP BY o.org_id, o.name
		ORDER BY note_count DESC
	`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var oid, name string
		var count int
		rows.Scan(&oid, &name, &count)
		fmt.Printf("Org: %-30s | ID: %s | Notes: %d\n", name, oid, count)
	}

	fmt.Println("\n--- Green Notes with NO Organization matching organizations table ---")
	rows, err = db.Query(`
		SELECT DISTINCT org_id, COUNT(*) 
		FROM green_notes 
		WHERE org_id NOT IN (SELECT org_id FROM organizations)
		GROUP BY org_id
	`)
	if err == nil {
		for rows.Next() {
			var oid string
			var count int
			rows.Scan(&oid, &count)
			fmt.Printf("Orphan OrgID: %s | Notes: %d\n", oid, count)
		}
	}
}
