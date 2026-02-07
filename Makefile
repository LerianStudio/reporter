# Project Root Makefile.
# Coordinates all component Makefiles and provides centralized commands.

# Define the root directory of the project
ROOT_DIR := $(shell pwd)
SERVICE_NAME := reporter
BIN_DIR := ./.bin
ARTIFACTS_DIR := ./artifacts

# Docker command detection (supports both docker compose and docker-compose)
DOCKER_VERSION := $(shell docker version --format '{{.Server.Version}}' 2>/dev/null || echo "0")
DOCKER_MIN_VERSION := 20.10.13
DOCKER_CMD := $(shell \
	if [ "$(shell printf '%s\n' "$(DOCKER_MIN_VERSION)" "$(DOCKER_VERSION)" | sort -V | head -n1)" = "$(DOCKER_MIN_VERSION)" ]; then \
		echo "docker compose"; \
	else \
		echo "docker-compose"; \
	fi \
)

# Component directories
INFRA_DIR := ./components/infra
MANAGER_DIR := ./components/manager
WORKER_DIR := ./components/worker

# Define a list of all component directories for easier iteration
BACKEND_COMPONENTS := $(WORKER_DIR) $(MANAGER_DIR)
COMPONENTS := $(INFRA_DIR) $(WORKER_DIR) $(MANAGER_DIR)

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
	@echo "  make lint                        - Run linting (includes format and imports)"
	@echo "  make format                      - Format code with gofumpt"
	@echo "  make imports                     - Organize imports with goimports"
	@echo "  make tidy                        - Clean dependencies in root directory"
	@echo "  make check-tests                 - Verify test coverage for components"
	@echo "  make sec                         - Run security checks (gosec + govulncheck)"
	@echo "  make sec-gosec                   - Run gosec security scanner"
	@echo "  make sec-govulncheck             - Run govulncheck vulnerability scanner"
	@echo ""
	@echo ""
	@echo "Git Hook Commands:"
	@echo "  make setup-git-hooks             - Install and configure git hooks"
	@echo "  make check-hooks                 - Verify git hooks installation status"
	@echo "  make check-envs                  - Check if github hooks are installed and secret env files are not exposed"
	@echo ""
	@echo ""
	@echo "Setup Commands:"
	@echo "  make check-tools                 - Verify all required development tools are installed"
	@echo "  make dev-setup                   - Install all development tools and set up environment"
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
	@echo "Code Generation Commands:"
	@echo "  make generate                    - Run code generation (go generate)"
	@echo "  make generate-mocks              - Generate mock files"
	@echo ""
	@echo ""
	@echo "Documentation Commands:"
	@echo "  make generate-docs               - Generate Swagger documentation for all services"
	@echo "  make serve-docs                  - Serve Swagger UI locally at http://localhost:8080"
	@echo ""
	@echo ""
	@echo "Test Suite Aliases:"
	@echo "  make test-unit                   - Run Go unit tests (exclude ./tests/**)"
	@echo "  make test-integration            - Run Go integration tests (brings up backend)"
	@echo "  make test-fuzzy                  - Run fuzz/robustness tests (brings up backend)"
	@echo "  make test-chaos                  - Run chaos/resilience tests (brings up backend)"
	@echo "  make test-property               - Run property-based tests"
	@echo ""
	@echo "Coverage Commands:"
	@echo "  make coverage-unit               - Run unit tests with coverage report"
	@echo ""
	@echo ""

#-------------------------------------------------------
# Git Hook Commands
#-------------------------------------------------------

.PHONY: setup-git-hooks
setup-git-hooks:
	$(call print_title,"Installing and configuring git hooks")
	@sh ./scripts/setup-git-hooks.sh
	@echo "[ok] Git hooks installed successfully ✔️"


.PHONY: check-hooks
check-hooks:
	$(call print_title,"Verifying git hooks installation status")
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
		echo "[ok] All git hooks are properly installed ✔️"; \
	else \
		echo "[error] Some git hooks are missing. Run 'make setup-git-hooks' to fix. ❌"; \
		exit 1; \
	fi

.PHONY: check-envs
check-envs:
	$(call print_title,"Checking if github hooks are installed and secret env files are not exposed")
	@sh ./scripts/check-envs.sh
	@echo "[ok] Environment check completed ✔️"

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
	@echo "Building backend components..."
	@for dir in $(BACKEND_COMPONENTS); do \
		component_name=$$(basename $$dir); \
		echo "Building $$component_name..."; \
		(cd $$dir && $(MAKE) build) || exit 1; \
	done
	@echo "[ok] All components built successfully"

.PHONY: cover
cover:
	$(call print_title,Generating test coverage report)
	@echo "Note: PostgreSQL repository tests are excluded from coverage metrics."
	@echo "See coverage report for details on why and what is being tested."
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

