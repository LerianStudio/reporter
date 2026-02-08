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
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	ValkeyPassword = "reporter-pass"
)

// ValkeyContainer wraps a Valkey/Redis testcontainer with connection info.
type ValkeyContainer struct {
	*redis.RedisContainer
	Address  string
	Host     string
	Port     string
	Password string
}

// StartValkey creates and starts a Valkey container.
func StartValkey(ctx context.Context, networkName, image string) (*ValkeyContainer, error) {
	if image == "" {
		image = "valkey/valkey:latest"
	}

	container, err := redis.Run(ctx,
		image,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Networks: []string{networkName},
				NetworkAliases: map[string][]string{
					networkName: {"valkey", "redis", "reporter-valkey"},
				},
				Cmd: []string{"redis-server", "--requirepass", ValkeyPassword},
				HostConfigModifier: func(hc *container.HostConfig) {
					hc.PortBindings = FixedPortBindings(map[nat.Port]string{
						"6379/tcp": HostPortValkey,
					})
				},
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("start valkey container: %w", err)
	}

	// Get host (port is fixed, no need for MappedPort)
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get valkey host: %w", err)
	}

	address := fmt.Sprintf("redis://%s:%s", host, HostPortValkey)

	return &ValkeyContainer{
		RedisContainer: container,
		Address:        address,
		Host:           host,
		Port:           HostPortValkey,
		Password:       ValkeyPassword,
	}, nil
}

// Restart stops and starts the Valkey container, refreshing connection info.
// Port mappings are fixed so they remain stable across restarts.
func (v *ValkeyContainer) Restart(ctx context.Context, delay time.Duration) error {
	if err := v.Stop(ctx, nil); err != nil {
		return fmt.Errorf("stop valkey: %w", err)
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	if err := v.Start(ctx); err != nil {
		return fmt.Errorf("start valkey: %w", err)
	}

	// Host may change after restart
	host, err := v.RedisContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("refresh valkey host: %w", err)
	}

	v.Host = host
	v.Address = fmt.Sprintf("redis://%s:%s", host, HostPortValkey)
	// Ports are fixed - no need to re-read mapped ports

	return nil
}
