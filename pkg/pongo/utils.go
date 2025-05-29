package pongo

import "strings"

// getNestedField retrieves a value from a nested map based on a dot-separated path. Returns the value and a boolean indicating success.
func getNestedField(m map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")

	var current any = m

	for _, part := range parts {
		if currMap, ok := current.(map[string]any); ok {
			current, ok = currMap[part]
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}
