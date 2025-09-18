# Changelog

All notable changes to this project will be documented in this file.

## [v2.0.1-beta.2] - 2025-09-18

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.1-beta.1...v2.0.1-beta.2)
Contributors: arthurkz, lerian-studio

### 🔧 Maintenance
- **Updated Build System:** The Go version and linting tools have been upgraded in our GitHub workflow. This ensures that our code is compatible with the latest Go features and adheres to updated linting standards, resulting in a more robust and maintainable codebase.
- **Library and Dependency Updates:** We've updated the `lib-commons` library and Go language version to enhance security, performance, and compatibility. This proactive maintenance ensures the project remains stable and benefits from the latest improvements in these libraries.
- **Changelog Refresh:** The changelog has been updated to provide a clear and concise history of recent changes, helping users track the project's evolution and understand the context of updates.


## [v2.0.1-beta.1] - 2025-09-18

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0...v2.0.1-beta.1)
Contributors: Gabriel Ferreira, lerian-studio

### 🔄 Changes
- **Frontend & Deployment**: The frontend Dockerfile has been updated to streamline the build process, aligning with current best practices. This change reduces potential setup issues, making deployment more efficient and less error-prone for developers.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been revised to accurately reflect recent updates, ensuring all stakeholders have access to the latest project developments and modifications.

### 🔧 Maintenance
- **Branch Management**: The develop branch was recreated to resolve inconsistencies and align with project standards. This update affects multiple components, including auth, backend, build, config, database, dependencies, docs, frontend, and test, ensuring a clean and organized development workflow.


## [v2.1.0-beta.13] - 2025-09-16

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.12...v2.1.0-beta.13)
Contributors: Caio Alexandre Troti Caetano, Gabriel Castro, lerian-studio

### ⚠️ Breaking Changes
- **Codebase Cleanup**: This release removes deprecated functions and modules across various components, which may impact existing integrations. Users should review their implementations to ensure compatibility with the updated codebase. Please refer to the migration guide for detailed steps on updating your integrations.

### ✨ Features
- **Schema Management with Zod v4**: We've upgraded to Zod v4, enhancing validation processes across multiple components. This change improves support for multiple languages and locales, making your application more accessible to a global audience.
- **New Library Integrations**: The addition of `sindarian-ui` and `sindarian-server` libraries enhances both frontend and backend capabilities, offering a more robust UI framework and improved server-side performance.

### 🐛 Bug Fixes
- **Translation Fixes**: We've resolved issues with missing translations in the frontend, ensuring all interface elements are properly localized for users in different regions. This fix enhances accessibility and user satisfaction.

### ⚡ Performance
- **Security Enhancements**: Implemented `noopener noreferrer` attributes for external links, preventing potential phishing attacks and ensuring safer navigation for users.

### 🔄 Changes
- **Enhanced Template System**: The new calendar filter feature in the template system allows for more precise date-based filtering, improving user experience by making it easier to find relevant templates quickly.

### 📚 Documentation
- **Changelog Update**: The changelog has been revised to accurately reflect recent updates and improvements, providing users with a clear history of changes and enhancements.

### 🔧 Maintenance
- **Dependency Management**: Updated `package.json` to include new libraries and dependencies, ensuring the project remains up-to-date with the latest tools and frameworks.


## [v2.1.0-beta.12] - 2025-09-15

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.11...v2.1.0-beta.12)
Contributors: arthurkz, lerian-studio

### 🐛 Bug Fixes
- **Backend/Database**: Improved validation of mapped fields in MongoDB schemas. This fix enhances data integrity by ensuring schema fields are correctly validated, reducing runtime errors and boosting system reliability. Users can now expect more consistent and accurate data representation across the application.

### 🔧 Maintenance
- **Release Management**: Updated the CHANGELOG to include recent changes and improvements. This update ensures users have access to the latest information about software updates, fostering better transparency and communication regarding the project's development progress.


## [v2.1.0-beta.11] - 2025-09-12

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.10...v2.1.0-beta.11)
Contributors: LF Barrile, arthurkz, lerian-studio

### ✨ Features
- **Enhanced Security**: We've improved security by adding a Common Vulnerabilities and Exposures (CVE) entry to the Trivy ignore list. This helps manage known vulnerabilities more effectively, ensuring a safer environment for your projects.

### 🔧 Maintenance
- **Build System Optimization**: The image build configuration has been updated to streamline the development workflow across the frontend, backend, and build components. This change promotes a more efficient and consistent build process, making it easier for developers to maintain and deploy updates.
- **Updated Documentation**: The CHANGELOG has been revised to reflect recent changes and improvements. This ensures you have access to the latest information about updates, enhancing transparency and clarity in the project's development history.


## [v2.1.0-beta.10] - 2025-09-12

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.9...v2.1.0-beta.10)
Contributors: arthurkz, lerian-studio

### 🐛 Bug Fixes
- **Backend/Database**: Fixed an issue with the discovery of mapped fields in MongoDB collections. This improvement ensures accurate data mapping and prevents potential data retrieval errors, enhancing the stability of your database interactions.

### 🔧 Maintenance
- **Release Management**: Updated the CHANGELOG to provide users with the latest information on updates and fixes, ensuring transparency and ease of access to release details.
- **Database/Test**: Added unit tests for `datasource.mongodb.go`, increasing test coverage and ensuring the MongoDB data source functions correctly. This enhancement improves the reliability and maintainability of the database component.


## [v2.1.0-beta.9] - 2025-09-10

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.8...v2.1.0-beta.9)
Contributors: arthurkz, lerian-studio

### 🔧 Maintenance
- **Smart Templates Image Update**: We've rebuilt the smart templates image to incorporate the latest configurations and dependencies. This ensures that all components, including build, config, and frontend, are aligned with recent optimizations. Users will benefit from consistent performance and improved reliability.
- **Changelog Update**: The CHANGELOG has been refreshed to accurately reflect recent updates. This provides users and developers with a clear and comprehensive record of changes, aiding in better tracking of project progress.


## [v2.1.0-beta.8] - 2025-09-10

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.7...v2.1.0-beta.8)
Contributors: arthurkz, lerian-studio

### ✨ Features
- **PDF Support for Reports**: You can now export reports as PDFs, making it easier to share and document your findings. This feature simplifies the process of disseminating information in a widely-accepted format. Check the updated documentation for guidance on using this feature.

### 🐛 Bug Fixes
- **Secure Temporary Files**: We've improved the security of temporary file creation, reducing the risk of unauthorized access and data leaks. This fix ensures your data is handled safely and securely.
- **Reliable Build Process**: Resolved issues in the 'garble' build process that could cause failures. This fix streamlines the development workflow, allowing for smoother project builds.

### ⚡ Performance
- **Faster PDF Downloads**: Experience significantly reduced wait times when downloading large PDF documents. This improvement enhances user experience by speeding up access to your reports.

### 📚 Documentation
- **Updated Changelog**: The changelog has been revised to reflect recent updates, ensuring transparency and keeping you informed about the latest changes and enhancements.

### 🔧 Maintenance
- **Build System Enhancements**: Improved the build process for the 'garble' component, reducing potential errors and supporting faster development cycles.


## [v2.1.0-beta.7] - 2025-09-08

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.6...v2.1.0-beta.7)
Contributors: arthurkz, lerian-studio

### ⚡ Performance
- **Streamlined Data Processing**: We've optimized backend operations by removing unnecessary scale references and int64 conversions. This refactor results in improved system efficiency, allowing users to experience faster processing times and more reliable data handling across various environments.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to ensure users have access to the latest information on project progress and updates, maintaining transparency and a clear history of changes.


## [v2.1.0-beta.6] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.5...v2.1.0-beta.6)
Contributors: arthurkz, lerian-studio

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest changes and improvements. This provides you with clear and current information on project updates, helping you track progress and understand modifications with ease.

### 🔧 Maintenance
- **Config Updates**: We've generated new images for both the worker and manager components. This ensures that you benefit from the latest configurations and dependencies, enhancing the stability and performance of your system.


## [v2.1.0-beta.5] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.4...v2.1.0-beta.5)
Contributors: arthurkz, lerian-studio

### 🐛 Bug Fixes
- **Improved Data Consistency**: Resolved an issue with MongoDB field mapping, ensuring all fields are correctly retrieved. This fix enhances data reliability, allowing users to access complete and accurate information without interruptions.
- **Code Quality Enhancements**: Adjusted the database component according to lint recommendations. These improvements bolster code maintainability and stability, reducing potential errors and enhancing developer efficiency.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to accurately reflect recent changes and improvements, providing users and developers with a clear record of the project's progress and updates.


## [v2.1.0-beta.4] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.3...v2.1.0-beta.4)
Contributors: arthurkz, lerian-studio

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to reflect the latest system changes. This ensures transparency and provides users and developers with the most current information about updates and modifications, supporting better understanding and easier tracking of software evolution.

### 🔧 Maintenance
- **Streamlined Backend Development**: We've automated the generation of backend component images. This change reduces setup times and enhances reliability, allowing developers to focus more on innovation and less on configuration. Expect smoother deployment processes and consistent performance across environments.


## [v2.1.0-beta.3] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.2...v2.1.0-beta.3)
Contributors: LF Barrile, lerian-studio

### 🐛 Bug Fixes
- **Build/Frontend**: Resolved an issue with the GitOps Firmino update process, leading to smoother integration and deployment workflows. This fix enhances the reliability of the build and deployment pipeline, reducing potential downtime and errors during updates.

### 🔧 Maintenance
- **Release Management**: Updated the CHANGELOG to reflect recent changes and improvements. This ensures that users have access to clear and accurate information about the latest software updates and fixes.


## [v2.1.0-beta.2] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.1.0-beta.1...v2.1.0-beta.2)
Contributors: LF Barrile, lerian-studio

### 📚 Documentation
- **Golang Image Update**: The documentation now includes an updated Golang image to ensure compatibility with the Garble tool. This change is crucial for developers who use Garble for code obfuscation, enhancing security and privacy in their projects.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been revised to accurately reflect recent updates and improvements. This ensures all stakeholders have access to the latest project information, aiding in effective project management and communication.


