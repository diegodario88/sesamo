package store_test

import (
	"embed"
	"log"
	"testing"

	"github.com/diegodario88/sesamo/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type StoreTestSuite struct {
	suite.Suite
	db *sqlx.DB
}

func (storeTestSuite *StoreTestSuite) SetupTest() {
	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal(err)
	}

	testConnString := config.MustGetEnv("TEST_DATABASE_URL")
	storeTestSuite.db = sqlx.MustConnect("postgres", testConnString)

	goose.SetBaseFS(embedMigrations)

	if err = goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err = goose.Up(storeTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
	if err = goose.Reset(storeTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
	if err = goose.Up(storeTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
