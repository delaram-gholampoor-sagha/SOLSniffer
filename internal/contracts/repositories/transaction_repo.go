package repositories

import (
	"context"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/models/entity"
)

type Transaction interface {
	Save(ctx context.Context, transaction *entity.Transaction) error
}
