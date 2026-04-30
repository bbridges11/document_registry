package bootstrap

import (
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/approval"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/deprecation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/document"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/middleware"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/stakeholder"
	userhttp "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/user"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/validation"
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/version"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/bbridges_11/document-registry/internal/platform/server"
	"go.uber.org/zap"
)

func wireHTTP(cfg *config.Config, log *zap.Logger, _ *infrastructure, handlers *httpHandlers) *server.HTTPServer {
	if cfg.Server.Driver != config.ServerDriverHTTP {
		return nil
	}

	log.Info("initializing server", zap.String("driver", string(cfg.Server.Driver)))
	httpServer := server.NewHTTPServer(cfg.Server.HTTP, log)
	e := httpServer.Echo()

	userValidation := middleware.NewUserValidationMiddleware(handlers.UserQueries)
	userValidationMW := userValidation.ValidateUserExists()

	adminCheck := middleware.NewAdminCheckMiddleware(handlers.UserQueries)
	adminOnlyMW := adminCheck.RequireAdmin()

	log.Info("registering HTTP handlers")

	handlers.Health.RegisterRoutes(e)
	handlers.Document.RegisterRoutes(e, &document.MiddlewareConfig{UserValidation: userValidationMW})
	handlers.Version.RegisterRoutes(e, &version.MiddlewareConfig{UserValidation: userValidationMW})
	handlers.Stakeholder.RegisterRoutes(e, &stakeholder.MiddlewareConfig{UserValidation: userValidationMW})
	handlers.Approval.RegisterRoutes(e, &approval.MiddlewareConfig{UserValidation: userValidationMW})
	handlers.Deprecation.RegisterRoutes(e, &deprecation.MiddlewareConfig{UserValidation: userValidationMW})
	handlers.User.RegisterRoutes(e, &userhttp.MiddlewareConfig{UserValidation: userValidationMW, AdminOnly: adminOnlyMW})
	handlers.Validation.RegisterRoutes(e, &validation.MiddlewareConfig{UserValidation: userValidationMW})

	// Publication routes - admin only
	adminGroup := e.Group("", userValidationMW, adminOnlyMW)
	handlers.Publication.RegisterRoutes(adminGroup)

	// Subscription routes - admin only
	handlers.Subscription.RegisterRoutes(adminGroup)

	return httpServer
}
