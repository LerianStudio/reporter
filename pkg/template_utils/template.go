package template_utils

import (
	"regexp"
	"strconv"
	"strings"
)

// MappedFieldsOfTemplate Map all fields of template file creating a map[string]map[string][]string
func MappedFieldsOfTemplate(templateFile string) map[string]map[string][]string {
	variableMap := map[string][]string{}

	// Regex to capture blocks (ex: {% for t in midaz_transaction.transaction %})
	forRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s+in\s+([^\s%]+)\s*-?%}`)
	forMatches := forRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range forMatches {
		variable := match[1]
		path := CleanPath(match[2])

		// If the placeholder be: midaz_transaction.transaction
		if len(path) >= 2 {
			parent := path[0]
			child := path[1]
			variableMap[variable] = []string{parent, child}
		} else {
			// Fallback
			variableMap[variable] = path
		}
	}

	result := map[string]any{}

	// Regex for fields {{ t.id }}, {{ user.name }}, etc
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

			// If start with (ex: t, op, etc)
			if loopPath, ok := variableMap[parts[0]]; ok {
				fullPath := append([]string{}, loopPath...) // clone
				insertField(result, fullPath, parts[1])
			} else {
				insertField(result, parts[:len(parts)-1], parts[len(parts)-1])
			}
		}
	}

	return normalizeStructure(result)
}

func normalizeStructure(input map[string]interface{}) map[string]map[string][]string {
	result := make(map[string]map[string][]string)

	for topKey, topVal := range input {
		section := make(map[string][]string)

		if m, ok := topVal.(map[string]interface{}); ok {
			for subKey, subVal := range m {
				switch v := subVal.(type) {
				case []interface{}:
					for _, item := range v {
						switch itemVal := item.(type) {
						case string:
							section[subKey] = append(section[subKey], itemVal)
						case map[string]interface{}:
							for nestedKey := range itemVal {
								section[subKey] = append(section[subKey], nestedKey)
							}
						}
					}
				case map[string]interface{}: // Caso especial como em "transaction": { "metadata": [...] }
					section[subKey] = append(section[subKey], getMapKeys(v)...)
				}
			}
		}

		result[topKey] = section
	}

	return result
}

func getMapKeys(m map[string]interface{}) []string {
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

// hydrateArray Adjust array of any if you have a metadata like this, (ex: "transaction": { "metadata": [ "mensagem" ] })
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

// extractFieldsFromExpression Get all fields of expression tags
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

// insertField insere um campo em uma estrutura de mapa aninhada com base no caminho fornecido
func insertField(m map[string]any, path []string, field string) {
	if len(path) == 0 {
		return
	}

	/*
		if len(path) >= 2 {
			parent := path[len(path)-2]
			last := path[len(path)-1]

			// List of subfields that should be grouped
			groupFields := map[string][]string{
				"address": {"line1", "line2", "city", "state", "zip", "country"},
			}

			// If actual field is "line1", "city", etc e o pai é "address", vamos registrar só "address"
			if subs, ok := groupFields[parent]; ok {
				for _, sub := range subs {
					if sub == last {
						// Remove the last element and use the principal field
						insertField(m, path[:len(path)-2], parent)
						return
					}
				}
			}
		} */

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
	var clean []string
	for _, p := range parts {
		base := strings.Split(p, "[")[0]
		if _, err := strconv.Atoi(base); err == nil {
			continue
		}
		clean = append(clean, base)
	}
	return clean
}
