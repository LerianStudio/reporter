package mongodb

import (
	"sort"
	"strings"
)

// FilterNestedFields removes nested fields when their parent field is already in the list.
// MongoDB projections cannot have both a parent field and its nested fields simultaneously.
// For example, if "related_parties" is in the list, "related_parties.document" should be removed.
// This prevents MongoDB projection conflicts and Pongo2 path collision errors.
func FilterNestedFields(fields []string) []string {
	if len(fields) == 0 {
		return fields
	}

	// Build a set of parent fields (fields without dots)
	parentFields := make(map[string]struct{})

	for _, field := range fields {
		if !strings.Contains(field, ".") {
			parentFields[field] = struct{}{}
		}
	}

	// Filter out nested fields whose parent is already in the list
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.Contains(field, ".") {
			parent := strings.Split(field, ".")[0]
			if _, parentExists := parentFields[parent]; parentExists {
				// Skip this nested field, parent already includes it
				continue
			}
		}

		result = append(result, field)
	}

	// Sort for consistent ordering
	sort.Strings(result)

	return result
}

// ValidateFieldsInSchemaMongo validate if all fields exist on mongo DB collection
// For nested fields (e.g., "related_parties.document"), it validates that the parent field exists
// since MongoDB is schemaless and nested structures cannot be fully validated at schema level
func ValidateFieldsInSchemaMongo(expectedFields []string, schema CollectionSchema, countIfTableExist *int32) (missing []string) {
	columnSet := make(map[string]struct{}, len(schema.Fields))
	for _, col := range schema.Fields {
		columnSet[strings.ToLower(col.Name)] = struct{}{}
	}

	for _, field := range expectedFields {
		*countIfTableExist++

		fieldLower := strings.ToLower(field)

		// Check if field exists directly
		if _, exists := columnSet[fieldLower]; exists {
			continue
		}

		// For nested fields (e.g., "related_parties.document"), check if parent field exists
		// MongoDB is schemaless, so if the parent field exists, we trust the nested structure
		if strings.Contains(field, ".") {
			parentField := strings.ToLower(strings.Split(field, ".")[0])
			if _, parentExists := columnSet[parentField]; parentExists {
				continue
			}
		}

		missing = append(missing, field)
	}

	return
}
