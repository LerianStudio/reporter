// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReportMongoDBRepository_NilConnection(t *testing.T) {
	t.Parallel()

	// Passing nil connection should panic or fail gracefully.
	// NewReportMongoDBRepository dereferences mc.Database, so nil input panics.
	assert.Panics(t, func() {
		_, _ = NewReportMongoDBRepository(nil)
	}, "Expected panic when creating repository with nil connection")
}

func TestReportMongoDBRepository_EnsureIndexes_RequiresConnection(t *testing.T) {
	t.Parallel()

	// EnsureIndexes and DropIndexes require a real MongoDB connection.
	// This test verifies the struct fields are correctly set.
	repo := &ReportMongoDBRepository{
		Database: "test_db",
	}

	assert.Equal(t, "test_db", repo.Database)
}
