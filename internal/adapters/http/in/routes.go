package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/log"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
)

func NewRoutes(lg log.Logger, tl *opentelemetry.Telemetry, exampleHandler *ExampleHandler) *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	tlMid := http.NewTelemetryMiddleware(tl)

	f.Use(tlMid.WithTelemetry(tl))
	f.Use(cors.New())
	f.Use(http.WithHTTPLogging(http.WithCustomLogger(lg)))

	// Example routes
	f.Post("/v1/example", http.WithBody(new(model.CreateExampleInput), exampleHandler.CreateExample))
	f.Get("/v1/example/:id", ParseUUIDPathParameters, exampleHandler.GetExampleByID)
	f.Get("/v1/example", exampleHandler.GetAllExample)
	f.Patch("/v1/example/:id", ParseUUIDPathParameters, http.WithBody(new(model.UpdateExampleInput), exampleHandler.UpdateExample))
	f.Delete("/v1/example/:id", ParseUUIDPathParameters, exampleHandler.DeleteExampleByID)

	// Doc Swagger
	f.Get("/swagger/*", WithSwaggerEnvConfig(), fiberSwagger.WrapHandler)

	f.Use(tlMid.EndTracingSpans)

	return f
}