## [v2.1.0-beta.1] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.75...v2.1.0-beta.1)
Contributors: Clara Tersi, arthurkz, lerian-studio

### ✨ Features
- **Advanced Filtering for Data Queries**: Users can now apply complex filtering logic to their data queries, enabling more precise and efficient data retrieval. This enhancement significantly boosts the flexibility and analytical power of your data tools.

### 🐛 Bug Fixes
- **Code Quality Improvements**: Resolved various code discrepancies and linting issues, enhancing overall code quality and maintainability. This ensures a more stable and reliable experience for users.

### 📚 Documentation
- **Expanded Filter System Documentation**: Added comprehensive examples and implementation plans for the new advanced filter system, including practical templates for financial reports. This helps users quickly understand and apply the new features to their specific needs.

### 🔧 Maintenance
- **Configuration Updates**: The `.env.example` files have been updated to reflect changes in version v2.1.0, ensuring all configurations are current and accurate for deployment.
- **Development Environment Streamlining**: Updated `.gitignore` to exclude unnecessary files, reducing clutter and streamlining the development process.


## [v2.0.0-beta.75] - 2025-09-04

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.74...v2.0.0-beta.75)
Contributors: LF Barrile, lerian-studio

### ✨ Features
- **Streamlined Development Process**: We've introduced new Gitflow steps to our build and frontend components. This enhancement provides a more structured approach to managing feature development, bug fixes, and releases. Users will experience a more organized workflow, reducing the potential for errors and improving overall efficiency.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to reflect the latest changes and improvements, ensuring that all documentation is up-to-date. This helps users stay informed about new features and enhancements.


## [v2.0.0-beta.74] - 2025-08-29

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.73...v2.0.0-beta.74)
Contributors: arthurkz, lerian-studio

### 🔧 Maintenance
- **Build Process Update**: Improved the build process to ensure the creation of a new image. This enhancement is vital for maintaining consistency and reliability in the build pipeline, allowing developers to integrate and deploy the latest updates seamlessly.
- **Changelog Documentation**: Updated the changelog to accurately reflect recent changes. This ensures that all stakeholders have access to the latest information, improving transparency and communication within the project.


## [v2.0.0-beta.73] - 2025-08-29

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.72...v2.0.0-beta.73)
Contributors: arthurkz, lerian-studio

### 🔧 Maintenance
- **Streamlined Build Process**: A new configuration file has been introduced to selectively skip unnecessary steps in the build pipeline. This change enhances the efficiency of development cycles by reducing redundant operations, leading to a cleaner and more streamlined workflow.
- **Updated Documentation**: The CHANGELOG has been updated to accurately reflect recent modifications. This ensures that all changes are well-documented, providing users and developers with a clear understanding of the project's evolution and context for the latest updates.


## [v2.0.0-beta.72] - 2025-08-28

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.71...v2.0.0-beta.72)
Contributors: arthurkz, lerian-studio

### ⚡ Performance
- **Backend Pipeline Update**: The backend pipeline now generates a new image incorporating the latest code changes and dependency updates. This enhancement ensures improved stability and performance, providing a smoother and more reliable user experience. Regular updates also maintain compatibility with current software standards and security practices.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to accurately reflect recent improvements and changes. This ensures users can easily track software evolution, understand new features, and stay informed about any enhancements. Keeping documentation current is crucial for effective version tracking and user transparency.

### 🔧 Maintenance
- **Backend and Dependencies**: Regular updates to the backend and dependencies have been performed to ensure the application remains stable and secure. These updates help maintain compatibility with the latest software standards, providing indirect benefits to users by enhancing overall application reliability.


## [v2.0.0-beta.71] - 2025-08-28

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.70...v2.0.0-beta.71)
Contributors: arthurkz, lerian-studio

### ✨ Features
- **Arithmetic Operations in Templates**: You can now perform calculations directly within your templates. This enhancement streamlines workflows by reducing the need for external processing, allowing for more dynamic and flexible template expressions. [See updated documentation for examples]

### 🐛 Bug Fixes
- **Testing Code Cleanup**: Removed unnecessary comments from test files, making the codebase cleaner and easier to navigate. This helps developers focus on the core logic and functionality without distractions.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to reflect the latest changes, ensuring you have access to the most current information about updates and improvements.

### 🔧 Maintenance
- **Code Quality Enhancements**: General improvements have been made to the codebase, contributing to the stability and reliability of the software. While these changes are not directly visible, they ensure a robust development environment.


## [v2.0.0-beta.70] - 2025-08-22

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.69...v2.0.0-beta.70)
Contributors: arthurkz, lerian-studio

### ✨ Features
- **CRM Database Connection**: Seamlessly integrate CRM data into your reports and templates. This feature empowers users to make more informed, data-driven decisions by incorporating comprehensive CRM insights directly into their workflow.

### 🐛 Bug Fixes
- **Data Encryption Reliability**: Resolved issues with data encryption errors, ensuring robust protection of sensitive information and enhancing overall system reliability.
- **Telemetry Accuracy**: Fixed errors in OpenTelemetry function implementations, ensuring consistent and accurate data collection for better system monitoring and diagnostics.

### ⚡ Performance
- **Enhanced Monitoring**: Optimized OpenTelemetry functions to improve system monitoring and observability, enabling quicker issue resolution and system optimization.

### 📚 Documentation
- **Updated Guidance**: Revised documentation to include new features and improvements, providing clear instructions on leveraging the latest capabilities for maximum benefit.

### 🔧 Maintenance
- **Changelog Update**: Ensured the changelog reflects all recent updates, maintaining transparency and keeping users informed about the latest enhancements.


## [v2.0.0-beta.69] - 2025-08-14

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.68...v2.0.0-beta.69)
Contributors: LF Barrile, lerian-studio

### ✨ Features
- **Enhanced GitOps Documentation**: We've introduced comprehensive validation for the Firmino GitOps flow. This update ensures a more accurate and reliable deployment process, making it easier for users to manage their infrastructure as code. The documentation now includes detailed guidelines and best practices, simplifying the adoption of GitOps principles and reducing potential errors.

### 📚 Documentation
- **Improved Deployment Guides**: The documentation now provides clearer, more detailed instructions for implementing the Firmino GitOps flow. This helps users streamline their deployment processes and adopt best practices with confidence.

### 🔧 Maintenance
- **Updated Changelog**: The changelog now accurately reflects recent updates and improvements, ensuring users have access to the latest information about new features and fixes. This transparency aids in smoother transitions and adoption of updates.


## [v2.0.0-beta.68] - 2025-08-14

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.67...v2.0.0-beta.68)
Contributors: LF Barrile, lerian-studio

### ✨ Features
- **Automated Configuration Updates with Firmino Integration**: Experience a streamlined process for keeping configurations up-to-date across your build, config, and frontend components. This automation reduces manual intervention and potential errors, providing a more seamless and reliable configuration management process.

### 🔄 Changes
- **Enhanced GitOps Integration**: We have improved the configuration update mechanism by incorporating Firmino GitOps practices. This ensures that configuration changes are automatically propagated and managed through GitOps workflows, enhancing consistency and reliability across deployments. Users will benefit from improved deployment efficiency and reduced configuration drift.

### 📚 Documentation
- **Changelog Update**: We've updated the changelog to reflect recent changes and improvements. This ensures that users and developers have access to the latest information regarding updates and enhancements, facilitating better understanding and tracking of project progress.

### 🔧 Maintenance
- **General Code Maintenance**: Minor code adjustments and optimizations have been made to improve the overall stability and performance of the plugin.


## [v2.0.0-beta.67] - 2025-08-13

