// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

const (
	MongoUser     = "reporter"
	MongoPassword = "reporter"
	MongoDatabase = "reporter"
)

// MongoDBContainer wraps a MongoDB testcontainer with connection info.
type MongoDBContainer struct {
	*mongodb.MongoDBContainer
	ConnectionString string
	Host             string
	Port             string
}

// StartMongoDB creates and starts a MongoDB container.
func StartMongoDB(ctx context.Context, networkName, image string) (*MongoDBContainer, error) {
	if image == "" {
		image = "mongo:latest"
	}

	container, err := mongodb.Run(ctx,
		image,
		mongodb.WithUsername(MongoUser),
		mongodb.WithPassword(MongoPassword),
		testcontainers.WithEnv(map[string]string{
			"MONGO_INITDB_DATABASE": MongoDatabase,
		}),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Networks: []string{networkName},
				NetworkAliases: map[string][]string{
					networkName: {"mongodb", "reporter-mongodb"},
				},
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("start mongodb container: %w", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("get mongodb connection string: %w", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("get mongodb host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("get mongodb port: %w", err)
	}

	return &MongoDBContainer{
		MongoDBContainer: container,
		ConnectionString: connStr,
		Host:             host,
		Port:             mappedPort.Port(),
	}, nil
}

// Restart stops and starts the MongoDB container.
func (m *MongoDBContainer) Restart(ctx context.Context, delay time.Duration) error {
	if err := m.Stop(ctx, nil); err != nil {
		return fmt.Errorf("stop mongodb: %w", err)
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	if err := m.Start(ctx); err != nil {
		return fmt.Errorf("start mongodb: %w", err)
	}

	return nil
}
