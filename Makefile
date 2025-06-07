.PHONY: all build test test-integration clean migrate-up migrate-down docker-up docker-down lint help

# Go related variables
BINARY_NAME=xyz-finance
MAIN_PACKAGE=./cmd/api
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

# Docker related variables
DOCKER_COMPOSE=docker-compose

# Migration related variables
MIGRATE=migrate
DB_URL=postgres://xyz_user:xyz_password@localhost:5432/xyz_db?sslmode=disable

# Default target
all: clean build

## Build:
build: ## Build the application
	@echo "Building..."
	@go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

## Test:
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -cover ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

## Database:
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	@migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	@migrate -path ./migrations -database "$(DB_URL)" down

migrate-force: ## Force database migration version
	@echo "Forcing database migration version..."
	@migrate -path ./migrations -database "$(DB_URL)" force $(version)

## Docker:
docker-up: ## Start all docker containers
	@echo "Starting docker containers..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all docker containers
	@echo "Stopping docker containers..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View docker container logs
	@echo "Viewing docker logs..."
	@$(DOCKER_COMPOSE) logs -f

docker-build: ## Build docker images
	@echo "Building docker images..."
	@$(DOCKER_COMPOSE) build

## Development:
run: ## Run the application locally
	@echo "Running application..."
	@go run $(MAIN_PACKAGE)

watch: ## Run the application with hot reload
	@echo "Running with hot reload..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@air

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

## Clean:
clean: ## Clean build files
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@rm -f coverage.out coverage.html

## Help:
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Default target when just running make
.DEFAULT_GOAL := help 