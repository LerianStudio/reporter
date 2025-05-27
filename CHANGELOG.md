# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0-beta.6] - 2025-05-26

### 🐛 Bug Fixes
- Correct Makefile commands to ensure proper build process
=======
## [v1.1.0-beta.5] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG for recent changes
- Test build workflow configuration to ensure CI setup is validated
=======
## [v1.1.0-beta.4] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes

### 🔧 Continuous Integration
- Test build workflow for continuous integration improvements
=======
## [v1.1.0-beta.3] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes
- Test build workflow to ensure continuous integration functionality
=======
## [v1.1.0-beta.2] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG for recent changes to reflect the latest updates and improvements.
- Test build workflow for continuous integration improvements, enhancing the reliability and efficiency of the CI/CD pipeline.
=======
## [v1.1.0-beta.1] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes
- Adjust configuration names for improved clarity
- Update Helm configuration to align with new standards

### 🔧 Continuous Integration
- Test build workflow to ensure pipeline reliability
- Add build pipeline configuration to automate deployment processes
=======
## [v1.0.0-beta.20] - 2025-05-26

### ✨ Features
- Add 'contains' tag to validate substring presence, enhancing template validation capabilities.

### 🐛 Bug Fixes
- Correct code formatting issues by running lint, ensuring consistent code style.
- Implement regex for aggregation block validation, improving accuracy in data processing.
- Adjust regex for 'with' filter block to fix incorrect filtering behavior.
- Validate mapped fields in data structures to prevent runtime errors.
- Ensure existence of collections in MongoDB schemas, avoiding potential data retrieval issues.
- Adjust placeholders in loops with regex to fix template rendering errors.
=======

## [v1.0.0-beta.19] - 2025-05-23

### 🐛 Bug Fixes
- Update `.releaserc` file to use the correct semantic release plugin, ensuring proper release process configuration

### 🔧 Maintenance
- Consolidate CHANGELOG updates for improved clarity and consistency
=======
## [v1.0.1-feat.6] - 2025-05-16

### 🔧 Maintenance
- Implement test build workflow for continuous integration

## [v1.0.1-feat.5] - 2025-05-16

### 🔧 Maintenance
- Implement test build workflow in CI process to automate and ensure reliability of the build process.

## [v1.0.1-feat.3] - 2025-05-16

### 🔧 Maintenance
- Add test build workflow for continuous integration to enhance CI/CD pipeline efficiency.


## [v1.0.1-feat.2] - 2025-05-16

### 🔧 Maintenance
- Add test build workflow to the continuous integration pipeline

## [v1.0.1-feat.1] - 2025-05-16

### 🔧 Maintenance
- Test build workflow for continuous integration to ensure robust deployment processes

## [v1.0.0] - 2025-05-16

### 🔧 Maintenance
- Update Go module dependencies to the latest versions for improved security and performance.
- Refactor JSON annotations for error structs, enhancing code readability and maintainability.

## [v1.0.0-beta.18] - 2025-05-15

### 🐛 Bug Fixes
- Establish network for plugin fees on worker to ensure proper functionality and connectivity

## [v1.0.0-beta.17] - 2025-05-15

### 🐛 Bug Fixes
- Establish network connection for plugin fees to ensure correct functionality and prevent connectivity issues.

## [v1.0.0-beta.16] - 2025-05-15

### 🐛 Bug Fixes
- Resolve linting issues in the codebase to maintain code quality

### 🔧 Maintenance
- Add 'createdAt' filter to the templates endpoint list, enhancing the functionality and allowing more precise template management

## [v1.0.0-beta.15] - 2025-05-14

### 🐛 Bug Fixes
- Organize code by running lint to improve code quality and maintainability.

### 🔧 Maintenance
- Standardize environment variable names for plugin consistency, enhancing code readability and maintainability.

## [v1.0.0-beta.14] - 2025-05-13

### 🐛 Bug Fixes
- Adjust filter validation for report generation to ensure correct functionality

### 📚 Documentation
- Generate new Swagger documentation to improve API clarity and usability

## [v1.0.0-beta.13] - 2025-05-13

### 🐛 Bug Fixes
- Correct the type casting of the `x-retry-count` variable to ensure proper retry logic functionality.
- Limit message retry attempts to three in the consumer to prevent excessive retries and reduce resource usage.

### 🔧 Maintenance
- Improve code structure and comments for report generation to enhance readability and maintainability.
- Update changelog documentation to reflect recent changes.

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
