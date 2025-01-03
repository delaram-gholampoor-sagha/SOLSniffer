package transaction

import (
	"context"
	"fmt"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/models/entity"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionRepository struct {
	collection *mongo.Collection
}

func NewTransactionRepository(db *mongo.Client) *TransactionRepository {
	return &TransactionRepository{
		collection: db.Database("solsniffer").Collection("transactions"),
	}
}

func (r *TransactionRepository) Save(ctx context.Context, transaction *entity.Transaction) error {
	_, err := r.collection.InsertOne(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to save transaction: %v", err)
	}
	return nil
}
