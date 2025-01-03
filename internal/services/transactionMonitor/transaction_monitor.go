package transactionMonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/tokenTransactionProcessor"
)

type Service struct {
	client             *client.Client
	transactionService *tokenTransactionProcessor.Service
}

func New(transactionService *tokenTransactionProcessor.Service) *Service {
	return &Service{
		client:             client.NewClient(rpc.MainnetRPCEndpoint),
		transactionService: transactionService,
	}
}

func (t *Service) ProcessMessage(ctx context.Context, message []byte) error {
	var logResult struct {
		Params struct {
			Result struct {
				Signature string `json:"signature"`
			} `json:"result"`
		} `json:"params"`
	}
	if err := json.Unmarshal(message, &logResult); err != nil {
		return fmt.Errorf("failed to parse WebSocket log message: %w", err)
	}

	// Process the transaction signature
	signature := logResult.Params.Result.Signature
	return t.processTransaction(ctx, signature)
}

func (t *Service) processTransaction(ctx context.Context, signature string) error {
	txDetails, err := t.client.GetTransaction(ctx, signature)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction details for signature %s: %w", signature, err)
	}
	if txDetails == nil {
		return fmt.Errorf("transaction %s not found", signature)
	}

	if err := t.transactionService.ProcessTransaction(ctx, txDetails); err != nil {
		return fmt.Errorf("failed to process transaction %s: %w", signature, err)
	}

	log.Infof("Transaction %s processed successfully", signature)
	return nil
}
