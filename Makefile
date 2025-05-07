# Project Root Makefile.
# Coordinates all component Makefiles and provides centralized commands.

# Define the root directory of the project
PROJECT_ROOT := $(shell pwd)

service_name := plugin-smart-templates
bin_dir := ./.bin

BLUE := \033[36m
NC := \033[0m

DOCKER_VERSION := $(shell docker version --format '{{.Server.Version}}')
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
COMPONENTS := $(INFRA_DIR) $(WORKER_DIR) $(MANAGER_DIR)

# Include shared color definitions and utility functions
#include $(PROJECT_ROOT)/pkg/shell/makefile_colors.mk
#include $(PROJECT_ROOT)/pkg/shell/makefile_utils.mk
#include $(PROJECT_ROOT)/pkg/shell/makefile_template.mk

# Display available commands
.PHONY: info
info:
	@echo "                                                                                                                                       "
	@echo "                                                                                                                                       "
	@echo "To run a specific command inside the audit container using make, you can execute:                                                     "
	@echo "                                                                                                                                       "
	@echo "make audit COMMAND=\"any\"                                                                                                            "
	@echo "                                                                                                                                       "
	@echo "This command will run the specified command inside the container. Replace \"any\" with the desired command you want to execute. "
	@echo "                                                                                                                         "
	@echo "## Docker commands:"
	@echo "                                                                                                                         "
	@echo "  COMMAND=\"build\"                                Builds all Docker images defined in docker-compose.yml."
	@echo "  COMMAND=\"up\"                                   Starts and runs all services defined in docker-compose.yml."
	@echo "  COMMAND=\"start\"                                Starts existing containers defined in docker-compose.yml without creating them."
	@echo "  COMMAND=\"stop\"                                 Stops running containers defined in docker-compose.yml without removing them."
	@echo "  COMMAND=\"down\"                                 Stops and removes containers, networks, and volumes defined in docker-compose.yml."
	@echo "  COMMAND=\"destroy\"                              Stops and removes containers, networks, and volumes (including named volumes) defined in docker-compose.yml."
	@echo "  COMMAND=\"restart\"                              Stops and removes containers, networks, and volumes, then starts all services in detached mode."
	@echo "  COMMAND=\"logs\"                                 Shows the last 100 lines of logs and follows live log output for services defined in docker-compose.yml."
	@echo "  COMMAND=\"logs-api\"                             Shows the last 100 lines of logs and follows live log output for the audit service defined in docker-compose.yml."
	@echo "  COMMAND=\"ps\"                                   Lists the status of containers defined in docker-compose.yml."
	@echo "                                                                                                                         "
	@echo "## App commands:"
	@echo "                                                                                                                         "
	@echo "  COMMAND=\"generate-docs\" 						  Generates Swagger API documentation and an OpenAPI Specification."
	@echo "                                                                                                                                       "
	@echo "                                                                                                                                       "

.PHONY: up
up:
	$(call title1,"Starting all services with Docker Compose")
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "$(CYAN)Starting services in $$dir...$(NC)"; \
			(cd $$dir && $(MAKE) up) || exit 1; \
		fi; \
	done
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All services started successfully$(GREEN) ✔️$(NC)"

.PHONY: start
start:
	@docker compose -f docker-compose.yml start $(c)

.PHONY: down
down:
	$(call title1,"Stopping all services with Docker Compose")
	@for dir in $(COMPONENTS); do \
		component_name=$$(basename $$dir); \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "$(CYAN)Stopping services in component: $(BOLD)$$component_name$(NC)"; \
			(cd $$dir && (docker compose -f docker-compose.yml down 2>/dev/null || docker-compose -f docker-compose.yml down)) || exit 1; \
		else \
			echo "$(YELLOW)No docker-compose.yml found in $$component_name, skipping$(NC)"; \
		fi; \
	done
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All services stopped successfully$(GREEN) ✔️$(NC)"

.PHONY: destroy
destroy:
	@$(DOCKER_CMD) -f docker-compose.yml down -v $(c)

.PHONY: stop
stop:
	@$(DOCKER_CMD) -f docker-compose.yml stop $(c)

.PHONY: restart
restart:
	make stop && \
    make up

.PHONY: ps
ps:
	@$(DOCKER_CMD) -f docker-compose.yml ps

.PHONY: generate-docs
generate-docs:
	@swag init -g ./cmd/app/main.go -d ./ -o ./api --parseDependency --parseInternal
	@docker run --rm -v $(pwd):/local --user $(shell id -u):$(shell id -g) openapitools/openapi-generator-cli:v5.1.1 generate -i /local/api/swagger.json -g openapi-yaml -o /local/api
	@mv ./api/openapi/openapi.yaml ./api/openapi.yaml
	@rm -rf ./api/README.md ./api/.openapi-generator* ./api/openapi

.PHONY: setup-git-hooks
setup-git-hooks:
	@echo "$(BLUE)Setting up git hooks...$(NC)"
	./make.sh "setupGitHooks"

.PHONY: check-hooks
check-hooks:
	@echo "$(BLUE)Checking git hooks status...$(NC)"
	./make.sh "checkHooks"

.PHONY: tidy
tidy:
	@echo "$(BLUE)Running go mod tidy...$(NC)"
	go mod tidy

.PHONY: sec
sec:
	@echo "$(BLUE)Running security checks...$(NC)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "$(RED)Error: gosec is not installed$(NC)"; \
		echo "$(MAGENTA)To install: go install github.com/securego/gosec/v2/cmd/gosec@latest$(NC)"; \
		exit 1; \
	fi
	gosec ./...

.PHONY: test
test:
	@echo "$(BLUE)Running tests...$(NC)"
		@if ! command -v go >/dev/null 2>&1; then \
		echo "$(RED)Error: go is not installed$(NC)"; \
		exit 1; \
	fi
	go test -v ./... ./...

#-------------------------------------------------------
# Setup Commands
#-------------------------------------------------------

.PHONY: set-env
set-env:
	$(call title1,"Setting up environment files")
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/.env.example" ] && [ ! -f "$$dir/.env" ]; then \
			echo "$(CYAN)Creating .env in $$dir from .env.example$(NC)"; \
			cp "$$dir/.env.example" "$$dir/.env"; \
		elif [ ! -f "$$dir/.env.example" ]; then \
			echo "$(YELLOW)Warning: No .env.example found in $$dir$(NC)"; \
		else \
			echo "$(GREEN).env already exists in $$dir$(NC)"; \
		fi; \
	done
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Environment files set up successfully$(GREEN) ✔️$(NC)"

.PHONY: clean-docker
clean-docker:
	$(call title1,"Cleaning all Docker resources")
	@for dir in $(COMPONENTS); do \
		if [ -f "$$dir/docker-compose.yml" ]; then \
			echo "$(CYAN)Cleaning Docker resources in $$dir...$(NC)"; \
			(cd $$dir && $(MAKE) clean-docker) || exit 1; \
		fi; \
	done
	@echo "$(YELLOW)Pruning system-wide Docker resources...$(NC)"
	@docker system prune -f
	@echo "$(YELLOW)Pruning system-wide Docker volumes...$(NC)"
	@docker volume prune -f
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All Docker resources cleaned successfully$(GREEN) ✔️$(NC)"