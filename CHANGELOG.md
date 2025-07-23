# Changelog

All notable changes to this project will be documented in this file.

## [v2.0.0-beta.4] - 2025-07-23

This release focuses on enhancing the reliability and accuracy of template processing, ensuring a smoother and more consistent user experience.

### 🐛 Bug Fixes
- **Improved Template Mapping**: Resolved issues with template mapping fields to enhance data accuracy and prevent errors during execution. Users will experience more reliable template functionality, reducing the likelihood of incorrect data representation.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to provide users and developers with the latest information on project changes and improvements, supporting better project tracking and understanding.

This changelog highlights the critical bug fixes that improve the reliability of the plugin-smart-templates, ensuring users have a clear understanding of how these changes enhance their experience. The maintenance update ensures that users are well-informed about the project's progress.

## [v2.0.0-beta.3] - 2025-07-23

This release focuses on enhancing the development workflow and updating documentation to provide a more streamlined and efficient experience for developers.

### 🔧 Maintenance
- **Build and Configuration Enhancements**: We've optimized the build pipeline for the frontend, which improves the development workflow. These changes streamline the build process, making it more efficient and easier for developers to manage dependencies and configurations. While these updates do not directly impact end-users, they lay the groundwork for smoother future enhancements.

### 📚 Documentation
- **Changelog Updates**: The changelog has been meticulously updated to reflect all recent changes and improvements. This ensures transparency and aids both developers and users in understanding the project's evolution, providing a reliable reference for future developments.

This release is part of our ongoing commitment to maintain a robust and up-to-date project infrastructure, ensuring a solid foundation for future enhancements and features.

## [v2.0.0-beta.2] - 2025-07-23

This release focuses on enhancing the security and reliability of the plugin-smart-templates, ensuring a safer and more efficient user experience.

### 🔄 Changes
- **Security Prioritization**: We've adjusted the severity level of npm audit analysis to prioritize critical security issues. This means that the most impactful vulnerabilities are addressed first, enhancing the overall security of your application.

### 🔧 Maintenance
- **Security Update**: A vulnerability related to brace-expansion has been resolved through an automatic fix. This update fortifies the application's defenses against potential exploits, contributing to a more secure environment.
- **Documentation Update**: The CHANGELOG has been updated to reflect recent improvements and changes, keeping users informed and documentation current.

This update ensures that your application remains secure and well-documented, with a focus on addressing critical vulnerabilities efficiently.

## [v2.0.0-beta.1] - 2025-07-23

This major release of plugin-smart-templates introduces enhanced security features, improved documentation organization, and critical bug fixes, ensuring a more robust and user-friendly experience.

