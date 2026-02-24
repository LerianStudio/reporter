//go:build property

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/stretchr/testify/require"
)

// TestProperty_RateLimitConfig_RejectsZeroValues verifies that Config.Validate
// always rejects configurations where any rate limit field is zero.
func TestProperty_RateLimitConfig_RejectsZeroValues(t *testing.T) {
	t.Parallel()

	property := func(selector uint8) bool {
		cfg := validManagerConfig()

		// Set one field to zero based on selector
		switch selector % 3 {
		case 0:
			cfg.RateLimitGlobal = 0
		case 1:
			cfg.RateLimitExport = 0
		case 2:
			cfg.RateLimitDispatch = 0
		}

		err := cfg.Validate()

		// Must always return an error
		return err != nil
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: zero rate limit accepted by validation")
}

// TestProperty_RateLimitConfig_RejectsNegativeValues verifies that
// Config.Validate always rejects configurations with negative rate limits.
// The error message must reference all three RATE_LIMIT fields.
func TestProperty_RateLimitConfig_RejectsNegativeValues(t *testing.T) {
	t.Parallel()

	property := func(value int32) bool {
		// Only test negative values
		if value >= 0 {
			return true
		}

		cfg := validManagerConfig()
		cfg.RateLimitGlobal = int(value)
		cfg.RateLimitExport = int(value)
		cfg.RateLimitDispatch = int(value)

		err := cfg.Validate()

		// Must always return an error for negative values
		if err == nil {
			return false
		}

		// Error message must reference all three rate limit fields
		errMsg := err.Error()

		return strings.Contains(errMsg, "RATE_LIMIT_GLOBAL") &&
			strings.Contains(errMsg, "RATE_LIMIT_EXPORT") &&
			strings.Contains(errMsg, "RATE_LIMIT_DISPATCH")
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: negative rate limits accepted by validation")
}

// TestProperty_RateLimitConfig_AcceptsAllPositiveValues verifies that
// Config.Validate never produces rate-limit-specific errors when all rate
// limit fields are positive integers.
func TestProperty_RateLimitConfig_AcceptsAllPositiveValues(t *testing.T) {
	t.Parallel()

	property := func(globalMax, exportMax, dispatchMax uint16) bool {
		// Clamp to valid upper bounds so the property tests only positive
		// values within the allowed range.
		cfg := validManagerConfig()
		cfg.RateLimitGlobal = int(globalMax%uint16(constant.RateLimitMaxGlobal)) + 1
		cfg.RateLimitExport = int(exportMax%uint16(constant.RateLimitMaxExport)) + 1
		cfg.RateLimitDispatch = int(dispatchMax%uint16(constant.RateLimitMaxDispatch)) + 1

		err := cfg.Validate()
		// If there is an error, it must NOT be about rate limits
		if err != nil {
			errMsg := err.Error()
			return !strings.Contains(errMsg, "RATE_LIMIT")
		}

		return true
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: positive rate limits rejected by validation")
}

// TestProperty_RateLimitConfig_IndependentFieldValidation verifies that each
// rate limit field is validated independently. Setting one field to an invalid
// value must produce an error mentioning that specific field.
func TestProperty_RateLimitConfig_IndependentFieldValidation(t *testing.T) {
	t.Parallel()

	property := func(invalidValue int32, fieldIndex uint8) bool {
		// Only test non-positive values
		if invalidValue > 0 {
			return true
		}

		cfg := validManagerConfig()

		fieldMap := map[uint8]string{
			0: "RATE_LIMIT_GLOBAL",
			1: "RATE_LIMIT_EXPORT",
			2: "RATE_LIMIT_DISPATCH",
		}

		idx := fieldIndex % 3

		switch idx {
		case 0:
			cfg.RateLimitGlobal = int(invalidValue)
		case 1:
			cfg.RateLimitExport = int(invalidValue)
		case 2:
			cfg.RateLimitDispatch = int(invalidValue)
		}

		err := cfg.Validate()
		if err == nil {
			return false
		}

		errMsg := err.Error()

		// The invalid field must be mentioned in the error
		return strings.Contains(errMsg, fieldMap[idx])
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: fields not validated independently")
}

// TestProperty_RateLimitConfig_ValidationNeverPanics verifies that
// Config.Validate never panics for any combination of rate limit values,
// including extreme values like MaxInt and MinInt.
func TestProperty_RateLimitConfig_ValidationNeverPanics(t *testing.T) {
	t.Parallel()

	property := func(globalMax, exportMax, dispatchMax int) bool {
		cfg := validManagerConfig()
		cfg.RateLimitGlobal = globalMax
		cfg.RateLimitExport = exportMax
		cfg.RateLimitDispatch = dispatchMax

		// Must not panic -- either returns nil or an error
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate panicked for config{global=%d, export=%d, dispatch=%d}: %v",
					globalMax, exportMax, dispatchMax, r)
			}
		}()

		_ = cfg.Validate()

		return true
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	require.NoError(t, err, "Property violated: Validate panicked")
}
