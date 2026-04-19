package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbridges_11/document-registry/internal/bootstrap"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() (err error) {
	defer err2.Handle(&err)

	ctx := context.Background()

	// Initialize application
	app := try.To1(bootstrap.New(ctx))

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := app.Start(); err != nil {
			app.Logger.Error("server error: " + err.Error())
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	app.Logger.Info("shutdown signal received")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	try.To(app.Shutdown(shutdownCtx))

	return nil
}
