package template_utils

import (
	"regexp"
	"strings"
)

func MappedFieldsOfTemplate(templateFile string) map[string]any {
	// Variable present on for loops of template
	variableMap := map[string][]string{}

	// Process for loops of template
	forRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s+in\s+([^\s%]+)\s*-?%}`)

	forMatches := forRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range forMatches {
		variable := match[1]
		path := CleanPath(match[2])
		variableMap[variable] = path
	}

	// Process {{ ... }}
	fieldRegex := regexp.MustCompile(`{{\s*([\w.\[\]_]+)\s*}}`)
	fieldMatches := fieldRegex.FindAllStringSubmatch(templateFile, -1)

	result := map[string]any{}

	for _, match := range fieldMatches {
		expr := match[1]
		parts := CleanPath(expr)

		if len(parts) < 2 {
			continue
		}

		if loopPath, ok := variableMap[parts[0]]; ok {
			// ex: t.id → loopPath = organization.transaction.account
			last := loopPath[len(loopPath)-1]
			insertField(result, append(loopPath[:len(loopPath)-1], last), parts[1])
		} else {
			// ex: organization.user.legal_name
			insertField(result, parts[:len(parts)-1], parts[len(parts)-1])
		}
	}

	return result
}

// insertField inserts a field into a nested map structure based on the given path
func insertField(m map[string]any, path []string, field string) {
	current := m

	for i, p := range path {
		if i == len(path)-1 {
			if _, ok := current[p]; !ok {
				current[p] = []string{}
			}

			current[p] = appendIfMissing(current[p].([]string), field)
		} else {
			if _, ok := current[p]; !ok {
				current[p] = map[string]any{}
			}

			current = current[p].(map[string]any)
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
