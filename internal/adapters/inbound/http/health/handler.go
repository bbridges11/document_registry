package health

import (
	"context"
	"time"

	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HealthChecker is a function that checks the health of a dependency
type HealthChecker func(context.Context) error

// Handler handles health and readiness endpoints
type Handler struct {
	runner      *dbPostgres.Runner
	authz       outbound.AuthorizationService
	userService outbound.UserService
	logger      *zap.Logger
}

func NewHandler(
	runner *dbPostgres.Runner,
	authz outbound.AuthorizationService,
	userService outbound.UserService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		runner:      runner,
		authz:       authz,
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers health-related routes
func (h *Handler) RegisterRoutes(e *echo.Echo, middleWareFuncs ...echo.MiddlewareFunc) {
	e.GET("/health", h.Health)
	e.GET("/live", h.Liveness)
	e.GET("/ready", h.Readiness)
}

// Health returns basic health status
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "ok",
	})
}

// Liveness returns liveness probe status
func (h *Handler) Liveness(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "alive",
	})
}

// Readiness returns readiness probe with dependency checks
func (h *Handler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if err := h.checkPostgres(ctx); err != nil {
		results["postgres"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		results["postgres"] = "healthy"
	}

	// // Check User Service
	// if err := h.userService.HealthCheck(ctx); err != nil {
	// 	results["user_service"] = "unhealthy: " + err.Error()
	// 	allHealthy = false
	// } else {
	// 	results["user_service"] = "healthy"
	// }

	// Check OpenFGA
	if err := h.authz.HealthCheck(ctx); err != nil {
		results["openfga"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		results["openfga"] = "healthy"
	}

	if allHealthy {
		return c.JSON(200, map[string]any{
			"status": "ready",
			"checks": results,
		})
	}

	return c.JSON(503, map[string]any{
		"status": "not ready",
		"checks": results,
	})
}

func (h *Handler) checkPostgres(ctx context.Context) error {
	return h.runner.Pool().Ping(ctx)
}
