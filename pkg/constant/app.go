// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package constant

const ApplicationName = "reporter"

// ErrFileAccepted is the Fiber error message when no file is associated with the given form key.
const ErrFileAccepted = "there is no uploaded file associated with the given key"

// DefaultPasswordPlaceholder is the placeholder value that must be replaced before production use.
const DefaultPasswordPlaceholder = "CHANGE_ME"

// RedactPlaceholder is the replacement value for masked credentials in connection strings.
const RedactPlaceholder = "REDACTED"

// MaxAggregateBalanceCollectionSize is the maximum number of items allowed in a collection
// to prevent resource exhaustion attacks in the aggregate_balance tag.
const MaxAggregateBalanceCollectionSize = 100000
