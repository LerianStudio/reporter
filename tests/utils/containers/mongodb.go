// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
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
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.PortBindings = FixedPortBindings(map[nat.Port]string{
						"27017/tcp": HostPortMongo,
					})
				},
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("start mongodb container: %w", err)
	}

	// Get host (port is fixed, no need for MappedPort)
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get mongodb host: %w", err)
	}

	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		MongoUser, MongoPassword, host, HostPortMongo, MongoDatabase)

	return &MongoDBContainer{
		MongoDBContainer: container,
		ConnectionString: connStr,
		Host:             host,
		Port:             HostPortMongo,
	}, nil
}

// Restart stops and starts the MongoDB container.
// Port mappings are fixed so they remain stable across restarts.
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
