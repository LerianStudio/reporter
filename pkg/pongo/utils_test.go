// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNestedField(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"email": "clara@example.com",
			},
		},
		"flat": "value",
	}

	t.Run("existing nested field", func(t *testing.T) {
		v, ok := getNestedField(data, "user.profile.email")
		assert.True(t, ok)
		assert.Equal(t, "clara@example.com", v)
	})

	t.Run("existing flat field", func(t *testing.T) {
		v, ok := getNestedField(data, "flat")
		assert.True(t, ok)
		assert.Equal(t, "value", v)
	})

	t.Run("nonexistent nested field", func(t *testing.T) {
		_, ok := getNestedField(data, "user.profile.phone")
		assert.False(t, ok)
	})

	t.Run("invalid intermediate path", func(t *testing.T) {
		_, ok := getNestedField(data, "user.profile.email.username")
		assert.False(t, ok)
	})

	t.Run("nonexistent top-level field", func(t *testing.T) {
		_, ok := getNestedField(data, "missing")
		assert.False(t, ok)
	})
}
