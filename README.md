![banner](image/README/reporter_banner.png)

<div align="center">

[![Latest Release](https://img.shields.io/github/v/release/LerianStudio/reporter?include_prereleases)](https://github.com/LerianStudio/reporter/releases)
[![License](https://img.shields.io/badge/license-Elastic%20License%202.0-4c1.svg)](LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/lerianstudio/reporter)](https://goreportcard.com/report/github.com/lerianstudio/reporter)
[![Discord](https://img.shields.io/badge/Discord-Lerian%20Studio-%237289da.svg?logo=discord)](https://discord.gg/DnhqKwkGv3)

</div>

# Lerian Reporter

A service for managing and generating customizable reports using templates. Reporter connects directly to your databases (PostgreSQL and MongoDB) and renders reports in multiple formats (HTML, PDF, CSV, XML, TXT).

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Data Sources](#data-sources)
- [Templates](#templates)
- [API Reference](#api-reference)
- [Development](#development)
- [Contributing](#contributing)
- [Security](#security)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## Overview

Reporter is a report generation service that:

- **Manages templates** using [Pongo2](https://github.com/flosch/pongo2) (Django-like templating for Go)
- **Connects to multiple databases** (PostgreSQL and MongoDB) configured via environment variables
- **Generates reports** in various formats: HTML, PDF, CSV, XML, TXT
- **Processes asynchronously** using RabbitMQ for scalable report generation
- **Stores files** in S3-compatible storage (AWS S3, SeaweedFS, MinIO)

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           REPORTER                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐   │
│  │   Manager   │ ──────► │  RabbitMQ   │ ──────► │   Worker    │   │
│  │  (REST API) │         │   Queue     │         │ (Generator) │   │
│  └─────────────┘         └─────────────┘         └──────┬──────┘   │
│         │                                                │          │
│         │                                                │          │
│         ▼                                                ▼          │
│  ┌─────────────┐                                 ┌─────────────┐   │
│  │   MongoDB   │                                 │ Data Sources│   │
│  │  (metadata) │                                 │ PostgreSQL  │   │
│  └─────────────┘                                 │  MongoDB    │   │
│         │                                        └─────────────┘   │
│         │                                                │          │
│         ▼                                                ▼          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Object Storage (S3)                       │   │
│  │         AWS S3 / SeaweedFS / MinIO (Templates & Reports)     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Description |
|-----------|-------------|
| **Manager** | REST API for templates and reports CRUD. Receives report requests and publishes to queue. |
| **Worker** | Consumes messages from RabbitMQ, queries data sources, renders templates, and stores results. |
| **MongoDB** | Stores metadata for templates and reports. |
| **RabbitMQ** | Message queue for asynchronous report generation. |
| **Object Storage** | S3-compatible storage (AWS S3, SeaweedFS, MinIO) for templates and generated reports. |
| **Redis/Valkey** | Caching layer for data source schemas. |

## Quick Start

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Make

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/LerianStudio/reporter.git
   cd reporter
   ```

2. **Set up environment files:**
   ```bash
   make set-env
   ```

3. **Start all services:**
   ```bash
   make up
   ```

4. **Access the API:**
   - API: http://localhost:4005
   - Swagger UI: http://localhost:4005/swagger/index.html

## Configuration

### Environment Variables

Reporter uses environment variables for configuration. Copy `.env.example` files and adjust as needed:

```bash
# In each component directory
cp .env.example .env
```

Key configurations:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Manager API port | `4005` |
| `MONGO_HOST` | MongoDB hostname | `reporter-mongodb` |
| `RABBITMQ_HOST` | RabbitMQ hostname | `reporter-rabbitmq` |
| `LOG_LEVEL` | Log verbosity | `debug` |

### Object Storage (S3-compatible)

Reporter supports S3-compatible object storage for templates and generated reports:

| Variable | Description | Default |
|----------|-------------|---------|
| `OBJECT_STORAGE_ENDPOINT` | S3 endpoint URL | `http://reporter-seaweedfs:8333` |
| `OBJECT_STORAGE_REGION` | AWS region | `us-east-1` |
| `OBJECT_STORAGE_ACCESS_KEY_ID` | Access key ID | - |
| `OBJECT_STORAGE_SECRET_KEY` | Secret access key | - |
| `OBJECT_STORAGE_BUCKET` | Bucket name | `reporter-storage` |
| `OBJECT_STORAGE_USE_PATH_STYLE` | Use path-style URLs | `true` |
| `OBJECT_STORAGE_DISABLE_SSL` | Disable SSL | `true` |

**Supported providers:** AWS S3, SeaweedFS S3, MinIO, and other S3-compatible services.

## Data Sources

Reporter connects directly to external databases to fetch data for reports. Configure data sources using the `DATASOURCE_*` environment variables pattern:

```bash
# Pattern: DATASOURCE_<NAME>_<PROPERTY>

# PostgreSQL Example
DATASOURCE_MYDB_CONFIG_NAME=my_database
DATASOURCE_MYDB_HOST=postgres-host
DATASOURCE_MYDB_PORT=5432
DATASOURCE_MYDB_USER=username
DATASOURCE_MYDB_PASSWORD=password
DATASOURCE_MYDB_DATABASE=dbname
DATASOURCE_MYDB_TYPE=postgresql
DATASOURCE_MYDB_SSLMODE=disable
DATASOURCE_MYDB_SCHEMAS=public,sales,inventory  # Multi-schema support

# MongoDB Example
DATASOURCE_MYMONGO_CONFIG_NAME=my_mongo
DATASOURCE_MYMONGO_HOST=mongo-host
DATASOURCE_MYMONGO_PORT=27017
DATASOURCE_MYMONGO_USER=username
DATASOURCE_MYMONGO_PASSWORD=password
DATASOURCE_MYMONGO_DATABASE=dbname
DATASOURCE_MYMONGO_TYPE=mongodb
DATASOURCE_MYMONGO_SSL=false
```

### Supported Databases

| Database | Type Value | Notes |
|----------|------------|-------|
| PostgreSQL | `postgresql` | Supports SSL modes |
| MongoDB | `mongodb` | Supports replica sets |

### Features

- **Automatic schema discovery** - Reporter introspects database schemas
- **Multi-schema support** - Query tables across multiple PostgreSQL schemas (e.g., `public`, `sales`, `inventory`)
- **Connection pooling** - Configurable pool sizes for performance
- **Circuit breaker** - Automatic failover for unavailable data sources
- **Health checking** - Background monitoring of data source availability

## Templates

Templates use [Pongo2](https://github.com/flosch/pongo2) syntax (similar to Django/Jinja2).

### Example Template

```django
{% for row in my_database.users %}
Name: {{ row.name }}
Email: {{ row.email }}
{% endfor %}
```

### Accessing Data

Data is available in templates using the pattern:
```
{{ datasource_config_name.table_name }}
```

For multi-schema databases, use explicit schema syntax:
```
{{ datasource_config_name:schema_name.table_name }}
```

Example with multiple schemas:
```django
{# Access table from public schema #}
{% for account in midaz_onboarding:public.account %}
  Account: {{ account.id }} - {{ account.name }}
{% endfor %}

{# Access table from payment schema #}
{% for transfer in midaz_onboarding:payment.transfers %}
  Transfer: {{ transfer.id }} - {{ transfer.amount }}
{% endfor %}
```

### Output Formats

| Format | Extension | Use Case |
|--------|-----------|----------|
| HTML | `.html` | Web reports, dashboards |
| PDF | `.pdf` | Printable documents |
| CSV | `.csv` | Data export, spreadsheets |
| XML | `.xml` | Regulatory reports, integrations |
| TXT | `.txt` | Plain text reports |

### Custom Filters

Reporter extends Pongo2 with additional filters for report generation. See `pkg/pongo/filters.go` for available filters.

## API Reference

### Endpoints

#### Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/manager/v1/templates` | Create template |
| `GET` | `/manager/v1/templates` | List templates |
| `GET` | `/manager/v1/templates/{id}` | Get template by ID |
| `PATCH` | `/manager/v1/templates/{id}` | Update template |
| `DELETE` | `/manager/v1/templates/{id}` | Delete template |

#### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/manager/v1/reports` | Generate report |
| `GET` | `/manager/v1/reports` | List reports |
| `GET` | `/manager/v1/reports/{id}` | Get report by ID |

#### Data Sources

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/manager/v1/data-sources` | List configured data sources |
| `GET` | `/manager/v1/data-sources/{id}` | Get data source schema |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |

### Report Generation Message

Reports are generated asynchronously via RabbitMQ:

- **Exchange:** `reporter.generate-report.exchange`
- **Queue:** `reporter.generate-report.queue`
- **Routing Key:** `reporter.generate-report.key`

```json
{
  "templateId": "019538ee-deee-769c-8859-cbe84fce9af7",
  "reportId": "019615d3-c1f6-7b1d-add4-6912b76cc4f2",
  "outputFormat": "html",
  "mappedFields": {
    "my_database": {
      "users": ["id", "name", "email"],
      "orders": ["id", "total", "created_at"]
    }
  }
}
```

### Report Request with Filters

You can filter data when generating reports. The filter supports multi-schema references:

```json
{
  "templateId": "019538ee-deee-769c-8859-cbe84fce9af7",
  "filters": {
    "midaz_onboarding": {
      "organization": {
        "id": {
          "eq": ["019c10b7-073e-7056-a494-40f54a838404"]
        }
      },
      "public.account": {
        "organization_id": {
          "eq": ["019c10b7-073e-7056-a494-40f54a838404"]
        }
      }
    }
  }
}
```

#### Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals (supports multiple values as OR) | `{"eq": ["value1", "value2"]}` |
| `gt` | Greater than | `{"gt": [100]}` |
| `gte` | Greater than or equal | `{"gte": [100]}` |
| `lt` | Less than | `{"lt": [100]}` |
| `lte` | Less than or equal | `{"lte": [100]}` |
| `in` | In list | `{"in": ["a", "b", "c"]}` |
| `notIn` | Not in list | `{"notIn": ["x", "y"]}` |
| `between` | Between two values | `{"between": [10, 100]}` |

### Swagger Documentation

Full API documentation is available at:
```
http://localhost:4005/swagger/index.html
```

## Development

### Project Structure

```
reporter/
├── components/
│   ├── manager/          # REST API service
│   ├── worker/           # Report generation worker
│   └── infra/            # Infrastructure (Docker Compose)
├── pkg/                  # Shared packages
│   ├── pongo/            # Template engine extensions
│   ├── postgres/         # PostgreSQL adapter
│   ├── mongodb/          # MongoDB adapter
│   ├── seaweedfs/        # Legacy SeaweedFS HTTP adapter
│   └── storage/          # S3-compatible storage adapter
├── docs/                 # Documentation
└── tests/                # Test suites
```

### Commands

```bash
# Start all services
make up

# Stop all services
make down

# Run tests
make test-unit

# Run linters
make lint

# Generate Swagger docs
make generate-docs

# View logs
make logs
```

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# Property tests
make test-property

# Fuzzy tests
make test-fuzzy
```

## Contributing

We welcome contributions to Reporter. Here's how you can help:

### Getting Started

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linters (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Guidelines

- Follow Go best practices and idioms
- Write tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic
- Use meaningful commit messages

### Pull Request Process

1. Ensure all tests pass
2. Update the README if needed
3. Request review from maintainers
4. Address review feedback
5. Squash commits if requested

## Security

### Reporting Vulnerabilities

If you discover a security vulnerability, please report it privately:

1. **Do NOT** open a public issue
2. Email security concerns to the maintainers
3. Include detailed steps to reproduce
4. Allow time for the issue to be addressed before public disclosure

### Security Best Practices

When deploying Reporter:

- Use strong passwords for all services
- Enable SSL/TLS for database connections in production
- Restrict network access to internal services
- Rotate credentials regularly
- Keep dependencies updated

## Code of Conduct

### Our Standards

We are committed to providing a welcoming and inclusive environment. We expect all participants to:

- Be respectful and inclusive
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards others

### Unacceptable Behavior

- Harassment, discrimination, or personal attacks
- Trolling or inflammatory comments
- Publishing others' private information
- Other conduct inappropriate in a professional setting

### Enforcement

Project maintainers may remove, edit, or reject contributions that do not align with this Code of Conduct. Repeated violations may result in a ban from the project.

## Community & Support

- If you want to raise anything to the attention of the community, open a Discussion in our [GitHub](https://github.com/LerianStudio/reporter/discussions).
- Follow us on [Twitter](https://twitter.com/LerianStudio), [Instagram](https://www.instagram.com/lerian.studio/) and [Linkedin](https://www.linkedin.com/company/lerianstudio/) for the latest news and announcements.


## License

This project is licensed under the **Elastic License 2.0**.

You are free to use, modify, and distribute this software, but you may not provide it to third parties as a hosted or managed service.

See the [LICENSE](LICENSE) file for full details.

---

## References

- [Pongo2 Template Engine](https://github.com/flosch/pongo2)
- [SeaweedFS Documentation](https://github.com/seaweedfs/seaweedfs)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Reporter Guide](https://docs.lerian.studio/en/reporter/what-is-reporter) 
- [API Guide Reporter](https://docs.lerian.studio/en/reference/reporter/upload-template)
