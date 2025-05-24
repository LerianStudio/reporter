# Plugin Smart Templates Makefile

# Component-specific variables
SERVICE_NAME := plugin-smart-templates
BIN_DIR := ./.bin
ARTIFACTS_DIR := ./artifacts

# Ensure artifacts directory exists
$(shell mkdir -p $(ARTIFACTS_DIR))

# Docker version detection for compatibility
DOCKER_VERSION := $(shell docker version --format '{{.Server.Version}}' 2>/dev/null || echo "0.0.0")
DOCKER_MIN_VERSION := 20.10.13

DOCKER_CMD := $(shell \
	if [ "$(shell printf '%s\n' "$(DOCKER_MIN_VERSION)" "$(DOCKER_VERSION)" | sort -V | head -n1)" = "$(DOCKER_MIN_VERSION)" ]; then \
		echo "docker compose"; \
	else \
		echo "docker-compose"; \
	fi \
)

#-------------------------------------------------------
# Help and Information
#-------------------------------------------------------

.PHONY: help
help:
	@echo ""
	@echo "Smart Templates Service Commands"
	@echo ""
	@echo "Core Commands:"
	@echo "  make help                        - Display this help message"
	@echo "  make build                       - Build the component"
	@echo "  make test                        - Run tests"
	@echo "  make clean                       - Clean build artifacts"
	@echo "  make run                         - Run the application"
	@echo ""
	@echo "Code Quality Commands:"
	@echo "  make lint                        - Run linting tools"
	@echo "  make format                      - Format code"
	@echo "  make tidy                        - Clean dependencies"
	@echo "  make sec                         - Run security checks"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make up                          - Start services with Docker Compose"
	@echo "  make down                        - Stop services with Docker Compose"
	@echo "  make start                       - Start existing containers"
	@echo "  make stop                        - Stop running containers"
	@echo "  make restart                     - Restart all containers"
	@echo "  make logs                        - Show logs for all services"
	@echo "  make logs-api                    - Show logs for smart-templates service"
	@echo "  make ps                          - List container status"
	@echo "  make destroy                     - Remove containers, networks, and volumes"
	@echo ""
	@echo "Smart Templates-Specific Commands:"
	@echo "  make generate-docs               - Generate Swagger API documentation"
	@echo "  make grpc-example-gen            - Generate gRPC example code"
	@echo ""
	@echo "Git and Setup Commands:"
	@echo "  make setup-git-hooks             - Install git hooks"
	@echo "  make check-hooks                 - Check git hooks status"
	@echo "  make set-env                     - Set up environment files"
	@echo ""

.DEFAULT_GOAL := help

#-------------------------------------------------------
# Build Commands
#-------------------------------------------------------

