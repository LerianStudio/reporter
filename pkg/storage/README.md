# Object Storage Package

This package provides a unified abstraction layer for object storage operations in Reporter, supporting multiple storage backends.

## Supported Backends

- **SeaweedFS HTTP API** - Fast distributed file storage with native TTL support
- **AWS S3** - Amazon's object storage service
- **MinIO** - High-performance S3-compatible object storage
- **SeaweedFS S3 API** - SeaweedFS with S3 compatibility

## Quick Start

### Using SeaweedFS HTTP (Default)

```go
import "github.com/LerianStudio/reporter/v4/pkg/storage"

config := storage.Config{
    Type:              storage.StorageTypeSeaweedFS,
    Bucket:            "templates",
    SeaweedFSEndpoint: "http://localhost:8888",
}

client, err := storage.NewStorageClient(ctx, config)
if err != nil {
    log.Fatal(err)
}

// Upload file
key, err := client.Upload(ctx, "template.tpl", reader, "text/plain")

// Download file
data, err := client.Download(ctx, "template.tpl")

// Upload with TTL (SeaweedFS only)
key, err := client.UploadWithTTL(ctx, "report.pdf", reader, "application/pdf", "7d")
```

### Using S3-Compatible Storage

```go
config := storage.Config{
    Type:              storage.StorageTypeS3,
    Bucket:            "reports",
    S3Endpoint:        "http://localhost:9000", // Optional for AWS S3
    S3Region:          "us-east-1",
    S3AccessKeyID:     "minioadmin",
    S3SecretAccessKey: "minioadmin",
    S3UsePathStyle:    true, // Required for MinIO
    S3DisableSSL:      false,
}

client, err := storage.NewStorageClient(ctx, config)
```

## Interface

```go
type ObjectStorage interface {
    // Upload stores content from a reader
    Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
    
    // UploadWithTTL stores content with expiration (SeaweedFS only)
    UploadWithTTL(ctx context.Context, key string, reader io.Reader, contentType string, ttl string) (string, error)
    
    // Download retrieves content
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    
    // Delete removes an object
    Delete(ctx context.Context, key string) error
    
    // Exists checks if an object exists
    Exists(ctx context.Context, key string) (bool, error)
    
    // GeneratePresignedURL creates time-limited download URL
    GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
```

## Environment Configuration

### SeaweedFS Mode

```bash
STORAGE_TYPE=seaweedfs
SEAWEEDFS_HOST=localhost
SEAWEEDFS_FILER_PORT=8888
STORAGE_TEMPLATE_BUCKET=templates
STORAGE_REPORT_BUCKET=reports
```

### S3 Mode

```bash
STORAGE_TYPE=s3
S3_ENDPOINT=http://localhost:9000  # Optional for AWS S3
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=your-key
S3_SECRET_ACCESS_KEY=your-secret
S3_USE_PATH_STYLE=true  # true for MinIO/SeaweedFS, false for AWS
S3_DISABLE_SSL=false
STORAGE_TEMPLATE_BUCKET=templates
STORAGE_REPORT_BUCKET=reports
```

## TTL Support

### SeaweedFS HTTP Mode
Full TTL support via upload API:
- Format: `3m`, `4h`, `5d`, `6w`, `7M`, `8y`
- Applied per-object during upload
- Automatic deletion after expiration

### S3 Mode
- TTL parameter is ignored (logged as warning)
- Use bucket lifecycle policies instead
- Configure at bucket level, not per-object

**Example MinIO lifecycle:**
```bash
mc ilm add minio/reports --expiry-days 180
```

## Architecture

```
pkg/storage/
├── ports.go              # ObjectStorage interface
├── s3_client.go          # S3-compatible implementation (AWS SDK v2)
├── seaweedfs_adapter.go  # SeaweedFS HTTP wrapper
├── config.go             # Factory pattern for creating clients
└── s3_client_test.go     # Unit tests
```

## Testing

```bash
# Run unit tests
go test ./pkg/storage/... -v

# Run integration tests (requires storage service)
go test ./pkg/storage/... -v -short=false
```

## Migration Guide

See [STORAGE_MIGRATION.md](../../docs/STORAGE_MIGRATION.md) for detailed migration instructions.

## Performance Considerations

- **SeaweedFS HTTP**: Lowest latency, highest throughput for local deployments
- **MinIO**: Excellent performance, recommended for on-premise S3 compatibility
- **AWS S3**: Best for cloud-native applications, higher latency but unlimited scalability

## Error Handling

All methods return standard Go errors with descriptive messages:
- `ErrBucketRequired` - Missing bucket name
- `ErrKeyRequired` - Missing object key
- `ErrObjectNotFound` - Object doesn't exist
- `ErrTTLNotSupported` - TTL used with S3 backend (warning only)

## Examples

### Complete Upload/Download Cycle

```go
ctx := context.Background()

// Create client
config := storage.Config{
    Type:              storage.StorageTypeS3,
    Bucket:            "templates",
    S3Endpoint:        "http://localhost:9000",
    S3Region:          "us-east-1",
    S3AccessKeyID:     "minioadmin",
    S3SecretAccessKey: "minioadmin",
    S3UsePathStyle:    true,
}

client, err := storage.NewStorageClient(ctx, config)
if err != nil {
    log.Fatal(err)
}

// Upload
content := []byte("Hello, Storage!")
key, err := client.Upload(ctx, "test.txt", bytes.NewReader(content), "text/plain")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Uploaded to: %s\n", key)

// Check existence
exists, err := client.Exists(ctx, key)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File exists: %v\n", exists)

// Download
reader, err := client.Download(ctx, key)
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

data, err := io.ReadAll(reader)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Downloaded: %s\n", string(data))

// Generate presigned URL (valid for 1 hour)
url, err := client.GeneratePresignedURL(ctx, key, 1*time.Hour)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Presigned URL: %s\n", url)

// Delete
err = client.Delete(ctx, key)
if err != nil {
    log.Fatal(err)
}
fmt.Println("File deleted")
```

## License

Copyright © 2024 Lerian Studio
