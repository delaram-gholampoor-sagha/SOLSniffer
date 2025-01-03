package main

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/platform/application"
	"os"
	"os/signal"
	"syscall"
)

var config *configs.Config

func init() {
	cfg, err := configs.Load("configs/config.yml")
	if err != nil {
		log.Fatalf("Failed to load configs")
	}
	config = cfg
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	app, err := application.NewApplication(ctx, config)
	if err != nil {
		log.Fatalf("Application setup failed")
	}

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownSignal
		log.Infof("Shutting down application...")
		cancel()
	}()

	err = app.Run(ctx)
	if err != nil {
		log.Fatalf("Application encountered an error")
	}

	if err := app.Shutdown(ctx); err != nil {
		log.Errorf("Error during application shutdown")
	}
	log.Infof("Application terminated gracefully")
}
