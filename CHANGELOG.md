# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0-beta.6] - 2025-07-08

This release enhances the user experience by automating changelog generation and improving system compatibility and reliability, all while maintaining backward compatibility.

### ✨ Features  
- **Automated Changelog Generation**: We've introduced a new feature that automatically generates changelogs using GPT for each release. This ensures you receive detailed and accurate updates, keeping you informed about the latest changes effortlessly.

### 🐛 Bug Fixes
- **Improved File Handling**: Fixed an issue in the backend where files containing script tags could cause unexpected behavior. This enhancement ensures smoother file processing, leading to a more robust and reliable system.
- **Codebase Cleanup**: Removed outdated feature branches, reducing clutter and potential confusion. This makes it easier for developers to focus on active projects and streamline development workflows.

### 🔄 Changes
- **Updated RabbitMQ and Redis**: Aligned our implementations with the latest updates, enhancing compatibility and performance. This change boosts the reliability of our message queuing and caching systems, ensuring smoother operations.
- **Revised License Implementation**: Updated to comply with the latest standards, providing clarity on usage rights and ensuring adherence to current licensing requirements.

### 📚 Documentation
- **Documentation Updates**: We've updated our documentation to reflect changes in RabbitMQ, Redis, and licensing implementations. This ensures that you have access to the most current and accurate information, supporting effective use and integration of our software.

### 🔧 Maintenance
- **Testing Enhancements**: Updated tests to verify the functionality of the latest changes, maintaining high standards of code quality and reliability.

This release is designed to enhance your experience by simplifying the update process, improving system performance, and ensuring compliance with the latest standards.

## [v0.0.0-beta.25] - 2025-05-27

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes, ensuring documentation is current and accurate.

### 🔧 Continuous Integration
- Test build workflow to ensure continuous integration setup is functioning correctly, maintaining pipeline reliability.

=======
## [v0.0.0-beta.24] - 2025-05-26

### 🐛 Bug Fixes
- Correct Makefile commands to ensure proper build process
=======
## [v0.0.0-beta.23] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG for recent changes
- Test build workflow configuration to ensure CI setup is validated
=======

## [v0.0.0-beta.22] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes

### 🔧 Continuous Integration
- Test build workflow for continuous integration improvements
=======

## [v0.0.0-beta.21] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes
- Test build workflow to ensure continuous integration functionality
=======

## [v0.0.0-beta.20] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG for recent changes to reflect the latest updates and improvements.
- Test build workflow for continuous integration improvements, enhancing the reliability and efficiency of the CI/CD pipeline.
=======

## [v0.0.0-beta.19] - 2025-05-26

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes
- Adjust configuration names for improved clarity
- Update Helm configuration to align with new standards

### 🔧 Continuous Integration
- Test build workflow to ensure pipeline reliability
- Add build pipeline configuration to automate deployment processes
=======

## [v0.0.0-beta.18] - 2025-05-26

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

## [v0.0.0-beta.17] - 2025-05-23

### 🐛 Bug Fixes
- Update `.releaserc` file to use the correct semantic release plugin, ensuring proper release process configuration

### 🔧 Maintenance
- Consolidate CHANGELOG updates for improved clarity and consistency
=======
## [v0.0.0-beta.16] - 2025-05-16

### 🔧 Maintenance
- Implement test build workflow for continuous integration

## [v0.0.0-beta.15] - 2025-05-16

### 🔧 Maintenance
- Implement test build workflow in CI process to automate and ensure reliability of the build process.

## [v0.0.0-beta.14] - 2025-05-16

### 🔧 Maintenance
- Add test build workflow for continuous integration to enhance CI/CD pipeline efficiency.


## [v0.0.0-beta.13] - 2025-05-16

### 🔧 Maintenance
- Add test build workflow to the continuous integration pipeline

## [v0.0.0-beta.12] - 2025-05-16

### 🔧 Maintenance
- Test build workflow for continuous integration to ensure robust deployment processes

## [v0.0.0-beta.11] - 2025-05-16

### 🔧 Maintenance
- Update Go module dependencies to the latest versions for improved security and performance.
- Refactor JSON annotations for error structs, enhancing code readability and maintainability.

## [v0.0.0-beta.10] - 2025-05-15

### 🐛 Bug Fixes
- Establish network for plugin fees on worker to ensure proper functionality and connectivity

## [v0.0.0-beta.9] - 2025-05-15

### 🐛 Bug Fixes
- Establish network connection for plugin fees to ensure correct functionality and prevent connectivity issues.

## [v0.0.0-beta.8] - 2025-05-15

### 🐛 Bug Fixes
- Resolve linting issues in the codebase to maintain code quality

### 🔧 Maintenance
- Add 'createdAt' filter to the templates endpoint list, enhancing the functionality and allowing more precise template management

## [v0.0.0-beta.7] - 2025-05-14

### 🐛 Bug Fixes
- Organize code by running lint to improve code quality and maintainability.

### 🔧 Maintenance
- Standardize environment variable names for plugin consistency, enhancing code readability and maintainability.

## [v0.0.0-beta.6] - 2025-05-13

### 🐛 Bug Fixes
- Adjust filter validation for report generation to ensure correct functionality

### 📚 Documentation
- Generate new Swagger documentation to improve API clarity and usability

## [v0.0.0-beta.5] - 2025-05-13

### 🐛 Bug Fixes
- Correct the type casting of the `x-retry-count` variable to ensure proper retry logic functionality.
- Limit message retry attempts to three in the consumer to prevent excessive retries and reduce resource usage.

### 🔧 Maintenance
- Improve code structure and comments for report generation to enhance readability and maintainability.
- Update changelog documentation to reflect recent changes.

## [v0.0.0-beta.4] - 2025-05-09

### 🐛 Bug Fixes
- Correct retry value conversion logic to ensure accurate retry attempts.
- Improve message reprocessing on queue to enhance reliability and performance.
- Adjust code to pass lint checks, ensuring code quality and consistency.

### 🔧 Maintenance
- Update CHANGELOG to reflect recent changes and improvements.

## [v0.0.0-beta.3] - 2025-05-08

### ✨ Features
- Add comments to the field mapping function for better understanding and maintainability.

### 🐛 Bug Fixes
- Adjust unit tests for improved reliability, ensuring tests run correctly.
- Update field mapping to support block structures and adjust report naming to include report ID for enhanced functionality.

### 🔧 Maintenance
- Set GitHub token value for CI/CD configuration to streamline deployment processes.
- Apply linting to improve code quality and maintain consistent coding standards.

## [v0.0.0-beta.2] - 2025-05-07

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

## [v0.0.0-beta.1] - 2025-05-05

### 🔧 Maintenance
- Normalize GitHub Actions release workflow to ensure consistency across CI/CD processes.
- Configure new jobs in the GitHub Actions pipeline to enhance automation and streamline the integration process.
- Rename repository and update GitHub Actions job names for improved clarity and alignment with project naming conventions.
