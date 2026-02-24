// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"net/url"

	"github.com/LerianStudio/reporter/pkg/constant"
)

// RedactConnectionString masks credentials in a connection URI.
// It replaces the username and password with "REDACTED" to prevent accidental
// credential leakage in logs. Returns "[invalid-uri]" if parsing fails.
func RedactConnectionString(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return "[invalid-uri]"
	}

	if u.User != nil {
		u.User = url.UserPassword(constant.RedactPlaceholder, constant.RedactPlaceholder)
	}

	return u.String()
}
