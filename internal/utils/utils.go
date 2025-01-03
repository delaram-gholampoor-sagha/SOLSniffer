package utils

import (
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
)

// ConvertToClientTransaction converts a types.Transaction and related data to a client.Transaction
func ConvertToClientTransaction(
	tx *types.Transaction,
	meta *client.TransactionMeta,
	accountKeys []common.PublicKey,
	slot uint64,
	blockTime *int64,
) *client.Transaction {
	return &client.Transaction{
		Slot:        slot,
		Meta:        meta,
		Transaction: *tx,
		BlockTime:   blockTime,
		AccountKeys: accountKeys,
	}
}
