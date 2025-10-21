# Midaz Test Infrastructure Makefile

SERVICE_NAME := midaz-test-infra
DOCKER_COMPOSE_FILE := docker-compose.infra.yml
ENV_FILE := .env

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
CYAN := \033[0;36m
BOLD := \033[1m
NC := \033[0m

# Docker compose command
DOCKER_CMD := docker compose

# Define title function
define title1
	@echo ""
	@echo "-----------------------------------"
	@echo "   📝 $(1)  "
	@echo "-----------------------------------"
endef

#-------------------------------------------------------
# Help Command
#-------------------------------------------------------

.PHONY: help
help:
	@echo ""
	@echo "$(BOLD)Midaz E2E Test Infrastructure Commands$(NC)"
	@echo ""
	@echo "$(BOLD)Main Commands:$(NC)"
	@echo "  make first-run           - 🌟 First time setup (does everything)"
	@echo "  make up-infra            - Start infrastructure services (postgres, mongo, redis, rabbitmq)"
	@echo "  make up                  - Alias for up-infra"
	@echo "  make up-all              - Start infrastructure + Midaz services"
	@echo "  make down                - Stop and remove all containers and networks"
	@echo "  make down-all            - Stop infrastructure + Midaz services"
	@echo "  make restart             - Restart infrastructure only"
	@echo "  make restart-all         - 🔄 Restart EVERYTHING (infra + services)"
	@echo "  make logs                - Show logs for all services"
	@echo "  make ps                  - List container status"
	@echo "  make status              - 📊 Quick status of all services"
	@echo ""
	@echo "$(BOLD)Service-specific Commands:$(NC)"
	@echo "  make up-postgres         - Start only PostgreSQL services"
	@echo "  make up-mongo            - Start only MongoDB services"
	@echo "  make up-redis            - Start only Redis service"
	@echo "  make up-rabbitmq         - Start only RabbitMQ service"
	@echo "  make up-onboarding       - Start only Onboarding service"
	@echo "  make up-transaction      - Start only Transaction service"
	@echo "  make up-console          - Start only Console service"
	@echo "  make up-nginx            - Start Nginx reverse proxy"
	@echo ""
	@echo "$(BOLD)Onboarding Commands:$(NC)"
	@echo "  make start-onboarding    - Start onboarding service"
	@echo "  make stop-onboarding     - Stop onboarding service"
	@echo "  make logs-onboarding     - Show onboarding service logs"
	@echo "  make restart-onboarding  - Restart onboarding service"
	@echo ""
	@echo "$(BOLD)Transaction Commands:$(NC)"
	@echo "  make start-transaction   - Start transaction service"
	@echo "  make stop-transaction    - Stop transaction service"
	@echo "  make logs-transaction    - Show transaction service logs"
	@echo "  make restart-transaction - Restart transaction service"
	@echo ""
	@echo "$(BOLD)Console Commands:$(NC)"
	@echo "  make start-console       - Start console service"
	@echo "  make stop-console        - Stop console service"
	@echo "  make logs-console        - Show console service logs"
	@echo "  make restart-console     - Restart console service"
	@echo ""
	@echo "$(BOLD)Nginx Commands:$(NC)"
	@echo "  make up-nginx            - Start Nginx reverse proxy"
	@echo "  make stop-nginx          - Stop Nginx service"
	@echo "  make down-nginx          - Stop and remove Nginx service"
	@echo "  make logs-nginx          - Show Nginx logs"
	@echo "  make restart-nginx       - Restart Nginx service"
	@echo "  make reload-nginx        - Reload Nginx configuration"
	@echo ""
	@echo "$(BOLD)Pre-Seed Commands:$(NC)"
	@echo "  make pre-seed            - Populate Midaz with test data"
	@echo "  make clean-seed          - Remove generated seed data file"
	@echo ""
	@echo "$(BOLD)E2E Test Commands:$(NC)"
	@echo "  make test-setup          - Install Playwright browsers (first time only)"
	@echo "  make test-e2e            - Run all E2E tests"
	@echo "  make test-e2e-phase1     - Run E2E tests (holders + alias accounts)"
	@echo "  make test-e2e-full       - Run tests + cleanup (⭐ recommended)"
	@echo "  make test-e2e-ui         - Run E2E tests with UI mode (visual debug)"
	@echo "  make test-e2e-debug      - Run E2E tests in debug mode"
	@echo "  make test-e2e-report     - Show test report (screenshots & videos)"
	@echo "  make test-cleanup        - Clean up test data (holders, aliases)"
	@echo ""
	@echo "$(BOLD)Clean Commands:$(NC)"
	@echo "  make clean               - Remove all volumes and containers"
	@echo "  make clean-volumes       - Remove only volumes (preserves containers)"
	@echo "  make rebuild             - Clean and rebuild infrastructure only"
	@echo "  make rebuild-all         - 🔄 Clean and rebuild EVERYTHING (infra + services)"
	@echo ""
	@echo "$(BOLD)Setup Commands:$(NC)"
	@echo "  make setup-databases     - Apply database schemas"
	@echo "  make first-run           - Complete first-time setup"
	@echo ""
	@echo "$(BOLD)Health Commands:$(NC)"
	@echo "  make health              - Check health of all services"
	@echo "  make wait-ready          - Wait for all services to be ready"
	@echo ""

#-------------------------------------------------------
# Main Docker Commands
#-------------------------------------------------------

.PHONY: up-infra
up-infra:
	$(call title1,"Starting Midaz E2E Test Infrastructure")
	@$(DOCKER_CMD) --env-file $(ENV_FILE) -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure started successfully$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Waiting for services to be healthy...$(NC)"
	@sleep 5
	@make health

.PHONY: up
up: up-infra

.PHONY: up-all
up-all:
	$(call title1,"Starting Complete E2E Stack")
	@make up
	@echo "$(CYAN)Waiting for infrastructure to be ready...$(NC)"
	@sleep 10
	@make up-onboarding
	@sleep 5
	@make up-transaction
	@sleep 5
	@make up-console
	@sleep 5
	@make up-nginx
	@sleep 2
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All services started successfully$(GREEN) ✔️$(NC)"

.PHONY: down
down:
	$(call title1,"Stopping Midaz E2E Test Infrastructure")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) down
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure stopped successfully$(GREEN) ✔️$(NC)"

.PHONY: down-all
down-all:
	$(call title1,"Stopping Complete E2E Stack")
	@$(DOCKER_CMD) --env-file .env -f docker-compose.infra.yml down 2>/dev/null || true
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml down 2>/dev/null || true
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml down 2>/dev/null || true
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml down 2>/dev/null || true
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml down 2>/dev/null || true
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All services stopped successfully$(GREEN) ✔️$(NC)"

.PHONY: restart
restart:
	$(call title1,"Restarting Midaz E2E Test Infrastructure")
	@make down
	@make up
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure restarted successfully$(GREEN) ✔️$(NC)"

.PHONY: restart-all
restart-all:
	$(call title1,"Restarting Complete Stack (Infrastructure + All Services)")
	@echo "$(CYAN)Stopping all services...$(NC)"
	@make down-all
	@echo "$(CYAN)Starting infrastructure...$(NC)"
	@make up
	@sleep 10
	@echo "$(CYAN)Starting all Midaz services...$(NC)"
	@make up-onboarding
	@sleep 5
	@make up-transaction
	@sleep 5
	@make up-console
	@sleep 5
	@echo "$(CYAN)Verifying health...$(NC)"
	@make health-all
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Complete stack restarted successfully!$(GREEN) ✔️$(NC)"

.PHONY: stop
stop:
	$(call title1,"Stopping containers")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) stop
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Containers stopped successfully$(GREEN) ✔️$(NC)"

.PHONY: start
start:
	$(call title1,"Starting existing containers")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) start
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Containers started successfully$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Service-specific Commands
#-------------------------------------------------------