.PHONY: cover-html
cover-html:
	$(call print_title,"Generating HTML test coverage report")
	@PACKAGES=$$(go list ./... | grep -v -f ./scripts/coverage_ignore.txt); \
	go test -coverprofile=$(ARTIFACTS_DIR)/coverage.out $$PACKAGES
	@go tool cover -html=$(ARTIFACTS_DIR)/coverage.out -o $(ARTIFACTS_DIR)/coverage.html
	@echo "Coverage report generated at $(ARTIFACTS_DIR)/coverage.html"
	@echo ""
	@echo "Coverage Summary:"
	@echo "----------------------------------------"
	@go tool cover -func=$(ARTIFACTS_DIR)/coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@echo "----------------------------------------"
	@echo "Open $(ARTIFACTS_DIR)/coverage.html in your browser to view detailed coverage report"

#-------------------------------------------------------
# Test Coverage Commands
#-------------------------------------------------------

.PHONY: check-tests
check-tests:
	$(call print_title,"Verifying test coverage")
	@if find . -name "*.go" -type f | grep -q .; then \
		echo "Running test coverage check..."; \
		go test -coverprofile=coverage.tmp ./... > /dev/null 2>&1; \
		if [ -f coverage.tmp ]; then \
			coverage=$$(go tool cover -func=coverage.tmp | grep total | awk '{print $$3}'); \
			echo "Test coverage: $$coverage"; \
			rm coverage.tmp; \
		else \
			echo "No coverage data generated"; \
		fi; \
	else \
		echo "No Go files found, skipping test coverage check"; \
	fi

#-------------------------------------------------------
# Code Quality Commands
#-------------------------------------------------------

.PHONY: lint
lint: format imports
	$(call print_title,"Running linters")
	@if find . -name "*.go" -type f | grep -q .; then \
		if ! command -v golangci-lint >/dev/null 2>&1; then \
			echo "Installing golangci-lint..."; \
			go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		fi; \
		golangci-lint run --fix ./... --verbose; \
		echo "[ok] Linting completed successfully ✔️"; \
	else \
		echo "No Go files found, skipping linting"; \
	fi

.PHONY: format
format:
	$(call print_title,"Formatting code with gofumpt")
	@if ! command -v gofumpt >/dev/null 2>&1; then \
		echo "Installing gofumpt..."; \
		go install mvdan.cc/gofumpt@latest; \
	fi
	@gofumpt -w .
	@echo "[ok] Formatting completed successfully ✔️"

.PHONY: imports
imports:
	$(call print_title,"Organizing imports with goimports")
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@goimports -w .
	@echo "[ok] Imports organized successfully ✔️"

.PHONY: tidy
tidy:
	$(call print_title,"Update and Cleaning dependencies")
	@go get -u ./...
	@go mod tidy
	@echo "[ok] Dependencies updated and cleaned successfully ✔️"

#-------------------------------------------------------
# Security Commands
#-------------------------------------------------------

.PHONY: sec-gosec
sec-gosec:
	$(call print_title,"Running gosec security scanner")
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@if find ./components ./pkg -name "*.go" -type f | grep -q .; then \
		echo "Running gosec on components/ and pkg/ folders..."; \
		gosec -quiet ./components/... ./pkg/...; \
		echo "[ok] gosec completed ✔️"; \
	else \
		echo "No Go files found, skipping gosec"; \
	fi

.PHONY: sec-govulncheck
sec-govulncheck:
	$(call print_title,"Running govulncheck vulnerability scanner")
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@if find . -name "*.go" -type f | grep -q .; then \
		echo "Running govulncheck..."; \
		govulncheck ./...; \
		echo "[ok] govulncheck completed ✔️"; \
	else \
		echo "No Go files found, skipping govulncheck"; \
	fi

.PHONY: sec
sec:
	$(call print_title,"Running security checks")
	@$(MAKE) sec-gosec
	@$(MAKE) sec-govulncheck
	@echo "[ok] All security checks completed ✔️"

#-------------------------------------------------------
# Clean Commands
#-------------------------------------------------------

.PHONY: clean
clean:
	@./scripts/clean-artifacts.sh

.PHONY: clean-docker
clean-docker:
	$(call print_title,Cleaning all Docker resources)
	@echo "Cleaning backend Docker resources..."
	@for dir in $(BACKEND_COMPONENTS); do \
		component_name=$$(basename $$dir); \
		echo "Cleaning $$component_name Docker resources..."; \
		(cd $$dir && $(MAKE) clean-docker 2>/dev/null || true); \
	done
	@echo "Cleaning infrastructure Docker resources..."
	@cd $(INFRA_DIR) && $(MAKE) clean-docker 2>/dev/null || true
	@echo "Pruning system-wide Docker resources..."
	@docker system prune -f
	@echo "Pruning system-wide Docker volumes..."
	@docker volume prune -f
	@echo "[ok] All Docker resources cleaned successfully"

#-------------------------------------------------------
# Docker Commands
#-------------------------------------------------------

.PHONY: run
run:
	$(call print_title,"Running the manager application")
	@cd components/manager && go run cmd/app/main.go
	@echo "[ok] Application started successfully ✔️"

