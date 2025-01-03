package application

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	repositoriescontracts "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/contracts/repositories"

	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/contracts/services"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/platform/monitoring"

	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"

	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/repositories/transaction"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/tokenTransactionProcessor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitorCoordinator"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/webSocket"

	"github.com/prometheus/client_golang/prometheus"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type App struct {
	config *configs.Config

	Monitoring         map[string]services.Monitoring
	MonitoringRegistry *prometheus.Registry

	Client struct {
		WebSocketManager *webSocket.Manager
	}

	Services struct {
		TokenProcessor                *tokenTransactionProcessor.Service
		TransactionMonitor            *transactionMonitor.Service
		TransactionMonitorCoordinator *transactionMonitorCoordinator.Service
	}

	Repositories struct {
		Transaction repositoriescontracts.Transaction
	}

	Database struct {
		Mongo *mongo.Client
	}
}

func NewApplication(ctx context.Context, config *configs.Config) (*App, error) {
	app := &App{
		config: config,
	}

	err := app.setupLogger()
	if err != nil {
		return nil, err
	}

	// Register Database
	if err := app.registerDatabase(); err != nil {
		return nil, err
	}

	// Register Repositories
	app.registerRepositories()

	app.registerMonitoring()

	// Register TokenTransactionProcessor Service
	app.registerTokenTransactionProcessor()

	// Register TransactionMonitor Service
	app.registerTransactionMonitor()

	// Register TransactionMonitorCoordinator Service
	if err := app.registerTransactionMonitorCoordinator(); err != nil {
		return nil, err
	}

	// Register WebSocket Manager
	if err := app.registerWebSocketManager(); err != nil {
		return nil, err
	}

	go app.monitorServices(ctx)

	return app, nil
}

func (a *App) registerDatabase() error {
	err := retry.Do(
		func() error {
			db, err := mongo.Connect(context.Background(), options.Client().ApplyURI(a.config.Database.URI))
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.Ping(ctx, nil); err != nil {
				return err
			}

			a.Database.Mongo = db
			return nil
		},
		retry.Attempts(3),                   // Retry up to 3 times
		retry.Delay(2*time.Second),          // Fixed delay between attempts
		retry.DelayType(retry.BackOffDelay), // Exponential backoff
		retry.OnRetry(func(n uint, err error) {
			log.Warnf("Retrying database connection (attempt %d)", n+1)
		}),
	)

	if err != nil {
		log.Errorf("Failed to connect to MongoDB after retries")
		return err
	}

	log.Infof("Connected to MongoDB")
	return nil
}

func (a *App) registerRepositories() {
	a.Repositories.Transaction = transaction.NewTransactionRepository(a.Database.Mongo)
	log.Infof("Repositories registered")
}

func (a *App) registerTokenTransactionProcessor() {
	a.Services.TokenProcessor = tokenTransactionProcessor.New(
		a.Repositories.Transaction,
		a.config.Services.Tokens,
		a.config.Services.Wallets,
	)
	log.Infof("Token Transaction Processor service registered")
}

func (a *App) registerTransactionMonitor() {
	transactionMonitor := transactionMonitor.New(a.Services.TokenProcessor)

	a.Services.TransactionMonitor = transactionMonitor
	log.Infof("Transaction Monitor service registered")

}

func (a *App) registerTransactionMonitorCoordinator() error {
	coordinator, err := transactionMonitorCoordinator.New(
		a.Services.TransactionMonitor,
		a.Client.WebSocketManager,
	)
	if err != nil {
		log.Errorf("Failed to initialize transaction monitor coordinator")
		return err
	}

	a.Services.TransactionMonitorCoordinator = coordinator
	log.Infof("Transaction Monitor Coordinator service registered")
	return nil
}

func (a *App) registerWebSocketManager() error {

	err := retry.Do(
		func() error {
			manager, err := webSocket.New(a.config.WebSocket.Scheme, a.config.WebSocket.Host, a.config.WebSocket.Path)
			if err != nil {
				return err
			}

			a.Client.WebSocketManager = manager
			return nil
		},
		retry.Attempts(5),                 // Retry up to 5 times
		retry.Delay(1*time.Second),        // Fixed delay of 1 second
		retry.DelayType(retry.FixedDelay), // Use a fixed delay between attempts
		retry.OnRetry(func(n uint, err error) {
			log.Warnf("Retrying WebSocket manager initialization (attempt %d)", n+1)
		}),
	)

	if err != nil {
		log.Errorf("Failed to initialize WebSocket manager after retries")
		return err
	}

	log.Infof("WebSocket Manager service registered")
	return nil
}

func (a *App) monitorServices(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("Stopping service health monitoring...")
			return
		case <-ticker.C:
			if err := a.Database.Mongo.Ping(ctx, nil); err != nil {
				log.Warnf("MongoDB is not reachable; attempting reconnection...")
				_ = a.registerDatabase()
			}

			if !a.Client.WebSocketManager.IsConnected() {
				log.Warnf("WebSocket disconnected; attempting reconnection...")
				_ = a.registerWebSocketManager()
			}
		}
	}
}

const (
	AppMonitoring = "app_monitoring"
)

func (a *App) registerMonitoring() {
	a.Monitoring[AppMonitoring] = monitoring.NewPrometheusAppMonitor()

	registry := prometheus.NewRegistry()
	for _, m := range a.Monitoring {
		registry.MustRegister(m.GetRegistry())
	}

	a.MonitoringRegistry = registry
}
func (a *App) setupLogger() error {
	err := log.Register(a.config.App)
	if err != nil {
		return err
	}
	log.Infof("Logger registered successfully")
	return nil

}

func (a *App) Run(ctx context.Context) error {
	log.Infof("Starting application...")

	err := retry.Do(
		func() error {
			if err := a.Services.TransactionMonitorCoordinator.Start(ctx); err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(3),                   // Retry up to 3 times
		retry.Delay(2*time.Second),          // Delay between retries
		retry.DelayType(retry.BackOffDelay), // Exponential backoff
		retry.OnRetry(func(n uint, err error) {
			log.Warnf("Retrying TransactionMonitorCoordinator start (attempt %d)", n+1)
		}),
	)

	if err != nil {
		log.Errorf("Failed to start TransactionMonitorCoordinator after retries")
		return err
	}

	log.Infof("Transaction monitor coordinator started")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Infof("Shutting down application...")

	if err := a.Services.TransactionMonitorCoordinator.Stop(ctx); err != nil {
		log.Errorf("Failed to stop transaction monitor coordinator")
		return err
	}

	if err := a.Database.Mongo.Disconnect(ctx); err != nil {
		log.Errorf("Failed to disconnect from MongoDB")
		return err
	}

	log.Infof("Application shutdown completed successfully")
	return nil
}
