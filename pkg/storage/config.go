package storage

// Config holds configuration for all storage providers
type Config struct {
	// Storage provider selection
	Provider string // "s3", "seaweedfs", or "" (defaults to seaweedfs)

	// S3 Configuration
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3Endpoint        string // Optional: for MinIO, LocalStack, etc.
	S3ForcePathStyle  bool   // Optional: for MinIO, LocalStack, etc.
	S3TemplateBucket  string // Optional: separate bucket for templates
	S3ReportBucket    string // Optional: separate bucket for reports

	// SeaweedFS Configuration
	SeaweedFSHost      string
	SeaweedFSFilerPort string
}