.PHONY: up-postgres
up-postgres:
	$(call title1,"Starting PostgreSQL services")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) up -d midaz-postgres-primary-test midaz-postgres-replica-test
	@echo "$(GREEN)$(BOLD)[ok]$(NC) PostgreSQL services started$(GREEN) ✔️$(NC)"

.PHONY: up-mongo
up-mongo:
	$(call title1,"Starting MongoDB services")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) up -d midaz-mongodb-test midaz-mongodb-test-init
	@echo "$(GREEN)$(BOLD)[ok]$(NC) MongoDB services started$(GREEN) ✔️$(NC)"

.PHONY: up-redis
up-redis:
	$(call title1,"Starting Redis service")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) up -d midaz-valkey-test
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Redis service started$(GREEN) ✔️$(NC)"

.PHONY: up-rabbitmq
up-rabbitmq:
	$(call title1,"Starting RabbitMQ service")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) up -d midaz-rabbitmq-test
	@echo "$(GREEN)$(BOLD)[ok]$(NC) RabbitMQ service started$(GREEN) ✔️$(NC)"

.PHONY: up-onboarding
up-onboarding:
	$(call title1,"Starting Onboarding service")
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml up -d
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Onboarding service started$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Waiting for onboarding to be healthy...$(NC)"
	@sleep 5

