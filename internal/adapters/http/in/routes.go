package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"plugin-template-engine/pkg/log"
	"plugin-template-engine/pkg/net/http"
	"plugin-template-engine/pkg/opentelemetry"
)

func NewRoutes(lg log.Logger, tl *opentelemetry.Telemetry, templateHandler *TemplateHandler) *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	tlMid := http.NewTelemetryMiddleware(tl)

	f.Use(tlMid.WithTelemetry(tl))
	f.Use(cors.New())
	f.Use(http.WithHTTPLogging(http.WithCustomLogger(lg)))

	// Example routes
	f.Post("/v1/template", templateHandler.CreateTemplatePongo2)

	// Doc Swagger
	f.Get("/swagger/*", WithSwaggerEnvConfig(), fiberSwagger.WrapHandler)

	f.Use(tlMid.EndTracingSpans)

	return f
}
