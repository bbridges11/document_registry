.PHONY: help
.DEFAULT_GOAL := help

# Variables
DOCKER_COMPOSE := docker-compose
DOCKER := docker
APP_NAME := document-registry
OPENFGA_CONTAINER := openfga
POSTGRES_CONTAINER := postgres
LOCALSTACK_CONTAINER := localstack

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Document Registry - Makefile Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-30s$(NC) %s\n", $$1, $$2}'

##@ Development

.PHONY: run
run: ## Run the application locally
	@echo "$(BLUE)Starting Document Registry...$(NC)"
	go run cmd/api/main.go

.PHONY: build
build: ## Build the application binary
	@echo "$(BLUE)Building Document Registry...$(NC)"
	go build -o bin/$(APP_NAME) cmd/api/main.go
	@echo "$(GREEN)✓ Binary created at bin/$(APP_NAME)$(NC)"

.PHONY: test
test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "$(BLUE)Tidying go.mod...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ go.mod tidied$(NC)"

##@ Docker - Individual Services

.PHONY: docker-openfga
docker-openfga: ## Start OpenFGA container
	@echo "$(BLUE)Starting OpenFGA...$(NC)"
	$(DOCKER) run -d \
		--name $(OPENFGA_CONTAINER) \
		-p 8081:8080 \
		-p 8082:8081 \
		-p 3000:3000 \
		openfga/openfga run
	@echo "$(GREEN)✓ OpenFGA started$(NC)"
	@echo "  - API: http://localhost:8081"
	@echo "  - Playground: http://localhost:3000"

.PHONY: docker-openfga-stop
docker-openfga-stop: ## Stop OpenFGA container
	@echo "$(YELLOW)Stopping OpenFGA...$(NC)"
	-$(DOCKER) stop $(OPENFGA_CONTAINER) 2>/dev/null
	-$(DOCKER) rm $(OPENFGA_CONTAINER) 2>/dev/null
	@echo "$(GREEN)✓ OpenFGA stopped$(NC)"

.PHONY: docker-openfga-logs
docker-openfga-logs: ## Show OpenFGA logs
	$(DOCKER) logs -f $(OPENFGA_CONTAINER)

.PHONY: docker-postgres
docker-postgres: ## Start PostgreSQL container
	@echo "$(BLUE)Starting PostgreSQL...$(NC)"
	$(DOCKER) run -d \
		--name $(POSTGRES_CONTAINER) \
		-p 5432:5432 \
		-e POSTGRES_DB=document_registry \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		postgres:16
	@echo "$(GREEN)✓ PostgreSQL started$(NC)"
	@echo "  - Host: localhost:5432"
	@echo "  - Database: document_registry"

.PHONY: docker-postgres-stop
docker-postgres-stop: ## Stop PostgreSQL container
	@echo "$(YELLOW)Stopping PostgreSQL...$(NC)"
	-$(DOCKER) stop $(POSTGRES_CONTAINER) 2>/dev/null
	-$(DOCKER) rm $(POSTGRES_CONTAINER) 2>/dev/null
	@echo "$(GREEN)✓ PostgreSQL stopped$(NC)"

.PHONY: docker-postgres-logs
docker-postgres-logs: ## Show PostgreSQL logs
	$(DOCKER) logs -f $(POSTGRES_CONTAINER)

.PHONY: docker-localstack
docker-localstack: ## Start LocalStack (S3) container
	@echo "$(BLUE)Starting LocalStack...$(NC)"
	$(DOCKER) run -d \
		--name $(LOCALSTACK_CONTAINER) \
		-p 4566:4566 \
		-e SERVICES=s3 \
		-e DEFAULT_REGION=us-east-1 \
		localstack/localstack
	@echo "$(GREEN)✓ LocalStack started$(NC)"
	@echo "  - S3 Endpoint: http://localhost:4566"

.PHONY: docker-localstack-stop
docker-localstack-stop: ## Stop LocalStack container
	@echo "$(YELLOW)Stopping LocalStack...$(NC)"
	-$(DOCKER) stop $(LOCALSTACK_CONTAINER) 2>/dev/null
	-$(DOCKER) rm $(LOCALSTACK_CONTAINER) 2>/dev/null
	@echo "$(GREEN)✓ LocalStack stopped$(NC)"

