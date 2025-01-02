package transactionMonitorCoordinator

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/enums"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/webSocket"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	webSocketManager *webSocket.Manager
	service          *transactionMonitor.Service
}

func New(service *transactionMonitor.Service, webSocketManager *webSocket.Manager) (*Service, error) {
	return &Service{
		webSocketManager: webSocketManager,
		service:          service,
	}, nil
}

func (c *Service) Start(ctx context.Context) error {
	log.Info("Starting transaction monitor coordinator...")

	// Subscribe to logs
	subscriptionID, err := c.webSocketManager.Subscribe(ctx, enums.LogsSubscribe)
	if err != nil {
		return err
	}
	log.Infof("Subscribed to logs with subscription ID: %s", subscriptionID)

	// Run message listening in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("Transaction monitor coordinator stopped")
				_ = c.webSocketManager.Unsubscribe(ctx, enums.LogsUnsubscribe)
				return
			default:
				message, err := c.webSocketManager.ReadMessage()
				if err != nil {
					log.WithError(err).Error("Error reading WebSocket message")
					continue
				}
				if err := c.service.ProcessMessage(ctx, message); err != nil {
					log.WithError(err).Error("Failed to process WebSocket message")
				}
			}
		}
	}()
	return nil
}

func (c *Service) Stop(ctx context.Context) error {
	log.Info("Stopping transaction monitor coordinator...")

	if err := c.webSocketManager.Unsubscribe(ctx, enums.LogsUnsubscribe); err != nil {
		log.WithError(err).Error("Failed to unsubscribe from logs")
		return err
	}

	if err := c.webSocketManager.Close(); err != nil {
		log.WithError(err).Error("Failed to close WebSocket connection")
		return err
	}

	log.Info("Transaction monitor coordinator stopped successfully")
	return nil
}