.PHONY: up-transaction
up-transaction:
	$(call title1,"Starting Transaction service")
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml up -d
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Transaction service started$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Waiting for transaction to be healthy...$(NC)"
	@sleep 5

.PHONY: up-console
up-console:
	$(call title1,"Starting Console service")
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml up -d
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Console service started$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Waiting for console to be healthy...$(NC)"
	@sleep 5

#-------------------------------------------------------
# Logs Commands
#-------------------------------------------------------

.PHONY: logs
logs:
	$(call title1,"Showing logs for all services")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) logs --tail=100 -f

.PHONY: logs-postgres
logs-postgres:
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) logs --tail=50 -f midaz-postgres-primary-test midaz-postgres-replica-test

.PHONY: logs-mongo
logs-mongo:
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) logs --tail=50 -f midaz-mongodb-test midaz-mongodb-test-init

.PHONY: logs-redis
logs-redis:
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) logs --tail=50 -f midaz-valkey-test

.PHONY: logs-rabbitmq
logs-rabbitmq:
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) logs --tail=50 -f midaz-rabbitmq-test

.PHONY: logs-onboarding
logs-onboarding:
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml logs --tail=100 -f midaz-onboarding-test

.PHONY: logs-transaction
logs-transaction:
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml logs --tail=100 -f midaz-transaction-test

.PHONY: logs-console
logs-console:
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml logs --tail=100 -f midaz-console-test

#-------------------------------------------------------
# Status Commands
#-------------------------------------------------------

.PHONY: ps
ps:
	$(call title1,"Listing container status")
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) ps

