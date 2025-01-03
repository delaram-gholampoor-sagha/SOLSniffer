package request

import (
	"encoding/json"
	"fmt"
)

type TransactionLog struct {
	Params struct {
		Result struct {
			Signature string `json:"signature"`
		} `json:"result"`
	} `json:"params"`
}

func ParseTransactionLog(message []byte) (*TransactionLog, error) {
	var txLog TransactionLog
	if err := json.Unmarshal(message, &txLog); err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket transaction log: %w", err)
	}
	return &txLog, nil
}
