# Plugin Smart Templates

## Overview

The Smart Templates Plugin is a service designed to manage and generate customizable reports using predefined templates.

## Quick Start

1. **Clone the Repository:**
    ```bash
    git clone https://github.com/LerianStudio/plugin-smart-templates.git
    ```

2. **Install Dependencies:**
    ```bash
    cd plugin-smart-templates
    go mod tidy
    ```

3. **Run the Server:**
    ```bash
   make up
    ```
   
4. **Access the API:**
   Visit `http://localhost:4005` to interact with the API.

## Components

### Manager

Responsible for managing templates and reports, the Service provides a complete CRUD for creating, listing, updating, and deleting templates as well as generating and retrieving reports. 
It exposes a RESTful API with full documentation available via Swagger at:

📄 http://localhost:4005/swagger/index.html

### Worker

Responsible for report generation, the worker is initialized whenever there are messages of this type in the RabbitMQ queue.
Based on the fields requested in the report, it connects to the respective databases and performs queries dynamically.

## Generate report RabbitMQ message

- Exchange: `smart-templates.generate-report.exchange`
- Queue: `smart-templates.generate-report.queue`
- Key: `smart-templates.generate-report.key`

```json
{
   "templateId": "019538ee-deee-769c-8859-cbe84fce9af7",
   "reportId": "019615d3-c1f6-7b1d-add4-6912b76cc4f2",
   "ledgerId": ["01963aba-18c3-77a5-adcc-18028fc7420d"],
   "outputFormat": "html",
   "mappedFields": {
      "onboarding": {
         "organization": ["legal_name"],
         "ledger": ["name", "status"]
      }
   }
}
```

The field mapping should be:
```json
{
   "mappedFields": [
      {
         "<database-name>": {
            "<table-name>": ["<field-name>, <field-name>"]
         }
      }
   ]
}
```

## File storage with MinIO

We use MinIO to store both the template files and the generated reports in their final format.

When starting the MinIO container using the project’s docker-compose, it uses the minio/mc image, which is the official image of the MinIO Client. This is a CLI utility similar to awscli, used to interact with MinIO servers.
The CLI image is used to create a user with upload and read permissions, which will be used by the service and the worker. It also creates two buckets: one for templates and another for the generated reports.

## Swagger Documentation

The Plugin includes Swagger documentation that helps in visualizing and interacting with the API endpoints. You can access the documentation by running the project and navigating to `http://localhost:4005/swagger/index.html`.

## References

- https://github.com/flosch/pongo2/blob/master/template_tests/filters.tpl