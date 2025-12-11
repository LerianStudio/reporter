# Project Root Makefile.
# Coordinates all component Makefiles and provides centralized commands.

# Define the root directory of the project
ROOT_DIR := $(shell pwd)

# Define the root directory of the project
SERVICE_NAME := reporter
BIN_DIR := ./.bin
ARTIFACTS_DIR := ./artifacts

# Choose docker compose command depending on installed version
DOCKER_CMD := $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; else echo "docker-compose"; fi)
export DOCKER_CMD

# Component directories
INFRA_DIR := ./components/infra
MANAGER_DIR := ./components/manager
WORKER_DIR := ./components/worker
FRONT_END_DIR := ./components/frontend

# Define a list of all component directories for easier iteration
BACKEND_COMPONENTS := $(WORKER_DIR) $(MANAGER_DIR)
COMPONENTS := $(INFRA_DIR) $(WORKER_DIR) $(MANAGER_DIR) $(FRONT_END_DIR)

# Include shared utility functions
# Define common utility functions
define print_title
	@echo ""
	@echo "------------------------------------------"
	@echo "   📝 $(1)  "
	@echo "------------------------------------------"
endef


# Check if a command is available
define check_command
	@which $(1) > /dev/null || (echo "Error: $(1) is required but not installed. $(2)" && exit 1)
endef

# Check if environment files exist
define check_env_files
	@missing=false; \
	for dir in $(COMPONENTS); do \
		if [ ! -f "$$dir/.env" ]; then \
			missing=true; \
			break; \
		fi; \
	done; \
	if [ "$$missing" = "true" ]; then \
		echo "Environment files are missing. Running set-env command first..."; \
		$(MAKE) set-env; \
	fi
endef

# ------------------------------------------------------
# Test configuration
# ------------------------------------------------------
TEST_MANAGER_URL ?= http://localhost:4005
TEST_HEALTH_WAIT ?= 60

define wait_for_services
	bash -c 'echo "Waiting for services to become healthy..."; \
	sleep 40; \
	for i in $$(seq 1 $(TEST_HEALTH_WAIT)); do \
	  if curl -fsS $(TEST_MANAGER_URL)/health >/dev/null 2>&1; then \
	    echo "Services are up"; exit 0; \
	  fi; \
	  sleep 1; \
	done; \
	echo "[error] Services not healthy after $(TEST_HEALTH_WAIT)s"; exit 1'
endef

MK_DIR := $(abspath mk)

include $(MK_DIR)/tests.mk

#-------------------------------------------------------
# Help Command
#-------------------------------------------------------

help:
	@echo ""
	@echo ""
	@echo "Reporter Commands"
	@echo ""
	@echo ""
	@echo "Core Commands:"
	@echo "  make help                        - Display this help message"
	@echo "  make test                        - Run tests on all components"
	@echo "  make build                       - Build all components"
	@echo "  make clean                       - Clean all build artifacts"
	@echo "  make cover                       - Run test coverage"
	@echo ""
	@echo ""
	@echo "Code Quality Commands:"
	@echo "  make lint                        - Run linting on all components"
	@echo "  make format                      - Format code in all components"
	@echo "  make tidy                        - Clean dependencies in root directory"
	@echo "  make check-tests                 - Verify test coverage for components"
	@echo "  make sec                         - Run security checks using gosec"
	@echo ""
	@echo ""
	@echo "Git Hook Commands:"
	@echo "  make setup-git-hooks             - Install and configure git hooks"
	@echo "  make check-hooks                 - Verify git hooks installation status"
	@echo "  make check-envs                  - Check if github hooks are installed and secret env files are not exposed"
	@echo ""
	@echo ""
	@echo "Setup Commands:"
	@echo "  make set-env                     - Copy .env.example to .env for all components"
	@echo ""
	@echo ""
	@echo "Service Commands:"
	@echo "  make up                           - Start all services with Docker Compose"
	@echo "  make down                         - Stop all services with Docker Compose"
	@echo "  make start                        - Start all containers"
	@echo "  make stop                         - Stop all containers"
	@echo "  make restart                      - Restart all containers"
	@echo "  make rebuild-up                   - Rebuild and restart all services"
	@echo "  make clean-docker                 - Clean all Docker resources (containers, networks, volumes)"
	@echo "  make logs                         - Show logs for all services"
	@echo ""
	@echo ""
	@echo "Documentation Commands:"
	@echo "  make generate-docs               - Generate Swagger documentation for all services"
	@echo ""
	@echo ""
	@echo "Test Suite Aliases:"
	@echo "  make test-unit                   - Run Go unit tests (exclude ./tests/**)"
	@echo "  make test-integration            - Run Go integration tests (brings up backend)"
	@echo "  make test-e2e                    - Run Apidog E2E tests (brings up backend)"
	@echo "  make test-fuzzy                  - Run fuzz/robustness tests (brings up backend)"
	@echo "  make test-fuzz-engine            - Run go fuzz engine on fuzzy tests (brings up backend)"
	@echo "  make test-chaos                  - Run chaos/resilience tests (brings up backend)"
	@echo "  make test-property               - Run property-based tests"
	@echo ""
	@echo ""

