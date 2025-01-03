package repositories

import "context"

type BackfillTransactionRepository interface {
	GetLastProcessedBlock(ctx context.Context) (int64, error)
	UpdateLastProcessedBlock(ctx context.Context, blockHeight int64) error
}
