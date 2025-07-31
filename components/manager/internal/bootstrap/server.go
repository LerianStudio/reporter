package bootstrap

import (
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libCommonsLicense "github.com/LerianStudio/lib-commons/v2/commons/license"
	libCommonsLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	libCommonsOtel "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	libCommonsServer "github.com/LerianStudio/lib-commons/v2/commons/server"
	libLicense "github.com/LerianStudio/lib-license-go/v2/middleware"
	"github.com/gofiber/fiber/v2"
)

// Server represents the http server for Ledger services.
type Server struct {
	app           *fiber.App
	serverAddress string
	license       *libCommonsLicense.ManagerShutdown
	logger        libCommonsLog.Logger
	telemetry     libCommonsOtel.Telemetry
}

// ServerAddress returns is a convenience method to return the server address.
func (s *Server) ServerAddress() string {
	return s.serverAddress
}

// NewServer creates an instance of Server.
func NewServer(cfg *Config, app *fiber.App, logger libCommonsLog.Logger, telemetry *libCommonsOtel.Telemetry, licenseClient *libLicense.LicenseClient) *Server {
	return &Server{
		app:           app,
		serverAddress: cfg.ServerAddress,
		license:       licenseClient.GetLicenseManagerShutdown(),
		logger:        logger,
		telemetry:     *telemetry,
	}
}

// Run runs the server.
func (s *Server) Run(l *libCommons.Launcher) error {
	libCommonsServer.NewServerManager(s.license, &s.telemetry, s.logger).
		WithHTTPServer(s.app, s.serverAddress).
		StartWithGracefulShutdown()

	return nil
}
