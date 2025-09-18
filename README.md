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

## File storage with SeaweedFS

We use SeaweedFS (Filer + Volume + Master) to store both template files and generated reports. Access is done via Filer HTTP API.

### Authentication (JWT)

The Filer HTTP endpoints can be protected with JWT (HS256). This project enables JWT by generating a `security.toml` at container startup (entrypoint) using the following environment variables:

- `SEAWEEDFS_JWT_SIGNING_READ`: secret used to validate tokens for read operations (GET/HEAD)
- `SEAWEEDFS_JWT_SIGNING_WRITE`: secret used to validate tokens for write operations (POST/PUT/DELETE)

Generate strong secrets (32 bytes) and encode in base64, for example:

```bash
openssl rand -base64 32
```

Set the same values in Infra (Filer) and in Manager/Worker `.env` files. The application generates short-lived tokens automatically when calling the Filer.

## Swagger Documentation

The Plugin includes Swagger documentation that helps in visualizing and interacting with the API endpoints. You can access the documentation by running the project and navigating to `http://localhost:4005/swagger/index.html`.

## References

- https://github.com/flosch/pongo2/blob/master/template_tests/filters.tpl