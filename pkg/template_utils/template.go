// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template_utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg/constant"
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
	case "pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// MappedFieldsOfTemplate analyzes a template file and returns a nested map of variable paths and their associated fields.
func MappedFieldsOfTemplate(templateFile string) map[string]map[string][]string {
	variableMap := regexBlockForOnPlaceholder(templateFile)
	resultRegex := regexBlockWithOnPlaceholder(variableMap, templateFile)
	regexBlockForWithFilterOnPlaceholder(resultRegex, variableMap, templateFile)

	// Regex for fields {{ ... }}
	fieldRegex := regexp.MustCompile(`{{\s*(.*?)\s*}}`)

	fieldMatches := fieldRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range fieldMatches {
		expr := match[1]

		// Skip expressions that contain DIMP filters - they will be processed by regexBlockDIMPFiltersOnPlaceholder
		if strings.Contains(expr, "|where:") || strings.Contains(expr, "|sum:") || strings.Contains(expr, "|count:") {
			// For DIMP filter expressions, only extract the base path fields (before the pipe)
			// The filter fields will be extracted by regexBlockDIMPFiltersOnPlaceholder
			basePart := strings.Split(expr, "|")[0]
			basePart = strings.TrimSpace(basePart)

			// Don't process if it's just a collection path (will be handled by DIMP function)
			basePathParts := CleanPath(basePart)
			if len(basePathParts) == 2 {
				// This is just datasource.collection, skip it - DIMP function will handle
				continue
			}
		}

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
	regexBlockSetOnPlaceholder(templateFile, resultRegex, variableMap)
	regexBlockAggregationBlocksOnPlaceholder(templateFile, resultRegex, variableMap)
	regexBlockCalcOnPlaceholder(templateFile, resultRegex, variableMap)
	regexBlockDIMPFiltersOnPlaceholder(templateFile, resultRegex, variableMap)

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

// regexBlockIfOnPlaceholder parses a template file to process "if" blocks and updates a nested map with extracted field mappings.
// It identifies fields used in conditional statements, cleans their paths, and inserts them into the resultRegex map structure.
func regexBlockSetOnPlaceholder(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	setRegex := regexp.MustCompile(`{%-?\s*set\s+(.*?)\s*-?%}`)

	setMatches := setRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range setMatches {
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

// regexBlockIfOnPlaceholder parses a template file to process "if" blocks and updates a nested map with extracted field mappings.
// It identifies fields used in conditional statements, cleans their paths, and inserts them into the resultRegex map structure.
func regexBlockAggregationBlocksOnPlaceholder(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	aggrRegexes := []*regexp.Regexp{
		regexp.MustCompile(`{%-?\s*count_by\s+(.*?)\s*-?%}`),
		regexp.MustCompile(`{%-?\s*sum_by\s+(.*?)\s*-?%}`),
		regexp.MustCompile(`{%-?\s*avg_by\s+(.*?)\s*-?%}`),
		regexp.MustCompile(`{%-?\s*min_by\s+(.*?)\s*-?%}`),
		regexp.MustCompile(`{%-?\s*max_by\s+(.*?)\s*-?%}`),
	}

	var matches [][]string
	for _, re := range aggrRegexes {
		matches = append(matches, re.FindAllStringSubmatch(templateFile, -1)...)
	}

	for _, match := range matches {
		expr := match[1]

		args := extractFieldsFromExpressionOfAggregation(expr)
		if len(args) == 0 {
			continue
		}

		mainPath := CleanPath(strings.TrimSpace(args[0]))
		if len(mainPath) < 2 {
			continue
		}

		variableMap[mainPath[1]] = mainPath

		for _, arg := range args[1:] {
			// Skip quoted string literals (values like "cacc", 'value', etc.)
			trimmedArg := strings.TrimSpace(arg)
			if isQuotedString(trimmedArg) {
				continue
			}

			argPath := CleanPath(arg)

			switch {
			case len(argPath) < 2:
				insertField(resultRegex, mainPath, arg)
			case variableMap[argPath[0]] != nil:
				insertField(resultRegex, variableMap[argPath[0]], argPath[1])
			default:
				insertField(resultRegex, argPath[:len(argPath)-1], argPath[len(argPath)-1])
			}
		}
	}
}

// regexBlockDIMPFiltersOnPlaceholder parses a template file to process DIMP filters (where, sum, count)
// and updates the nested map with extracted field mappings.
// It identifies fields used in filter expressions like |where:"field:value", |sum:"field", |count:"field:value"
func regexBlockDIMPFiltersOnPlaceholder(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	processDIMPExpressions(templateFile, resultRegex, variableMap)
	processDIMPForLoops(templateFile, resultRegex)
}

// processDIMPExpressions processes {{ }} expressions containing DIMP filters
func processDIMPExpressions(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	exprRegex := regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)
	exprMatches := exprRegex.FindAllStringSubmatch(templateFile, -1)

	for _, exprMatch := range exprMatches {
		expr := exprMatch[1]

		if !containsDIMPFilter(expr) {
			continue
		}

		basePath := extractDIMPBasePath(expr, variableMap)
		if basePath == nil {
			continue
		}

		ensureMapStructure(resultRegex, basePath)
		extractFieldsFromDIMPFilters(expr, resultRegex, basePath)
	}
}

// processDIMPForLoops processes for loops containing DIMP filters
func processDIMPForLoops(templateFile string, resultRegex map[string]any) {
	forFilterRegex := regexp.MustCompile(`{%-?\s*for\s+\w+\s+in\s+([a-zA-Z_][\w.]*)\s*\|\s*(where|sum|count)\s*:\s*"([^"]+)"`)
	forMatches := forFilterRegex.FindAllStringSubmatch(templateFile, -1)

	for _, match := range forMatches {
		collection := match[1]
		filterType := match[2]
		filterArg := match[3]

		collectionParts := CleanPath(collection)
		if len(collectionParts) < 2 {
			continue
		}

		basePath := collectionParts[:2]
		ensureMapStructure(resultRegex, basePath)
		extractFieldFromFilterArg(resultRegex, basePath, filterType, filterArg)
	}
}

// containsDIMPFilter checks if an expression contains DIMP filters
func containsDIMPFilter(expr string) bool {
	return strings.Contains(expr, "|where:") || strings.Contains(expr, "|sum:") || strings.Contains(expr, "|count:")
}

func extractDIMPBasePath(expr string, variableMap map[string][]string) []string {
	parts := strings.Split(expr, "|")
	if len(parts) < 2 {
		return nil
	}

	baseCollection := strings.TrimSpace(parts[0])
	collectionParts := CleanPath(baseCollection)

	if len(collectionParts) == 0 {
		return nil
	}

	if loopPath, ok := variableMap[collectionParts[0]]; ok {
		return loopPath
	}

	if len(collectionParts) >= 2 {
		return collectionParts[:2]
	}

	return nil
}

// extractFieldsFromDIMPFilters extracts fields from all DIMP filters in an expression
func extractFieldsFromDIMPFilters(expr string, resultRegex map[string]any, basePath []string) {
	parts := strings.Split(expr, "|")
	filterArgRegex := regexp.MustCompile(`^(where|sum|count)\s*:\s*"([^"]+)"`)

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		filterMatch := filterArgRegex.FindStringSubmatch(part)

		if filterMatch == nil {
			continue
		}

		extractFieldFromFilterArg(resultRegex, basePath, filterMatch[1], filterMatch[2])
	}
}

// extractFieldFromFilterArg extracts a field from a filter argument and inserts it into the result
func extractFieldFromFilterArg(resultRegex map[string]any, basePath []string, filterType, filterArg string) {
	switch filterType {
	case "where", "count":
		colonIdx := strings.Index(filterArg, ":")
		if colonIdx > 0 {
			field := filterArg[:colonIdx]
			insertFieldToPath(resultRegex, basePath, field)
		}
	case "sum":
		field := strings.Trim(filterArg, `"' `)
		if field != "" {
			insertFieldToPath(resultRegex, basePath, field)
		}
	}
}

// ensureMapStructure ensures that the nested map structure exists for the given path
// For path ["datasource", "collection"], it creates: resultRegex["datasource"]["collection"] = []any{}
func ensureMapStructure(m map[string]any, path []string) {
	if len(path) < 2 {
		return
	}

	datasource := path[0]
	collection := path[1]

	// Ensure datasource map exists
	if _, ok := m[datasource]; !ok {
		m[datasource] = map[string]any{}
	}

	// Get or create the datasource map
	dsMap, ok := m[datasource].(map[string]any)
	if !ok {
		dsMap = map[string]any{}
		m[datasource] = dsMap
	}

	// Ensure collection array exists
	if _, ok := dsMap[collection]; !ok {
		dsMap[collection] = []any{}
	}
}

// insertFieldToPath inserts a field into the nested structure at the given path
// For path ["datasource", "collection"] and field "status", it adds "status" to the collection's field list
func insertFieldToPath(m map[string]any, path []string, field string) {
	if len(path) < 2 {
		return
	}

	datasource := path[0]
	collection := path[1]

	// Get the datasource map
	dsMap, ok := m[datasource].(map[string]any)
	if !ok {
		return
	}

	// Get the collection's field list and append
	switch v := dsMap[collection].(type) {
	case []any:
		dsMap[collection] = appendIfMissingAny(v, field)
	case nil:
		dsMap[collection] = []any{field}
	}
}

// regexBlockCalcOnPlaceholder parses a template file to process "calc" blocks and updates a nested map with extracted field mappings.
// It identifies fields used in calculation expressions, cleans their paths, and inserts them into the resultRegex map structure.
func regexBlockCalcOnPlaceholder(templateFile string, resultRegex map[string]any, variableMap map[string][]string) {
	calcRegex := regexp.MustCompile(`{%-?\s*calc\s+(.*?)\s*-?%}`)

	calcMatches := calcRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range calcMatches {
		expr := match[1]
		fieldPaths := extractCalcFromExpression(expr)

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
// It also handles for loops with filters like: {% for acc in collection|where:"field:value" %}
func regexBlockForOnPlaceholder(templateFile string) map[string][]string {
	variableMap := map[string][]string{}

	// Regex for block for - updated to capture collection with optional filters
	// Matches: {% for var in collection %} or {% for var in collection|filter:"arg" %}
	forRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s+in\s+([a-zA-Z_][\w.]*(?:\s*\|[^%]+)?)\s*-?%}`)

	forMatches := forRegex.FindAllStringSubmatch(templateFile, -1)
	for _, match := range forMatches {
		variable := match[1]
		fullExpr := match[2]

		// Extract base collection path (before any filter)
		// e.g., "midaz_onboarding.account|where:\"type:cacc\"" -> "midaz_onboarding.account"
		basePath := extractBasePathFromFilterExpr(fullExpr)
		path := CleanPath(basePath)

		if len(path) == 2 {
			variableMap[variable] = []string{path[0], path[1]}
		} else if len(path) > 2 {
			variableMap[variable] = []string{path[0], path[1], path[2]}
		} else {
			variableMap[variable] = path
		}
	}

	// Resolve nested variable references (e.g., when inner loop iterates over parent loop variable's field)
	resolveNestedVariables(variableMap)

	return variableMap
}

// resolveNestedVariables resolves nested loop variable references in variableMap.
// When a variable's path starts with another loop variable, it expands the full path.
// Example: if variableMap["alias"] = ["plugin_crm", "aliases"] and
// variableMap["related_party"] = ["alias", "related_parties"], this function
// resolves it to variableMap["related_party"] = ["plugin_crm", "aliases", "related_parties"]
func resolveNestedVariables(variableMap map[string][]string) {
	maxIterations := len(variableMap) // Prevent infinite loops

	for i := 0; i < maxIterations; i++ {
		resolved := true

		for varName, path := range variableMap {
			if len(path) == 0 {
				continue
			}

			// Check if first element of path is another loop variable
			if parentPath, exists := variableMap[path[0]]; exists && path[0] != varName {
				// Expand: replace path[0] with parentPath, keep rest
				newPath := make([]string, 0, len(parentPath)+len(path)-1)
				newPath = append(newPath, parentPath...)
				newPath = append(newPath, path[1:]...)
				variableMap[varName] = newPath
				resolved = false
			}
		}

		if resolved {
			break
		}
	}
}

// extractBasePathFromFilterExpr extracts the base collection path from an expression that may contain filters.
// e.g., "midaz_onboarding.account|where:\"type:cacc\"" -> "midaz_onboarding.account"
func extractBasePathFromFilterExpr(expr string) string {
	// Find the first pipe character (filter separator)
	pipeIdx := strings.Index(expr, "|")
	if pipeIdx > 0 {
		return strings.TrimSpace(expr[:pipeIdx])
	}

	return strings.TrimSpace(expr)
}

// regexBlockForWithFilterOnPlaceholder processes "for" loops with the filter function in a template file, updating nested data structures.
// It extracts variable mappings, assigns paths, and inserts filtered parameter data into the result map.
func regexBlockForWithFilterOnPlaceholder(result map[string]any, variableMap map[string][]string, templateFile string) {
	withRegex := regexp.MustCompile(`{%-?\s*for\s+(\w+)\s*in\s*filter\(\s*([^)]+)\s*\)[^\%]+`)

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
}

// regexBlockWithOnPlaceholder parses a template file to process "with" statements and updates `variableMap` with mapped variables.
// The function extracts filters, processes their arguments, and organizes nested data into a structured map for use.
// It cleans paths, maps targets to their corresponding variables, and inserts additional parameters where applicable.
func regexBlockWithOnPlaceholder(variableMap map[string][]string, templateFile string) map[string]any {
	result := map[string]any{}
	withRegex1 := regexp.MustCompile(`{%-?\s*with\s+(\w+)\s*=\s*filter\(\s*([^)]+)\s*\)[^\%]+`)
	withRegex2 := regexp.MustCompile(`{%-?\s*with\s+(\w+)\s*=\s*([^\s%]+)\s*-?%}`)

	withMatches := withRegex1.FindAllStringSubmatch(templateFile, -1)
	withMatches2 := withRegex2.FindAllStringSubmatch(templateFile, -1)

	// Aggregate both sets of matches
	if withMatches2 != nil {
		withMatches = append(withMatches, withMatches2...)
	}

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

// extractFieldsFromExpressionOfAggregation parses an aggregation expression and extracts key fields as a slice of strings.
func extractFieldsFromExpressionOfAggregation(expr string) []string {
	result := make([]string, 0)
	re := regexp.MustCompile(`^\s*(\S+)\s+if\s+(\S+)\s*==\s*(\S+)\s*$`)
	matches := re.FindStringSubmatch(expr)

	if len(matches) == 4 {
		result = []string{matches[1], matches[2], matches[3]}
	} else {
		re = regexp.MustCompile(`^\s*(\S+)\s+by\s+"([^"]+)"\s+if\s+(\S+)\s*==\s*(\S+)`)
		matches = re.FindStringSubmatch(expr)

		if len(matches) == 5 {
			result = []string{matches[1], matches[2], matches[3], matches[4]}
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

// extractCalcFromExpression extracts object.field patterns from a calculation expression
func extractCalcFromExpression(expr string) []string {
	return extractIfFromExpression(expr)
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
							// Recursively flatten nested fields with prefix
							nestedFields := flattenNestedFields(itemVal, "")
							section[subKey] = append(section[subKey], nestedFields...)
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

// flattenNestedFields recursively extracts all fields from a nested map structure,
// prefixing nested field names with their parent keys (e.g., "related_parties.role").
func flattenNestedFields(m map[string]any, prefix string) []string {
	var fields []string

	for key, val := range m {
		fieldName := key
		if prefix != "" {
			fieldName = prefix + "." + key
		}

		switch v := val.(type) {
		case []any:
			// Add the key itself as a field
			fields = append(fields, fieldName)
			// Also extract nested string fields
			for _, item := range v {
				switch itemVal := item.(type) {
				case string:
					fields = append(fields, fieldName+"."+itemVal)
				case map[string]any:
					// Recursively flatten deeper nested structures
					nested := flattenNestedFields(itemVal, fieldName)
					fields = append(fields, nested...)
				}
			}
		case map[string]any:
			// Recursively process nested maps
			nested := flattenNestedFields(v, fieldName)
			fields = append(fields, nested...)
		case string:
			fields = append(fields, fieldName)
		}
	}

	return fields
}

// getMapKeys retrieves all keys from a given map and returns them as a slice of strings.
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
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
			// Skip if it looks like a filter argument (contains quotes) or is too short
			if strings.Contains(sub, `"`) || strings.Contains(sub, `'`) {
				continue
			}

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

// isQuotedString checks if a string is a quoted literal (starts and ends with quotes)
func isQuotedString(s string) bool {
	if len(s) < 2 {
		return false
	}

	return (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')
}

// ValidateNoScriptTag checks if the template contains <script> tags (case-insensitive) and returns an error if found.
func ValidateNoScriptTag(templateFile string) error {
	lower := strings.ToLower(templateFile)
	if strings.Contains(lower, "<script>") || strings.Contains(lower, "</script>") {
		return constant.ErrScriptTagDetected
	}

	return nil
}
