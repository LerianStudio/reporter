package pongo

import (
	"context"
	"testing"

	"github.com/LerianStudio/lib-commons/v2/commons/zap"
	"github.com/stretchr/testify/assert"
)

func TestRenderFromBytes_Success(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte("Hello, {{ person._.0.name }}!")

	data := map[string]map[string][]map[string]any{
		"person": {
			"_": {
				{"name": "World"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", out)
}

func TestRenderFromBytes_SyntaxError(t *testing.T) {
	r := NewTemplateRenderer()
	logger := zap.InitializeLogger()
	tpl := []byte("Hello, {{ name !")
	data := map[string]map[string][]map[string]any{
		"name": {
			"_": {
				{"name": "World"},
			},
		},
	}

	out, err := r.RenderFromBytes(context.Background(), tpl, data, logger)
	assert.Error(t, err)
	assert.Empty(t, out)
}
