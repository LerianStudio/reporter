// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"context"
	"time"

	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/net/http"
	"github.com/LerianStudio/reporter/pkg/storage"

	middlewareAuth "github.com/LerianStudio/lib-auth/v2/auth/middleware"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	mongoDB "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	commonsHttp "github.com/LerianStudio/lib-commons/v2/commons/net/http"
	"github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

const (
	applicationName       = "reporter"
	templateResource      = "templates"
	reportResource        = "reports"
	dataSourceResource    = "data-source"
	readinessCheckTimeout = 2 * time.Second
)

// ReadinessDeps holds the dependency connections needed for the /ready endpoint.
type ReadinessDeps struct {
	MongoConnection    *mongoDB.MongoConnection
	RabbitMQConnection *libRabbitmq.RabbitMQConnection
	RedisConnection    *libRedis.RedisConnection
	StorageClient      storage.ObjectStorage
}

// NewRoutes creates a new fiber router with the specified handlers and middleware.
func NewRoutes(lg log.Logger, tl *opentelemetry.Telemetry, templateHandler *TemplateHandler, reportHandler *ReportHandler, dataSourceHandler *DataSourceHandler, auth *middlewareAuth.AuthClient, deps *ReadinessDeps) *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return commonsHttp.HandleFiberError(ctx, err)
		},
	})
	tlMid := commonsHttp.NewTelemetryMiddleware(tl)

	f.Use(tlMid.WithTelemetry(tl))
	f.Use(cors.New())
	f.Use(commonsHttp.WithHTTPLogging(commonsHttp.WithCustomLogger(lg)))

	// Plugin templates routes
	// Template routes
	f.Post("/v1/templates", auth.Authorize(applicationName, templateResource, "post"), templateHandler.CreateTemplate)
	f.Patch("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "patch"), ParsePathParametersUUID, templateHandler.UpdateTemplateByID)
	f.Get("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "get"), ParsePathParametersUUID, templateHandler.GetTemplateByID)
	f.Get("/v1/templates", auth.Authorize(applicationName, templateResource, "get"), templateHandler.GetAllTemplates)
	f.Delete("/v1/templates/:id", auth.Authorize(applicationName, templateResource, "delete"), ParsePathParametersUUID, templateHandler.DeleteTemplateByID)

	// Report routes
	f.Post("/v1/reports", auth.Authorize(applicationName, reportResource, "post"), http.WithBody(new(model.CreateReportInput), reportHandler.CreateReport))
	f.Get("/v1/reports/:id/download", auth.Authorize(applicationName, reportResource, "get"), ParsePathParametersUUID, reportHandler.GetDownloadReport)
	f.Get("/v1/reports/:id", auth.Authorize(applicationName, reportResource, "get"), ParsePathParametersUUID, reportHandler.GetReport)
	f.Get("/v1/reports", auth.Authorize(applicationName, reportResource, "get"), reportHandler.GetAllReports)

	// Data source routes
	f.Get("/v1/data-sources", auth.Authorize(applicationName, dataSourceResource, "get"), dataSourceHandler.GetDataSourceInformation)
	f.Get("/v1/data-sources/:dataSourceId", auth.Authorize(applicationName, dataSourceResource, "get"), dataSourceHandler.GetDataSourceInformationByID)

	// Doc Swagger
	f.Get("/swagger/*", WithSwaggerEnvConfig(), fiberSwagger.WrapHandler)

	// Health
	f.Get("/health", commonsHttp.Ping)

	// Readiness - checks all dependency connections
	f.Get("/ready", readinessHandler(deps))

	// Version
	f.Get("/version", commonsHttp.Version)

	f.Use(tlMid.EndTracingSpans)

	return f
}

// dependencyResult represents the health status of a single dependency in the readiness check.
type dependencyResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// readinessHandler returns a Fiber handler that checks all dependency connections.
// Each dependency is checked with a 2-second timeout. Returns 200 if all healthy, 503 otherwise.
func readinessHandler(deps *ReadinessDeps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		httpStatus := fiber.StatusOK
		results := make(map[string]*dependencyResult)

		// Check MongoDB
		results["mongodb"] = checkMongoDB(deps.MongoConnection)

		// Check RabbitMQ
		results["rabbitmq"] = checkRabbitMQ(deps.RabbitMQConnection)

		// Check Redis/Valkey
		results["redis"] = checkRedis(deps.RedisConnection)

		// Check Storage (S3/SeaweedFS)
		results["storage"] = checkStorage(deps.StorageClient)

		for _, result := range results {
			if result.Status != "ready" {
				httpStatus = fiber.StatusServiceUnavailable

				break
			}
		}

		overallStatus := "ready"
		if httpStatus == fiber.StatusServiceUnavailable {
			overallStatus = "not_ready"
		}

		return commonsHttp.JSONResponse(c, httpStatus, fiber.Map{
			"status":       overallStatus,
			"dependencies": results,
		})
	}
}

// checkMongoDB pings the MongoDB connection with a timeout.
func checkMongoDB(conn *mongoDB.MongoConnection) *dependencyResult {
	if conn == nil {
		return &dependencyResult{Status: "not_ready", Message: "connection not configured"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), readinessCheckTimeout)
	defer cancel()

	db, err := conn.GetDB(ctx)
	if err != nil {
		return &dependencyResult{Status: "not_ready", Message: "failed to get connection"}
	}

	if err = db.Ping(ctx, nil); err != nil {
		return &dependencyResult{Status: "not_ready", Message: "ping failed"}
	}

	return &dependencyResult{Status: "ready"}
}

// checkRabbitMQ verifies the RabbitMQ connection is alive.
func checkRabbitMQ(conn *libRabbitmq.RabbitMQConnection) *dependencyResult {
	if conn == nil {
		return &dependencyResult{Status: "not_ready", Message: "connection not configured"}
	}

	if !conn.Connected || conn.Connection == nil || conn.Connection.IsClosed() {
		return &dependencyResult{Status: "not_ready", Message: "connection is closed"}
	}

	if !conn.HealthCheck() {
		return &dependencyResult{Status: "not_ready", Message: "health check failed"}
	}

	return &dependencyResult{Status: "ready"}
}

// checkRedis pings the Redis/Valkey connection with a timeout.
func checkRedis(conn *libRedis.RedisConnection) *dependencyResult {
	if conn == nil {
		return &dependencyResult{Status: "not_ready", Message: "connection not configured"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), readinessCheckTimeout)
	defer cancel()

	client, err := conn.GetClient(ctx)
	if err != nil {
		return &dependencyResult{Status: "not_ready", Message: "failed to get client"}
	}

	if _, err = client.Ping(ctx).Result(); err != nil {
		return &dependencyResult{Status: "not_ready", Message: "ping failed"}
	}

	return &dependencyResult{Status: "ready"}
}

// checkStorage verifies the S3/SeaweedFS storage connection by checking bucket existence.
func checkStorage(client storage.ObjectStorage) *dependencyResult {
	if client == nil {
		return &dependencyResult{Status: "not_ready", Message: "storage client not configured"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), readinessCheckTimeout)
	defer cancel()

	// Attempt to check existence of a non-existent key as a connectivity test.
	// This exercises the S3 API path and confirms the bucket/endpoint is reachable.
	_, err := client.Exists(ctx, ".readiness-check")
	if err != nil {
		return &dependencyResult{Status: "not_ready", Message: "storage connectivity check failed"}
	}

	return &dependencyResult{Status: "ready"}
}
