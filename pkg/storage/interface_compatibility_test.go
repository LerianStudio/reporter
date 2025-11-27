package storage

import (
	"context"
	"testing"

	reportS3 "github.com/LerianStudio/reporter/v4/pkg/storage/s3/report"
	templateS3 "github.com/LerianStudio/reporter/v4/pkg/storage/s3/template"
	reportSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs/template"
)

// TestInterfaceCompatibility verifies that our generic interfaces
// are compatible with both S3 and SeaweedFS implementations
func TestTemplateRepositoryCompatibility(t *testing.T) {
	// Test that SeaweedFS implementation satisfies our generic interface
	var _ TemplateRepository = (*templateSeaweedFS.SimpleRepository)(nil)

	// Test that S3 implementation satisfies our generic interface
	var _ TemplateRepository = (*templateS3.SimpleRepository)(nil)
}

func TestReportRepositoryCompatibility(t *testing.T) {
	// Test that SeaweedFS implementation satisfies our generic interface
	var _ ReportRepository = (*reportSeaweedFS.SimpleRepository)(nil)

	// Test that S3 implementation satisfies our generic interface
	var _ ReportRepository = (*reportS3.SimpleRepository)(nil)
}

// TestInterfaceMethodSignatures ensures method signatures are identical
func TestTemplateRepositoryMethodSignatures(t *testing.T) {
	// Create a function that accepts our generic interface
	testFunc := func(repo TemplateRepository) {
		ctx := context.Background()

		// These calls should compile without issues
		_, _ = repo.Get(ctx, "test")
		_ = repo.Put(ctx, "test", "text/plain", []byte("data"))
	}

	// This test passes if it compiles - no runtime assertions needed
	_ = testFunc
}

func TestReportRepositoryMethodSignatures(t *testing.T) {
	// Create a function that accepts our generic interface
	testFunc := func(repo ReportRepository) {
		ctx := context.Background()

		// These calls should compile without issues
		_, _ = repo.Get(ctx, "test")
		_ = repo.Put(ctx, "test", "application/pdf", []byte("data"), "1h")
	}

	// This test passes if it compiles - no runtime assertions needed
	_ = testFunc
}
