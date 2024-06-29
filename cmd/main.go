package main

import (
	"fmt"
	"log"

	"github.com/diegodario88/sesamo/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DBConnString string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	DBConnString = config.MustGetEnv("DATABASE_URL")
}

func main() {
	db := sqlx.MustConnect("postgres", DBConnString)
	var version string

	err := db.QueryRow("select version()").Scan(&version)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
