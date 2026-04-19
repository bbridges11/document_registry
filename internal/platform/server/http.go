package server

import (
	"context"
	"fmt"

	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	echo   *echo.Echo
	config config.HTTPConfig
	logger *zap.Logger
}

func NewHTTPServer(cfg config.HTTPConfig, logger *zap.Logger) *HTTPServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	return &HTTPServer{
		echo:   e,
		config: cfg,
		logger: logger,
	}
}

func (s *HTTPServer) Echo() *echo.Echo {
	return s.echo
}

func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info("starting HTTP server", zap.String("address", addr))
	return s.echo.Start(addr)
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.echo.Shutdown(ctx)
}
