package repositories

import (
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/repositories/transaction"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repositories struct {
	TransactionRepository *transaction.TransactionRepository
}

func NewRepositories(db *mongo.Client) *Repositories {
	return &Repositories{
		TransactionRepository: transaction.NewTransactionRepository(db),
	}
}
