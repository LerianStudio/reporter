//go:build fuzz

// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"strings"
	"testing"
)

// FuzzRateLimitConfig_Validate tests that the Config.Validate method never
// panics when given arbitrary integer values for rate limit configuration
// fields. The fuzzer generates random integers for GlobalMax, ExportMax,
// and DispatchMax. Validate must return either nil or an error -- never panic.
func FuzzRateLimitConfig_Validate(f *testing.F) {
	// Seed corpus: 7 entries across all required categories
	// Category 1 (Valid): default production values
	f.Add(100, 10, 50)
	// Category 2 (Empty/zero): all zeros should fail validation
	f.Add(0, 0, 0)
	// Category 3 (Negative): all negative should fail validation
	f.Add(-1, -1, -1)
	// Category 4 (Boundary): max int32
	f.Add(2147483647, 2147483647, 2147483647)
	// Category 5 (Boundary): min int32
	f.Add(-2147483648, -2147483648, -2147483648)
	// Category 6 (Mixed): one valid, others invalid
	f.Add(100, 0, -5)
	// Category 7 (Boundary): minimum positive
	f.Add(1, 1, 1)

	f.Fuzz(func(t *testing.T, globalMax, exportMax, dispatchMax int) {
		// Create a minimally valid config (required fields populated) so
		// that only rate limit validation is exercised.
		cfg := validManagerConfig()
		cfg.RateLimitGlobal = globalMax
		cfg.RateLimitExport = exportMax
		cfg.RateLimitDispatch = dispatchMax

		// Must not panic -- either returns nil or an error
		err := cfg.Validate()

		// Verify consistency: if any value is <= 0, validation must fail
		anyNonPositive := globalMax <= 0 || exportMax <= 0 || dispatchMax <= 0
		if anyNonPositive && err == nil {
			t.Fatalf("expected validation error for non-positive rate limits: global=%d export=%d dispatch=%d",
				globalMax, exportMax, dispatchMax)
		}

		// If all values are positive, rate limit validation should not produce
		// rate-limit-specific errors.
		if !anyNonPositive && err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "RATE_LIMIT") {
				t.Fatalf("unexpected rate limit validation error for positive values: global=%d export=%d dispatch=%d err=%v",
					globalMax, exportMax, dispatchMax, err)
			}
		}
	})
}

// FuzzRateLimitConfig_IndividualField tests that each rate limit field is
// independently validated when given arbitrary integer values. The fuzzer
// generates a random integer that is applied to each field one at a time.
func FuzzRateLimitConfig_IndividualField(f *testing.F) {
	// Seed corpus: 7 entries across all required categories
	// Category 1 (Valid): standard value
	f.Add(100)
	// Category 2 (Zero): boundary at zero
	f.Add(0)
	// Category 3 (Negative): simple negative
	f.Add(-1)
	// Category 4 (Boundary): large positive
	f.Add(1000000)
	// Category 5 (Boundary): minimum positive
	f.Add(1)
	// Category 6 (Boundary): large negative
	f.Add(-1000000)
	// Category 7 (Boundary): min int
	f.Add(-2147483648)

	f.Fuzz(func(t *testing.T, value int) {
		// Test each field independently
		fieldEnvMap := map[string]string{
			"global":   "RATE_LIMIT_GLOBAL",
			"export":   "RATE_LIMIT_EXPORT",
			"dispatch": "RATE_LIMIT_DISPATCH",
		}

		for field, envName := range fieldEnvMap {
			cfg := validManagerConfig()

			switch field {
			case "global":
				cfg.RateLimitGlobal = value
			case "export":
				cfg.RateLimitExport = value
			case "dispatch":
				cfg.RateLimitDispatch = value
			}

			// Must not panic
			err := cfg.Validate()

			if value <= 0 && err == nil {
				t.Fatalf("expected validation error for %s=%d", field, value)
			}

			if value > 0 && err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, envName) {
					t.Fatalf("unexpected rate limit error for valid %s=%d: %v", field, value, err)
				}
			}
		}
	})
}
