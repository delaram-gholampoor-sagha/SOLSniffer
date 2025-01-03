package solanaClient

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
)

// SolanaClient wraps the client.Client and provides methods to interact with Solana.
type SolanaClient struct {
	client *client.Client
}

func New() *SolanaClient {
	return &SolanaClient{
		client: client.NewClient(rpc.MainnetRPCEndpoint),
	}
}

// GetBlockHeight retrieves the current block height from Solana
func (sc *SolanaClient) GetBlockHeight(ctx context.Context) (int64, error) {
	blockDetails, err := sc.client.GetBlock(ctx, uint64(0)) // Use slot 0 to get the latest block
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest block details: %w", err)
	}
	return *blockDetails.BlockHeight, nil
}

// GetBlock retrieves block details by slot
func (sc *SolanaClient) GetBlock(ctx context.Context, slot uint64) (*client.Block, error) {
	return sc.client.GetBlock(ctx, slot)
}

// GetTransaction fetches transaction details by signature
func (sc *SolanaClient) GetTransaction(ctx context.Context, signature string) (*client.Transaction, error) {
	return sc.client.GetTransaction(ctx, signature)
}
