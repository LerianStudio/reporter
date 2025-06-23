package bootstrap

import (
	"github.com/LerianStudio/lib-commons/commons/log"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/LerianStudio/lib-commons/commons/shutdown"
	libLicense "github.com/LerianStudio/lib-license-go/middleware"
	"github.com/gofiber/fiber/v2"
	"plugin-smart-templates/pkg"
)

// Server represents the http server for Ledger services.
type Server struct {
	app           *fiber.App
	serverAddress string
	license       *shutdown.LicenseManagerShutdown
	logger        log.Logger
	telemetry     opentelemetry.Telemetry
}

// ServerAddress returns is a convenience method to return the server address.
func (s *Server) ServerAddress() string {
	return s.serverAddress
}

// NewServer creates an instance of Server.
func NewServer(cfg *Config, app *fiber.App, logger log.Logger, telemetry *opentelemetry.Telemetry, licenseClient *libLicense.LicenseClient) *Server {
	return &Server{
		app:           app,
		serverAddress: cfg.ServerAddress,
		license:       licenseClient.GetLicenseManagerShutdown(),
		logger:        logger,
		telemetry:     *telemetry,
	}
}

// Run runs the server.
func (s *Server) Run(l *pkg.Launcher) error {
	shutdown.StartServerWithGracefulShutdown(
		s.app,
		s.license,
		&s.telemetry,
		s.ServerAddress(),
		s.logger,
	)

	return nil
}
