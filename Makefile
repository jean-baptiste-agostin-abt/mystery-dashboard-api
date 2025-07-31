# Mystery Factory API Makefile
# Provides targets for building, testing, linting, and running the application

# Variables
APP_NAME := mysteryfactory-api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Directories
BUILD_DIR := ./bin
DOCKER_DIR := ./docker
MIGRATIONS_DIR := ./migrations

# Docker settings
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := $(VERSION)
DOCKER_REGISTRY := your-registry.com

# Database settings
DB_HOST := localhost
DB_PORT := 3306
DB_NAME := mysteryfactory
DB_USER := root
DB_PASSWORD := password
DATABASE_DSN := "$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?charset=utf8mb4&parseTime=True&loc=Local"

.PHONY: help build test lint clean run migrate docker-build docker-run docker-push dev setup deps check format vet security

# Default target
all: clean deps lint test build

# Help target
help: ## Show this help message
	@echo "Mystery Factory API - Available targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

# Build targets
build: ## Build the application binary
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

build-local: ## Build the application for local development
	@echo "Building $(APP_NAME) for local development..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server
	@echo "Local build complete: $(BUILD_DIR)/$(APP_NAME)"

# Development targets
dev: ## Run the application in development mode with hot reload
	@echo "Starting development server..."
	@air -c .air.toml || go run ./cmd/server

run: build-local ## Build and run the application locally
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

# Testing targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report generated: coverage.html"

test-short: ## Run tests without race detection (faster)
	@echo "Running short tests..."
	@go test -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -tags=integration -v ./...

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Code quality targets
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run --config .golangci.yml || echo "golangci-lint not installed, skipping..."
	@go vet ./...
	@gofmt -l . | grep -v vendor | tee /dev/stderr | test -z "$$(cat)"

format: ## Format code
	@echo "Formatting code..."
	@gofmt -w .
	@goimports -w . || echo "goimports not installed, skipping..."

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

security: ## Run security checks
	@echo "Running security checks..."
	@gosec ./... || echo "gosec not installed, skipping..."

# Dependency management
deps: ## Download and tidy dependencies
	@echo "Managing dependencies..."
	@go mod download
	@go mod tidy
	@go mod verify

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Database targets
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@go run ./cmd/migrate -dsn=$(DATABASE_DSN) -dir=$(MIGRATIONS_DIR)

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-up: ## Apply all pending migrations
	@echo "Applying migrations..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_DSN) up

migrate-down: ## Rollback one migration
	@echo "Rolling back migration..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_DSN) down 1

migrate-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_DSN) drop -f
	@migrate -path $(MIGRATIONS_DIR) -database $(DATABASE_DSN) up

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built successfully"

docker-run: ## Run application in Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 \
		-e DATABASE_DSN=$(DATABASE_DSN) \
		-e JWT_SECRET=your-jwt-secret \
		-e AWS_REGION=us-east-1 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image to $(DOCKER_REGISTRY)..."
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	@docker-compose down

# Setup and installation targets
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install github.com/cosmtrek/air@latest
	@go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Development tools installed"

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(APP_NAME)..."
	@go install $(LDFLAGS) ./cmd/server

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache
	@go clean -testcache

check: lint test ## Run all checks (lint + test)

generate: ## Generate code (mocks, etc.)
	@echo "Generating code..."
	@go generate ./...

# Monitoring and metrics
metrics: ## Start Prometheus and Grafana for monitoring
	@echo "Starting monitoring stack..."
	@docker-compose -f docker-compose.monitoring.yml up -d

metrics-down: ## Stop monitoring stack
	@echo "Stopping monitoring stack..."
	@docker-compose -f docker-compose.monitoring.yml down

# Environment management
env-example: ## Create .env.example file
	@echo "Creating .env.example..."
	@echo "# Server Configuration" > .env.example
	@echo "PORT=8080" >> .env.example
	@echo "ENVIRONMENT=development" >> .env.example
	@echo "SERVICE_NAME=mysteryfactory-api" >> .env.example
	@echo "READ_TIMEOUT=30" >> .env.example
	@echo "WRITE_TIMEOUT=30" >> .env.example
	@echo "IDLE_TIMEOUT=120" >> .env.example
	@echo "" >> .env.example
	@echo "# Database Configuration" >> .env.example
	@echo "DATABASE_DSN=root:password@tcp(localhost:3306)/mysteryfactory?charset=utf8mb4&parseTime=True&loc=Local" >> .env.example
	@echo "" >> .env.example
	@echo "# JWT Configuration" >> .env.example
	@echo "JWT_SECRET=your-super-secret-jwt-key-change-this-in-production" >> .env.example
	@echo "JWT_EXPIRATION=3600" >> .env.example
	@echo "" >> .env.example
	@echo "# Logging Configuration" >> .env.example
	@echo "LOG_LEVEL=info" >> .env.example
	@echo "" >> .env.example
	@echo "# OpenTelemetry Configuration" >> .env.example
	@echo "JAEGER_ENDPOINT=http://localhost:14268/api/traces" >> .env.example
	@echo "" >> .env.example
	@echo "# AWS Configuration" >> .env.example
	@echo "AWS_REGION=us-east-1" >> .env.example
	@echo "AWS_ACCESS_KEY_ID=your-access-key" >> .env.example
	@echo "AWS_SECRET_ACCESS_KEY=your-secret-key" >> .env.example
	@echo "S3_BUCKET=your-s3-bucket" >> .env.example
	@echo "" >> .env.example
	@echo "# Multi-tenant Configuration" >> .env.example
	@echo "DEFAULT_TENANT_ID=default" >> .env.example
	@echo ".env.example created"

# Release targets
release: clean deps lint test build ## Prepare a release build
	@echo "Release $(VERSION) ready"

# Show build info
info: ## Show build information
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"