.PHONY: build
build:
	@echo "Building Smart Templates component..."
	@if [ -d "cmd" ]; then \
		for dir in cmd/*/; do \
			if [ -f "$$dir/main.go" ]; then \
				echo "Building $$dir..."; \
				go build -o $(BIN_DIR)/$$(basename $$dir) ./$$dir; \
			fi; \
		done; \
	else \
		echo "No cmd directory found, building current directory..."; \
		go build -o $(BIN_DIR)/$(SERVICE_NAME) .; \
	fi
	@echo "[ok] Build completed successfully"

.PHONY: run
run:
	@echo "Running the application..."
	@if [ -f "$(BIN_DIR)/app" ]; then \
		$(BIN_DIR)/app; \
	elif [ -f "cmd/app/main.go" ]; then \
		go run cmd/app/main.go; \
	else \
		echo "No main application found"; \
		exit 1; \
	fi

#-------------------------------------------------------
# Test Commands
#-------------------------------------------------------

.PHONY: test
test:
	@echo "Running tests..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Error: go is not installed"; \
		exit 1; \
	fi
	@go test -v ./...
	@echo "[ok] Tests completed successfully"

.PHONY: cover
cover:
	@echo "Generating test coverage report..."
	@go test -coverprofile=$(ARTIFACTS_DIR)/coverage.out ./...
	@go tool cover -func=$(ARTIFACTS_DIR)/coverage.out
	@echo "[ok] Coverage report generated"

.PHONY: cover-html
cover-html:
	@echo "Generating HTML test coverage report..."
	@go test -coverprofile=$(ARTIFACTS_DIR)/coverage.out ./...
	@go tool cover -html=$(ARTIFACTS_DIR)/coverage.out -o $(ARTIFACTS_DIR)/coverage.html
	@echo "Coverage report generated at $(ARTIFACTS_DIR)/coverage.html"
	@echo ""
	@echo "Coverage Summary:"
	@echo "----------------------------------------"
	@go tool cover -func=$(ARTIFACTS_DIR)/coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@echo "----------------------------------------"
	@echo "Open $(ARTIFACTS_DIR)/coverage.html in your browser to view detailed coverage report"

#-------------------------------------------------------
# Code Quality Commands
#-------------------------------------------------------

.PHONY: lint
lint:
	@echo "Running linter and performance checks..."
	@if [ -x "./make.sh" ]; then \
		./make.sh "lint"; \
	else \
		if find . -name "*.go" -type f | grep -q .; then \
			if ! command -v golangci-lint >/dev/null 2>&1; then \
				echo "Installing golangci-lint..."; \
				go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
			fi; \
			golangci-lint run ./...; \
		else \
			echo "No Go files found, skipping linting"; \
		fi; \
	fi
	@echo "[ok] Linting completed successfully"

.PHONY: format
format:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "[ok] Formatting completed successfully"

.PHONY: tidy
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "[ok] Dependencies cleaned successfully"

.PHONY: sec
sec:
	@echo "Running security checks..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@if find . -name "*.go" -type f | grep -q .; then \
		gosec -quiet ./...; \
		echo "[ok] Security checks completed"; \
	else \
		echo "No Go files found, skipping security checks"; \
	fi

#-------------------------------------------------------
# Clean Commands
#-------------------------------------------------------

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)/* $(ARTIFACTS_DIR)/*
	@echo "[ok] Artifacts cleaned successfully"

#-------------------------------------------------------
# Docker Commands
#-------------------------------------------------------

.PHONY: up
up: set-env
	@echo "Starting all services..."
	@$(DOCKER_CMD) -f docker-compose.yml up --build -d
	@echo "[ok] All services started successfully"

.PHONY: start
start:
	@echo "Starting existing containers..."
	@$(DOCKER_CMD) -f docker-compose.yml start $(c)
	@echo "[ok] Containers started successfully"

.PHONY: down
down:
	@echo "Stopping and removing containers..."
	@$(DOCKER_CMD) -f docker-compose.yml down $(c)
	@echo "[ok] Services stopped successfully"

.PHONY: stop
stop:
	@echo "Stopping running containers..."
	@$(DOCKER_CMD) -f docker-compose.yml stop $(c)
	@echo "[ok] Containers stopped successfully"

.PHONY: restart
restart:
	@echo "Restarting all services..."
	@make stop && make up
	@echo "[ok] Services restarted successfully"

.PHONY: destroy
destroy:
	@echo "Removing containers, networks, and volumes..."
	@$(DOCKER_CMD) -f docker-compose.yml down -v $(c)
	@echo "[ok] Resources cleaned successfully"

.PHONY: logs
logs:
	@echo "Showing logs for all services..."
	@$(DOCKER_CMD) -f docker-compose.yml logs --tail=100 -f $(c)

.PHONY: logs-api
logs-api:
	@echo "Showing logs for smart-templates service..."
	@$(DOCKER_CMD) -f docker-compose.yml logs --tail=100 -f smart-templates

.PHONY: ps
ps:
	@echo "Listing container status..."
	@$(DOCKER_CMD) -f docker-compose.yml ps

#-------------------------------------------------------
# Documentation Commands
#-------------------------------------------------------

.PHONY: generate-docs
generate-docs:
	@echo "Generating Swagger API documentation..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g ./cmd/app/main.go -d ./ -o ./api --parseDependency --parseInternal
	@docker run --rm -v ./:/local --user $(shell id -u):$(shell id -g) openapitools/openapi-generator-cli:v5.1.1 generate -i /local/api/swagger.json -g openapi-yaml -o /local/api
	@mv ./api/openapi/openapi.yaml ./api/openapi.yaml
	@rm -rf ./api/README.md ./api/.openapi-generator* ./api/openapi
	@echo "[ok] Swagger API documentation generated successfully"

#-------------------------------------------------------
# Smart Templates-Specific Commands  
#-------------------------------------------------------

.PHONY: grpc-example-gen
grpc-example-gen:
	@echo "Generating gRPC example code..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "Error: protoc is not installed"; \
		exit 1; \
	fi
	@if [ -d "pkg/proto" ]; then \
		protoc --go_out=. --go-grpc_out=. pkg/proto/*.proto; \
	else \
		echo "No proto directory found, skipping gRPC generation"; \
	fi
	@echo "[ok] gRPC example code generated successfully"

#-------------------------------------------------------
# Git Hook Commands
#-------------------------------------------------------

.PHONY: setup-git-hooks
setup-git-hooks:
	@echo "Setting up git hooks..."
	@if [ -x "./make.sh" ]; then \
		./make.sh "setupGitHooks"; \
	else \
		echo "make.sh script not found or not executable"; \
	fi
	@echo "[ok] Git hooks setup completed"

.PHONY: check-hooks
check-hooks:
	@echo "Checking git hooks status..."
	@if [ -x "./make.sh" ]; then \
		./make.sh "checkHooks"; \
	else \
		echo "make.sh script not found or not executable"; \
	fi
	@echo "[ok] Git hooks check completed"

#-------------------------------------------------------
# Environment Setup Commands
#-------------------------------------------------------

.PHONY: set-env
set-env:
	@echo "Setting up environment files..."
	@if [ -f ".env.example" ] && [ ! -f ".env" ]; then \
		cp ".env.example" ".env"; \
		echo "Created .env file from .env.example"; \
	elif [ ! -f ".env.example" ]; then \
		echo "Warning: No .env.example found"; \
	else \
		echo ".env file already exists"; \
	fi
	@echo "[ok] Environment files setup completed"

#-------------------------------------------------------
# Development Setup Commands
#-------------------------------------------------------

.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@make set-env
	@make tidy
	@make test
	@echo "[ok] Development environment setup completed"
	@echo "You're ready to start developing! Here are some useful commands:"
	@echo "  make build         - Build the component"
	@echo "  make test          - Run tests"  
	@echo "  make up            - Start services"
	@echo "  make generate-docs - Generate API documentation"