.PHONY: docker-localstack-logs
docker-localstack-logs: ## Show LocalStack logs
	$(DOCKER) logs -f $(LOCALSTACK_CONTAINER)

##@ Docker - Combined Operations

.PHONY: docker-up
docker-up: docker-postgres docker-openfga docker-localstack ## Start all infrastructure containers
	@echo ""
	@echo "$(GREEN)✓ All infrastructure started$(NC)"
	@echo ""
	@echo "$(BLUE)Service URLs:$(NC)"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - OpenFGA API: http://localhost:8081"
	@echo "  - OpenFGA Playground: http://localhost:3000"
	@echo "  - LocalStack S3: http://localhost:4566"
	@echo ""
	@echo "$(YELLOW)Next: Run 'make openfga-setup' to create store$(NC)"

.PHONY: docker-down
docker-down: docker-postgres-stop docker-openfga-stop docker-localstack-stop ## Stop all infrastructure containers
	@echo "$(GREEN)✓ All infrastructure stopped$(NC)"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart all infrastructure containers

.PHONY: docker-clean
docker-clean: docker-down ## Stop and remove all containers, volumes, and networks
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	-$(DOCKER) volume prune -f
	-$(DOCKER) network prune -f
	@echo "$(GREEN)✓ Docker resources cleaned$(NC)"

.PHONY: docker-ps
docker-ps: ## Show running containers
	@echo "$(BLUE)Running containers:$(NC)"
	@$(DOCKER) ps --filter "name=$(POSTGRES_CONTAINER)" --filter "name=$(OPENFGA_CONTAINER)" --filter "name=$(LOCALSTACK_CONTAINER)" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

##@ OpenFGA Setup

.PHONY: openfga-setup
openfga-setup: ## Create OpenFGA store (run after docker-openfga)
	@echo "$(BLUE)Creating OpenFGA store...$(NC)"
	@STORE_ID=$$(curl -s -X POST http://localhost:8081/stores \
		-H "Content-Type: application/json" \
		-d '{"name": "document-registry"}' | grep -o '"id":"[^"]*' | cut -d'"' -f4); \
	if [ -z "$$STORE_ID" ]; then \
		echo "$(RED)✗ Failed to create store$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✓ Store created successfully$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Add this to your .env file:$(NC)"; \
		echo "OPENFGA_STORE_ID=$$STORE_ID"; \
		echo ""; \
		echo "$(BLUE)Or run:$(NC)"; \
		echo "export OPENFGA_STORE_ID=$$STORE_ID"; \
	fi

.PHONY: openfga-list-stores
openfga-list-stores: ## List all OpenFGA stores
	@echo "$(BLUE)OpenFGA Stores:$(NC)"
	@curl -s http://localhost:8081/stores | jq '.'

.PHONY: openfga-health
openfga-health: ## Check OpenFGA health
	@echo "$(BLUE)Checking OpenFGA health...$(NC)"
	@curl -s http://localhost:8081/healthz && echo "$(GREEN)✓ OpenFGA is healthy$(NC)" || echo "$(RED)✗ OpenFGA is not responding$(NC)"

.PHONY: openfga-playground
openfga-playground: ## Open OpenFGA Playground in browser
	@echo "$(BLUE)Opening OpenFGA Playground...$(NC)"
	@open http://localhost:3000 || xdg-open http://localhost:3000 || echo "Open http://localhost:3000 in your browser"

##@ Database Operations

.PHONY: migrate-install
migrate-install: ## Install golang-migrate CLI
	@echo "$(BLUE)Installing golang-migrate...$(NC)"
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)✓ golang-migrate installed$(NC)"
	@echo "$(YELLOW)Note: Ensure $$GOPATH/bin is in your PATH$(NC)"

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)✗ Error: NAME is required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-create NAME=add_users_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)...$(NC)"
	migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created$(NC)"

