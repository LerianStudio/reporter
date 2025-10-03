package helpers

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
	RabbitContainer string
	MongoContainer  string
	RedisContainer  string
	MinioContainer  string

	// Domain/testing context
	DefaultOrgID string
}

func LoadEnvironment() Environment {
	mgr := getenvDefault("MANAGER_URL", "http://localhost:4005")
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

		RabbitContainer: getenvDefault("RABBIT_CONTAINER", "plugin-smart-templates-rabbitmq"),
		MongoContainer:  getenvDefault("MONGO_CONTAINER", "plugin-smart-templates-mongodb"),
		RedisContainer:  getenvDefault("REDIS_CONTAINER", "plugin-smart-templates-valkey"),
		MinioContainer:  getenvDefault("MINIO_CONTAINER", "plugin-smart-templates-minio"),

		DefaultOrgID: firstNonEmpty(
			os.Getenv("X_ORGANIZATION_ID"),
			os.Getenv("ORGANIZATION_ID"),
			os.Getenv("ORG_ID"),
		),
	}

	return env
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}

	return ""
}
