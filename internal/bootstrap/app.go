package bootstrap

import (
	"context"
	"log"

	"github.com/lainio/err2"
	"go.uber.org/zap"

	"github.com/bbridges_11/document-registry/internal/platform/config"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	platformEvents "github.com/bbridges_11/document-registry/internal/platform/events"
	"github.com/bbridges_11/document-registry/internal/platform/server"
)

type App struct {
	Config     *config.Config
	Logger     *zap.Logger
	HTTPServer *server.HTTPServer
	EventBus   *platformEvents.Bus
	Runner     *dbPostgres.Runner
}

func New(ctx context.Context) (app *App, err error) {
	defer err2.Handle(&err, func(err error) error {
		log.Println(err.Error())
		return err
	})

	cfg, logger := wireConfigAndLogger()
	infra := wireInfrastructure(ctx, cfg, logger)
	handlers := wireApplication(ctx, cfg, logger, infra)
	httpServer := wireHTTP(cfg, logger, infra, handlers)

	logger.Info("application initialized successfully")

	return &App{
		Config:     cfg,
		Logger:     logger,
		HTTPServer: httpServer,
		EventBus:   infra.EventBus,
		Runner:     infra.Runner,
	}, nil
}

func (a *App) Start() (err error) {
	defer err2.Handle(&err)

	a.Logger.Info("starting application")

	if a.Config.Server.Driver == config.ServerDriverHTTP {
		return a.HTTPServer.Start()
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) (err error) {
	defer err2.Handle(&err)

	a.Logger.Info("shutting down application")

	if a.EventBus != nil {
		if stopErr := a.EventBus.Stop(); stopErr != nil {
			return stopErr
		}
	}

	if a.HTTPServer != nil {
		if shutdownErr := a.HTTPServer.Shutdown(ctx); shutdownErr != nil {
			return shutdownErr
		}
	}

	if a.Runner != nil {
		a.Runner.Pool().Close()
	}

	a.Logger.Info("application shutdown complete")
	return nil
}
