service_name := plugin-template-engine-manager
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

# Display available commands
.PHONY: info
info:
	@echo "                                                                                                                                       "
	@echo "                                                                                                                                       "
	@echo "To run a specific command inside the audit container using make, you can execute:                                                     "
	@echo "                                                                                                                                       "
	@echo "make audit COMMAND=\"any\"                                                                                                            "
	@echo "                                                                                                                                       "
	@echo "This command will run the specified command inside the audit container. Replace \"any\" with the desired command you want to execute. "
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

# Docker Compose Commands
.PHONY: up
up:
	make set-env
	@echo "$(BLUE)Starting all services...$(NC)"
	@$(DOCKER_CMD) -f docker-compose.yml up --build -d
	@echo "$(BLUE)All services started successfully$(NC)"

.PHONY: start
start:
	@docker compose -f docker-compose.yml start $(c)

.PHONY: down
down:
	@$(DOCKER_CMD) -f docker-compose.yml down $(c)

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
	@docker run --rm -v $(pwd):/local --user $(shell id -u):$(shell id -g) openapitools/openapi-generator-cli:v5.1.1 generate -i ./manager/api/swagger.json -g openapi-yaml -o ./manager/api
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

.PHONY: lint
lint:
	$(call title1,"Running linters")
	@if find . -name "*.go" -type f | grep -q .; then \
		if ! command -v golangci-lint >/dev/null 2>&1; then \
			echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
			go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		fi; \
		golangci-lint run --fix ./... --verbose; \
		echo "$(GREEN)$(BOLD)[ok]$(NC) Linting completed successfully$(GREEN) ✔️$(NC)"; \
	else \
		echo "$(YELLOW)No Go files found, skipping linting$(NC)"; \
	fi

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

.PHONY: set-env
set-env:
	@echo "$(BLUE)Setting up environment files...$(NC)"
	cp -r ./.env.example ./.env
	@echo "$(BLUE)Environment files created successfully$(NC)"