### ⚠️ Breaking Changes
- **Dependency Updates**: This release includes updates to several dependencies that may not be backward compatible. Users must review and update their configurations accordingly. Please ensure your build scripts and dependency management settings are adjusted to accommodate the new APIs and behaviors. [Migration Guide](#)

### ✨ Features
- **Secure Pipeline Configurations**: We've introduced the ability to use private libraries in pipeline configurations, significantly enhancing security and customization options for deployment processes. This feature allows seamless integration of proprietary components, ensuring a secure build pipeline.
- **GitHub Token Support**: Docker configurations now support GitHub tokens, simplifying authentication and access management for GitHub-hosted resources. This enhancement improves security and integration ease.

### 🐛 Bug Fixes
- **Consistent Data Presentation**: Resolved issues with inconsistent data formatting across modules, ensuring reliable and uniform data handling, which enhances user experience.
- **Configuration Parsing**: Fixed errors in datasource name extraction and property names, preventing runtime issues and ensuring accurate configuration parsing.

### 📚 Documentation
- **Template Examples**: Template examples have been moved to the documentation directory, making them easier to find and use. This reorganization aids developers in quickly accessing resources for efficient development.
- **OpenAPI Documentation**: Updated OpenAPI documentation to reflect the latest changes, providing clear and comprehensive guidance for users.

### 🔧 Maintenance
- **Dependency Updates**: Regular updates to project dependencies ensure compatibility and leverage improvements in underlying libraries, maintaining security and performance standards.
- **Repository Management**: Updated `.gitignore` to exclude unnecessary files, streamlining repository management and reducing clutter.

This release is designed to enhance the overall functionality, security, and usability of the plugin-smart-templates project, focusing on delivering a reliable and seamless user experience.

This changelog provides a clear, user-focused overview of the changes in version 2.0.0, highlighting the benefits and necessary actions for users while maintaining a professional and accessible tone.

## [v1.1.0-beta.9] - 2025-07-18

This release brings significant structural improvements and integration efforts, enhancing developer productivity and ensuring a smoother user experience.

### ✨ Features  
- **Frontend & Backend Integration**: Experience a more cohesive and seamless user interface with the integration of front-end components into the backend architecture. This change simplifies interactions and ensures a harmonious operation across the application.

### 🔄 Changes
- **Environment Configuration**: Setting up and deploying the application is now easier and less error-prone, thanks to streamlined environment variable adjustments. This improvement reduces setup time and minimizes configuration errors.
- **Folder Structure Optimization**: Enjoy a more organized and navigable codebase with our comprehensive folder structure reorganization. This change enhances code maintainability, making it easier for developers to locate and manage files efficiently.

### 📚 Documentation
- **Documentation Enhancements**: Updated documentation provides clearer guidance on new workflows and environment configurations, helping developers understand and utilize the latest changes effectively.

### 🔧 Maintenance
- **Dependency Updates**: We've updated libraries and dependencies, particularly in the frontend components, ensuring compatibility with the latest security patches and performance improvements for a more stable and secure application.

This release focuses on setting a solid foundation for future feature development and enhancements, with changes designed to improve both developer productivity and user experience.

## [v1.1.0-beta.8] - 2025-07-15

This release introduces a major overhaul of the release process, enhancing consistency and efficiency. Users will benefit from a more predictable and streamlined deployment experience.

### ⚠️ Breaking Changes
- **Standardized Release Flow**: The release process has been restructured to include a new flow with a built-in hotfix mechanism. This change requires users to update their deployment scripts or workflows to align with the new structure. Please review the updated documentation for detailed migration guidance.

### ✨ Features  
- **Enhanced Release Process**: The introduction of a standardized release flow ensures a more reliable and efficient deployment process. This new structure supports quick hotfixes, enabling faster resolution of issues and minimizing downtime. Users can now enjoy a smoother update experience with improved release management.

### 📚 Documentation
- **Changelog Update**: The changelog has been thoroughly updated to reflect the latest changes and improvements, providing users with clear and comprehensive information about the plugin's evolution.

### 🔧 Maintenance
- **Improved Documentation**: We've enhanced the documentation to better support users in understanding and adapting to the new release flow. This includes detailed instructions and examples to assist in the transition.

In this release, the focus has been on refining the release process to ensure smoother and more predictable deployments. Users are encouraged to familiarize themselves with the new release flow and update their processes accordingly.

## [v1.1.0-beta.7] - 2025-07-14

This release enhances the security and configurability of the plugin-smart-templates, introducing SSL support and improved environment variable management, along with crucial bug fixes and documentation updates.

### ✨ Features  
- **SSL Connection Support**: Enhance your data security with the new SSL connection capabilities. This feature ensures encrypted communication, protecting your data during transmission. Check the updated documentation for setup guidance.

### 🐛 Bug Fixes
- **Smart Templates Compatibility**: Resolved compatibility issues by updating the smart-templates version. This fix ensures you have access to the latest features and stability improvements.
- **Code Quality Improvements**: Adjustments to meet linting standards enhance code readability and maintainability, reducing potential errors.

### 🔄 Changes
- **Environment Variable Management**: The `.env.example` file now supports more flexible configuration options, simplifying the setup of environment-specific variables to better suit your needs.

### 📚 Documentation
- **Updated Setup Guides**: Documentation has been refreshed to reflect the latest changes, including SSL setup instructions and updated smart-templates version information.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been meticulously updated to provide a clear record of recent changes, ensuring you are informed of all enhancements and fixes.

These updates collectively improve the security, configurability, and reliability of the plugin-smart-templates, offering you a more robust and user-friendly experience.

This changelog highlights the most significant changes in version 1.2.0, focusing on user benefits and the impact of each update. It is structured to provide clear, concise information that helps users understand and leverage the new features and improvements effectively.

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
