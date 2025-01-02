package main

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/platform/application"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var config *configs.Config

func init() {
	cfg, err := configs.Load("configs/config.yml")
	if err != nil {
		log.WithError(err).Fatal("Failed to load configs")
	}
	config = cfg
}

func main() {
	app, err := application.NewApplication(config)
	if err != nil {
		log.WithError(err).Fatal("Application setup failed")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Capture shutdown signals
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownSignal
		log.Info("Shutting down application...")
		cancel()
	}()

	err = app.Run(ctx)
	if err != nil {
		log.WithError(err).Fatal("Application encountered an error")
	}

	if err := app.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error during application shutdown")
	}
	log.Info("Application terminated gracefully")
}
