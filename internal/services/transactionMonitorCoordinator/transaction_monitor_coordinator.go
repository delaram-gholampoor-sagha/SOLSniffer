package transactionMonitorCoordinator

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/enums"
	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/webSocket"
)

type Service struct {
	webSocketManager *webSocket.Manager
	service          *transactionMonitor.Service
}

func New(service *transactionMonitor.Service, webSocketManager *webSocket.Manager) *Service {
	return &Service{
		webSocketManager: webSocketManager,
		service:          service,
	}
}

func (c *Service) Start(ctx context.Context) error {
	log.Infof("Starting transaction monitor coordinator...")

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
				log.Infof("Transaction monitor coordinator stopped")
				_ = c.webSocketManager.Unsubscribe(ctx, enums.LogsUnsubscribe)
				return
			default:
				message, err := c.webSocketManager.ReadMessage()
				if err != nil {
					log.Errorf("Error reading WebSocket message")
					continue
				}
				if err := c.service.ProcessMessage(ctx, message); err != nil {
					log.Errorf("Failed to process WebSocket message")
				}
			}
		}
	}()
	return nil
}

func (c *Service) Stop(ctx context.Context) error {
	log.Infof("Stopping transaction monitor coordinator...")

	if err := c.webSocketManager.Unsubscribe(ctx, enums.LogsUnsubscribe); err != nil {
		log.Errorf("Failed to unsubscribe from logs")
		return err
	}

	if err := c.webSocketManager.Close(); err != nil {
		log.Errorf("Failed to close WebSocket connection")
		return err
	}

	log.Infof("Transaction monitor coordinator stopped successfully")
	return nil
}
