package transaction

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type TransactionRepository struct {
	collection *mongo.Collection
}

func NewTransactionRepository(db *mongo.Client) *TransactionRepository {
	return &TransactionRepository{
		collection: db.Database("solsniffer").Collection("transactions"),
	}
}

func (r *TransactionRepository) Save(ctx context.Context, hash, source, destination string, amount float64, token string, timestamp time.Time) error {
	doc := bson.M{
		"hash":        hash,
		"source":      source,
		"destination": destination,
		"amount":      amount,
		"token_mint":  token,
		"timestamp":   timestamp,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save transaction: %v", err)
	}
	return nil
}
