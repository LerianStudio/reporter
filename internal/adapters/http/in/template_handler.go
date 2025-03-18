package in

import (
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"plugin-template-engine/internal/services"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/net/http"
	"plugin-template-engine/pkg/opentelemetry"
	"regexp"
	"strings"
)

type TemplateHandler struct {
	Service *services.UseCase
}

// CreateTemplatePongo2 is a method that creates an example.
//
//	@Summary		Create an Example
//	@Description	Create an Example with the input payload
//	@Tags			Example
//	@Accept			json
//	@Produce		json
//	@Router			/v1/template [post]
func (ex *TemplateHandler) CreateTemplatePongo2(c *fiber.Ctx) error {
	ctx := c.UserContext()

	logger := pkg.NewLoggerFromContext(ctx)
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "handler.create_template_pongo2")
	defer span.End()

	c.SetUserContext(ctx)

	logger.Info("Request to create template with Pongo2")

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
	}

	logger.Infof("Successfully created create template %v", res)

	return http.OK(c, nil)
	/*
			re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.]+)\s*}}`)
			matches := re.FindAllStringSubmatch(file, -1)

			var keys []string
			for _, match := range matches {
				if len(match) > 1 {
					keys = append(keys, match[1])
				}
			}

			fmt.Println(keys)

		tpl, err := pongo2.FromString(file)
		if err != nil {
			panic(err)
		}

		/*
			!--- como é mapeado o executa ---!
			out, err := tpl.Execute(pongo2.Context{
				"onboarding": map[string]interface{}{
					"organization": map[string]interface{}{
						"legal_name": "Minha Empresa Ltda",
					},
				},
			})



		out, err := tpl.Execute(pongo2.Context{"test": "test"})
		if err != nil {
			panic(err)
		}
		fmt.Println(out)

		template := pongo2.Must(tpl, err)

		fmt.Println(template)

		/*
			response, err := ex.ExampleCommand.CreateExample(ctx, payload)
			if err != nil {
				opentelemetry.HandleSpanError(&span, "Failed to create example on services", err)

				return http.WithError(c, err)
			}
	*/
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
