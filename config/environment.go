package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Variables = initConfig()

type environment struct {
	DatabaseUrl            string
	TestDatabaseUrl        string
	Port                   int64
	JwtSecret              string
	JwtExpirationInSeconds int64
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: Environment variable %s is not set!", key)
	}

	return value
}

func initConfig() environment {
	err := godotenv.Load("/app/.env")

	if err != nil {
		log.Fatal(err)
	}

	stringPort := mustGetEnv("HTTP_SERVER_PORT")
	intPort, err := strconv.ParseInt(stringPort, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	stringJwtExpiration := mustGetEnv("JWT_EXPIRATION_IN_SECONDS")
	intJwtExpiration, err := strconv.ParseInt(stringJwtExpiration, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	return environment{
		DatabaseUrl:            mustGetEnv("DATABASE_URL"),
		TestDatabaseUrl:        mustGetEnv("TEST_DATABASE_URL"),
		Port:                   intPort,
		JwtSecret:              mustGetEnv("JWT_SECRET"),
		JwtExpirationInSeconds: intJwtExpiration,
	}
}
