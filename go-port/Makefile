# PAMPAX Go Port - Makefile
# Convenience commands for development

.PHONY: help build test clean install-tools migrate-up migrate-down migrate-new sqlc-generate

help: ## Show this help message
	@echo "PAMPAX Go Port - Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the pampax binary
	@echo "Building pampax..."
	@go build -o pampax ./cmd/pampax

run: ## Run pampax without building (usage: make run ARGS="index .")
	@go run ./cmd/pampax $(ARGS)

test: ## Run all tests
	@echo "Running all tests..."
	@go test ./...

test-verbose: ## Run all tests with verbose output
	@go test -v ./...

test-unit: ## Run unit tests only
	@go test -v ./test/unit/...

test-compat: ## Run compatibility tests only
	@go test -v ./test/compat/...

test-cover: ## Run tests with coverage report
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f pampax
	@rm -f coverage.out coverage.html
	@go clean

install-tools: ## Install development tools (sqlc, dbmate)
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installing dbmate..."
	@go install github.com/amacneil/dbmate/v2@latest
	@echo "Tools installed successfully!"

sqlc-generate: ## Generate database query code with sqlc
	@echo "Generating sqlc code..."
	@cd internal/db && sqlc generate
	@echo "sqlc generation complete!"

migrate-up: ## Run database migrations (up)
	@echo "Running migrations..."
	@dbmate up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@dbmate down

migrate-status: ## Show migration status
	@dbmate status

migrate-new: ## Create new migration (usage: make migrate-new NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please provide NAME variable (e.g., make migrate-new NAME=create_users_table)"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	@dbmate new $(NAME)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "Running golangci-lint..."
	@golangci-lint run

check: fmt vet ## Run format and vet checks

all: clean deps check test build ## Run all checks and build
