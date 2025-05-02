package mongodb

import (
	"strings"
)

// ValidateFieldsInSchemaMongo validate if all fields exist on mongo DB collection
func ValidateFieldsInSchemaMongo(expectedFields []string, schema CollectionSchema) (missing []string) {
	columnSet := make(map[string]struct{}, len(schema.Fields))
	for _, col := range schema.Fields {
		columnSet[strings.ToLower(col.Name)] = struct{}{}
	}

	for _, field := range expectedFields {
		if _, exists := columnSet[strings.ToLower(field)]; !exists {
			missing = append(missing, field)
		}
	}

	return
}
