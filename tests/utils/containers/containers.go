// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// Fixed host ports matching production (components/infra/.env.example).
// Using fixed ports ensures containers survive restart without changing
// the address the Manager process connects to.
const (
	HostPortRabbitAMQP = "3005"
	HostPortRabbitMgmt = "3006"
	HostPortMongo      = "5708"
	HostPortValkey     = "5705"
	HostPortSeaweedS3  = "8333"
	HostPortSeaweedAdm = "9333"
)

// FixedPortBindings creates a nat.PortMap from container-port -> host-port pairs.
func FixedPortBindings(mappings map[nat.Port]string) nat.PortMap {
	pm := nat.PortMap{}
	for containerPort, hostPort := range mappings {
		pm[containerPort] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: hostPort},
		}
	}

	return pm
}

// TestInfrastructure holds all test containers and provides connection information.
type TestInfrastructure struct {
	MongoDB   *MongoDBContainer
	RabbitMQ  *RabbitMQContainer
	SeaweedFS *SeaweedFSContainer
	Valkey    *ValkeyContainer

	network *testcontainers.DockerNetwork
	mu      sync.Mutex
}

// InfrastructureConfig holds configuration for container startup.
type InfrastructureConfig struct {
	MongoImage   string
	RabbitImage  string
	SeaweedImage string
	ValkeyImage  string
	NetworkName  string
	StartTimeout time.Duration
}

// DefaultConfig returns default configuration for test infrastructure.
func DefaultConfig() *InfrastructureConfig {
	return &InfrastructureConfig{
		MongoImage:   "mongo:latest",
		RabbitImage:  "rabbitmq:4.0-management-alpine",
		SeaweedImage: "chrislusf/seaweedfs:3.97",
		ValkeyImage:  "valkey/valkey:latest",
		NetworkName:  "reporter-test-network",
		StartTimeout: 120 * time.Second,
	}
}

// StartInfrastructure starts all required containers for testing.
// Containers are started in parallel for faster startup.
func StartInfrastructure(ctx context.Context) (*TestInfrastructure, error) {
	return StartInfrastructureWithConfig(ctx, DefaultConfig())
}

// StartInfrastructureWithConfig starts all containers with custom configuration.
func StartInfrastructureWithConfig(ctx context.Context, cfg *InfrastructureConfig) (*TestInfrastructure, error) {
	// Create network for container communication
	net, err := network.New(ctx,
		network.WithDriver("bridge"),
	)
	if err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}

	networkName := net.Name

	infra := &TestInfrastructure{
		network: net,
	}

	// Start containers in parallel
	var wg sync.WaitGroup

	errCh := make(chan error, 4)

	// MongoDB
	wg.Add(1)

	go func() {
		defer wg.Done()

		mongo, err := StartMongoDB(ctx, networkName, cfg.MongoImage)
		if err != nil {
			errCh <- fmt.Errorf("mongodb: %w", err)
			return
		}

		infra.mu.Lock()
		infra.MongoDB = mongo
		infra.mu.Unlock()
	}()

	// RabbitMQ
	wg.Add(1)

	go func() {
		defer wg.Done()

		rabbit, err := StartRabbitMQ(ctx, networkName, cfg.RabbitImage)
		if err != nil {
			errCh <- fmt.Errorf("rabbitmq: %w", err)
			return
		}

		infra.mu.Lock()
		infra.RabbitMQ = rabbit
		infra.mu.Unlock()
	}()

	// SeaweedFS
	wg.Add(1)

	go func() {
		defer wg.Done()

		seaweed, err := StartSeaweedFS(ctx, networkName, cfg.SeaweedImage)
		if err != nil {
			errCh <- fmt.Errorf("seaweedfs: %w", err)
			return
		}

		infra.mu.Lock()
		infra.SeaweedFS = seaweed
		infra.mu.Unlock()
	}()

	// Valkey
	wg.Add(1)

	go func() {
		defer wg.Done()

		valkey, err := StartValkey(ctx, networkName, cfg.ValkeyImage)
		if err != nil {
			errCh <- fmt.Errorf("valkey: %w", err)
			return
		}

		infra.mu.Lock()
		infra.Valkey = valkey
		infra.mu.Unlock()
	}()

	// Wait for all containers to start
	wg.Wait()
	close(errCh)

	// Check for errors
	errs := make([]error, 0, 4)
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		// Cleanup any started containers
		_ = infra.Stop(ctx)
		return nil, fmt.Errorf("failed to start containers: %v", errs)
	}

	return infra, nil
}

// Stop terminates all containers and cleans up resources.
func (i *TestInfrastructure) Stop(ctx context.Context) error {
	var errs []error

	if i.MongoDB != nil {
		if err := i.MongoDB.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("mongodb terminate: %w", err))
		}
	}

	if i.RabbitMQ != nil {
		if err := i.RabbitMQ.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("rabbitmq terminate: %w", err))
		}
	}

	if i.SeaweedFS != nil {
		if err := i.SeaweedFS.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("seaweedfs terminate: %w", err))
		}
	}

	if i.Valkey != nil {
		if err := i.Valkey.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("valkey terminate: %w", err))
		}
	}

	if i.network != nil {
		if err := i.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("network remove: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// ConnectionConfig returns all connection strings for services.
type ConnectionConfig struct {
	MongoURI       string
	MongoHost      string
	MongoPort      string
	RabbitURL      string
	RabbitHost     string
	RabbitPort     string
	RabbitMgmtPort string
	S3Endpoint     string
	S3Host         string
	S3Port         string
	RedisHost      string
	RedisPort      string
	RedisAddr      string
}

// GetConnectionConfig returns connection configuration for all services.
func (i *TestInfrastructure) GetConnectionConfig() *ConnectionConfig {
	cfg := &ConnectionConfig{}

	if i.MongoDB != nil {
		cfg.MongoURI = i.MongoDB.ConnectionString
		cfg.MongoHost = i.MongoDB.Host
		cfg.MongoPort = i.MongoDB.Port
	}

	if i.RabbitMQ != nil {
		cfg.RabbitURL = i.RabbitMQ.AmqpURL
		cfg.RabbitHost = i.RabbitMQ.Host
		cfg.RabbitPort = i.RabbitMQ.AmqpPort
		cfg.RabbitMgmtPort = i.RabbitMQ.MgmtPort
	}

	if i.SeaweedFS != nil {
		cfg.S3Endpoint = i.SeaweedFS.S3Endpoint
		cfg.S3Host = i.SeaweedFS.Host
		cfg.S3Port = i.SeaweedFS.S3Port
	}

	if i.Valkey != nil {
		cfg.RedisHost = i.Valkey.Host
		cfg.RedisPort = i.Valkey.Port
		cfg.RedisAddr = i.Valkey.Address
	}

	return cfg
}
