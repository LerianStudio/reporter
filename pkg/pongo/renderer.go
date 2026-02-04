// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/LerianStudio/lib-commons/v2/commons/log"

	"github.com/flosch/pongo2/v6"
	"github.com/shopspring/decimal"
)

// TemplateRenderer handles rendering templates using pongo2
type TemplateRenderer struct{}

// NewTemplateRenderer creates a new TemplateRenderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// RenderFromBytes renders a template from bytes using the provided data context
func (r *TemplateRenderer) RenderFromBytes(ctx context.Context, templateBytes []byte, data map[string]map[string][]map[string]any, logger log.Logger) (string, error) {
	// Pre-process template to convert schema syntax (database:schema.table) to Pongo2 compatible syntax
	processedTemplate := preprocessSchemaReferences(string(templateBytes))

	tpl, err := pongo2.FromString(processedTemplate)
	if err != nil {
		logger.Errorf("Error parsing template: %s", err.Error())
		return "", err
	}

	pongoCtx := pongo2.Context{
		// Counter storage scoped to this render (prevents race conditions between concurrent renders)
		CounterContextKey: NewCounterStorage(),
		"filter": func(collection any, field string, value any) []map[string]any {
			var result []map[string]any

			items, ok := collection.([]map[string]any)
			if !ok {
				return result
			}

			for _, item := range items {
				if v, ok := item[field]; ok && fmt.Sprintf("%v", v) == fmt.Sprintf("%v", value) {
					result = append(result, item)
				}
			}

			return result
		},
		"contains": func(str1 any, str2 any) bool {
			s1 := strings.ToUpper(fmt.Sprintf("%v", str1))
			s2 := strings.ToUpper(fmt.Sprintf("%v", str2))
			return strings.Contains(s1, s2)
		},
	}
	for k, v := range data {
		pongoCtx[k] = v
	}

	out, err := tpl.Execute(pongoCtx)
	if err != nil {
		logger.Errorf("Error executing template: %s", err.Error())
		return "", err
	}

	cleaned := cleanNumericOutput(out)

	return cleaned, nil
}

// preprocessSchemaReferences converts explicit schema syntax (database:schema.table) to Pongo2 dot notation.
// Example: "pix_btg:payment.transfers" becomes "pix_btg.payment__transfers"
// The schema.table is converted to schema__table (double underscore) to create a valid Pongo2 identifier.
// This requires the data storage to use the same key format (schema__table).
func preprocessSchemaReferences(template string) string {
	// Pattern matches: word:word.word (database:schema.table pattern)
	// Captures: (database):(schema).(table)
	schemaPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*):([a-zA-Z_][a-zA-Z0-9_]*)\.([a-zA-Z_][a-zA-Z0-9_]*)`)

	// Replace database:schema.table with database.schema__table
	// Note: Use ${n} syntax to avoid Go interpreting $2__ as a named group
	return schemaPattern.ReplaceAllString(template, `${1}.${2}__${3}`)
}

// cleanNumericOutput removes trailing zeros from numeric values in the output
func cleanNumericOutput(output string) string {
	// First, protect XML declarations from being modified
	xmlDeclarationRegex := regexp.MustCompile(`<\?xml[^>]*version="[^"]*"[^>]*\?>`)
	xmlDeclarations := xmlDeclarationRegex.FindAllString(output, -1)

	protectedOutput := output

	for i, declaration := range xmlDeclarations {
		placeholder := fmt.Sprintf("__XML_DECLARATION_%d__", i)
		protectedOutput = strings.Replace(protectedOutput, declaration, placeholder, 1)
	}

	re := regexp.MustCompile(`\b\d+\.\d*0+\b`)
	cleaned := re.ReplaceAllStringFunc(protectedOutput, func(match string) string {
		return cleanNumericString(match)
	})

	for i, declaration := range xmlDeclarations {
		placeholder := fmt.Sprintf("__XML_DECLARATION_%d__", i)
		cleaned = strings.Replace(cleaned, placeholder, declaration, 1)
	}

	return cleaned
}

// cleanNumericString removes trailing zeros from a numeric string
func cleanNumericString(s string) string {
	s = strings.TrimSpace(s)

	if dec, err := decimal.NewFromString(s); err == nil {
		return dec.String()
	}

	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	return s
}
