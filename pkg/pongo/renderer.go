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
	tpl, err := pongo2.FromBytes(templateBytes)
	if err != nil {
		logger.Errorf("Error parsing template: %s", err.Error())
		return "", err
	}

	pongoCtx := pongo2.Context{
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

// cleanNumericOutput removes trailing zeros from numeric values in the output
func cleanNumericOutput(output string) string {
	re := regexp.MustCompile(`\b\d+\.\d*0+\b`)

	cleaned := re.ReplaceAllStringFunc(output, func(match string) string {
		return cleanNumericString(match)
	})

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