.PHONY: status
status:
	$(call title1,"Quick Status Check")
	@echo "$(BOLD)Infrastructure:$(NC)"
	@docker ps --filter "name=midaz" --filter "name=postgres\|mongo\|valkey\|rabbit\|otel" --format "  {{.Names}}: {{.Status}}" | head -6
	@echo ""
	@echo "$(BOLD)Midaz Services:$(NC)"
	@docker ps --filter "name=midaz-onboarding-test\|midaz-transaction-test\|midaz-console-test" --format "  {{.Names}}: {{.Status}}"
	@echo ""
	@echo "$(BOLD)Health Status:$(NC)"
	@echo -n "  Onboarding:  " && (curl -sf http://localhost:3100/health >/dev/null 2>&1 && echo "$(GREEN)✅ UP$(NC)" || echo "$(YELLOW)❌ DOWN$(NC)")
	@echo -n "  Transaction: " && (curl -sf http://localhost:3101/health >/dev/null 2>&1 && echo "$(GREEN)✅ UP$(NC)" || echo "$(YELLOW)❌ DOWN$(NC)")
	@echo -n "  Console:     " && (curl -sf http://localhost:8081/api/admin/health/ready >/dev/null 2>&1 && echo "$(GREEN)✅ UP$(NC)" || echo "$(YELLOW)❌ DOWN$(NC)")

.PHONY: health
health:
	$(call title1,"Checking service health")
	@echo "$(CYAN)PostgreSQL Primary:$(NC)"
	@docker exec midaz-postgres-primary-test pg_isready -U crm_test -p 5801 2>/dev/null && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"
	@echo "$(CYAN)PostgreSQL Replica:$(NC)"
	@docker exec midaz-postgres-replica-test pg_isready -U crm_test -p 5802 2>/dev/null && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"
	@echo "$(CYAN)MongoDB:$(NC)"
	@docker exec midaz-mongodb-test mongosh --host midaz-mongodb-test --port 5803 --eval "db.adminCommand('ping')" >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"
	@echo "$(CYAN)Redis:$(NC)"
	@docker exec midaz-valkey-test redis-cli -p 5804 -a crm_test_password ping >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"
	@echo "$(CYAN)RabbitMQ:$(NC)"
	@docker exec midaz-rabbitmq-test rabbitmq-diagnostics -q ping >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"

.PHONY: health-onboarding
health-onboarding:
	@echo "$(CYAN)Onboarding Service:$(NC)"
	@curl -sf http://localhost:3100/health >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"

.PHONY: health-transaction
health-transaction:
	@echo "$(CYAN)Transaction Service:$(NC)"
	@curl -sf http://localhost:3101/health >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"

.PHONY: health-console
health-console:
	@echo "$(CYAN)Console Service:$(NC)"
	@curl -sf http://localhost:8081/api/admin/health/ready >/dev/null 2>&1 && echo "$(GREEN)✔️ Healthy$(NC)" || echo "$(YELLOW)⚠️  Not ready$(NC)"

.PHONY: health-all
health-all:
	@make health
	@make health-onboarding
	@make health-transaction
	@make health-console

.PHONY: wait-ready
wait-ready:
	$(call title1,"Waiting for all services to be ready")
	@echo "$(CYAN)This may take up to 90 seconds...$(NC)"
	@bash -c ' \
		timeout=90; \
		elapsed=0; \
		while [ $$elapsed -lt $$timeout ]; do \
			if docker exec midaz-postgres-primary-test pg_isready -U crm_test -p 5801 >/dev/null 2>&1 && \
			   docker exec midaz-mongodb-test mongosh --host midaz-mongodb-test --port 5803 --eval "db.adminCommand(\"ping\")" >/dev/null 2>&1 && \
			   docker exec midaz-valkey-test redis-cli -p 5804 -a crm_test_password ping >/dev/null 2>&1 && \
			   docker exec midaz-rabbitmq-test rabbitmq-diagnostics -q ping >/dev/null 2>&1; then \
				echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure is ready$(GREEN) ✔️$(NC)"; \
				echo "$(CYAN)Checking if PostgreSQL is HEALTHY (not just ready)...$(NC)"; \
				if docker inspect midaz-postgres-primary-test | grep -q "\"Status\": \"healthy\""; then \
					echo "$(GREEN)$(BOLD)[ok]$(NC) PostgreSQL is healthy$(GREEN) ✔️$(NC)"; \
					exit 0; \
				else \
					echo "$(YELLOW)PostgreSQL is ready but not yet healthy, waiting...$(NC)"; \
				fi; \
			fi; \
			echo "$(YELLOW)Waiting... ($$elapsed/$$timeout seconds)$(NC)"; \
			sleep 5; \
			elapsed=$$((elapsed + 5)); \
		done; \
		echo "$(YELLOW)$(BOLD)[warning]$(NC) Timeout reached. Some services may not be ready.$(NC)"; \
		make health; \
		exit 1'

#-------------------------------------------------------
# Clean Commands
#-------------------------------------------------------

.PHONY: clean
clean:
	$(call title1,"Cleaning all infrastructure resources")
	@echo "$(YELLOW)Stopping and removing containers, networks, and volumes...$(NC)"
	@$(DOCKER_CMD) -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure cleaned successfully$(GREEN) ✔️$(NC)"

.PHONY: clean-volumes
clean-volumes:
	$(call title1,"Removing all volumes")
	@echo "$(YELLOW)Removing test volumes...$(NC)"
	@docker volume rm postgres-test-data postgres-replica-test-data redis-test-data grafana-test-data mongodb-test-data rabbitmq-test-data rabbitmq-test-logs 2>/dev/null || true
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Volumes removed successfully$(GREEN) ✔️$(NC)"

.PHONY: rebuild
rebuild:
	$(call title1,"Rebuilding infrastructure from scratch")
	@make clean
	@make up
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Infrastructure rebuilt successfully$(GREEN) ✔️$(NC)"

.PHONY: setup-databases
setup-databases:
	$(call title1,"Setting up databases schemas")
	@echo "$(CYAN)Creating databases...$(NC)"
	@docker exec midaz-postgres-primary-test psql -U crm_test -d postgres -c "CREATE DATABASE onboarding;" 2>/dev/null || echo "  Database onboarding already exists"
	@docker exec midaz-postgres-primary-test psql -U crm_test -d postgres -c "CREATE DATABASE transaction;" 2>/dev/null || echo "  Database transaction already exists"
	@echo "$(GREEN)✔️ Databases created$(NC)"
	@echo "$(CYAN)Applying onboarding schema...$(NC)"
	@docker cp configs/postgres/onboarding-schema.sql midaz-postgres-primary-test:/tmp/onboarding-schema.sql
	@docker exec midaz-postgres-primary-test psql -U crm_test -d onboarding -f /tmp/onboarding-schema.sql >/dev/null 2>&1
	@docker exec midaz-postgres-primary-test psql -U crm_test -d onboarding -c "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint PRIMARY KEY, dirty boolean); INSERT INTO schema_migrations (version, dirty) VALUES (6, false) ON CONFLICT (version) DO UPDATE SET dirty = false;" >/dev/null 2>&1
	@echo "$(GREEN)✔️ Onboarding schema applied$(NC)"
	@echo "$(CYAN)Applying transaction schema...$(NC)"
	@docker cp configs/postgres/transaction-schema.sql midaz-postgres-primary-test:/tmp/transaction-schema.sql
	@docker exec midaz-postgres-primary-test psql -U crm_test -d transaction -f /tmp/transaction-schema.sql >/dev/null 2>&1
	@docker exec midaz-postgres-primary-test psql -U crm_test -d transaction -c "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint PRIMARY KEY, dirty boolean); INSERT INTO schema_migrations (version, dirty) VALUES (13, false) ON CONFLICT (version) DO UPDATE SET dirty = false;" >/dev/null 2>&1
	@echo "$(GREEN)✔️ Transaction schema applied$(NC)"
	@echo "$(GREEN)$(BOLD)[ok]$(NC) All database schemas configured$(GREEN) ✔️$(NC)"

.PHONY: first-run
first-run:
	$(call title1,"First Run - Complete Setup")
	@echo "$(CYAN)This will set up everything from scratch...$(NC)"
	@echo "$(YELLOW)Cleaning previous environment...$(NC)"
	@make clean 2>/dev/null || true
	@echo "$(CYAN)Starting infrastructure...$(NC)"
	@make up
	@echo "$(CYAN)Waiting for PostgreSQL to be fully healthy (30s)...$(NC)"
	@sleep 30
	@echo "$(CYAN)Verifying PostgreSQL health...$(NC)"
	@docker exec midaz-postgres-primary-test pg_isready -U crm_test -p 5801 || (echo "$(YELLOW)PostgreSQL not ready, waiting more...$(NC)" && sleep 10)
	@echo "$(CYAN)Setting up database schemas...$(NC)"
	@make setup-databases
	@echo "$(CYAN)Starting Midaz services...$(NC)"
	@make up-onboarding
	@sleep 8
	@make up-transaction
	@sleep 8
	@make up-console
	@sleep 8
	@echo "$(CYAN)Verifying health...$(NC)"
	@make health-all
	@echo ""
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Complete setup finished!$(GREEN) ✔️$(NC)"
	@echo ""
	@make info

.PHONY: rebuild-all
rebuild-all:
	$(call title1,"Rebuilding Complete Stack (Infrastructure + All Services)")
	@echo "$(YELLOW)⚠️  This will stop and rebuild EVERYTHING$(NC)"
	@echo "$(YELLOW)⚠️  All data will be lost$(NC)"
	@echo "$(CYAN)Stopping all services...$(NC)"
	@make down-all
	@echo "$(CYAN)Cleaning all volumes and containers...$(NC)"
	@make clean
	@echo "$(CYAN)Starting infrastructure...$(NC)"
	@make up
	@echo "$(CYAN)Waiting for PostgreSQL to be fully healthy (30s)...$(NC)"
	@sleep 30
	@docker exec midaz-postgres-primary-test pg_isready -U crm_test -p 5801 || (echo "$(YELLOW)PostgreSQL not ready, waiting more...$(NC)" && sleep 10)
	@echo "$(CYAN)Setting up database schemas...$(NC)"
	@make setup-databases
	@echo "$(CYAN)Starting all Midaz services...$(NC)"
	@make up-onboarding
	@sleep 8
	@make up-transaction
	@sleep 8
	@make up-console
	@sleep 8
	@echo "$(CYAN)Verifying health of all services...$(NC)"
	@make health-all
	@echo ""
	@echo "$(CYAN)Populating Midaz with test data (pre-seed)...$(NC)"
	@./pre-seed.sh
	@echo ""
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Complete stack rebuilt successfully!$(GREEN) ✔️$(NC)"
	@echo ""
	@echo "$(CYAN)Current status:$(NC)"
	@make info
	@echo ""
	@echo "$(YELLOW)Pre-seed IDs saved to .env.seed$(NC)"
	@echo "$(YELLOW)Run 'source .env.seed' to use the generated IDs$(NC)"
	@echo ""
	@echo "$(GREEN)$(BOLD)✨ Stack is ready!$(NC)$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Development Helpers
#-------------------------------------------------------

.PHONY: shell-postgres
shell-postgres:
	@docker exec -it midaz-postgres-primary-test psql -U crm_test -d crm_test

.PHONY: shell-mongo
shell-mongo:
	@docker exec -it midaz-mongodb-test mongosh --host midaz-mongodb-test --port 5803 -u crm_test -p crm_test_password

.PHONY: shell-redis
shell-redis:
	@docker exec -it midaz-valkey-test redis-cli -p 5804 -a crm_test_password

.PHONY: shell-rabbitmq
shell-rabbitmq:
	@echo "$(CYAN)RabbitMQ Management UI: http://localhost:3104$(NC)"
	@echo "$(CYAN)Username: crm_test$(NC)"
	@echo "$(CYAN)Password: crm_test_password$(NC)"

#-------------------------------------------------------
# Onboarding Service Commands
#-------------------------------------------------------

.PHONY: start-onboarding
start-onboarding:
	$(call title1,"Starting Onboarding service")
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml start
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Onboarding service started$(GREEN) ✔️$(NC)"

.PHONY: stop-onboarding
stop-onboarding:
	$(call title1,"Stopping Onboarding service")
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml stop
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Onboarding service stopped$(GREEN) ✔️$(NC)"

.PHONY: restart-onboarding
restart-onboarding:
	$(call title1,"Restarting Onboarding service")
	@$(DOCKER_CMD) --env-file .env.onboarding -f docker-compose.onboarding.yml restart
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Onboarding service restarted$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Transaction Service Commands
#-------------------------------------------------------

.PHONY: start-transaction
start-transaction:
	$(call title1,"Starting Transaction service")
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml start
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Transaction service started$(GREEN) ✔️$(NC)"

.PHONY: stop-transaction
stop-transaction:
	$(call title1,"Stopping Transaction service")
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml stop
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Transaction service stopped$(GREEN) ✔️$(NC)"

.PHONY: restart-transaction
restart-transaction:
	$(call title1,"Restarting Transaction service")
	@$(DOCKER_CMD) --env-file .env.transaction -f docker-compose.transaction.yml restart
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Transaction service restarted$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Console Service Commands
#-------------------------------------------------------

.PHONY: start-console
start-console:
	$(call title1,"Starting Console service")
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml start
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Console service started$(GREEN) ✔️$(NC)"

.PHONY: stop-console
stop-console:
	$(call title1,"Stopping Console service")
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml stop
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Console service stopped$(GREEN) ✔️$(NC)"

.PHONY: restart-console
restart-console:
	$(call title1,"Restarting Console service")
	@$(DOCKER_CMD) --env-file .env.console -f docker-compose.console.yml restart
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Console service restarted$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Nginx Service Commands
#-------------------------------------------------------

.PHONY: up-nginx
up-nginx:
	$(call title1,"Starting Nginx Reverse Proxy")
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml up -d
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Nginx service started$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Nginx is now routing:$(NC)"
	@echo "  / → Console (http://localhost:8080)"

.PHONY: stop-nginx
stop-nginx:
	$(call title1,"Stopping Nginx service")
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml stop
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Nginx service stopped$(GREEN) ✔️$(NC)"

.PHONY: down-nginx
down-nginx:
	$(call title1,"Stopping and removing Nginx service")
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml down
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Nginx service removed$(GREEN) ✔️$(NC)"

.PHONY: logs-nginx
logs-nginx:
	$(call title1,"Nginx Logs")
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml logs -f

.PHONY: restart-nginx
restart-nginx:
	$(call title1,"Restarting Nginx service")
	@$(DOCKER_CMD) --env-file .env.nginx -f docker-compose.nginx.yml restart
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Nginx service restarted$(GREEN) ✔️$(NC)"

.PHONY: reload-nginx
reload-nginx:
	$(call title1,"Reloading Nginx configuration")
	@docker exec midaz-nginx-test nginx -s reload
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Nginx configuration reloaded$(GREEN) ✔️$(NC)"

#-------------------------------------------------------
# Pre-Seed Commands
#-------------------------------------------------------

.PHONY: pre-seed
pre-seed:
	$(call title1,"Running Pre-Seed Script")
	@echo "$(CYAN)Populating Midaz with test data...$(NC)"
	@./pre-seed.sh
	@echo ""
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Pre-seed completed$(GREEN) ✔️$(NC)"
	@echo "$(CYAN)Generated IDs saved to .env.seed$(NC)"
	@echo ""
	@echo "$(YELLOW)To use the generated IDs:$(NC)"
	@echo "  source .env.seed"

.PHONY: clean-seed
clean-seed:
	$(call title1,"Cleaning Seed Data")
	@if [ -f .env.seed ]; then \
		rm -f .env.seed; \
		echo "$(GREEN)$(BOLD)[ok]$(NC) Removed .env.seed file$(GREEN) ✔️$(NC)"; \
	else \
		echo "$(YELLOW)No .env.seed file found$(NC)"; \
	fi

#-------------------------------------------------------
# E2E Test Commands
#-------------------------------------------------------

.PHONY: test-setup
test-setup:
	$(call title1,"Installing Playwright Browsers")
	@echo "$(CYAN)Installing Chromium browser...$(NC)"
	@cd ../../ && npx playwright install chromium
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Playwright browsers installed$(GREEN) ✔️$(NC)"

.PHONY: test-e2e
test-e2e:
	$(call title1,"Running E2E Tests")
	@echo "$(CYAN)Running all E2E tests...$(NC)"
	@cd ../../ && npm run test:e2e
	@echo "$(GREEN)$(BOLD)[ok]$(NC) E2E tests completed$(GREEN) ✔️$(NC)"

.PHONY: test-e2e-phase1
test-e2e-phase1:
	$(call title1,"Running Phase 1 E2E Tests")
	@echo "$(CYAN)Ensuring Nginx is running...$(NC)"
	@docker ps | grep midaz-nginx-test > /dev/null || (echo "$(YELLOW)Starting Nginx...$(NC)" && make up-nginx)
	@echo "$(CYAN)Running E2E tests (holders + alias accounts)...$(NC)"
	@cd ../../ && npx playwright test holders-list holders-create-natural-person holders-create-legal-person holders-edit holders-delete alias-create alias-create-from-tab alias-create-scenarios alias-edit
	@echo "$(GREEN)$(BOLD)[ok]$(NC) E2E tests completed$(GREEN) ✔️$(NC)"

.PHONY: test-e2e-phase1-ui
test-e2e-phase1-ui:
	$(call title1,"Running Phase 1 UI Tests (complex - need selector adjustment)")
	@echo "$(YELLOW)⚠️  UI tests require selector refinement$(NC)"
	@echo "$(CYAN)Running UI tests...$(NC)"
	@cd ../../ && npx playwright test phase1-01 phase1-02 phase1-03
	@echo "$(GREEN)$(BOLD)[ok]$(NC) UI tests completed$(GREEN) ✔️$(NC)"

.PHONY: test-e2e-ui
test-e2e-ui:
	$(call title1,"Running E2E Tests - UI Mode")
	@echo "$(CYAN)Opening Playwright UI...$(NC)"
	@cd ../../ && npx playwright test --ui

.PHONY: test-e2e-debug
test-e2e-debug:
	$(call title1,"Running E2E Tests - Debug Mode")
	@echo "$(CYAN)Running tests in debug mode...$(NC)"
	@cd ../../ && npx playwright test --debug

.PHONY: test-e2e-report
test-e2e-report:
	$(call title1,"Opening E2E Test Report")
	@cd ../../ && npx playwright show-report

.PHONY: test-cleanup
test-cleanup:
	$(call title1,"Cleaning Up Playwright Test Data")
	@echo "$(CYAN)Removendo dados criados pelo Playwright (preserva pre-seed)...$(NC)"
	@./cleanup-playwright-data.sh
	@echo "$(GREEN)$(BOLD)[ok]$(NC) Playwright data cleaned up$(GREEN) ✔️$(NC)"

.PHONY: test-e2e-full
test-e2e-full:
	$(call title1,"Running E2E Tests with Cleanup")
	@echo "$(CYAN)Executando testes...$(NC)"
	@make test-e2e-phase1
	@echo ""
	@echo "$(CYAN)Limpando dados de teste...$(NC)"
	@make test-cleanup
	@echo ""
	@echo "$(GREEN)$(BOLD)[ok]$(NC) E2E cycle complete$(GREEN) ✔️$(NC)"
	@echo ""
	@echo "$(CYAN)Resumo:$(NC)"
	@echo "  ✅ Testes executados"
	@echo "  ✅ Dados limpos"
	@echo "  ✅ Ambiente pronto para nova execução"

#-------------------------------------------------------
# Info Commands
#-------------------------------------------------------

.PHONY: info
info:
	@echo ""
	@echo "$(BOLD)Midaz E2E Test Infrastructure$(NC)"
	@echo ""
	@echo "$(BOLD)Services:$(NC)"
	@echo "  PostgreSQL Primary:  localhost:5801 (user: crm_test)"
	@echo "  PostgreSQL Replica:  localhost:5802 (user: crm_test)"
	@echo "  MongoDB:             localhost:5803 (user: crm_test)"
	@echo "  Redis/Valkey:        localhost:5804 (password: crm_test_password)"
	@echo "  RabbitMQ AMQP:       localhost:3103"
	@echo "  RabbitMQ Management: http://localhost:3104 (user: crm_test)"
	@echo "  Grafana:             http://localhost:3200 (user: crm_test)"
	@echo ""
	@echo "$(BOLD)Midaz Services:$(NC)"
	@echo "  Onboarding API:      http://localhost:3100"
	@echo "  Onboarding Swagger:  http://localhost:3100/swagger/index.html"
	@echo "  Transaction API:     http://localhost:3101"
	@echo "  Transaction Swagger: http://localhost:3101/swagger/index.html"
	@echo "  Console UI:          http://localhost:8081"
	@echo ""
	@echo "$(BOLD)Nginx Reverse Proxy (Optional):$(NC)"
	@echo "  Nginx:               http://localhost:8080 (if started)"
	@echo "  Console via Nginx:   http://localhost:8080/"
	@echo ""

