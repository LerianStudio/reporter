package pongo

import (
	"context"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	"github.com/flosch/pongo2/v6"
	"html"
)

func init() {
	if err := pongo2.RegisterFilter("xmlattr", func(input *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsValue(html.EscapeString(input.String())), nil
	}); err != nil {
		panic("Failed to register XML attribute filter: " + err.Error())
	}
}

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

	pongoCtx := pongo2.Context{}
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
