// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package helpers

import (
	"os"
)

// AuthHeadersWithOrg returns default headers including Authorization and X-Request-Id.
// If TEST_AUTH_HEADER is set, its value is used for Authorization.
func AuthHeadersWithOrg(orgID string) map[string]string {
	hdr := map[string]string{
		"Content-Type":      "application/json",
		"X-Organization-Id": orgID,
	}
	if v := os.Getenv("TEST_AUTH_HEADER"); v != "" {
		hdr["Authorization"] = v
	} else {
		hdr["Authorization"] = "Bearer test"
	}

	return hdr
}
