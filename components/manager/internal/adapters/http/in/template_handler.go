package in

import (
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/components/manager/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/model"
	"plugin-template-engine/pkg/net/http"
	"strings"
)

type TemplateHandler struct {
	Service *services.UseCase
}

// CreateTemplate is a method that creates a template.
//
//	@Summary		Create a Template for reports
//	@Description	Create a Template for reports with the input payload
//	@Tags			Template
//	@Accept			json
//	@Produce		json
//	@Param			pack				body		model.CreateTemplateInput	true	"Template Input"
//	@Success		201					{object}	template.Template
//	@Router			/v1/template [post]
func (th *TemplateHandler) CreateTemplate(t any, c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_template")
	defer span.End()

	c.SetUserContext(ctx)

	logger.Info("Request to create template")
	organizationID := c.Locals("X-Organization-Id").(uuid.UUID)
	payload := t.(*model.CreateTemplateInput)
	logger.Infof("Request to create a pack with details: %#v", payload)

	err := opentelemetry.SetSpanAttributesFromStruct(&span, "payload", payload)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)

		return http.WithError(c, err)
	}

	templateOut, err := th.Service.CreateTemplate(ctx, payload, organizationID)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to create pack on command", err)

		return http.WithError(c, err)
	}

	logger.Infof("Successfully created create template %v", templateOut)

	return http.OK(c, templateOut)

	/*
		file, errFile := http.GetFileFromHeader(c)
		if errFile != nil {
			opentelemetry.HandleSpanError(&span, "Failed to get file from Header", errFile)

			logger.Error("Failed to get file from Header: ", errFile.Error())

			return http.WithError(c, errFile)
		}

		err := opentelemetry.SetSpanAttributesFromStruct(&span, "file", file)
		if err != nil {
			opentelemetry.HandleSpanError(&span, "Failed to convert payload to JSON string", err)

			return http.WithError(c, err)
		}

		tpl, err := pongo2.FromString(file)
		if err != nil {
			return http.WithError(c, err)
		}

		re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.]+)\s*}}`)
		matches := re.FindAllStringSubmatch(file, -1)

		var keysFile []string

		for _, match := range matches {
			if len(match) > 1 {
				keysFile = append(keysFile, match[1])
			}
		}

		ctxPongo := pongo2.Context{}

		setValueKeyWithDot(ctxPongo, keysFile[0], "Minha Empresa Ltda")
		setValueKeyWithDot(ctxPongo, keysFile[1], 12345)

		res, err := tpl.Execute(ctxPongo)
		if err != nil {
			return http.WithError(c, err)
		}*/
}

func setValueKeyWithDot(context pongo2.Context, key string, value any) {
	keys := strings.Split(key, ".") // divido a chave que tem os .
	lastKey := keys[len(keys)-1]    // pego a ultima parte que indica o valor

	currentMap := context
	for _, k := range keys[:len(keys)-1] {
		if _, exists := currentMap[k]; !exists {
			currentMap[k] = make(map[string]any)
		}

		currentMap = currentMap[k].(map[string]any)
	}

	currentMap[lastKey] = value
}
