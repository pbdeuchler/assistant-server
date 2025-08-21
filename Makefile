# Assistant Server Makefile

.PHONY: help build test test-unit test-integration clean lint fmt tidy docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/assistant-server ./cmd/main.go

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -v ./... -short

test-integration: ## Run integration tests with Docker PostgreSQL
	./scripts/test-integration.sh

clean: ## Clean build artifacts
	rm -rf bin/
	docker-compose -f docker-compose.test.yml down -v || true

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

fmt: ## Format code
	go fmt ./...

tidy: ## Tidy dependencies
	go mod tidy

docker-up: ## Start test database
	docker-compose -f docker-compose.test.yml up -d

docker-down: ## Stop test database
	docker-compose -f docker-compose.test.yml down -v

# Development targets
dev-setup: docker-up ## Set up development environment
	@echo "Development environment ready!"
	@echo "Test database running on localhost:5433"
	@echo "Connection string: postgres://test_user:test_password@localhost:5433/assistant_test?sslmode=disable"

dev-clean: docker-down ## Clean development environment

# CI targets
ci-test: ## Run tests in CI environment
	@echo "Running CI tests..."
	go test -v ./... -race -coverprofile=coverage.out
	./scripts/test-integration.sh