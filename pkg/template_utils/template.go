package template_utils

import (
	"strings"
)

// InsertField inserts a field into a nested map structure based on the given path
func InsertField(m map[string]interface{}, path []string, field string) {
	current := m
	for i, p := range path {
		if i == len(path)-1 {
			if _, ok := current[p]; !ok {
				current[p] = []string{}
			}
			current[p] = appendIfMissing(current[p].([]string), field)
		} else {
			if _, ok := current[p]; !ok {
				current[p] = map[string]interface{}{}
			}
			current = current[p].(map[string]interface{})
		}
	}
}

// appendIfMissing appends a value to a slice only if it doesn't already exist
func appendIfMissing(slice []string, val string) []string {
	for _, item := range slice {
		if item == val {
			return slice
		}
	}
	return append(slice, val)
}

// CleanPath fix the fields if you have array annotation
func CleanPath(path string) []string {
	parts := strings.Split(path, ".")
	for i, p := range parts {
		parts[i] = strings.Split(p, "[")[0]
	}
	return parts
}
