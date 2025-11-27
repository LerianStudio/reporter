package storage

import "context"

// TemplateRepository defines the interface for template storage operations
// This interface is provider-agnostic and can be implemented by any storage backend
type TemplateRepository interface {
	// Get retrieves a template by its object name
	Get(ctx context.Context, objectName string) ([]byte, error)

	// Put stores a template with the given object name, content type, and data
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
}

// ReportRepository defines the interface for report storage operations
// This interface is provider-agnostic and can be implemented by any storage backend
type ReportRepository interface {
	// Put stores a report with the given object name, content type, data, and TTL
	Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error

	// Get retrieves a report by its object name
	Get(ctx context.Context, objectName string) ([]byte, error)
}
