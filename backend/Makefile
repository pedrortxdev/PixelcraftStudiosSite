.PHONY: help run build test clean migrate-up migrate-down

# Variables
APP_NAME=pixelcraft-api
MAIN_PATH=cmd/api/main.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the application in development mode
	@echo "🚀 Starting Pixelcraft API..."
	go run $(MAIN_PATH)

build: ## Build the application binary
	@echo "🔨 Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: ./$(APP_NAME)"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

clean: ## Remove build artifacts
	@echo "🧹 Cleaning..."
	rm -f $(APP_NAME)
	rm -f $(APP_NAME).exe
	@echo "✅ Clean complete"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

migrate-up: ## Apply database migrations
	@echo "⬆️  Applying database migrations..."
	psql -U pedro -d pixelcraft -f database/schema.sql
	@echo "✅ Migrations applied"

migrate-down: ## Rollback database migrations (drops tables)
	@echo "⚠️  WARNING: This will drop all tables!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		psql -U pedro -d pixelcraft -c "DROP TABLE IF EXISTS users CASCADE;"; \
		echo "✅ Tables dropped"; \
	fi

dev: ## Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
	@echo "🔥 Starting with hot reload..."
	air

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest
