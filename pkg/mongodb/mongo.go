// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package mongodb

import (
	"strings"
)

// ValidateFieldsInSchemaMongo validate if all fields exist on mongo DB collection
func ValidateFieldsInSchemaMongo(expectedFields []string, schema CollectionSchema, countIfTableExist *int32) (missing []string) {
	columnSet := make(map[string]struct{}, len(schema.Fields))
	for _, col := range schema.Fields {
		columnSet[strings.ToLower(col.Name)] = struct{}{}
	}

	for _, field := range expectedFields {
		*countIfTableExist++

		if _, exists := columnSet[strings.ToLower(field)]; !exists {
			missing = append(missing, field)
		}
	}

	return
}
