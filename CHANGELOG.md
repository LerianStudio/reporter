# Changelog

All notable changes to this project will be documented in this file.

## [v1.0.0-beta.18] - 2025-05-15

### 🐛 Bug Fixes
- Establish network for plugin fees on worker to ensure proper functionality and connectivity

## [v1.0.0-beta.12] - 2025-05-09

### 🐛 Bug Fixes
- Correct retry value conversion logic to ensure accurate retry attempts.
- Improve message reprocessing on queue to enhance reliability and performance.
- Adjust code to pass lint checks, ensuring code quality and consistency.

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes and improvements.

## [v1.0.0-beta.11] - 2025-05-08

### ✨ Features
- Add comments to the field mapping function for better understanding and maintainability.

### 🐛 Bug Fixes
- Adjust unit tests for improved reliability, ensuring tests run correctly.
- Update field mapping to support block structures and adjust report naming to include report ID for enhanced functionality.

### 🔧 Maintenance
- Set GitHub token value for CI/CD configuration to streamline deployment processes.
- Apply linting to improve code quality and maintain consistent coding standards.

## [v1.0.0-beta.9] - 2025-05-07

### 🐛 Bug Fixes
- Adjust Makefile for improved build process, enhancing build reliability.
- Adjust Docker Compose configuration for manager and infrastructure services to ensure correct service startup.

### 🔧 Maintenance
- Rename Go module to 'plugin-smart-templates' and update all imports. **Note: This is a breaking change and may affect existing integrations.**
- Update function comments for clarity, improving code readability and maintainability.

### 📚 Documentation
- Update Swagger documentation for API endpoints to reflect the latest changes and ensure accurate API usage.

### 🔧 Chore
- Update CHANGELOG with recent changes to maintain accurate project history.

## [v1.0.0-beta.8] - 2025-05-05

### 🔧 Maintenance
- Normalize GitHub Actions release workflow to ensure consistency across CI/CD processes.
- Configure new jobs in the GitHub Actions pipeline to enhance automation and streamline the integration process.
- Rename repository and update GitHub Actions job names for improved clarity and alignment with project naming conventions.
