TEST_MANAGER_URL ?= http://localhost:4005
TEST_HEALTH_WAIT ?= 60

# macOS ld64 workaround: newer ld emits noisy LC_DYSYMTAB warnings when linking test binaries with -race.
# If available, prefer Apple's classic linker to silence them.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  # Prefer classic mode to suppress LC_DYSYMTAB warnings on macOS.
  # Set DISABLE_OSX_LINKER_WORKAROUND=1 to disable this behavior.
  ifneq ($(DISABLE_OSX_LINKER_WORKAROUND),1)
    GO_TEST_LDFLAGS := -ldflags="-linkmode=external -extldflags=-ld_classic"
  else
    GO_TEST_LDFLAGS :=
  endif
else
  GO_TEST_LDFLAGS :=
endif

define wait_for_services
	bash -c 'echo "Waiting for services to become healthy..."; \
	sleep 60; \
	for i in $$(seq 1 $(TEST_HEALTH_WAIT)); do \
	  if curl -fsS $(TEST_MANAGER_URL)/health >/dev/null 2>&1; then \
	    echo "Services are up"; exit 0; \
	  fi; \
	  sleep 1; \
	done; \
	echo "[error] Services not healthy after $(TEST_HEALTH_WAIT)s"; exit 1'
endef

.PHONY: wait-for-services
wait-for-services:
	$(call wait_for_services)

# ------------------------------------------------------
# Test tooling configuration
# ------------------------------------------------------

TEST_REPORTS_DIR ?= ./reports
GOTESTSUM := $(shell command -v gotestsum 2>/dev/null)
RETRY_ON_FAIL ?= 0

.PHONY: tools tools-gotestsum
tools: tools-gotestsum ## Install helpful dev/test tools

tools-gotestsum:
	@if [ -z "$(GOTESTSUM)" ]; then \
		echo "Installing gotestsum..."; \
		GO111MODULE=on go install gotest.tools/gotestsum@latest; \
	else \
		echo "gotestsum already installed: $(GOTESTSUM)"; \
	fi

#-------------------------------------------------------
# Test Commands
#-------------------------------------------------------

.PHONY: test
test:
	@./scripts/run-tests.sh


#-------------------------------------------------------
# Test Suite Aliases
#-------------------------------------------------------

