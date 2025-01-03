package transaction

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MetadataRepository struct {
	collection *mongo.Collection
}

func NewMetadataRepository(db *mongo.Client) *MetadataRepository {
	return &MetadataRepository{
		collection: db.Database("solsniffer").Collection("metadata"),
	}
}

// GetLastProcessedBlock retrieves the last processed block height from the metadata collection.
func (r *MetadataRepository) GetLastProcessedBlock(ctx context.Context) (int64, error) {
	var result struct {
		BlockHeight int64 `bson:"block_height"`
	}
	err := r.collection.FindOne(ctx, bson.M{"_id": "last_processed_block"}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil // Return 0 if no document is found (meaning no block has been processed yet)
		}
		return 0, fmt.Errorf("failed to get last processed block: %v", err)
	}
	return result.BlockHeight, nil
}

// UpdateLastProcessedBlock updates the last processed block height.
func (r *MetadataRepository) UpdateLastProcessedBlock(ctx context.Context, blockHeight int64) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": "last_processed_block"},
		bson.M{"$set": bson.M{"block_height": blockHeight}},
		// Create the document if it doesn't exist
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to update last processed block: %v", err)
	}
	return nil
}
