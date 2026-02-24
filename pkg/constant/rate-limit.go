// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package constant

import "time"

// Rate Limiting Defaults
const (
	// RateLimitDefaultEnabled indicates whether rate limiting is enabled by default.
	RateLimitDefaultEnabled = true

	// RateLimitDefaultGlobalMax is the default maximum number of requests per
	// window for the global (catch-all) rate limit tier.
	RateLimitDefaultGlobalMax = 100

	// RateLimitDefaultExportMax is the default maximum number of requests per
	// window for the export (download) rate limit tier.
	RateLimitDefaultExportMax = 10

	// RateLimitDefaultDispatchMax is the default maximum number of requests per
	// window for the dispatch (create/write) rate limit tier.
	RateLimitDefaultDispatchMax = 50

	// RateLimitDefaultWindow is the default sliding window duration for all
	// rate limit tiers.
	RateLimitDefaultWindow = 60 * time.Second
)

// Rate Limiting Upper Bounds
const (
	// RateLimitMaxGlobal is the maximum allowed value for the global rate limit
	// tier. Values above this threshold indicate misconfiguration.
	RateLimitMaxGlobal = 10000

	// RateLimitMaxExport is the maximum allowed value for the export rate limit
	// tier. Export operations are resource-intensive, so this bound is lower.
	RateLimitMaxExport = 1000

	// RateLimitMaxDispatch is the maximum allowed value for the dispatch rate limit
	// tier. Write operations need a moderate upper bound to prevent abuse.
	RateLimitMaxDispatch = 5000
)
