// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libRabbitmq "github.com/LerianStudio/lib-commons/v2/commons/rabbitmq"
	libRedis "github.com/LerianStudio/lib-commons/v2/commons/redis"
)

// Service is the application glue where we put all top-level components to be used.
type Service struct {
	*Server
	log.Logger
	mongoConnection    *libMongo.MongoConnection
	rabbitMQConnection *libRabbitmq.RabbitMQConnection
	redisConnection    *libRedis.RedisConnection
}

// Run starts the application.
// This is the only necessary code to run an app in the main.go
func (app *Service) Run() {
	commons.NewLauncher(
		commons.WithLogger(app.Logger),
		commons.RunApp("HTTP Service", app.Server),
	).Run()

	// Graceful shutdown
	app.Info("Starting graceful shutdown...")

	// Close MongoDB connection
	if app.mongoConnection != nil && app.mongoConnection.DB != nil {
		app.Info("Closing MongoDB connection...")

		if err := app.mongoConnection.DB.Disconnect(context.Background()); err != nil {
			app.Errorf("Failed to close MongoDB connection: %v", err)
		} else {
			app.Info("MongoDB connection closed")
		}
	}

	// Close RabbitMQ connection
	if app.rabbitMQConnection != nil {
		app.Info("Closing RabbitMQ connection...")

		if app.rabbitMQConnection.Channel != nil {
			if err := app.rabbitMQConnection.Channel.Close(); err != nil {
				app.Errorf("Failed to close RabbitMQ channel: %v", err)
			}
		}

		if app.rabbitMQConnection.Connection != nil && !app.rabbitMQConnection.Connection.IsClosed() {
			if err := app.rabbitMQConnection.Connection.Close(); err != nil {
				app.Errorf("Failed to close RabbitMQ connection: %v", err)
			} else {
				app.Info("RabbitMQ connection closed")
			}
		}
	}

	// Close Redis connection
	if app.redisConnection != nil {
		app.Info("Closing Redis connection...")

		if err := app.redisConnection.Close(); err != nil {
			app.Errorf("Failed to close Redis connection: %v", err)
		} else {
			app.Info("Redis connection closed")
		}
	}

	app.Info("Graceful shutdown complete")
}