# Unit tests (exclude ./tests/** packages)
.PHONY: test-unit
test-unit:
	$(call print_title,Running Go unit tests (excluding ./tests/**))
	$(call check_command,go,"Install Go from https://golang.org/doc/install")
	@set -e; mkdir -p $(TEST_REPORTS_DIR)/unit; \
	pkgs=$$(go list ./... | awk '!/\/tests($|\/)/'); \
	if [ -z "$$pkgs" ]; then \
	  echo "No unit test packages found (outside ./tests)**"; \
	else \
	  if [ -n "$(GOTESTSUM)" ]; then \
		echo "Running unit tests with gotestsum"; \
		gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/unit/unit.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) $$pkgs || { \
		  if [ "$(RETRY_ON_FAIL)" = "1" ]; then \
			echo "Retrying unit tests once..."; \
			gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/unit/unit-rerun.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) $$pkgs; \
		  else \
			exit 1; \
		  fi; \
		}; \
	  else \
		go test -v -race -count=1 $(GO_TEST_LDFLAGS) $$pkgs; \
	  fi; \
	fi

.PHONY: test-integration
test-integration:
	$(call print_title,Running Go integration tests (with Docker stack))
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@set -e; mkdir -p $(TEST_REPORTS_DIR)/integration; \
	trap '$(MAKE) -s down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) up; \
	$(call wait_for_services); \
	if [ -n "$(GOTESTSUM)" ]; then \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/integration/integration.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/integration || { \
	    if [ "$(RETRY_ON_FAIL)" = "1" ]; then \
	      echo "Retrying integration tests once..."; \
	      TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/integration/integration-rerun.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/integration; \
	    else \
	      exit 1; \
	    fi; \
	  }; \
	else \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) go test -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/integration; \
	fi

.PHONY: test-e2e
test-e2e:
	$(call print_title,Running E2E tests with Apidog CLI (with Docker stack))
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@set -e; \
	trap '$(MAKE) -s down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) up; \
	$(call wait_for_services); \
	mkdir -p ./reports/e2e; \
	echo "Running Apidog CLI via npx against tests/e2e/local.apidog-cli.json"; \
	npx --yes apidog-cli@latest run ./tests/e2e/local.apidog-cli.json -r html,cli --out-dir ./reports/e2e

# Fuzzy/robustness tests
.PHONY: test-fuzzy
test-fuzzy:
	$(call print_title,Running fuzz/robustness tests - requires Docker stack)
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@set -e; mkdir -p $(TEST_REPORTS_DIR)/fuzzy; \
	trap '$(MAKE) -s down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) up; \
	$(call wait_for_services); \
	if [ -n "$(GOTESTSUM)" ]; then \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/fuzzy/fuzzy.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/fuzzy || { \
	    if [ "$(RETRY_ON_FAIL)" = "1" ]; then \
	      echo "Retrying fuzzy tests once..."; \
	      TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/fuzzy/fuzzy-rerun.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/fuzzy; \
	    else \
	      exit 1; \
	    fi; \
	  }; \
	else \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) go test -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/fuzzy; \
	fi

# Property-based tests
.PHONY: test-property
test-property:
	$(call print_title,Running property-based tests - requires Docker stack)
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@set -e; mkdir -p $(TEST_REPORTS_DIR)/property; \
	trap '$(MAKE) -s down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) up; \
	$(call wait_for_services); \
	if [ -n "$(GOTESTSUM)" ]; then \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL)  gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/property/property.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/property || { \
	    if [ "$(RETRY_ON_FAIL)" = "1" ]; then \
	      echo "Retrying property tests once..."; \
	      TEST_MANAGER_URL=$(TEST_MANAGER_URL)  gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/property/property-rerun.xml -- -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/property; \
	    else \
	      exit 1; \
	    fi; \
	  }; \
	else \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL)  go test -v -race -count=1 $(GO_TEST_LDFLAGS) ./tests/property; \
	fi

# Chaos tests
.PHONY: test-chaos
test-chaos:
	$(call print_title,Running chaos tests - requires Docker stack)
	$(call check_command,docker,"Install Docker from https://docs.docker.com/get-docker/")
	$(call check_env_files)
	@set -e; mkdir -p $(TEST_REPORTS_DIR)/chaos; \
	trap '$(MAKE) -s down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) up; \
	$(call wait_for_services); \
	if [ -n "$(GOTESTSUM)" ]; then \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/chaos/chaos.xml -- -v -race -timeout 30m -count=1 $(GO_TEST_LDFLAGS) ./tests/chaos || { \
	    if [ "$(RETRY_ON_FAIL)" = "1" ]; then \
	      echo "Retrying chaos tests once..."; \
	      TEST_MANAGER_URL=$(TEST_MANAGER_URL) gotestsum --format testname --junitfile $(TEST_REPORTS_DIR)/chaos/chaos-rerun.xml -- -v -race -timeout 30m -count=1 $(GO_TEST_LDFLAGS) ./tests/chaos; \
	    else \
	      exit 1; \
	    fi; \
	  }; \
	else \
	  TEST_MANAGER_URL=$(TEST_MANAGER_URL) go test -v -race -timeout 30m -count=1 $(GO_TEST_LDFLAGS) ./tests/chaos; \
	fi

# Run all test suites
.PHONY: test-all
test-all:
	$(call print_title,Running all tests)
	$(call print_title,Running unit tests)
	$(MAKE) test-unit
	$(call print_title,Running integration tests)
	$(MAKE) test-integration
	$(call print_title,Running chaos tests)
	$(MAKE) test-chaos
	$(call print_title,Running e2e tests)
	$(MAKE) test-e2e
	$(call print_title,Running property tests)
	$(MAKE) test-property
	$(call print_title,Running fuzzy tests)
	$(MAKE) test-fuzzy