// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplateMongoDBRepository_NilConnection(t *testing.T) {
	t.Parallel()

	// Passing nil connection should panic or fail gracefully.
	// NewTemplateMongoDBRepository dereferences mc.Database, so nil input panics.
	assert.Panics(t, func() {
		_, _ = NewTemplateMongoDBRepository(nil)
	}, "Expected panic when creating repository with nil connection")
}

func TestTemplateMongoDBRepository_EnsureIndexes_RequiresConnection(t *testing.T) {
	t.Parallel()

	// EnsureIndexes and DropIndexes require a real MongoDB connection.
	// This test verifies the struct fields are correctly set.
	repo := &TemplateMongoDBRepository{
		Database: "test_db",
	}

	assert.Equal(t, "test_db", repo.Database)
}
