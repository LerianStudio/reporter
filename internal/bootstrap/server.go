package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/log"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
)

// Server represents the http server for Ledger services.
type Server struct {
	app           *fiber.App
	serverAddress string
	log.Logger
	opentelemetry.Telemetry
}

// ServerAddress returns is a convenience method to return the server address.
func (s *Server) ServerAddress() string {
	return s.serverAddress
}

// NewServer creates an instance of Server.
func NewServer(cfg *Config, app *fiber.App, logger log.Logger, telemetry *opentelemetry.Telemetry) *Server {
	return &Server{
		app:           app,
		serverAddress: cfg.ServerAddress,
		Logger:        logger,
		Telemetry:     *telemetry,
	}
}

// Run runs the server.
func (s *Server) Run(l *pkg.Launcher) error {
	s.InitializeTelemetry(s.Logger)
	defer s.ShutdownTelemetry()

	defer func() {
		if err := s.Logger.Sync(); err != nil {
			s.Logger.Fatalf("Failed to sync logger: %s", err)
		}
	}()

	err := s.app.Listen(s.ServerAddress())
	if err != nil {
		return errors.Wrap(err, "failed to run the server")
	}

	return nil
}
