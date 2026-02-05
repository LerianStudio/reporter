// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package integration

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

	log.Println("Starting test infrastructure with testcontainers...")

	// Start infrastructure containers
	var err error
	testInfra, err = containers.StartInfrastructure(ctx)
	if err != nil {
		log.Fatalf("Failed to start infrastructure: %v", err)
	}

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
	log.Println("Running integration tests...")
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
