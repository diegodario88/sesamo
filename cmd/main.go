package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DBConnString string

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: Environment variable %s is not set!", key)
	}

	return value
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	DBConnString = MustGetEnv("DATABASE_URL")
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