[Compare changes](https://github.com/LerianStudio/plugin-smart-templates/compare/v2.0.0-beta.66...v2.0.0-beta.67)
Contributors: arthurkz, lerian-studio

### 🐛 Bug Fixes
- **Improved Error Validation**: Users now receive accurate feedback when template errors occur, enhancing the reliability of the template management system.
- **Clear 'Not Found' Messages**: When requested resources are unavailable, users will see a clear 'not found' message, improving understanding and transparency.
- **Persisted Report Filters**: Report filters are now saved across sessions, ensuring a seamless user experience without needing to reapply settings.
- **Invalid Filter Handling**: Invalid filter inputs are now met with a 'bad request' response, maintaining input integrity and preventing errors.
- **Enhanced Deletion Process**: Attempting to delete non-existent templates now results in a 'not found' error, providing robust feedback and improving system reliability.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to reflect recent changes and improvements, ensuring users are informed of the latest updates.


## [v2.0.0-beta.66] - 2025-08-08

This release enhances the build process for the frontend component, delivering faster build times and a more reliable deployment pipeline. Users will experience quicker updates and a smoother development workflow.

### ✨ Features  
- **Streamlined Build Configuration**: The frontend component now benefits from an optimized build process. This enhancement reduces build times and increases reliability, allowing users to receive updates more quickly and with fewer errors. This change makes managing deployments easier and more efficient.

### 📚 Documentation
- **Updated Changelog**: We've refreshed our documentation to ensure users have access to the latest information about updates and improvements. This helps maintain transparency and keeps all stakeholders informed about the project's progress.

### 🔧 Maintenance
- **Changelog Improvements**: The changelog has been updated to accurately reflect recent changes and enhancements, ensuring users are well-informed about the latest developments in the project.


This changelog is crafted to highlight the key benefits and improvements introduced in version 2.0.0 of the plugin-smart-templates project. It focuses on user impact, ensuring that users understand the value of the changes without delving into technical details.

## [v2.0.0-beta.65] - 2025-08-08

This release introduces a major update to the worker process API, enhancing scalability and performance. Users will experience improved task handling efficiency, but should review integration points due to breaking changes.

### ⚠️ Breaking Changes
- **API Update**: The worker process API has been overhauled for better scalability and performance. This change affects existing API endpoints and requires users to update their integration points. Please review the updated API documentation for guidance on transitioning to the new structure.

### ✨ Features  
- **Enhanced Worker Process Handling**: The new API significantly boosts task processing efficiency, particularly benefiting users handling high volumes of data. This improvement allows for faster and more reliable task execution.

### ⚡ Performance
- **Improved Scalability**: The updated API architecture enhances scalability, allowing for more efficient resource utilization and faster task processing. Users should notice a marked improvement in application responsiveness during high-load scenarios.

### 📚 Documentation
- **Updated API Guides**: Comprehensive updates to the documentation provide clear instructions on transitioning to the new API. This ensures users have the necessary information to adapt their systems smoothly and take full advantage of the new capabilities.

### 🔧 Maintenance
- **Build and Test Enhancements**: The build and testing frameworks have been updated to accommodate the new API changes, ensuring robust validation of all components. This results in a more stable and reliable development environment.

This changelog is structured to clearly communicate the key updates in version 3.0.0, focusing on the impact and benefits for users. Breaking changes are prominently highlighted with guidance for migration, while new features and performance improvements are described in terms of user value. Documentation updates and maintenance improvements are also noted to ensure users are well-informed about the changes.


## [v2.0.0] - 2025-08-08

This major release of plugin-smart-templates introduces enhanced security, improved performance, and significant updates to monitoring capabilities. Users should review breaking changes to ensure smooth integration.

### ⚠️ Breaking Changes
- **Ledger ID Removal:** The `ledgerId` has been removed from forms and reports, affecting data input and output. Users must update any scripts or integrations relying on `ledgerId` for continued functionality.
- **Release Flow Update:** A new standardized release flow is in place. Review the updated release documentation to adapt your workflows accordingly.
- **Conversion Script Removal:** OpenAPI to Postman conversion scripts have been discontinued. Users should seek alternative solutions for API documentation conversion.
- **API Documentation Update:** The `LedgerID` has been removed from Swagger definitions. Update your API clients to align with these changes.
- **Module Update:** The backend module has been updated to v2. Ensure your dependencies are compatible with the new versioning scheme.

### ✨ Features
- **Enhanced Monitoring:** OpenTelemetry tracing is now implemented for data sources and message consumers, providing improved observability across services.
- **SSL Support:** Added support for SSL connections, enhancing the security of data transmissions.
- **Caching for Data Retrieval:** Introduced caching for endpoint data retrieval by ID, significantly improving response times for repeated requests.
- **License Integration:** Implemented license management for worker and manager components, streamlining compliance and administration.

### 🐛 Bug Fixes
- **Navigation Fix:** Resolved an issue preventing access to account settings, restoring full navigation functionality.
- **Persistent Report Filters:** Fixed a bug where report filters were cleared when changing tabs, ensuring consistent filter application.
- **Authentication Reliability:** Corrected AuthClient initialization issues, enhancing user login experience.
- **Data Source Accuracy:** Addressed data source request issues by adding organization ID, ensuring accurate data retrieval.

### ⚡ Performance
- **Docker Build Optimization:** Optimized Dockerfile configurations, reducing build times and improving deployment efficiency.
- **Improved Report Generation:** Enhanced filtering and pagination in report generation, providing users with greater control and flexibility.

### 🔄 Changes
- **Console Layout Update:** The frontend console layout has been updated for improved user interface consistency and data management efficiency.

### 🗑️ Removed
- **Conversion Scripts:** OpenAPI to Postman conversion scripts have been removed. Users should transition to alternative documentation methods.

### 📚 Documentation
- **Documentation Overhaul:** Refactored documentation structure and removed unused XML fields, improving clarity and reducing clutter.

### 🔧 Maintenance
- **Dependency Updates:** Updated dependencies like `lib-commons`, `lib-auth`, and `fiber` to their latest stable versions, enhancing security and performance.
- **Telemetry Code Refactor:** Refactored telemetry code to remove redundant imports and enhance span attributes, improving code quality.

Users are encouraged to review these changes thoroughly to understand their impact and take necessary actions for a seamless transition.


## [v2.0.0-beta.64] - 2025-08-08

This release introduces a streamlined configuration process for release management, enhancing the efficiency and accuracy of future project updates.

### ✨ Features  
- **New Configuration for Release Management:** We've introduced a new configuration setup that simplifies managing release settings. This enhancement makes the release process more organized and predictable, reducing errors and improving overall project maintainability. Users can now configure their release settings more efficiently, ensuring smoother transitions between project versions.

### 📚 Documentation
- **Changelog Update:** The CHANGELOG has been updated to reflect the latest changes and improvements. This ensures that all users have access to current information about project updates, promoting better communication and transparency. Keeping the documentation up-to-date helps users track the project's evolution and understand the context of recent modifications.

### 🔧 Maintenance
- **Release Management Improvements:** Behind-the-scenes improvements have been made to the release management process, enhancing the overall workflow for developers. These changes indirectly benefit users by ensuring that future releases are handled more efficiently and accurately.

This changelog is structured to highlight the key feature introduced in this release, along with documentation updates and maintenance improvements. The focus is on the benefits and impact of the changes, ensuring users understand how these updates enhance their experience and project management capabilities.

## [v2.0.0-beta.63] - 2025-08-08

This release focuses on enhancing the user experience by ensuring consistent branding and improving documentation transparency.

### 🐛 Bug Fixes
- **Frontend/Backend**: Corrected the application title in the manifest to 'Smart Templates'. This update standardizes the branding across all interfaces, enhancing user recognition and trust in the application.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest changes, ensuring users and developers have access to the most current information about the software's evolution, aiding in transparency and communication.

### 🔧 Maintenance
- **Documentation**: Regular updates to the documentation ensure clarity and accessibility, supporting users in understanding the software's capabilities and changes.

This changelog provides a concise overview of the changes in version 2.0.0, focusing on user experience improvements and documentation updates. Each section highlights the value and impact of the changes, presented in a clear and accessible manner.

## [v2.0.0-beta.62] - 2025-08-07

This release brings a major update to the user interface, enhancing navigation and usability, alongside crucial bug fixes for improved reliability.

### ✨ Features  
- **Revamped Console Layout**: Experience a more intuitive and streamlined user interface with the redesigned console layout. This update enhances navigation efficiency, making it easier for users to interact with the console and access features quickly.

### 🐛 Bug Fixes
- **Accurate Report Filtering**: Resolved an issue affecting data filters during report generation, ensuring users receive consistent and expected results. This fix improves the reliability of report outputs.
- **Template Selector Improvement**: Fixed a bug in the template selector's filter functionality, allowing users to accurately filter and select templates without errors, thereby enhancing the selection process's usability.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to reflect recent changes and improvements, providing users with a current and informative overview of the latest updates.


## [v2.0.0-beta.61] - 2025-08-07

This release introduces a streamlined deployment process via DockerHub, enhancing the ease of building and deploying the plugin-smart-templates. Documentation updates ensure users can fully leverage these new capabilities.

### ✨ Features  
- **Simplified Deployment with DockerHub**: We've added DockerHub build configuration, allowing you to effortlessly build and deploy the plugin-smart-templates directly from DockerHub. This enhancement simplifies your integration and deployment workflow, making it easier to manage and deploy applications efficiently.

### 📚 Documentation
- **Updated Deployment Guides**: The documentation now includes detailed guidance on using the new DockerHub build configuration. This update ensures you have all the information needed to utilize the new feature, reducing the learning curve and enhancing usability.

### 🔧 Maintenance
- **Changelog Updates**: The CHANGELOG has been updated to reflect the latest changes, keeping you informed about the project's evolution and ensuring you have access to the most current information.


## [v2.0.0-beta.60] - 2025-08-07

This release focuses on enhancing the user interface and maintaining up-to-date project documentation, ensuring a smoother experience for developers and users alike.

### ✨ Features  
- **Streamlined Frontend Interface**: We've removed unnecessary comments from the frontend codebase. This enhancement improves code readability and maintainability, allowing developers to work more efficiently and potentially enhancing load times for end users.

### 📚 Documentation
- **Updated Changelog**: The changelog now accurately reflects recent updates and improvements. This ensures that both users and developers have a clear understanding of the project's progression, aiding in better communication and project management.

### 🔧 Maintenance
- **Codebase Clean-Up**: By refining the frontend code, we've laid the groundwork for faster feature development and easier bug fixes, indirectly benefiting all users by accelerating the delivery of future updates.

No breaking changes are included in this release, ensuring a seamless upgrade experience without requiring additional user actions.

## [v2.0.0-beta.59] - 2025-08-06

This release introduces significant enhancements to task distribution and maintenance, improving system performance and user convenience.

### ✨ Features
- **Optimized Task Distribution**: A new job distribution system now effectively splits tasks between backend and frontend components. This enhancement optimizes resource utilization, resulting in a more responsive and efficient user experience.

### ⚡ Performance
- **Codebase Streamlining**: Redundant code structures have been removed, leading to a cleaner codebase. This improvement enhances system performance and reduces the potential for future bugs, ensuring a smoother user experience.

### 🔄 Changes
- **Automatic Updates for Firmino**: Configuration changes now enable automatic updates for Firmino, ensuring users always have access to the latest features and security patches without manual intervention. This change enhances system reliability and minimizes downtime.

### 📚 Documentation
- **Changelog Updates**: The CHANGELOG has been updated to reflect recent changes and improvements, ensuring that users are informed about the latest updates and enhancements. This supports transparency and user awareness of ongoing project developments.

### 🔧 Maintenance
- **Documentation Maintenance**: Regular updates have been made to keep documentation current, supporting user understanding and engagement with the latest software features.

This changelog highlights the most significant changes and their benefits to users, ensuring clarity and accessibility for both technical and non-technical audiences.

## [v2.0.0-beta.58] - 2025-08-06

This release focuses on enhancing the user experience by fixing key bugs in the frontend, ensuring smoother navigation and more reliable application performance. We've also improved our testing framework to maintain high-quality standards.

### 🐛 Bug Fixes
- **Consistent Navigation Experience**: Resolved an issue where report filters were not being cleared when changing tabs. This fix ensures that users experience consistent behavior when navigating between different sections of the application, reducing confusion and improving usability.
- **Reliable URL Handling**: Corrected URL handling to address a navigation bug. This update enhances the reliability of the application by ensuring that users are directed to the correct pages without encountering errors.

### 🔧 Maintenance
- **Enhanced Testing Framework**: Updated unit tests to improve coverage and accuracy, ensuring that new and existing functionalities are thoroughly tested. This change enhances the robustness of the application, leading to fewer bugs and more reliable performance.
- **Changelog Update**: Updated the CHANGELOG to reflect recent changes and improvements, ensuring that users and developers have access to the latest information about updates and fixes, facilitating better understanding and tracking of project progress.

This changelog provides a clear and concise overview of the improvements made in version 2.0.0, focusing on the benefits to users and maintaining a professional tone suitable for public release notes.

## [v2.0.0-beta.57] - 2025-08-06

This release enhances the reporting capabilities of the plugin by increasing the template limit, allowing users to generate more comprehensive reports. Additionally, we have updated our documentation to ensure users have the latest information on improvements.

### ✨ Features  
- **Expanded Template Limit for Reports**: Users can now create more comprehensive and detailed reports without hitting previous constraints. This enhancement significantly improves the usability and flexibility of the reporting feature, catering to users with extensive data needs.

### 📚 Documentation
- **Changelog Updates**: The CHANGELOG has been updated to reflect recent changes and improvements, ensuring users and developers have access to the latest information about updates and fixes. This transparency aids in version tracking and understanding the evolution of the project.

### 🔧 Maintenance
- **Release Management**: Behind-the-scenes improvements have been made to streamline the release process, indirectly benefiting users by enhancing the overall stability and reliability of the plugin.


This changelog provides a clear and user-focused summary of the latest release, highlighting the key enhancements and maintenance updates. The structure is designed to be easily scannable, with a professional tone suitable for public release notes.

## [v2.0.0-beta.56] - 2025-08-06

This release of plugin-smart-templates introduces major enhancements to data management, reporting, and template systems, significantly improving user experience and efficiency.

### ✨ Features  
- **Enhanced Data Management**: Enjoy more robust and flexible data handling with improved integration across frontend and backend components. This streamlines data retrieval, boosting efficiency and user satisfaction.
- **Upgraded Reporting System**: Access more comprehensive and customizable report options, enabling better data analysis and informed decision-making.
- **Revamped Template System**: Experience more intuitive and versatile template creation and management, simplifying workflows and enhancing productivity.
- **Updated HTTP Library**: Benefit from improved performance, security, and compatibility with modern web standards, ensuring reliable and fast network communications.

### 🐛 Bug Fixes
- **Resolved Frontend Pagination Issues**: Navigate data accurately across multiple pages with fixed pagination, ensuring consistent and reliable data presentation.

### 🔄 Changes
- **Data Source Requests Improvement**: Added organization ID to data source requests, enhancing data security and relevance through precise filtering and access control.

### 🔧 Maintenance
- **Changelog Update**: The project changelog has been updated to reflect recent changes and improvements, ensuring users and developers have access to the latest information about the software's evolution.

These updates collectively enhance the plugin-smart-templates project by improving data handling, reporting, and template management, while also addressing critical bugs and maintaining the system's overall quality and performance.

## [v2.0.0-beta.55] - 2025-08-05

This release focuses on enhancing the reliability of user authentication processes, ensuring smoother login experiences and more stable configuration handling.

### 🐛 Bug Fixes
- **Improved Authentication Stability**: Resolved an issue with the initialization of the AuthClient. This fix enhances the reliability of user authentication, ensuring smoother login experiences and more stable configuration handling across the application.

### 🔧 Maintenance
- **Changelog Update**: Updated the CHANGELOG to reflect recent changes and improvements. This ensures all modifications are documented for transparency and future reference, aiding in better version tracking and user communication.

This changelog provides a concise overview of the key improvements in version 2.0.0, focusing on the benefits to users, such as enhanced authentication reliability. The maintenance update ensures users are informed about the documentation improvements, promoting transparency and effective communication.

## [v2.0.0-beta.54] - 2025-08-04

This release focuses on enhancing the accuracy and reliability of our documentation, ensuring users have access to the most current information and maintaining a clear record of project updates.

### 🐛 Bug Fixes
- **Documentation Path Detection**: Resolved an issue where changes to documentation paths were not being detected correctly. This fix ensures that any modifications are accurately tracked, improving the reliability of documentation updates and ensuring users always have access to the most up-to-date information.

### 📚 Documentation
- **Changelog Updates**: The CHANGELOG has been updated to reflect recent changes and improvements. This ensures users have a comprehensive and up-to-date record of all modifications, aiding in transparency and facilitating easier tracking of project evolution.

These updates are designed to improve the user experience by ensuring documentation is always current and changes are clearly communicated.

## [v2.0.0-beta.53] - 2025-08-04

This release enhances the efficiency and reliability of the template processing system, providing users with faster processing times and a more robust experience. 

### ⚡ Performance
- **Enhanced Condition Checking**: The logic used in both the build and frontend components has been optimized, resulting in improved efficiency and reliability. Users will experience faster processing times when working with smart templates, enhancing overall productivity.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest improvements, ensuring users and developers have access to current information about the software's evolution. This promotes transparency and ease of tracking project progress.

This changelog provides a clear, user-focused summary of the improvements made in version 2.0.0, highlighting the performance enhancements and maintenance updates that users will benefit from.

## [v2.0.0-beta.52] - 2025-08-04

This release focuses on enhancing data integrity and transparency, ensuring a more reliable and informed user experience.

### 🐛 Bug Fixes
- **Improved Data Management**: Resolved an issue in the backend where invalid data entries could be included in results. This fix enhances data accuracy and system stability, leading to fewer disruptions in your data-driven operations.

### 📚 Documentation
- **Changelog Update**: The project CHANGELOG has been updated to reflect recent changes and improvements. This ensures users are well-informed about the latest updates, fostering transparency and ease of access to historical modifications.

### 🔧 Maintenance
- **Backend Stability**: General maintenance improvements have been made to ensure ongoing system reliability and performance.


This changelog provides a concise overview of the most impactful changes in version 2.0.0, focusing on bug fixes and documentation updates that enhance user experience and system transparency.

## [v2.0.0-beta.51] - 2025-08-04

This release focuses on enhancing performance and reliability, providing users with a faster and smoother experience.

### ⚡ Performance
- **Improved Condition Checking**: We've optimized the logic used in both the build process and the frontend. Users will experience faster load times and more responsive interactions, enhancing overall satisfaction with the software.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to accurately reflect recent improvements, ensuring users have a clear and accessible history of changes.

This concise changelog communicates the key benefits of the release, focusing on how the improvements positively impact the user experience.

## [v2.0.0-beta.50] - 2025-08-04

This release focuses on improving system reliability by addressing environment variable handling issues and enhancing documentation transparency.

### 🐛 Bug Fixes
- **Consistent Environment Configuration**: Resolved issues with environment variable handling across configuration, documentation, and frontend components. This fix ensures that settings are consistently applied, reducing configuration errors and enhancing overall system reliability for users.

### 📚 Documentation
- **Updated Environment Variable Documentation**: Improved documentation to clearly outline environment variable usage and setup. This enhancement helps users configure their systems more accurately, minimizing potential errors.

### 🔧 Maintenance
- **Changelog Updates**: Revised the CHANGELOG to accurately reflect recent fixes and improvements. This update ensures users are well-informed about the latest changes, promoting transparency and ease of tracking project progress.

This changelog is crafted to emphasize the user benefits of the recent bug fixes and documentation improvements, ensuring users understand the enhancements and their impact on system reliability and configuration accuracy.

## [v2.0.0-beta.49] - 2025-08-01

This release focuses on enhancing the security and reliability of the authentication component through important dependency updates. While there are no new user-facing features, these updates ensure a more stable and secure experience for future releases.

### 🔧 Maintenance
- **Dependency Updates**: We have upgraded the `lib-auth` and `lib-commons` libraries to their latest stable versions. This update integrates the latest security patches and performance enhancements, providing a more secure and efficient authentication process.
  
- **Changelog Update**: The CHANGELOG file has been refreshed to include the latest changes, ensuring users have a clear view of the software's evolution and improvements.

These updates are part of our ongoing commitment to maintaining a robust and secure system, laying the groundwork for future enhancements and features.


## [v2.0.0-beta.48] - 2025-08-01

This release enhances user experience with improved data handling precision and a refreshed interface, while maintaining system integrity through regular updates and documentation improvements.

### ✨ Features  
- **Data Source Retrieval by ID**: You can now fetch data sources by their ID, allowing for precise and efficient data retrieval. This feature is particularly useful for users who need specific data points quickly, enhancing both speed and accuracy in data handling.

### 🔄 Changes
- **User Interface Update**: Enjoy a more intuitive and visually appealing layout that aligns with modern design standards. This update improves usability, making navigation and interaction more seamless.
- **Configuration Enhancements**: The configuration process is now more streamlined and user-friendly, thanks to updated versioning across multiple components. This ensures compatibility and leverages the latest features and security improvements from dependencies.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to reflect recent changes, providing clear and current documentation for users tracking project updates.

### 🔧 Maintenance
- **Routine Dependency Updates**: We've performed routine updates to dependencies to maintain system stability and security. This ensures all components are running on the latest, most secure versions, reducing potential vulnerabilities.

This release focuses on delivering a better user experience through enhanced functionality and a modernized interface, while also ensuring the system remains secure and up-to-date.

## [v2.0.0-beta.47] - 2025-08-01

This release focuses on enhancing the efficiency of our build process and improving project documentation to ensure clarity and transparency for all users.

### 🔧 Maintenance
- **Build Process Enhancement**: We've implemented a new workflow for building beta versions using firmino-lxc-runners. This upgrade enhances the build process by utilizing more efficient runners, which can lead to faster build times and improved resource management. This change is part of our ongoing efforts to streamline the development pipeline, ensuring more reliable and timely deployment of beta releases.
  
- **Documentation Update**: The CHANGELOG has been updated to accurately reflect recent changes and improvements. Keeping the changelog current ensures that users and developers have a clear understanding of the latest modifications and enhancements, facilitating better communication and transparency within the project.


This changelog is designed to provide users with a clear understanding of the latest updates, focusing on the benefits and impacts of the changes made in this release.

## [v2.0.0-beta.46] - 2025-07-31

This release of plugin-smart-templates introduces major enhancements to data integration and performance, improving user interaction and application efficiency.

### ✨ Features  
- **Seamless Data Integration**: Users can now select and interact with database inputs directly from the UI, providing a more intuitive and streamlined experience. This feature simplifies data handling and enhances productivity.

### 🐛 Bug Fixes
- **Reliable Template Rendering**: Resolved an issue affecting smart template rendering, ensuring users can create and utilize templates without encountering errors or unexpected behavior.

### ⚡ Performance
- **Optimized Backend Processing**: Refactored data sources to enhance backend speed and application responsiveness, resulting in a smoother user experience.

### 🔄 Changes
- **Updated Dependencies**: Library dependencies have been updated to improve code quality and maintainability, ensuring compatibility with the latest features and security patches.

### 🔧 Maintenance
- **Documentation Updates**: The changelog has been updated to reflect recent changes, providing clear and informative project history for users.

This changelog is designed to clearly communicate the value and impact of the v2.0.0 release, focusing on enhancements that improve user experience and application performance.

## [v2.0.0-beta.45] - 2025-07-31

This release introduces enhancements to the build process, improving security and efficiency for developers, along with updated documentation to ensure clarity and transparency.

### ⚡ Performance
- **Optimized Build Process**: The introduction of the 'garble' command in the Dockerfile enhances security by obfuscating code and reduces the final build size. This streamlines deployment and protects your application, making it easier and safer to manage.

### 📚 Documentation
- **Updated Changelog**: The changelog has been updated to accurately reflect recent changes and improvements, ensuring that all stakeholders have access to the latest information and can easily track project progress.

### 🔧 Maintenance
- **Documentation Maintenance**: Regular updates to documentation ensure that users have the most current and accurate information, supporting better decision-making and project transparency.


## [v2.0.0-beta.44] - 2025-07-31

This release focuses on a comprehensive refactor of the system's architecture, enhancing maintainability and paving the way for future feature expansions. Users will benefit from improved performance and easier integration of new functionalities.

### ✨ Improvements  
- **System Architecture Overhaul**: The backend, configuration, database, and testing modules have been refactored to improve code maintainability. This enhancement ensures a more robust system performance and simplifies the integration of future updates, providing a smoother user experience.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to accurately reflect recent changes and improvements, ensuring transparency and effective communication with users and developers. This update supports ongoing community engagement and awareness of system enhancements.

This changelog provides a clear, user-focused overview of the changes in version 2.0.0, emphasizing the benefits of the architectural refactor and maintaining transparency through updated documentation.

## [v2.0.0-beta.43] - 2025-07-31

This release of `plugin-smart-templates` introduces significant enhancements in performance monitoring and system maintainability, ensuring a more robust and efficient user experience.

### ✨ Features  
- **Enhanced Observability**: We've integrated OpenTelemetry tracing into our backend and database systems. This improvement allows you to monitor data flow and system performance more effectively, making it easier to diagnose issues and optimize your setup.

### 🔄 Changes
- **Consistent Telemetry Data**: We've standardized OpenTelemetry attributes and error handling across services. This change provides a more uniform and clear view of telemetry data, enhancing error diagnostics and trace readability.
- **Code Simplification**: Redundant imports and unnecessary code elements have been removed from the database, improving code readability and reducing maintenance efforts.

### 🔧 Maintenance
- **Dependency Updates**: We've streamlined the build process by removing and upgrading dependencies. Notably, `lib-commons` has been upgraded to v2.0.0, and `lib-auth` and `lib-license-go` have been updated to their latest beta versions. These updates ensure compatibility with the latest features and security enhancements.
- **Changelog Refresh**: The changelog has been updated to reflect recent changes, ensuring you have access to current and comprehensive documentation.

These updates collectively enhance the system's performance, maintainability, and observability, providing you with a more robust and efficient experience.

## [v2.0.0-beta.42] - 2025-07-30

This release focuses on enhancing the clarity and efficiency of the project repository, making it easier for contributors and users to navigate and stay informed.

### 📚 Documentation
- **Streamlined Repository**: We have removed the unused `.dockerignore` file. This cleanup reduces clutter, making the project structure more straightforward and easier to manage. Users and contributors will benefit from a cleaner repository, enhancing the overall development experience.

### 🔧 Maintenance
- **Updated Changelog**: The changelog has been refreshed to accurately reflect recent updates and improvements. This ensures that all users and developers are kept informed about the latest changes, promoting transparency and effective communication within the community.

This changelog highlights the key changes in version 2.0.0, focusing on the benefits of a cleaner repository and the importance of maintaining up-to-date documentation for user and developer engagement.

## [v2.0.0-beta.41] - 2025-07-30

This release focuses on enhancing the build process and updating documentation to improve efficiency and clarity for users.

### ⚡ Performance
- **Build System Optimization**: The Dockerfile configurations for the manager and worker components have been streamlined. This results in faster build times and improved deployment efficiency, allowing users to experience quicker updates and reduced resource consumption during the build process.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect recent changes and improvements, ensuring users have access to the latest information about updates and modifications. This facilitates better understanding and tracking of the project's evolution.

### 🔧 Maintenance
- **Documentation Improvements**: Behind-the-scenes updates to documentation ensure that users can easily find and understand the latest changes, contributing to a smoother user experience.


## [v2.0.0-beta.40] - 2025-07-29

This release introduces significant enhancements to the plugin-smart-templates project, focusing on improved network integration, user interface configuration, and dependency management to boost overall user experience and system reliability.

### ✨ Features  
- **Enhanced Network Integration**: The Smart Templates Manager and authentication system now support additional networks, allowing for seamless integration and improved connectivity across diverse environments. This flexibility enhances operational efficiency and scalability for users managing multiple network setups.

### 🔄 Changes
- **Smart Templates UI Configuration Update**: The environment configuration for the Smart Templates user interface has been updated to ensure smoother backend and frontend integration. This change enhances the overall user experience by providing a more reliable and intuitive interface.

### 📚 Documentation
- **Updated Guidance**: The documentation has been revised to reflect the latest configuration changes, offering users clear instructions for setup and adjustments. This ensures users can easily adapt to the new enhancements without confusion.

### 🔧 Maintenance
- **Dependency Update**: The `@lerianstudio/console-layout` package has been updated to version 1.5.3, ensuring compatibility with the latest features and security updates. This maintenance step maintains the stability and performance of the frontend components.
- **Changelog Revision**: The CHANGELOG has been updated to provide users with a comprehensive overview of recent updates, facilitating easy tracking of the software's evolution.

These updates collectively enhance the functionality, reliability, and user experience of the plugin-smart-templates project, ensuring users benefit from improved network integration, updated configurations, and maintained dependencies.

## [v2.0.0-beta.39] - 2025-07-29

This release simplifies the report structure, enhances documentation, and improves system reliability. Users should review integrations to accommodate these updates.

### ⚠️ Breaking Changes
- **Removal of `LedgerID`**: The `LedgerID` has been removed from the report structure, affecting backend, database, and frontend components. This change requires users to update their integrations and workflows that depend on `LedgerID`. Please review and modify your systems to ensure compatibility with the new structure.

### ✨ Features  
- **Streamlined Report Structure**: By removing `LedgerID`, the report generation process is now more efficient, reducing complexity in data handling and validation. This change enhances performance and simplifies maintenance for developers.

### 🐛 Bug Fixes
- **Updated Test Cases**: Test cases have been adjusted to remove dependencies on `LedgerID`, ensuring that tests accurately reflect the new system behavior. This improves test reliability and maintains comprehensive coverage.

### 📚 Documentation
- **Updated API Documentation**: Documentation has been revised to reflect the removal of `LedgerID` from swagger definitions and examples. This ensures that developers have accurate, up-to-date information when working with the API.

### 🔧 Maintenance
- **Codebase Cleanup**: Obsolete references to `LedgerID` have been removed from the backend and test components, enhancing code quality and maintainability. This cleanup reduces technical debt and prepares the system for future enhancements.


## [v2.0.0-beta.37] - 2025-07-28

This release focuses on enhancing the reliability and transparency of the plugin-smart-templates project, ensuring accurate data handling and providing clear documentation of changes.

### 🐛 Bug Fixes
- **Improved Data Handling**: Resolved an issue with form data request headers that could lead to incorrect data submission. This fix enhances the reliability of data processing across both frontend and backend components, ensuring user inputs are handled accurately and consistently.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to reflect the latest changes and improvements. This ensures users have access to a clear history of updates and fixes, enhancing transparency and understanding of the software's evolution.

These updates collectively contribute to a more stable and user-friendly experience, reducing errors and providing users with up-to-date information about the software's development.

## [v2.0.0-beta.36] - 2025-07-28

This release focuses on enhancing the development workflow with improved build processes and updated documentation, ensuring a more efficient and informed environment for developers.

### ⚡ Performance
- **Build Process Optimization**: Streamlined the build configuration by removing unnecessary parameters, potentially reducing build times and improving efficiency for developers. This enhancement makes the build process more straightforward, allowing developers to focus more on coding and less on configuration.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to include the latest changes and improvements. This ensures that all users and developers have access to the most recent information, facilitating better understanding and communication of project updates.

### 🔧 Maintenance
- **Behind-the-Scenes Enhancements**: Internal improvements in the build and documentation components aim to enhance the overall development experience without directly impacting end-user functionality.

This release does not introduce any breaking changes or new features but focuses on refining the development process and maintaining up-to-date project documentation.

## [v2.0.0-beta.35] - 2025-07-28

This release focuses on streamlining data handling and improving system efficiency by removing redundant fields, which requires user attention for continued functionality.

### ⚠️ Breaking Changes
- **Removal of `ledgerId` Field**: The `ledgerId` field has been removed from forms across authentication, backend, configuration, database, and frontend components. This change simplifies data management but requires users to update their configurations and code. 
  - **Migration Steps**:
    1. Review and update any API calls that include `ledgerId`.
    2. Modify database schemas to remove references to `ledgerId`.
    3. Adjust frontend forms to accommodate this change.
  - **Impact**: Ensures a more efficient system by eliminating unnecessary fields.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to accurately document recent changes, ensuring users have a clear history of updates and enhancements.

This release requires users to take action to maintain compatibility and avoid disruptions. Please review the breaking changes section carefully.

## [v2.0.0-beta.34] - 2025-07-28

This release enhances the efficiency of the build process, significantly reducing build times and improving the overall development workflow. Updated documentation ensures developers can easily adapt to these changes.

### ⚡ Performance
- **Build Process Optimization**: We've streamlined the build process, resulting in faster compilation and deployment times. This enhancement allows for quicker testing and iteration, improving developer productivity.

### 📚 Documentation
- **Updated Build Process Documentation**: The documentation has been revised to provide clear guidance on the new build procedures, helping developers understand and leverage the enhancements effectively.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG file has been updated to include recent changes, ensuring transparency and aiding users in tracking project evolution.

This changelog highlights the key improvements and documentation updates in version 2.0.0, focusing on the benefits and impact for users. The performance section emphasizes the enhanced build process, while the documentation section ensures developers have the necessary information to adapt to these changes. The maintenance section underscores the importance of keeping users informed about the project's progress.

## [v2.0.0-beta.33] - 2025-07-28

This release focuses on enhancing the stability and maintainability of the plugin-smart-templates, ensuring a more reliable and efficient user experience.

### 🐛 Bug Fixes
- Updated backend and database tests by removing deprecated parameters, leading to more accurate and reliable test results. This reduces the risk of encountering false positives or negatives, enhancing overall system reliability.

### 🔧 Maintenance
- Refactored backend code by replacing the collection parameter with constants. This simplification improves code maintainability and reduces potential errors, resulting in a more stable backend performance.
- Updated the CHANGELOG documentation to reflect recent changes, ensuring users are informed about the latest updates and improvements.

Each of these changes contributes to a more stable and maintainable system, enhancing the overall user experience with the plugin-smart-templates project.

## [v2.0.0-beta.32] - 2025-07-28

This release enhances the deployment and configuration processes, making them more streamlined and user-friendly. It also ensures that documentation is up-to-date to support these changes.

### ✨ Features
- **Simplified Deployment with Docker**: We've introduced an updated Docker configuration along with a new runtime environment script. This change simplifies the deployment process, ensuring consistency across different environments and making it easier for developers to set up and manage the application infrastructure. This enhancement reduces setup time and minimizes potential errors, allowing you to focus more on development and less on configuration.

### 📚 Documentation
- **Updated Setup Guides**: The documentation has been updated to reflect the new Docker configuration and runtime environment script. This ensures that both users and developers have the most current information for setting up and running the application, reducing setup time and potential errors.

### 🔧 Maintenance
- **Changelog Updates**: The CHANGELOG has been updated to include the latest changes, ensuring transparency and keeping users informed about the updates and improvements made in this release.


## [v2.0.0-beta.31] - 2025-07-28

This release focuses on simplifying the codebase and enhancing documentation processes, ensuring a smoother experience for users managing API documentation and transaction templates.

### ⚠️ Breaking Changes
- **Removal of OpenAPI to Postman Conversion**: This release removes the scripts and references for converting OpenAPI to Postman. Users relying on these scripts will need to explore alternative solutions for API documentation conversion. This change streamlines the codebase by removing outdated functionality, improving overall maintainability.

### 📚 Documentation
- **Refined Transaction Templates**: Unused XML fields have been removed, and value calculations have been improved within transaction templates. These changes make it easier for users to understand and implement transaction templates effectively.
- **Enhanced Documentation Processes**: Updates to the documentation generation processes and `docker-compose` configurations make it easier to manage and deploy documentation, ensuring users have access to the latest information.

### 🔧 Maintenance
- **Dependency Updates**: Upgraded `lib-commons` to version `1.19.0-beta.2`, ensuring compatibility with the latest features and security patches, thereby maintaining system stability and performance.
- **Repository Clean-Up**: Improved the `.gitignore` file to exclude unnecessary directories like `scripts/node_modules` and introduced a `.dockerignore` file to exclude docs and markdown files, helping maintain a cleaner repository.
- **Changelog Updates**: The CHANGELOG has been updated to reflect the recent changes, providing users with a clear understanding of the latest updates and modifications.

This release ensures that users of the plugin-smart-templates system benefit from a more streamlined and efficient experience, with improved documentation and a cleaner codebase.

## [v2.0.0-beta.30] - 2025-07-28

This release focuses on enhancing the documentation experience with a new context input feature, designed to improve user engagement and understanding.

### ✨ Features  
- **Enhanced Documentation**: A new context input feature has been added to the documentation, providing a more intuitive way for users to interact with and understand context-related functionalities. This enhancement aims to improve user engagement and comprehension, offering clearer guidance and examples.

### 📚 Documentation
- **Improved Clarity and Guidance**: The documentation now includes more detailed examples and explanations, making it easier for users to grasp complex concepts and utilize the software effectively.

### 🔧 Maintenance
- **Updated Changelog**: The changelog has been updated to reflect recent changes and improvements, ensuring users have access to the latest information about updates and features. This transparency aids in version tracking and user awareness.

No breaking changes were identified in this release, ensuring a smooth transition for users upgrading to the latest version.


## [v2.0.0-beta.29] - 2025-07-28

This release enhances the organization and usability of documentation for projects involving Go components, ensuring a smoother integration and improved clarity for developers.

### ✨ Features  
- **Improved Documentation for Go Projects**: We've introduced the ability to set context paths for Go files within our documentation. This enhancement allows developers to better organize and reference Go code, significantly improving the clarity and usability of documentation. Teams working with Go will find it easier to integrate their code into documentation workflows, streamlining their development process.

### 📚 Documentation
- **Enhanced Clarity and Usability**: The documentation now supports context paths for Go files, making it easier for developers to organize and reference their Go code. This improvement is particularly beneficial for teams that frequently work with Go, as it enhances the overall documentation experience and supports more efficient project management.

### 🔧 Maintenance
- **Updated Changelog**: We've refreshed the CHANGELOG to include the latest features and enhancements. This update ensures transparency and keeps all users informed about the most recent changes, supporting effective version tracking and communication.


## [v2.0.0-beta.28] - 2025-07-28

This release focuses on enhancing the documentation structure and ensuring up-to-date release information, significantly improving the developer experience with plugin-smart-templates.

### ✨ Features  
- **Enhanced Documentation for Go Files**: We've introduced context path settings for Go files within the documentation. This improvement helps developers quickly locate and understand the context of Go files, making it easier to navigate and comprehend the codebase.

### 📚 Documentation
- **Improved Code Organization**: The new documentation structure provides clearer organization and context for Go files, enhancing the overall clarity and usability for developers.

### 🔧 Maintenance
- **Updated Changelog**: The changelog has been updated to reflect the latest changes and improvements, ensuring transparency and easier tracking of project progress for both users and developers.


This changelog is designed to communicate the key updates in version 2.0.0 of the plugin-smart-templates project, focusing on the benefits these changes bring to users and developers.

## [v2.0.0-beta.27] - 2025-07-28

This release introduces a new context path configuration to enhance codebase organization and updates documentation for improved project tracking.

### ✨ Features  
- **Enhanced Codebase Modularity**: We've introduced a new context path configuration for Go files. This improvement enhances the modularity and organization of the project, making it easier for developers to manage and navigate the codebase. This change is designed to boost development efficiency and maintainability, providing a more streamlined workflow.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest enhancements and changes. This ensures all stakeholders have access to up-to-date information, facilitating better communication and project tracking.

### 🔧 Maintenance
- **Documentation Alignment**: Ensured that all documentation and frontend components are aligned with the new context path structure, providing a cohesive development experience and minimizing discrepancies across the project.


This changelog focuses on the key benefits and impacts of the new release, highlighting the improved modularity and documentation updates that enhance user experience and project management.

## [v2.0.0-beta.26] - 2025-07-28

This release introduces a significant enhancement to streamline development processes and improve documentation clarity, benefiting both developers and project stakeholders.

### ✨ Features  
- **Context Path Management**: Introduced a new feature that allows developers to set context paths for Go files across build, frontend, and documentation components. This enhancement improves code organization, making it easier to navigate and maintain the project. Developers can now manage context paths more efficiently, leading to a more streamlined and consistent development process.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest changes and improvements, ensuring that all team members and stakeholders have access to the most current information. This update enhances transparency and supports effective project management by keeping everyone informed about the project's progress.

### 🔧 Maintenance
- **Documentation Maintenance**: Regular updates and refinements to the documentation ensure that it remains accurate and useful, supporting ongoing development and collaboration efforts.


This changelog is structured to highlight the user-centric benefits and practical applications of the new features and documentation updates. The focus is on how these changes enhance the user experience and project management, presented in a clear and accessible manner.

## [v2.0.0-beta.25] - 2025-07-28

This release introduces a major enhancement to the build process, improving flexibility and organization for developers working with Go files.

### ✨ Features  
- **Enhanced Build Configurability**: You can now set the context path for Go files, allowing for more organized and efficient project structures. This feature empowers developers to customize their build processes, resulting in streamlined workflows and easier code management.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to include the latest features and improvements. This ensures all users have access to up-to-date information, promoting transparency and informed decision-making.

### 🔧 Maintenance
- **Documentation Refresh**: We've enhanced our documentation to reflect recent changes, supporting better user understanding and engagement with the latest project developments.

This changelog focuses on the key feature introduced in version 2.0.0, emphasizing its benefits for developers. The documentation updates are highlighted to ensure users are aware of the latest information available. The structure and language are designed to be accessible and informative, catering to both technical and non-technical users.

## [v2.0.0-beta.24] - 2025-07-27

This release enhances the organization and accessibility of documentation for Go files, providing a smoother experience for developers. Additionally, the changelog has been updated to ensure users are well-informed about the latest improvements.

### ✨ Features  
- **Improved Documentation for Go Files**: We've introduced context path settings that enhance the clarity and structure of documentation related to Go projects. This improvement makes it easier for developers to navigate and understand the codebase, ultimately boosting productivity and reducing time spent searching for information.

### 📚 Documentation
- **Enhanced Documentation Structure**: The documentation now includes context path settings for Go files, significantly improving how information is organized and accessed. This change is particularly beneficial for developers working with complex Go projects, as it simplifies the process of locating relevant documentation.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been thoroughly updated to reflect all recent changes and improvements. This ensures that users have the latest information at their fingertips, promoting transparency and aiding in effective version tracking.

---

This update does not include any breaking changes, ensuring a seamless upgrade experience for all users. Enjoy the enhanced documentation capabilities and stay informed with our updated changelog.

## [v2.0.0-beta.23] - 2025-07-27

This release introduces a significant enhancement to the build process, allowing for greater flexibility and ease of integration across different development environments.

### ✨ Features  
- **Flexible Context Path for Builds**: You can now set the context path for Go files during the build process. This enhancement offers developers more control over their development workflow, making it easier to integrate with various environments and streamline project setup.

### 📚 Documentation
- **Updated Changelog**: The changelog has been refreshed to include the latest updates, ensuring you have access to current information about the software's progress and improvements. This transparency helps you stay informed and make better decisions regarding your use of the plugin.

### 🔧 Maintenance
- **Release Management Improvements**: We've made behind-the-scenes updates to our release management process, ensuring that all documentation is up-to-date and accurately reflects the latest software changes. This commitment to maintenance supports a smoother user experience and enhances overall project reliability.


## [v2.0.0-beta.22] - 2025-07-27

This release focuses on enhancing the development environment by standardizing Docker structures, improving documentation, and streamlining the onboarding process for developers.

### ✨ Features  
- **Standardized Docker Structure**: We've unified the Docker setup across build, documentation, and frontend components. This change simplifies the development and deployment process, ensuring a consistent environment that reduces setup time and configuration errors. Developers will find it easier to onboard and maintain their development environments.

### 📚 Documentation
- **Updated Changelog**: The changelog has been refreshed to include the latest updates and improvements. This ensures that developers and users have access to current information about the project's progress, aiding in better tracking and understanding of changes.

### 🔧 Maintenance
- **Infrastructure Enhancement**: Behind-the-scenes improvements have been made to the development infrastructure, providing a more reliable and efficient environment for ongoing development efforts.

---

No breaking changes, bug fixes, or additional performance improvements were identified in this release. The focus remains on enhancing the development infrastructure and ensuring up-to-date documentation.


## [v2.0.0-beta.21] - 2025-07-25

This release introduces a streamlined deployment process with Docker, enhancing consistency and ease of setup for developers. Updated documentation ensures smooth adoption of these improvements.

### ✨ Features
- **Docker Deployment**: We've added a Dockerfile configuration to simplify the deployment process across various environments. This ensures consistent application behavior in production and enhances the development workflow, making it easier for developers to set up and deploy the application efficiently.

### 📚 Documentation
- **Updated Guides**: Our documentation now includes detailed instructions on using the new Docker setup. This update provides clear guidance for developers, reducing onboarding time and minimizing potential setup errors.

### 🔧 Maintenance
- **Changelog Updates**: We've updated the CHANGELOG to reflect recent changes and enhancements. This ensures that all stakeholders have access to the latest information about the project's progress and updates, supporting transparency and keeping project documentation current.

This changelog communicates the key benefits of the new Docker deployment feature, highlights the supporting documentation updates, and notes the maintenance work done to keep the changelog current. The focus is on how these changes improve the user experience and streamline the development process.

## [v2.0.0-beta.20] - 2025-07-25

This release introduces a streamlined deployment process with Docker support, enhancing setup consistency and ease of use. Documentation updates accompany these changes to guide users effectively.

### ✨ Features  
- **Dockerfile Configuration for Deployment**: We've added a Dockerfile to simplify the deployment of plugin-smart-templates. This feature ensures that your development and production environments are consistent, reducing setup time and potential configuration errors. Users can now deploy the application in a containerized environment, making it easier to manage and scale. Check the updated documentation for a step-by-step guide to using Docker with our plugin.

### 📚 Documentation
- **Docker Setup Guide**: Updated the documentation to include comprehensive instructions on setting up and using the new Docker configuration. This guide will help you quickly get started with the containerized deployment process, ensuring a smooth transition to the new setup.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been revised to reflect the latest changes and improvements, providing a clear and up-to-date record for users and developers to track the project's progress and updates.

These updates focus on improving the deployment process and maintaining clear communication with users through updated documentation. There are no breaking changes in this release, ensuring a smooth transition for users upgrading to the latest version.

## [v2.0.0-beta.19] - 2025-07-25

This release introduces Docker support to streamline deployment and updates the changelog for enhanced transparency. Enjoy a more efficient setup process with no breaking changes to worry about.

### ✨ Features
- **Docker Configuration**: We've added Docker support to simplify the deployment of plugin-smart-templates. This feature provides a consistent environment, reducing setup errors and improving deployment efficiency. Whether you're a developer or user, this enhancement ensures a smoother and more reliable setup experience.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to reflect the latest changes and improvements, ensuring you stay informed about the project's development progress and modifications.

### 🔧 Maintenance
- **Project Transparency**: By maintaining up-to-date documentation and changelogs, we enhance communication and transparency, allowing users to easily track the evolution of the plugin-smart-templates.

This changelog is designed to help you quickly understand the benefits of the latest release and how it impacts your use of plugin-smart-templates. Enjoy the new features and improvements with confidence, knowing there are no compatibility issues to address.

## [v2.0.0-beta.18] - 2025-07-25

This release introduces a streamlined deployment process for plugin-smart-templates using Docker, enhancing setup consistency and ease of use. Updated documentation supports these improvements, ensuring a smooth user experience.

### ✨ Features  
- **Dockerfile Configuration**: We've added a Dockerfile to simplify the deployment of plugin-smart-templates. This feature ensures a standardized environment across installations, making it easier for users to set up and maintain consistency. The accompanying documentation provides step-by-step guidance on using Docker, which is particularly beneficial for those seeking efficient and reproducible setups.

### 📚 Documentation
- **Enhanced Deployment Guide**: Updated the documentation to include detailed instructions on deploying using Docker. This addition helps users quickly understand and implement the new setup process, reducing the time and effort needed to get started.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest features and improvements, ensuring users have access to the most current information about the project's development. This transparency helps users stay informed about new capabilities and enhancements.

All changes in this release are backward-compatible, ensuring a seamless upgrade experience without requiring modifications to existing setups.

## [v2.0.0-beta.17] - 2025-07-25

This release introduces a major enhancement to streamline deployment and improve user experience through Docker integration, alongside essential documentation updates.

### ✨ Features  
- **Dockerfile Configuration**: We've added a Dockerfile to simplify the deployment process of the plugin-smart-templates project. This feature provides a consistent and standardized environment, making it easier for users to set up and run the application in both development and production. This reduces setup time and potential configuration errors, allowing users to focus more on their projects.

### 📚 Documentation
- **Enhanced Deployment Guide**: Updated documentation to include comprehensive instructions on using Docker with the plugin-smart-templates. This ensures users have clear guidance on setting up their environments efficiently, leveraging the new Docker support for a seamless experience.

### 🔧 Maintenance
- **Changelog Update**: We've updated the CHANGELOG to reflect the latest features and improvements. This ensures transparency and helps users and contributors stay informed about the project's evolution and ongoing maintenance efforts.

This changelog provides a concise overview of the new features and improvements, focusing on the benefits and impact for users. It ensures clarity and accessibility, making it easy for users to understand the changes and how they can leverage them in their workflows.

## [v2.0.0-beta.16] - 2025-07-25

This release introduces a streamlined deployment process with Dockerfile configuration, enhancing setup consistency and portability for developers. Additionally, updated documentation ensures users have access to the latest project insights.

### ✨ Features  
- **Dockerfile Configuration**: We've added a Dockerfile to simplify the deployment process. This ensures that developers can set up their environments quickly and consistently across different systems, reducing setup time and potential errors. This enhancement is particularly beneficial for teams working in varied environments, providing a reliable and portable solution.

### 📚 Documentation
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest changes and improvements. This ensures that all users and developers have access to the most current information about the project's evolution, enhancing transparency and communication.

### 🔧 Maintenance
- **Documentation Maintenance**: Regular updates to documentation improve clarity and accessibility for all users, ensuring that everyone benefits from the latest project developments and can easily navigate changes.

This release emphasizes enhancing the deployment process and maintaining up-to-date documentation, providing a more seamless experience for developers and users alike.

## [v2.0.0-beta.15] - 2025-07-25

This release introduces a streamlined setup process for plugin-smart-templates, enhancing ease of deployment and ensuring consistent configurations across systems.

### ✨ Features
- **Simplified Deployment with Docker**: We've added a new Dockerfile configuration, making it easier for developers to deploy plugin-smart-templates. This update ensures that your environment is consistently configured, reducing setup time and potential errors. Check out the updated documentation for a step-by-step guide on using this new feature.

### 📚 Documentation
- **Enhanced Setup Guides**: The documentation has been updated to reflect the new Dockerfile configuration. This includes a detailed guide to help you get started quickly and maintain the application effectively. Whether you're a new user or updating your existing setup, these improvements will simplify your workflow.

### 🔧 Maintenance
- **Changelog Updates**: We have updated the CHANGELOG to include the latest features and enhancements, ensuring you have access to the most current information about the project's development. This routine maintenance keeps our documentation accurate and helpful for all users.

This changelog provides a clear and concise overview of the changes in version 2.0.0, focusing on the benefits to users and ensuring that the information is accessible and actionable.

## [v2.0.0-beta.14] - 2025-07-25

This release focuses on simplifying the deployment process with Docker support, enhancing both ease of use and reliability for developers. No breaking changes ensure a smooth upgrade experience.

### ✨ Features  
- **Dockerfile Configuration**: Streamline your deployment with the new Dockerfile setup, allowing for consistent environments across development and production. This feature is particularly beneficial for maintaining uniformity and reducing environment-specific issues, making it easier to manage application dependencies and configurations.

### 📚 Documentation
- **Changelog Update**: The changelog has been updated to include the latest changes, ensuring you have access to the most current information about the software's evolution. This transparency helps you stay informed about new features and improvements.

### 🔧 Maintenance
- **Documentation Enhancements**: Improved documentation to support the new Dockerfile configuration, providing clear guidance for setting up and running the application in a containerized environment. This helps users quickly adapt to the new deployment process and leverage the benefits of containerization.

This changelog provides a concise overview of the new features and improvements, focusing on the benefits to users and ensuring they understand the impact and value of the changes.

## [v2.0.0-beta.13] - 2025-07-25

This release introduces a streamlined deployment process for the plugin-smart-templates, enhancing setup efficiency and reliability for users.

### ✨ Features  
- **Streamlined Deployment with Dockerfile**: We've introduced a Dockerfile configuration to simplify the deployment process of the plugin-smart-templates. This ensures a consistent setup environment, reducing the risk of configuration errors and enhancing portability across different systems. Users will find it easier to get started and maintain their setups with this new feature.

### 📚 Documentation
- **Updated Changelog**: The CHANGELOG has been updated to reflect recent changes and improvements, ensuring users have access to the latest information about updates and enhancements. This update facilitates better understanding and tracking of the project's evolution.

### 🔧 Maintenance
- **Documentation Enhancements**: Continued improvements to documentation ensure that users are well-informed about the latest features and changes, supporting a smoother user experience.

This changelog focuses on the key benefits and user impacts of the release, using clear and accessible language to communicate the value of the new Dockerfile configuration and documentation updates.

## [v2.0.0-beta.12] - 2025-07-25

This release introduces a new Dockerfile configuration, streamlining the deployment process and enhancing consistency across environments. Documentation updates accompany this change, ensuring a smooth transition for developers.

### ✨ Features  
- **Dockerfile Configuration**: We've added a new Dockerfile to simplify the deployment process. This enhancement allows for consistent environment setups, making it easier for developers to deploy and manage the plugin-smart-templates project. The updated documentation provides step-by-step guidance on using this new feature effectively.

### 📚 Documentation
- **Deployment Guide Update**: The documentation now includes detailed instructions on utilizing the new Dockerfile configuration. This update aims to assist developers in quickly adapting to the new deployment process without any disruption.

### 🔧 Maintenance
- **Changelog Update**: We've refreshed the CHANGELOG to ensure all stakeholders have access to a comprehensive log of updates. This update facilitates better tracking of project progress and changes over time.

In this release, we focused on enhancing the deployment process through a new Dockerfile configuration, which improves ease of use and reliability for developers. The documentation updates ensure users can quickly adapt to these changes, maintaining a seamless workflow.

## [v2.0.0-beta.11] - 2025-07-25

This release of plugin-smart-templates introduces a streamlined deployment process and enhanced documentation, making it easier for users to set up and manage their environments consistently.

### ✨ Features  
- **Dockerfile Configuration**: We have introduced a Dockerfile to simplify the deployment process. This feature provides a consistent environment for running the application, reducing setup time and potential configuration errors. The Dockerfile is accompanied by comprehensive documentation to guide users through the setup process, ensuring a smooth experience.

### 📚 Documentation
- **Comprehensive Setup Guide**: The new Dockerfile comes with detailed documentation that walks users through the setup process. This addition helps users understand the prerequisites and steps necessary to deploy the plugin-smart-templates efficiently.

### 🔧 Maintenance
- **Changelog Update**: The CHANGELOG has been updated to reflect the latest changes and improvements. This ensures that users and developers have access to the most current information about the project's evolution and can track updates over time.

Each of these changes contributes to a more efficient and user-friendly experience, particularly by enhancing the deployment process and maintaining clear documentation of project updates.

## [v2.0.0-beta.10] - 2025-07-25

This release enhances the documentation of Go files, making it easier for developers to navigate and understand the codebase. We've also updated the changelog to ensure transparency and keep all stakeholders informed.

### ✨ Features  
- **Improved Documentation Structure**: We've introduced context path settings for Go files, which significantly enhances the organization and accessibility of documentation. This update helps developers quickly locate and understand the context of Go files within the project, streamlining the development and maintenance process.

### 📚 Documentation
- **Context Path Settings for Go Files**: This improvement provides a clear structure for documentation, making it easier for developers to navigate the codebase and find relevant information efficiently.

### 🔧 Maintenance
- **Changelog Update**: The changelog has been updated to reflect the latest changes and improvements. This ensures that users and developers have access to the most current information about the project's progress, maintaining transparency and facilitating easier tracking of changes over time.

This changelog highlights the key improvements in version 2.0.0, focusing on the benefits to developers and stakeholders. The documentation enhancements improve the user experience by making it easier to navigate and understand the project, while the updated changelog maintains transparency and keeps everyone informed.

## [v2.0.0-beta.9] - 2025-07-25

This release enhances the flexibility of your Docker deployments and improves the transparency of our software updates through a comprehensive changelog.

### ✨ Features  
- **Expanded Docker Configuration Options**: We've added new configuration options to our Docker setup, allowing for more tailored and efficient deployment scenarios. This enhancement enables users to better integrate the application with diverse infrastructure environments, providing greater control and customization.

### 📚 Documentation
- **Enhanced Docker Setup Guide**: Our documentation now includes detailed instructions on the new Docker configuration options. This update ensures users can easily implement the latest features and optimize their deployment strategies.

### 🔧 Maintenance
- **Updated Changelog for Improved Transparency**: We've revised our changelog to offer a clearer and more comprehensive history of changes. This helps users stay informed about the software's evolution and better understand the improvements and features introduced in each release.

This changelog communicates the key updates in version 2.1.0, focusing on the benefits of enhanced Docker configuration options and improved documentation. The maintenance update ensures users have access to a transparent and detailed history of changes.

## [v2.0.0-beta.8] - 2025-07-25

This release focuses on enhancing the reliability of our deployment process and ensuring users have up-to-date information. A critical bug fix improves Docker deployment for the frontend, and documentation updates keep everyone informed.

### 🐛 Bug Fixes
- **Docker Deployment Reliability**: Resolved a configuration issue in the Dockerfile affecting the frontend component. This fix ensures smoother operation and fewer setup errors when deploying the application in Docker environments, enhancing overall deployment reliability for developers and users.

### 📚 Documentation
- **Changelog Updates**: The CHANGELOG has been updated to reflect recent changes, ensuring that users and developers have access to the latest information about software updates. This helps in better understanding and tracking the project's evolution.

### 🔧 Maintenance
- **Documentation Maintenance**: Regular updates to documentation ensure clarity and accuracy, providing users with reliable resources for understanding software changes and enhancements.

This changelog provides a concise yet informative overview of the changes in version 2.0.0, emphasizing the benefits to users and maintaining a professional tone throughout.

## [v2.0.0-beta.7] - 2025-07-24

This release introduces streamlined deployment through Docker, enhancing the consistency and ease of setting up environments for developers. Additionally, documentation updates ensure users have the latest information for seamless interaction with the project.

### ✨ Features  
- **Streamlined Deployment with Docker**: The introduction of a Dockerfile configuration allows for simplified and consistent deployment across different environments. This feature enables developers to easily set up and deploy the application, reducing setup time and potential configuration errors. Comprehensive documentation is provided to guide users through the Docker setup process.

### 📚 Documentation
- **Updated Changelog**: The changelog has been updated to reflect recent changes and enhancements, ensuring that all users and developers have access to the most current information. This update is part of our ongoing commitment to transparency and user support.

### 🔧 Maintenance
- **Routine Documentation Maintenance**: Regular updates have been made to the documentation to keep it accurate and useful for all users. This ensures that all project-related information remains current and accessible, facilitating better user interaction and understanding.

This changelog focuses on the key enhancements in deployment capabilities through Docker, highlighting the benefits for developers and ensuring users are informed of the latest documentation updates.

## [v2.0.0-beta.6] - 2025-07-23

This release introduces enhanced observability features and updates key dependencies to improve system monitoring and security.

### ✨ Features  
- **Improved Observability**: We've integrated OpenTelemetry span attributes across our backend and database services. This enhancement allows for better monitoring and traceability, providing users with deeper insights into system performance and operation flows.

### 🔧 Maintenance
- **Dependency Updates**: We've updated `lib-license-go` to version 1.23.0 and `fiber` to version 2.52.9. These updates ensure that our system is secure and compatible with the latest features, enhancing the overall stability and security of your experience.
- **Codebase Simplification**: Removed redundant OpenTelemetry import alias in template services, streamlining the code and reducing potential errors.

This release focuses on improving the reliability and observability of the system, ensuring you have the tools needed for effective monitoring and diagnostics.


## [v2.0.0-beta.5] - 2025-07-23

This release focuses on enhancing system stability and improving the user experience with key bug fixes and updated documentation.

### 🐛 Bug Fixes
- **Authentication Reliability**: Resolved issues with token management in the authentication system. Users will now experience more reliable and consistent login processes, reducing potential disruptions during access.
- **Configuration Clarity**: Updated configuration settings and documentation to align with backend improvements. This change helps developers set up their environments correctly, minimizing setup errors and improving overall efficiency.

### ⚡ Performance
- **Dependency Updates**: Enhanced system stability and performance by updating dependencies. This update incorporates the latest security patches and performance improvements, ensuring a smoother and safer user experience.

### 📚 Documentation
- **Configuration Guides**: Revised documentation to reflect the latest configuration settings, aiding developers in understanding and applying changes effectively. This ensures that all users have access to accurate and up-to-date information.

### 🔧 Maintenance
- **Changelog Updates**: The changelog has been meticulously updated to document all recent changes and improvements, ensuring transparency and providing a reliable reference for users and developers alike.

These updates collectively enhance the reliability, performance, and usability of the plugin-smart-templates project, ensuring a more seamless experience for both end-users and developers.

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
