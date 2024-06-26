.PHONY: build run test

build:
	@go build -o bin/sesamo cmd/main.go

run: build
	@./bin/sesamo

test:
	@go test -v ./...