.PHONY: db-migrate
db-migrate: ## Run database migrations (up)
	@echo "$(BLUE)Running database migrations...$(NC)"
	@if [ -z "$(POSTGRES_PASSWORD)" ]; then \
		echo "$(YELLOW)Using default password from .env.example$(NC)"; \
		POSTGRES_PASSWORD=postgres migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/document_registry?sslmode=disable" up; \
	else \
		migrate -path migrations -database "postgresql://postgres:$(POSTGRES_PASSWORD)@localhost:5432/document_registry?sslmode=disable" up; \
	fi
	@echo "$(GREEN)✓ Migrations applied successfully$(NC)"

.PHONY: db-migrate-down
db-migrate-down: ## Rollback last migration
	@echo "$(BLUE)Rolling back database migration...$(NC)"
	@if [ -z "$(POSTGRES_PASSWORD)" ]; then \
		echo "$(YELLOW)Using default password from .env.example$(NC)"; \
		POSTGRES_PASSWORD=postgres migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/document_registry?sslmode=disable" down 1; \
	else \
		migrate -path migrations -database "postgresql://postgres:$(POSTGRES_PASSWORD)@localhost:5432/document_registry?sslmode=disable" down 1; \
	fi
	@echo "$(GREEN)✓ Migration rolled back$(NC)"

.PHONY: db-migrate-down-all
db-migrate-down-all: ## Rollback all migrations (WARNING: destructive)
	@echo "$(RED)WARNING: This will rollback ALL migrations!$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to cancel, or wait 5 seconds to continue...$(NC)"
	@sleep 5
	@echo "$(BLUE)Rolling back all migrations...$(NC)"
	@if [ -z "$(POSTGRES_PASSWORD)" ]; then \
		POSTGRES_PASSWORD=postgres migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/document_registry?sslmode=disable" down -all; \
	else \
		migrate -path migrations -database "postgresql://postgres:$(POSTGRES_PASSWORD)@localhost:5432/document_registry?sslmode=disable" down -all; \
	fi
	@echo "$(GREEN)✓ All migrations rolled back$(NC)"

.PHONY: db-migrate-force
db-migrate-force: ## Force migration version (usage: make db-migrate-force VERSION=4)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)✗ Error: VERSION is required$(NC)"; \
		echo "$(YELLOW)Usage: make db-migrate-force VERSION=4$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Forcing migration to version $(VERSION)...$(NC)"
	@if [ -z "$(POSTGRES_PASSWORD)" ]; then \
		POSTGRES_PASSWORD=postgres migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/document_registry?sslmode=disable" force $(VERSION); \
	else \
		migrate -path migrations -database "postgresql://postgres:$(POSTGRES_PASSWORD)@localhost:5432/document_registry?sslmode=disable" force $(VERSION); \
	fi
	@echo "$(GREEN)✓ Migration version forced to $(VERSION)$(NC)"

.PHONY: db-migrate-version
db-migrate-version: ## Show current migration version
	@echo "$(BLUE)Current migration version:$(NC)"
	@if [ -z "$(POSTGRES_PASSWORD)" ]; then \
		POSTGRES_PASSWORD=postgres migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/document_registry?sslmode=disable" version; \
	else \
		migrate -path migrations -database "postgresql://postgres:$(POSTGRES_PASSWORD)@localhost:5432/document_registry?sslmode=disable" version; \
	fi

.PHONY: db-reset
db-reset: docker-postgres-stop docker-postgres ## Reset database (WARNING: deletes all data)
	@echo "$(RED)Database reset complete$(NC)"
	@sleep 2
	@echo "$(YELLOW)Waiting for PostgreSQL to be ready...$(NC)"
	@sleep 3

.PHONY: db-shell
db-shell: ## Connect to PostgreSQL shell
	@echo "$(BLUE)Connecting to PostgreSQL...$(NC)"
	$(DOCKER) exec -it $(POSTGRES_CONTAINER) psql -U postgres -d document_registry

##@ S3 Operations

.PHONY: s3-create-bucket
s3-create-bucket: ## Create S3 bucket in LocalStack
	@echo "$(BLUE)Creating S3 bucket...$(NC)"
	AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 s3 mb s3://document-registry
	@echo "$(GREEN)✓ Bucket created$(NC)"

.PHONY: s3-list-buckets
s3-list-buckets: ## List S3 buckets
	@echo "$(BLUE)S3 Buckets:$(NC)"
	AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 s3 ls

