.PHONY: help build test clean up down logs dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the plugin
	@echo "Building modifier plugin..."
	@go mod tidy
	@go build -v ./...

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts and Docker containers
	@echo "Cleaning up..."
	@docker-compose down -v
	@docker system prune -f

up: ## Start the stack
	@echo "Starting Traefik stack..."
	@docker-compose up -d
	@echo ""
	@echo "🚀 Stack started successfully!"
	@echo ""
	@echo "Services available:"
	@echo "  - Traefik Dashboard: http://localhost:8080"
	@echo "  - Chat Service:      http://chat.localhost"
	@echo "  - HTTPBin:          http://httpbin.localhost"
	@echo "  - Echo Service:     http://echo.localhost"
	@echo ""
	@echo "Run 'make test-plugin' to test the plugin"

down: ## Stop the stack
	@echo "Stopping Traefik stack..."
	@docker-compose down

logs: ## Show logs
	@docker-compose logs -f

dev: up ## Start development environment
	@echo "Development environment ready!"
	@echo "Run 'make test-plugin' to test changes"

test-plugin: ## Test the plugin functionality
	@echo "Testing modifier plugin..."
	@./examples/test-plugin.sh

restart: down up ## Restart the stack

status: ## Show stack status
	@docker-compose ps

plugin-logs: ## Show plugin-related logs from Traefik
	@docker-compose logs traefik | grep -i "modifier\|plugin"

# Development helpers
watch-logs: ## Watch Traefik logs for debugging
	@docker-compose logs -f traefik

rebuild: ## Rebuild and restart
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d