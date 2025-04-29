package pongo

import (
	"context"
	"fmt"

	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/flosch/pongo2/v6"
)

// TemplateRenderer handles rendering templates using pongo2
type TemplateRenderer struct{}

// NewTemplateRenderer creates a new TemplateRenderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// RenderFromBytes renders a template from bytes using the provided data context
func (r *TemplateRenderer) RenderFromBytes(ctx context.Context, templateBytes []byte, data map[string]map[string][]map[string]any) (string, error) {
	logger := libCommons.NewLoggerFromContext(ctx)

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
	}
	for k, v := range data {
		pongoCtx[k] = v
	}

	out, err := tpl.Execute(pongoCtx)
	if err != nil {
		logger.Errorf("Error executing template: %s", err.Error())
		return "", err
	}

	return out, nil
}
