package in

import (
	"plugin-smart-templates/pkg/model"
	"plugin-smart-templates/pkg/net/http"

	middlewareAuth "github.com/LerianStudio/lib-auth/auth/middleware"
	"github.com/LerianStudio/lib-commons/commons/log"
	commonsHttp "github.com/LerianStudio/lib-commons/commons/net/http"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

const (
	applicationName  = "plugin-smart-templates"
	templateResource = "templates"
	reportResource   = "reports"
)

// NewRoutes creates a new fiber router with the specified handlers and middleware.
func NewRoutes(lg log.Logger, tl *opentelemetry.Telemetry, templateHandler *TemplateHandler, reportHandler *ReportHandler, auth *middlewareAuth.AuthClient) *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	tlMid := commonsHttp.NewTelemetryMiddleware(tl)

	f.Use(tlMid.WithTelemetry(tl))
	f.Use(cors.New())
	f.Use(commonsHttp.WithHTTPLogging(commonsHttp.WithCustomLogger(lg)))

	// Plugin templates routes
	// Template routes
	f.Post("/v1/templates", auth.Authorize(applicationName, templateResource, "post"), ParseHeaderParameters, templateHandler.CreateTemplate)
	f.Patch("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "patch"), ParseHeaderParameters, ParsePathParameters, templateHandler.UpdateTemplateByID)
	f.Get("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "get"), ParseHeaderParameters, ParsePathParameters, templateHandler.GetTemplateByID)
	f.Get("/v1/templates", auth.Authorize(applicationName, templateResource, "get"), ParseHeaderParameters, templateHandler.GetAllTemplates)
	f.Delete("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "delete"), ParseHeaderParameters, ParsePathParameters, templateHandler.DeleteTemplateByID)

	// Report routes
	f.Post("/v1/reports", auth.Authorize(applicationName, reportResource, "post"), ParseHeaderParameters, http.WithBody(new(model.CreateReportInput), reportHandler.CreateReport))
	f.Get("/v1/reports/:id/download", auth.Authorize(applicationName, reportResource, "get"), ParseHeaderParameters, ParsePathParameters, reportHandler.GetDownloadReport)
	f.Get("/v1/reports/:id", auth.Authorize(applicationName, reportResource, "get"), ParseHeaderParameters, ParsePathParameters, reportHandler.GetReport)

	// Doc Swagger
	f.Get("/swagger/*", WithSwaggerEnvConfig(), fiberSwagger.WrapHandler)

	// Health
	f.Get("/health", commonsHttp.Ping)

	// Version
	f.Get("/version", commonsHttp.Version)

	f.Use(tlMid.EndTracingSpans)

	return f
}
