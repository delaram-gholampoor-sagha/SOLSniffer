package application

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	repositoriescontracts "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/contracts/repositories"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/repositories/transaction"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/tokenTransactionProcessor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/transactionMonitorCoordinator"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/webSocket"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type App struct {
	config *configs.Config
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

func NewApplication(config *configs.Config) (*App, error) {
	app := &App{
		config: config,
	}

	// Register Database
	if err := app.registerDatabase(); err != nil {
		return nil, err
	}

	// Register Repositories
	app.registerRepositories()

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

	return app, nil
}

func (a *App) registerDatabase() error {
	err := utils.Retry(context.Background(), func() error {
		db, err := mongo.Connect(context.Background(), options.Client().ApplyURI(a.config.MongoURI))
		if err != nil {
			return err
		}
		// Ping the database to ensure connectivity
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Ping(ctx, nil); err != nil {
			return err
		}
		a.Database.Mongo = db
		return nil
	},
		utils.WithMaxAttempts(3),
		utils.WithDelay(2*time.Second),
		utils.WithBackoff(func(attempt int) time.Duration {
			return time.Duration(attempt) * 2 * time.Second // Exponential backoff
		}),
	)

	if err != nil {
		log.WithError(err).Error("Failed to connect to MongoDB after retries")
		return err
	}

	log.Info("Connected to MongoDB")
	return nil
}

func (a *App) registerRepositories() {
	a.Repositories.Transaction = transaction.NewTransactionRepository(a.Database.Mongo)
	log.Info("Repositories registered")
}

func (a *App) registerTokenTransactionProcessor() {
	a.Services.TokenProcessor = tokenTransactionProcessor.New(
		a.Repositories.Transaction,
		a.config.Tokens,
		a.config.Wallets,
	)
	log.Info("Token Transaction Processor service registered")
}

func (a *App) registerTransactionMonitor() {
	transactionMonitor := transactionMonitor.New(a.Services.TokenProcessor)

	a.Services.TransactionMonitor = transactionMonitor
	log.Info("Transaction Monitor service registered")

}

func (a *App) registerTransactionMonitorCoordinator() error {
	coordinator, err := transactionMonitorCoordinator.New(
		a.Services.TransactionMonitor,
		a.Client.WebSocketManager,
	)
	if err != nil {
		log.WithError(err).Error("Failed to initialize transaction monitor coordinator")
		return err
	}

	a.Services.TransactionMonitorCoordinator = coordinator
	log.Info("Transaction Monitor Coordinator service registered")
	return nil
}

func (a *App) registerWebSocketManager() error {
	err := utils.Retry(context.Background(), func() error {
		manager, err := webSocket.New(a.config.WebSocketScheme, a.config.WebSocketHost, a.config.WebSocketPath)
		if err != nil {
			return err
		}
		a.Client.WebSocketManager = manager
		return nil
	},
		utils.WithMaxAttempts(5),
		utils.WithDelay(1*time.Second),
		utils.WithBackoff(func(attempt int) time.Duration {
			return time.Duration(attempt) * 500 * time.Millisecond // Linear backoff
		}),
	)

	if err != nil {
		log.WithError(err).Error("Failed to initialize WebSocket manager after retries")
		return err
	}

	log.Info("WebSocket Manager service registered")
	return nil
}

func (a *App) Run(ctx context.Context) error {
	log.Info("Starting application...")

	// Start TransactionMonitorCoordinator
	if err := a.Services.TransactionMonitorCoordinator.Start(ctx); err != nil {
		log.WithError(err).Error("Error while starting transaction monitor coordinator")
		return err
	}

	log.Info("Transaction monitor coordinator started")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Info("Shutting down application...")

	// Stop TransactionMonitorCoordinator
	if err := a.Services.TransactionMonitorCoordinator.Stop(ctx); err != nil {
		log.WithError(err).Error("Failed to stop transaction monitor coordinator")
		return err
	}

	// Disconnect Database
	if err := a.Database.Mongo.Disconnect(ctx); err != nil {
		log.WithError(err).Error("Failed to disconnect from MongoDB")
		return err
	}

	log.Info("Application shutdown completed successfully")
	return nil
}
