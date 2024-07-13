path=db/migrations
dsn="postgres://admin:admin@suindara.dev/sesamo?sslmode=disable"

.PHONY: build run test up down create db-status

build:
	@go build -o bin/sesamo cmd/main.go

run: build
	@./bin/sesamo

test:
	@go test -v ./...

up:
	@echo "Applying database migrations..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(dsn) GOOSE_MIGRATION_DIR=$(path) goose up

down:
	@echo "Reverting database migrations..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(dsn) GOOSE_MIGRATION_DIR=$(path) goose down

reset:
	@echo "Reseting database migrations..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(dsn) GOOSE_MIGRATION_DIR=$(path) goose reset

create:
	@echo "Creating new migration file..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(dsn) GOOSE_MIGRATION_DIR=$(path) goose create $(name) sql

db-status:
	@echo "Checking migration status..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(dsn) GOOSE_MIGRATION_DIR=$(path) goose status