#-------------------------------------------------------
# Git Hook Commands
#-------------------------------------------------------

.PHONY: setup-git-hooks
setup-git-hooks:
	$(call print_title,Installing and configuring git hooks)
	@sh ./scripts/setup-git-hooks.sh
	@echo "[ok] Git hooks installed successfully"


.PHONY: check-hooks
check-hooks:
	$(call print_title,Verifying git hooks installation status)
	@err=0; \
	for hook_dir in .githooks/*; do \
		hook_name=$$(basename $$hook_dir); \
		if [ ! -f ".git/hooks/$$hook_name" ]; then \
			echo "Git hook $$hook_name is not installed"; \
			err=1; \
		else \
			echo "Git hook $$hook_name is installed"; \
		fi; \
	done; \
	if [ $$err -eq 0 ]; then \
		echo "[ok] All git hooks are properly installed"; \
	else \
		echo "[error] Some git hooks are missing. Run 'make setup-git-hooks' to fix."; \
		exit 1; \
	fi

.PHONY: check-envs
check-envs:
	$(call print_title,Checking if github hooks are installed and secret env files are not exposed)
	@sh ./scripts/check-envs.sh
	@echo "[ok] Environment check completed"

#-------------------------------------------------------
# Setup Commands
#-------------------------------------------------------

.PHONY: set-env
set-env:
	$(call print_title,"Setting up environment files")
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/.env.example" ] && [ ! -f "$$dir/.env" ]; then \
			echo "Creating .env in $$dir from .env.example"; \
			cp "$$dir/.env.example" "$$dir/.env"; \
		elif [ ! -f "$$dir/.env.example" ]; then \
			echo "Warning: No .env.example found in $$dir"; \
		else \
			echo ".env already exists in $$dir"; \
		fi; \
	done
	@echo "[ok] Environment files set up successfully"

#-------------------------------------------------------
# Build Commands
#-------------------------------------------------------

.PHONY: build
build:
	$(call print_title,Building all components)
	@echo "[ok] All components built successfully"

.PHONY: cover
cover:
	$(call print_title,Generating test coverage report)
	$(call check_command,go,"Install Go from https://golang.org/doc/install")
	@sh ./scripts/coverage.sh
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"
	@echo ""
	@echo "Coverage Summary:"
	@echo "----------------------------------------"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@echo "----------------------------------------"
	@echo "Open coverage.html in your browser to view detailed coverage report"
	@echo "[ok] Coverage report generated successfully ✔️"

#-------------------------------------------------------
# Test Coverage Commands
#-------------------------------------------------------

.PHONY: check-tests
check-tests:
	$(call print_title,Verifying test coverage for components)
	@sh ./scripts/check-tests.sh
	@echo "[ok] Test coverage verification completed"

#-------------------------------------------------------
# Code Quality Commands
#-------------------------------------------------------

.PHONY: lint
lint:
	$(call print_title,Running linters on all components)
	@if find . -name "*.go" -type f | grep -q .; then \
		if ! command -v golangci-lint >/dev/null 2>&1; then \
			echo "Installing golangci-lint..."; \
			go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		fi; \
		golangci-lint run --fix ./... --verbose; \
		echo "[ok] Linting completed successfully"; \
	else \
		echo "No Go files found, skipping linting"; \
	fi

.PHONY: format
format:
	$(call print_title,Formatting code in all components)
	@go fmt ./...
	@echo "[ok] Formatting completed successfully"

.PHONY: tidy
tidy:
	$(call print_title,Cleaning dependencies in root directory)
	@echo "Tidying root go.mod..."
	@go mod tidy
	@echo "[ok] Dependencies cleaned successfully"

#-------------------------------------------------------
# Security Commands
#-------------------------------------------------------

.PHONY: sec
sec:
	$(call print_title,Running security checks using gosec)
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@if find ./components ./pkg -name "*.go" -type f | grep -q .; then \
		echo "Running security checks on components/ and pkg/ folders..."; \
		gosec ./components/... ./pkg/...; \
		echo "[ok] Security checks completed"; \
	else \
		echo "No Go files found, skipping security checks"; \
	fi

#-------------------------------------------------------
# Clean Commands
#-------------------------------------------------------

.PHONY: clean
clean:
	@./scripts/clean-artifacts.sh

#-------------------------------------------------------
# Docker Commands
#-------------------------------------------------------

.PHONY: run
run:
	$(call print_title,Running the application with .env config)
	@go run cmd/app/main.go .env
	@echo "[ok] Application started successfully"

.PHONY: build-docker
build-docker:
	$(call print_title,Building Docker images)
	@$(DOCKER_CMD) -f docker-compose.yml build $(c)
	@echo "[ok] Docker images built successfully"

.PHONY: up
up:
	$(call print_title,Starting all services with Docker Compose)
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@echo "Starting infrastructure services first..."
	@cd $(INFRA_DIR) && $(MAKE) up
	@echo "Starting backend components..."
	@for dir in $(BACKEND_COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Starting services in $$dir..."; \
			(cd $$dir && $(MAKE) up) || exit 1; \
		fi \
	done
	@echo "[ok] Services started successfully ✔️"

.PHONY: start
start:
	$(call print_title,Starting all containers)
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Starting containers in $$dir..."; \
			(cd $$dir && $(MAKE) start) || exit 1; \
		fi; \
	done
	@echo "[ok] All containers started successfully"

.PHONY: down
down:
	$(call print_title,Stopping all services with Docker Compose)
	@echo "Stopping components..."
	@for dir in $(BACKEND_COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Stopping services in $$dir..."; \
			(cd $$dir && $(MAKE) down) || exit 1; \
		fi \
	done
	@echo "Stopping infrastructure services..."
	@cd $(INFRA_DIR) && $(MAKE) down
	@echo "[ok] Services stopped successfully ✔️"

.PHONY: stop
stop:
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Stopping containers in $$dir..."; \
			(cd $$dir && $(MAKE) stop) || exit 1; \
		fi; \
	done
	@echo "[ok] All containers stopped successfully"

.PHONY: restart
restart:
	@make stop && make start
	@echo "[ok] All containers restarted successfully"

.PHONY: rebuild-up
rebuild-up:
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Rebuilding and restarting services in $$dir..."; \
			(cd $$dir && $(MAKE) rebuild-up) || exit 1; \
		fi; \
	done
	@echo "[ok] All services rebuilt and restarted successfully"

.PHONY: logs
logs:
	$(call print_title,Showing logs for all services)
	@for dir in $(COMPONENTS); do \
		component_name=$$(basename $$dir); \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Logs for component: $$component_name"; \
			(cd $$dir && ($(DOCKER_CMD) -f docker-compose.yml logs --tail=50 2>/dev/null || $(DOCKER_CMD) -f docker-compose.yml logs --tail=50)) || exit 1; \
			echo ""; \
		fi; \
	done

.PHONY: logs-api
logs-api:
	$(call print_title,Showing logs for reporter service)
	@$(DOCKER_CMD) -f docker-compose.yml logs --tail=100 -f reporter

.PHONY: ps
ps:
	$(call print_title,Listing container status)
	@$(DOCKER_CMD) -f docker-compose.yml ps

#-------------------------------------------------------
# Docs Commands
#-------------------------------------------------------

.PHONY: generate-docs-all
generate-docs-all:
	$(call print_title,Generating Swagger documentation for all services)
	$(call check_command,swag,"go install github.com/swaggo/swag/cmd/swag@latest")
	@echo "Verifying API documentation coverage..."
	@sh ./scripts/verify-api-docs.sh 2>/dev/null || echo "Warning: Some API endpoints may not be properly documented. Continuing with documentation generation..."
	@echo "Generating documentation for plugin component..."
	$(MAKE) generate-docs 2>&1 | grep -v "warning: "
	@echo "[ok] Swagger documentation generated successfully"


.PHONY: verify-api-docs
verify-api-docs:
	$(call print_title,Verifying API documentation coverage)
	@if [ -f "./scripts/package.json" ]; then \
		echo "Installing npm dependencies..."; \
		cd ./scripts && npm install; \
	fi
	@sh ./scripts/verify-api-docs.sh
	@echo "[ok] API documentation verification completed"

.PHONY: validate-api-docs-legacy
validate-api-docs-legacy:
	$(call print_title,Validating API documentation structure and implementation)
	@if [ -f "./scripts/package.json" ]; then \
		echo "Using npm to run validation..."; \
		cd ./scripts && npm run validate-all; \
	else \
		echo "No package.json found in scripts directory. Running traditional validation..."; \
		$(MAKE) verify-api-docs; \
	fi
	@echo "[ok] API documentation validation completed"

.PHONY: validate-plugin
validate-plugin:
	$(call print_title,Validating pluginAPI documentation)
	@if [ -f "./scripts/package.json" ]; then \
		echo "Installing npm dependencies..."; \
		cd ./scripts && npm install; \
	fi
	make validate-api-docs
	@echo "[ok] plugin API validation completed"


.PHONY: generate-docs
generate-docs:
	$(call print_title,Generating Swagger API documentation)
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g ./components/manager/cmd/app/main.go -d ./ -o ./components/manager/api --parseDependency --parseInternal
	@docker run --rm -v $(ROOT_DIR):/local --user $(shell id -u):$(shell id -g) openapitools/openapi-generator-cli:v5.1.1 generate -i /local/components/manager/api/swagger.json -g openapi-yaml -o /local/components/manager/api
	@mv ./components/manager/api/openapi/openapi.yaml ./components/manager/api/openapi.yaml
	@rm -rf ./components/manager/api/README.md ./components/manager/api/.openapi-generator* ./components/manager/api/openapi
	@if [ -f "$(ROOT_DIR)/scripts/package.json" ]; then \
		echo "Installing npm dependencies for validation..."; \
		cd $(ROOT_DIR)/scripts && npm install > /dev/null; \
	fi
	@echo "[ok] Swagger API documentation generated successfully"

.PHONY: validate-api-docs
validate-api-docs: generate-docs
	$(call print_title,Validating API documentation)
	@if [ -f "scripts/validate-api-docs.js" ] && [ -f "$(ROOT_DIR)/scripts/package.json" ]; then \
		echo "Validating API documentation structure..."; \
		cd $(ROOT_DIR)/scripts && node $(ROOT_DIR)/scripts/validate-api-docs.js; \
		echo "Validating API implementations..."; \
		cd $(ROOT_DIR)/scripts && node $(ROOT_DIR)/scripts/validate-api-implementations.js; \
		echo "[ok] API documentation validation completed"; \
	else \
		echo "Validation scripts not found. Skipping validation."; \
	fi

#-------------------------------------------------------
# Developer Helper Commands
#-------------------------------------------------------

.PHONY: dev-setup
dev-setup:
	$(call print_title,Setting up development environment for all components)
	@echo "Setting up git hooks..."
	@$(MAKE) setup-git-hooks
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		go install github.com/golang/mock/mockgen@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "Setting up environment..."
	@if [ -f .env.example ] && [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from template"; \
	fi
	@make tidy
	@make check-tests
	@make sec
	@echo "[ok] Development environment set up successfully for all components"