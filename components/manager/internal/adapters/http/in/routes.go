package in

import (
	"github.com/LerianStudio/lib-commons/commons/log"
	commonsHttp "github.com/LerianStudio/lib-commons/commons/net/http"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"plugin-template-engine/pkg/model"
	"plugin-template-engine/pkg/net/http"
)

func NewRoutes(lg log.Logger, tl *opentelemetry.Telemetry, templateHandler *TemplateHandler, reportHandler *ReportHandler) *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	tlMid := commonsHttp.NewTelemetryMiddleware(tl)

	f.Use(tlMid.WithTelemetry(tl))
	f.Use(cors.New())
	f.Use(commonsHttp.WithHTTPLogging(commonsHttp.WithCustomLogger(lg)))

	// Plugin templates routes
	// Template routes
	f.Post("/v1/templates", ParseHeaderParameters, templateHandler.CreateTemplate)
	f.Patch("/v1/templates/:id", ParseHeaderParameters, ParsePathParameters, templateHandler.UpdateTemplateByID)
	f.Get("/v1/templates/:id", ParseHeaderParameters, ParsePathParameters, templateHandler.GetTemplateByID)
	f.Get("/v1/templates", ParseHeaderParameters, templateHandler.GetAllTemplates)
	f.Delete("/v1/templates/:id", ParseHeaderParameters, ParsePathParameters, templateHandler.DeleteTemplateByID)

	// Report routes
	f.Post("/v1/reports", ParseHeaderParameters, http.WithBody(new(model.CreateReportInput), reportHandler.CreateReport))
	f.Get("/v1/reports/:id/download", ParseHeaderParameters, ParsePathParameters, reportHandler.GetDownloadReport)
	f.Get("/v1/reports/:id", ParseHeaderParameters, ParsePathParameters, reportHandler.GetReport)

	// Doc Swagger
	f.Get("/swagger/*", WithSwaggerEnvConfig(), fiberSwagger.WrapHandler)

	// Health
	f.Get("/health", commonsHttp.Ping)

	// Version
	f.Get("/version", commonsHttp.Version)

	f.Use(tlMid.EndTracingSpans)

	return f
}