.PHONY: build-docker
build-docker:
	$(call print_title,"Building Docker images")
	@$(DOCKER_CMD) -f components/manager/docker-compose.yml build $(c)
	@$(DOCKER_CMD) -f components/worker/docker-compose.yml build $(c)
	@echo "[ok] Docker images built successfully ✔️"

.PHONY: up
up:
	$(call print_title,"Starting services")
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
	$(call print_title,"Starting existing containers")
	@$(DOCKER_CMD) -f docker-compose.yml start $(c)
	@echo "[ok] Containers started successfully ✔️"

.PHONY: down
down:
	$(call print_title,"Stopping services")
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
	$(call print_title,"Restarting services")
	@make down && make up
	@echo "[ok] Backend services successfully ✔️"

.PHONY: rebuild-up
rebuild-up:
	$(call print_title,"Rebuilding and restarting services")
	@echo "Rebuilding infrastructure services..."
	@cd $(INFRA_DIR) && ($(DOCKER_CMD) -f docker-compose.yml build --no-cache && $(DOCKER_CMD) -f docker-compose.yml up -d --force-recreate)
	@echo "Rebuilding backend components..."
	@for dir in $(BACKEND_COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "Rebuilding services in $$dir..."; \
			(cd $$dir && $(DOCKER_CMD) -f docker-compose.yml build --no-cache && $(DOCKER_CMD) -f docker-compose.yml up -d --force-recreate) || exit 1; \
		fi; \
	done
	@echo "[ok] Services rebuilt and restarted successfully ✔️"

.PHONY: logs
logs:
	$(call print_title,"Showing logs for all services")
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
	$(call print_title,"Showing logs for reporter service")
	@$(DOCKER_CMD) -f docker-compose.yml logs --tail=100 -f reporter

.PHONY: ps
ps:
	$(call print_title,"Listing container status")
	@$(DOCKER_CMD) -f docker-compose.yml ps

#-------------------------------------------------------
# Docs Commands
#-------------------------------------------------------

.PHONY: serve-docs
serve-docs:
	@echo "Serving Swagger UI at http://localhost:8080"
	@docker run --rm -p 8080:8080 \
		-e SWAGGER_JSON=/api/swagger.json \
		-v $(shell pwd)/components/manager/api:/api \
		swaggerapi/swagger-ui

.PHONY: generate-docs
generate-docs:
	$(call print_title,"Generating Swagger API documentation")
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
	@echo "[ok] Swagger API documentation generated successfully ✔️"

.PHONY: validate-api-docs
validate-api-docs: generate-docs
	$(call print_title,"Validating API documentation")
	@if [ -f "scripts/validate-api-docs.js" ] && [ -f "$(ROOT_DIR)/scripts/package.json" ]; then \
		echo "Validating API documentation structure..."; \
		cd $(ROOT_DIR)/scripts && node $(ROOT_DIR)/scripts/validate-api-docs.js; \
		echo "Validating API implementations..."; \
		cd $(ROOT_DIR)/scripts && node $(ROOT_DIR)/scripts/validate-api-implementations.js; \
		echo "[ok] API documentation validation completed ✔️"; \
	else \
		echo "Validation scripts not found. Skipping validation."; \
	fi

#-------------------------------------------------------
# Code Generation Commands
#-------------------------------------------------------

.PHONY: generate
generate:
	$(call print_title,"Running code generation")
	@go generate ./...
	@echo "[ok] Code generation completed ✔️"

.PHONY: generate-mocks
generate-mocks:
	$(call print_title,"Generating mock files")
	@go generate ./...
	@echo "[ok] Mock generation completed ✔️"

#-------------------------------------------------------
# Developer Helper Commands
#-------------------------------------------------------

.PHONY: check-tools
check-tools:
	$(call print_title,"Checking required tools")
	@err=0; \
	for tool in go docker golangci-lint swag mockgen gosec govulncheck gofumpt goimports; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo "[ok] $$tool is installed"; \
		else \
			echo "[missing] $$tool is NOT installed"; \
			err=1; \
		fi; \
	done; \
	if [ $$err -eq 1 ]; then \
		echo ""; \
		echo "Run 'make dev-setup' to install missing tools"; \
		exit 1; \
	fi; \
	echo "[ok] All tools available ✔️"

.PHONY: dev-setup
dev-setup:
	$(call print_title,"Setting up development environment")
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
		go install go.uber.org/mock/mockgen@v0.6.0; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@if ! command -v gofumpt >/dev/null 2>&1; then \
		echo "Installing gofumpt..."; \
		go install mvdan.cc/gofumpt@latest; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@echo "Setting up environment..."
	@if [ -f .env.example ] && [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from template"; \
	fi
	@make tidy
	@make check-tests
	@make sec
	@echo "[ok] Development environment set up successfully ✔️"
	@echo "You're ready to start developing! Here are some useful commands:"
	@echo "  make build         - Build the component"
	@echo "  make test          - Run tests"
	@echo "  make up            - Start services"
	@echo "  make rebuild-up    - Rebuild and restart services during development"