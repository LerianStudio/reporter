// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package chaos

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/LerianStudio/reporter/v4/tests/helpers/containers"
	"github.com/LerianStudio/reporter/v4/tests/helpers/services"
)

var (
	testInfra   *containers.TestInfrastructure
	managerSvc  *services.ManagerService
	workerSvc   *services.WorkerService
	managerAddr string

	// Expose containers for chaos test manipulation
	MongoContainer    *containers.MongoDBContainer
	RabbitContainer   *containers.RabbitMQContainer
	SeaweedContainer  *containers.SeaweedFSContainer
	ValkeyContainer   *containers.ValkeyContainer
)

func TestMain(m *testing.M) {
	// Check if we should use testcontainers or existing infrastructure
	if os.Getenv("USE_EXISTING_INFRA") == "true" {
		// Use existing infrastructure (docker-compose)
		log.Println("Using existing infrastructure from docker-compose")
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("Starting test infrastructure with testcontainers for chaos tests...")

	// Start infrastructure containers
	var err error
	testInfra, err = containers.StartInfrastructure(ctx)
	if err != nil {
		log.Fatalf("Failed to start infrastructure: %v", err)
	}

	// Store container references for chaos manipulation
	MongoContainer = testInfra.MongoDB
	RabbitContainer = testInfra.RabbitMQ
	SeaweedContainer = testInfra.SeaweedFS
	ValkeyContainer = testInfra.Valkey

	log.Println("Infrastructure started successfully")

	// Create service configuration from containers
	cfg := services.NewConfigFromInfrastructure(testInfra)

	// Start Manager service
	log.Println("Starting Manager service...")
	managerSvc, err = services.StartManager(ctx, cfg)
	if err != nil {
		testInfra.Stop(ctx)
		log.Fatalf("Failed to start manager: %v", err)
	}
	managerAddr = managerSvc.Address()
	log.Printf("Manager started at %s", managerAddr)

	// Set environment variable for test helpers
	os.Setenv("MANAGER_URL", managerAddr)

	// Start Worker service
	log.Println("Starting Worker service...")
	workerSvc, err = services.StartWorker(ctx, cfg)
	if err != nil {
		managerSvc.Stop(ctx)
		testInfra.Stop(ctx)
		log.Fatalf("Failed to start worker: %v", err)
	}
	log.Println("Worker started successfully")

	// Run tests
	log.Println("Running chaos tests...")
	code := m.Run()

	// Cleanup
	log.Println("Cleaning up...")
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()

	if workerSvc != nil {
		workerSvc.Stop(cleanupCtx)
	}
	if managerSvc != nil {
		managerSvc.Stop(cleanupCtx)
	}
	if testInfra != nil {
		testInfra.Stop(cleanupCtx)
	}

	log.Println("Cleanup complete")
	os.Exit(code)
}

// GetManagerAddress returns the Manager service address for tests.
func GetManagerAddress() string {
	if managerAddr != "" {
		return managerAddr
	}
	// Fallback to environment variable or default
	if addr := os.Getenv("MANAGER_URL"); addr != "" {
		return addr
	}
	return "http://127.0.0.1:4005"
}

// RestartMongoDB restarts the MongoDB container for chaos testing.
func RestartMongoDB(delay time.Duration) error {
	if MongoContainer == nil {
		return nil // Using existing infra
	}
	return MongoContainer.Restart(context.Background(), delay)
}

// RestartRabbitMQ restarts the RabbitMQ container for chaos testing.
func RestartRabbitMQ(delay time.Duration) error {
	if RabbitContainer == nil {
		return nil // Using existing infra
	}
	return RabbitContainer.Restart(context.Background(), delay)
}

// RestartValkey restarts the Valkey/Redis container for chaos testing.
func RestartValkey(delay time.Duration) error {
	if ValkeyContainer == nil {
		return nil // Using existing infra
	}
	return ValkeyContainer.Restart(context.Background(), delay)
}

// StopMongoDB stops the MongoDB container for chaos testing.
func StopMongoDB() error {
	if MongoContainer == nil {
		return nil
	}
	return MongoContainer.Stop(context.Background(), nil)
}

// StartMongoDB starts the MongoDB container after being stopped.
func StartMongoDB() error {
	if MongoContainer == nil {
		return nil
	}
	return MongoContainer.Start(context.Background())
}

// StopRabbitMQ stops the RabbitMQ container for chaos testing.
func StopRabbitMQ() error {
	if RabbitContainer == nil {
		return nil
	}
	return RabbitContainer.Stop(context.Background(), nil)
}

// StartRabbitMQ starts the RabbitMQ container after being stopped.
func StartRabbitMQ() error {
	if RabbitContainer == nil {
		return nil
	}
	return RabbitContainer.Start(context.Background())
}
