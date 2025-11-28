# Reporter

## Overview

The Reporter is a service designed to manage and generate customizable reports using predefined templates.

## Quick Start

1. **Clone the Repository:**
    ```bash
    git clone https://github.com/LerianStudio/reporter.git
    ```

2. **Install Dependencies:**
    ```bash
    cd reporter
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

- Exchange: `reporter.generate-report.exchange`
- Queue: `reporter.generate-report.queue`
- Key: `reporter.generate-report.key`

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

## File Storage Configuration

The Reporter supports two storage providers for templates and generated reports: **SeaweedFS** (default) and **AWS S3**.

### Storage Provider Selection

Set the storage provider using the environment variable:

```bash
# Use SeaweedFS (default)
STORAGE_PROVIDER=seaweedfs

# Use AWS S3
STORAGE_PROVIDER=s3
```

### SeaweedFS Storage

We use SeaweedFS (Filer + Volume + Master) to store both template files and generated reports. Access is done via Filer HTTP API.

#### Configuration

Configure the following environment variables:

- `STORAGE_PROVIDER`: Set to `seaweedfs` or leave empty (default)
- `SEAWEEDFS_HOST`: SeaweedFS Filer hostname (default: `reporter-seaweedfs-filer`)
- `SEAWEEDFS_FILER_PORT`: SeaweedFS Filer port (default: `8888`)
- `SEAWEEDFS_TTL`: Time-to-live for stored files (default: `6M`)

#### Accessing SeaweedFS

**Development**: Access the Filer web interface directly at `http://localhost:8888/`

**Production**: Filer should be accessible only from Manager/Worker services within the private network.

### AWS S3 Storage

Alternative storage using Amazon S3 for scalable cloud storage of templates and reports.

#### Configuration

Configure the following environment variables:

- `STORAGE_PROVIDER`: Set to `s3`
- `S3_REGION`: AWS region (e.g., `us-east-1`, `us-west-2`)
- `S3_BUCKET`: S3 bucket name for storing files
- `S3_ACCESS_KEY_ID`: AWS access key ID
- `S3_SECRET_ACCESS_KEY`: AWS secret access key
- `S3_ENDPOINT`: Custom S3 endpoint (optional, for S3-compatible services)
- `S3_FORCE_PATH_STYLE`: Force path-style URLs (optional, default: `false`)

#### Example S3 Configuration

```bash
STORAGE_PROVIDER=s3
S3_REGION=us-east-1
S3_BUCKET=my-reporter-bucket
S3_ACCESS_KEY_ID=AKIA...
S3_SECRET_ACCESS_KEY=your-secret-key
```

#### S3 Bucket Structure

The Reporter organizes files in S3 with the following structure:

```text
my-reporter-bucket/
├── templates/
│   ├── template-uuid-1.tpl
│   └── template-uuid-2.tpl
└── reports/
    ├── report-uuid-1.pdf
    ├── report-uuid-2.html
    └── report-uuid-3.csv
```

#### S3 Permissions

Ensure your AWS credentials have the following S3 permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-reporter-bucket",
        "arn:aws:s3:::my-reporter-bucket/*"
      ]
    }
  ]
}
```

### Storage Provider Troubleshooting

#### Common Issues

#### 1. Storage Provider Not Recognized

The `STORAGE_PROVIDER` value is case-sensitive and must be lowercase:

```bash
# ✅ Correct
STORAGE_PROVIDER=s3
STORAGE_PROVIDER=seaweedfs

# ❌ Incorrect
STORAGE_PROVIDER=S3
STORAGE_PROVIDER=SeaweedFS
```

#### 2. S3 Connection Issues

- Verify AWS credentials have correct permissions
- Check S3 bucket exists and is accessible
- Ensure region matches bucket location
- For custom endpoints, verify `S3_ENDPOINT` and `S3_FORCE_PATH_STYLE` settings

#### 3. SeaweedFS Connection Issues

- Verify SeaweedFS services are running
- Check `SEAWEEDFS_HOST` and `SEAWEEDFS_FILER_PORT` are correct
- Ensure Filer is accessible from Manager/Worker containers

#### Switching Storage Providers

To switch between storage providers:

1. Update environment variables
2. Restart Manager and Worker services
3. Existing files remain in the previous storage (migration not automatic)

```bash
# Switch to S3
export STORAGE_PROVIDER=s3
export S3_REGION=us-east-1
export S3_BUCKET=my-bucket
# ... other S3 config

# Restart services
make down && make up
```

## Swagger Documentation

The Plugin includes Swagger documentation that helps in visualizing and interacting with the API endpoints. You can access the documentation by running the project and navigating to `http://localhost:4005/swagger/index.html`.

## References

- https://github.com/flosch/pongo2/blob/master/template_tests/filters.tpl