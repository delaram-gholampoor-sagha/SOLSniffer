package repositories

import (
	"context"
	"time"
)

type Transaction interface {
	Save(ctx context.Context, hash, source, destination string, amount float64, token string, timestamp time.Time) error
}
