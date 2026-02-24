# Reporter Changelog

## [1.1.0](https://github.com/LerianStudio/reporter/releases/tag/v1.1.0)

- **Features:**
  - Added Redis SetNX idempotency check for report creation.
  - Introduced readiness endpoint, handler constructors, config validation, and domain model constructors.

- **Fixes:**
  - Resolved critical bugs in report workflow, XSS validation, and Redis panic.
  - Improved error handling, nil guards, and log redaction.
  - Corrected WaitGroup Done placement in goroutine cleanup handlers.
  - Removed the required validation for env object storage endpoint.
  - Fixed health check on container and PDF report generation.

- **Improvements:**
  - Improved observability, type safety, and production hardening.
  - Enhanced test quality with build tags, env isolation, and chaos guards.
  - Improved code quality, observability, and split generate-report.go.
  - Standardized Docker Compose, Makefiles, response wrappers, and config.
  - Centralized os.Getenv calls and added thread-safe datasource map.

Contributors: @arthur.ribeiro, @arthurkz, @bruno.souza, @brunoblsouza, @dependabot[bot], @ferr3ira.gabriel, @jefferson.comff

[Compare changes](https://github.com/LerianStudio/reporter/compare/v1.0.0...v1.1.0)