.PHONY: s3-list-objects
s3-list-objects: ## List objects in document-registry bucket
	@echo "$(BLUE)Objects in document-registry bucket:$(NC)"
	AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 s3 ls s3://document-registry --recursive

##@ Testing & Validation

.PHONY: health
health: ## Check all service health endpoints
	@echo "$(BLUE)Checking service health...$(NC)"
	@echo ""
	@echo "$(YELLOW)PostgreSQL:$(NC)"
	@$(DOCKER) exec $(POSTGRES_CONTAINER) pg_isready -U postgres && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo ""
	@echo "$(YELLOW)OpenFGA:$(NC)"
	@curl -s http://localhost:8081/healthz > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Unhealthy$(NC)"
	@echo ""
	@echo "$(YELLOW)Document Registry:$(NC)"
	@curl -s http://localhost:8080/health > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Not running$(NC)"

.PHONY: ports
ports: ## Check if required ports are available
	@echo "$(BLUE)Checking port availability...$(NC)"
	@echo ""
	@echo "$(YELLOW)Port 8080 (Document Registry):$(NC)"
	@lsof -i :8080 > /dev/null 2>&1 && echo "$(RED)✗ In use$(NC)" || echo "$(GREEN)✓ Available$(NC)"
	@echo ""
	@echo "$(YELLOW)Port 8081 (OpenFGA):$(NC)"
	@lsof -i :8081 > /dev/null 2>&1 && echo "$(RED)✗ In use$(NC)" || echo "$(GREEN)✓ Available$(NC)"
	@echo ""
	@echo "$(YELLOW)Port 5432 (PostgreSQL):$(NC)"
	@lsof -i :5432 > /dev/null 2>&1 && echo "$(RED)✗ In use$(NC)" || echo "$(GREEN)✓ Available$(NC)"
	@echo ""
	@echo "$(YELLOW)Port 4566 (LocalStack):$(NC)"
	@lsof -i :4566 > /dev/null 2>&1 && echo "$(RED)✗ In use$(NC)" || echo "$(GREEN)✓ Available$(NC)"

##@ Full Stack Operations

.PHONY: dev
dev: docker-up ## Start full development environment (Docker + App)
	@echo ""
	@echo "$(GREEN)✓ Infrastructure started$(NC)"
	@echo ""
	@echo "$(YELLOW)Waiting for services to be ready...$(NC)"
	@sleep 5
	@echo ""
	@echo "$(BLUE)To start the application:$(NC)"
	@echo "  make run"
	@echo ""
	@echo "$(BLUE)Or run with auto-setup:$(NC)"
	@echo "  make dev-full"

.PHONY: dev-full
dev-full: docker-up openfga-setup s3-create-bucket ## Full dev setup (Docker + OpenFGA store + S3)
	@echo ""
	@echo "$(GREEN)✓ Development environment ready!$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Copy OPENFGA_STORE_ID to your .env file"
	@echo "  2. Run: make run"

.PHONY: stop
stop: docker-down ## Stop all services

.PHONY: restart
restart: docker-restart ## Restart all infrastructure

.PHONY: clean
clean: docker-clean ## Clean all Docker resources
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

##@ Quick Start

.PHONY: quickstart
quickstart: ## Quick start guide
	@echo "$(BLUE)Document Registry - Quick Start Guide$(NC)"
	@echo ""
	@echo "$(GREEN)1. Start infrastructure:$(NC)"
	@echo "   make docker-up"
	@echo ""
	@echo "$(GREEN)2. Setup OpenFGA:$(NC)"
	@echo "   make openfga-setup"
	@echo ""
	@echo "$(GREEN)3. Create S3 bucket:$(NC)"
	@echo "   make s3-create-bucket"
	@echo ""
	@echo "$(GREEN)4. Configure environment:$(NC)"
	@echo "   - Copy .env.example to .env"
	@echo "   - Add OPENFGA_STORE_ID from step 2"
	@echo ""
	@echo "$(GREEN)5. Run the application:$(NC)"
	@echo "   make run"
	@echo ""
	@echo "$(YELLOW)Or use all-in-one:$(NC)"
	@echo "   make dev-full"