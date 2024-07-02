package main

import (
	"database/sql"
	"log"

	api "github.com/diegodario88/sesamo/cmd/http"
	"github.com/diegodario88/sesamo/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	storage := sqlx.MustConnect("postgres", config.Variables.DatabaseUrl)
	healthcheckDb(storage.DB)
	api.NewServer(config.Variables.Port, storage.DB).Run()
}

func healthcheckDb(db *sql.DB) {
	err := db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database Successfully connected!")
}
