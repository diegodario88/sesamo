package db

import (
	"context"
	"embed"
	"time"

	"github.com/diegodario88/sesamo/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var Migrations embed.FS

func CreateStorageConn() (*sqlx.DB, error) {
	DB, err := setup(config.Variables.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	DB.SetMaxIdleConns(2)
	DB.SetMaxOpenConns(4)
	DB.SetConnMaxLifetime(time.Duration(30) * time.Minute)

	return DB, nil
}

func setup(uri string) (*sqlx.DB, error) {
	connConfig, _ := pgx.ParseConfig(uri)
	afterConnect := stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `
    SET SESSION "some.key" = 'somekey';
      CREATE TEMP TABLE IF NOT EXISTS sometable AS SELECT 212 id;
      `)
		if err != nil {
			return err
		}
		return nil
	})

	pgxdb := stdlib.OpenDB(*connConfig, afterConnect)
	goose.SetBaseFS(Migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(pgxdb, "migrations"); err != nil {
		panic(err)
	}

	return sqlx.NewDb(pgxdb, "pgx"), nil
}
