package template_utils

import (
	"regexp"
	"strconv"
	"strings"
)

// GetMimeType return a MIME type correctly based with outputFormat
func GetMimeType(outputFormat string) string {
	switch strings.ToLower(outputFormat) {
	case "xml":
		return "application/xml"
	case "html":
		return "text/html"
	case "csv":
		return "text/csv"
	case "txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// MappedFieldsOfTemplate analyzes a template file and returns a nested map of variable paths and their associated fields.
func MappedFieldsOfTemplate(templateFile string) map[string]map[string][]string {
	variableMap := regexBlockForOnPlaceholder(templateFile)
	resultRegex := regexBlockWithOnPlaceholder(variableMap, templateFile)

	// Regex for fields {{ ... }}
	fieldRegex := regexp.MustCompile(`{{\s*(.*?)\s*}}`)

	fieldMatches := fieldRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range fieldMatches {
		expr := match[1]
		fieldPaths := extractFieldsFromExpression(expr)

		for _, fieldExpr := range fieldPaths {
			parts := CleanPath(fieldExpr)
			if len(parts) < 2 {
				continue
			}

			if loopPath, ok := variableMap[parts[0]]; ok {
				fullPath := append([]string{}, loopPath...)
				insertField(resultRegex, fullPath, parts[1])
			} else {
				insertField(resultRegex, parts[:len(parts)-1], parts[len(parts)-1])
			}
		}
	}

	regexBlockIfOnPlaceholder(templateFile, resultRegex, variableMap)

	return normalizeStructure(resultRegex)
}

// regexBlockIfOnPlaceholder parses a template file to process "if" blocks and updates a nested map with extracted field mappings.
// It identifies fields used in conditional statements, cleans their paths, and inserts them into the resultRegex map structure.
func regexBlockIfOnPlaceholder(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	ifRegex := regexp.MustCompile(`{%-?\s*if\s+(.*?)\s*-?%}`)

	ifMatches := ifRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range ifMatches {
		expr := match[1]
		fieldPaths := extractIfFromExpression(expr)

		for _, fieldExpr := range fieldPaths {
			parts := CleanPath(fieldExpr)
			if len(parts) < 2 {
				continue
			}

			if loopPath, ok := variableMap[parts[0]]; ok {
				insertField(resultRegex, loopPath, parts[1])
			} else {
				insertField(resultRegex, parts[:len(parts)-1], parts[len(parts)-1])
			}
		}
	}
}

// regexBlockForOnPlaceholder parses a template file to extract variable mappings defined in for-loop blocks.
// It returns a map where keys are variables from the for loop, and values are their corresponding path segments.
func regexBlockForOnPlaceholder(templateFile string) map[string][]string {
	variableMap := map[string][]string{}

	// Regex for block for
	forRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s+in\s+([^\s%]+)\s*-?%}`)

	forMatches := forRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range forMatches {
		variable := match[1]
		path := CleanPath(match[2])

		if len(path) >= 2 {
			variableMap[variable] = []string{path[0], path[1]}
		} else {
			variableMap[variable] = path
		}
	}

	return variableMap
}

// regexBlockWithOnPlaceholder parses a template file to process "with" statements and updates `variableMap` with mapped variables.
// The function extracts filters, processes their arguments, and organizes nested data into a structured map for use.
// It cleans paths, maps targets to their corresponding variables, and inserts additional parameters where applicable.
func regexBlockWithOnPlaceholder(variableMap map[string][]string, templateFile string) map[string]any {
	result := map[string]any{}
	withRegex := regexp.MustCompile(`{%-?\s*with\s+(\w+)\s*=\s*filter\(\s*([^)]+)\s*\)[^\%]+`)

	withMatches := withRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range withMatches {
		assignedVar := match[1]
		args := match[2]
		argParts := strings.Split(args, ",")

		if len(argParts) > 0 {
			filterTarget := strings.TrimSpace(argParts[0])
			path := CleanPath(filterTarget)

			if len(path) >= 2 {
				variableMap[assignedVar] = []string{path[0], path[1]}

				for _, param := range argParts[1:] {
					param = strings.TrimSpace(param)
					cleanParam := strings.Trim(param, `"' `)

					if cleanParam == "" {
						continue
					}

					paramPath := CleanPath(cleanParam)

					if len(paramPath) < 2 {
						insertField(result, path, cleanParam)
						continue
					}

					if loopPath, ok := variableMap[paramPath[0]]; ok {
						insertField(result, loopPath, paramPath[1])
					} else {
						insertField(result, paramPath[:len(paramPath)-1], paramPath[len(paramPath)-1])
					}
				}
			}
		}
	}

	return result
}

// extractIfFromExpression extracts object.field patterns from a string expression,
// skipping numerical indices like `.0` in midaz_transaction.transaction.0.id.
func extractIfFromExpression(expr string) []string {
	// Regex: matches paths like `foo.bar.baz`, optionally with `.0` etc., but filters them after
	identifierRegex := regexp.MustCompile(`\b(?:[a-zA-Z_]\w*)(?:\.(?:[a-zA-Z_]\w*|\d+))+\b`)
	matches := identifierRegex.FindAllString(expr, -1)

	var results []string
	for _, match := range matches {
		parts := strings.Split(match, ".")
		var cleaned []string

		for _, part := range parts {
			// Skip purely numeric parts like "0"
			if _, err := strconv.Atoi(part); err == nil {
				continue
			}
			cleaned = append(cleaned, part)
		}

		if len(cleaned) > 1 {
			results = append(results, strings.Join(cleaned, "."))
		}
	}

	return results
}

