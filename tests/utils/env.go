// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package utils

import (
	"os"
	"strconv"
	"time"
)

type Environment struct {
	ManagerURL  string
	HTTPTimeout time.Duration
	ManageStack bool

	// Optional infra identifiers for chaos
	RabbitContainer    string
	WorkerContainer    string
	MongoContainer     string
	RedisContainer     string
	SeaweedFSContainer string
}

func LoadEnvironment() Environment {
	mgr := getenvDefault("MANAGER_URL", "http://127.0.0.1:4005")
	timeoutStr := getenvDefault("HTTP_TIMEOUT_SECS", "30")

	secs, _ := strconv.Atoi(timeoutStr)
	if secs <= 0 {
		secs = 30
	}

	manage := getenvDefault("MANAGE_STACK", "false") == "true"

	env := Environment{
		ManagerURL:  mgr,
		HTTPTimeout: time.Duration(secs) * time.Second,
		ManageStack: manage,

		RabbitContainer:    getenvDefault("RABBIT_CONTAINER", "reporter-rabbitmq"),
		WorkerContainer:    getenvDefault("WORKER_CONTAINER", "reporter-worker"),
		MongoContainer:     getenvDefault("MONGO_CONTAINER", "reporter-mongodb"),
		RedisContainer:     getenvDefault("REDIS_CONTAINER", "reporter-valkey"),
		SeaweedFSContainer: getenvDefault("SEAWEEDFS_CONTAINER", "reporter-seaweedfs-filer"),
	}

	return env
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}
