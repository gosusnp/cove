package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gosusnp/cove/api/db"
)

func main() {
	dbPath := os.Getenv("COVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cove.db"
	}

	database := db.Open(dbPath)
	defer database.Close()

	mux := http.NewServeMux()

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