// normalizeStructure convert input to a type pattern of mapped fields map[string]map[string][]string
func normalizeStructure(input map[string]any) map[string]map[string][]string {
	result := make(map[string]map[string][]string)

	for topKey, topVal := range input {
		section := make(map[string][]string)

		if m, ok := topVal.(map[string]any); ok {
			for subKey, subVal := range m {
				switch v := subVal.(type) {
				case []any:
					for _, item := range v {
						switch itemVal := item.(type) {
						case string:
							section[subKey] = append(section[subKey], itemVal)
						case map[string]any:
							for nestedKey := range itemVal {
								section[subKey] = append(section[subKey], nestedKey)
							}
						}
					}
				case map[string]any: // Caso especial como em "transaction": { "metadata": [...] }
					section[subKey] = append(section[subKey], getMapKeys(v)...)
				}
			}
		}

		result[topKey] = section
	}

	return result
}

// getMapKeys retrieves all keys from a given map and returns them as a slice of strings.
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// HydrateMappedFields Adjust the map[string]any if the value be an object, (ex: [ { "address": ["city"] } ])
func HydrateMappedFields(m map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			// If is a map with only one key, and value of the key is []any of strings, can collapse
			if len(val) == 1 && isSingleKeyWithStringArray(val) {
				for key := range val {
					result[k] = []string{key}
				}
			} else {
				// Uses recursion normally
				result[k] = HydrateMappedFields(val)
			}

		case []any:
			// Hydrate arrays with recursion
			result[k] = hydrateArray(val)

		default:
			result[k] = val
		}
	}

	return result
}

// hydrateArray Adjust an array of any if you have a metadata like this, (ex: "transaction": { "metadata": [ "message" ] })
func hydrateArray(arr []any) []any {
	var result []any

	for _, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			// Apply logic of collapse only if map has a key of []string
			if len(v) == 1 && isSingleKeyWithStringArray(v) {
				for key := range v {
					result = append(result, key)
				}
			} else {
				result = append(result, HydrateMappedFields(v))
			}
		default:
			result = append(result, v)
		}
	}

	return result
}

// isSingleKeyWithStringArray Validate if map is a string[] or any[]
func isSingleKeyWithStringArray(m map[string]any) bool {
	for _, v := range m {
		arr, ok := v.([]any)
		if !ok {
			return false
		}

		for _, item := range arr {
			if _, ok := item.(string); !ok {
				return false
			}
		}
	}

	return true
}

// extractFieldsFromExpression Get all valid object.property fields from expression
func extractFieldsFromExpression(expr string) []string {
	fields := []string{}

	parts := strings.Split(expr, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		subParts := strings.Split(part, ":")

		for _, sub := range subParts {
			sub = strings.TrimSpace(sub)
			if strings.Contains(sub, ".") {
				fields = append(fields, sub)
			}
		}
	}

	return fields
}

// insertField inserts a field into a nested map structure at a specified path, creating intermediate maps as needed.
func insertField(m map[string]any, path []string, field string) {
	if len(path) == 0 {
		return
	}

	// Pass throw struct normally
	current := m

	for i, p := range path {
		if i == len(path)-1 {
			val := current[p]
			switch cast := val.(type) {
			case nil:
				current[p] = []any{field}
			case []any:
				current[p] = appendIfMissingAny(cast, field)
			default:
				current[p] = []any{field}
			}
		} else {
			next := current[p]
			switch val := next.(type) {
			case map[string]any:
				current = val
			case []any:
				found := false

				for _, item := range val {
					if m2, ok := item.(map[string]any); ok {
						current = m2
						found = true

						break
					}
				}

				if !found {
					newMap := map[string]any{}
					current[p] = append(val, newMap)
					current = newMap
				}
			case nil:
				newMap := map[string]any{}
				current[p] = newMap
				current = newMap
			default:
				newMap := map[string]any{}
				current[p] = newMap
				current = newMap
			}
		}
	}
}

// appendIfMissingAny add field only if does not exist yet
func appendIfMissingAny(slice []any, val any) []any {
	switch v := val.(type) {
	case string:
		for _, item := range slice {
			if str, ok := item.(string); ok && str == v {
				return slice
			}
		}
	case map[string]any:
		for _, item := range slice {
			if m, ok := item.(map[string]any); ok {
				for key := range m {
					if _, exists := v[key]; exists {
						return slice
					}
				}
			}
		}
	}

	return append(slice, val)
}

// CleanPath remove indexes and brackets of paths like foo[0].bar or foo.0.bar
func CleanPath(path string) []string {
	parts := strings.Split(path, ".")

	clean := make([]string, 0, len(parts))

	for _, p := range parts {
		base := strings.Split(p, "[")[0]
		if _, err := strconv.Atoi(base); err == nil {
			continue
		}

		clean = append(clean, base)
	}

	return clean
}
