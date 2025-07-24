# Automatically include and export variables from .env file if it exists
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

.PHONY: all build clean lint test integration-test

## all: Default task, builds the binary.
all: build

## build: Compile the Go binary and make it executable.
build:
	@echo "🔨 Building esq..."
	@go build -ldflags="-s -w" -o esq . # <-- Compile the main package at the project root
	@chmod +x esq

## clean: Remove build artifacts.
clean:
	@echo "🧹 Cleaning up..."
	@rm -f esq

## lint: Run the Go linter.
lint:
	@echo "🔍 Linting..."
	@golangci-lint run

## test: Run unit tests (excluding integration tests).
test:
	@echo "🧪 Running unit tests..."
	@go test ./...

## integration-test: Build the binary and run integration tests.
integration-test: build
	@echo "🚀 Running integration tests..."
	@go test -v -tags=integration ./...

## start-dev: Start the docker containers (Elasticsearch and Kibana).
start-dev:
	@./start.sh

## stop-dev: Stop the docker containers
stop-dev:
	@./stop.sh