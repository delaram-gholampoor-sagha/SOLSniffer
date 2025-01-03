package transactionMonitor

import (
	"context"
	"fmt"
	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/models/request"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/tokenTransactionProcessor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/solanaClient"
)

type Service struct {
	solanaClient       *solanaClient.SolanaClient
	transactionService *tokenTransactionProcessor.Service
}

func New(solanaClient *solanaClient.SolanaClient, transactionService *tokenTransactionProcessor.Service) *Service {
	return &Service{
		solanaClient:       solanaClient,
		transactionService: transactionService,
	}
}

func (t *Service) ProcessMessage(ctx context.Context, message []byte) error {
	txLog, err := request.ParseTransactionLog(message)
	if err != nil {
		return err
	}

	// Process the transaction signature
	signature := txLog.Params.Result.Signature
	if err := t.processTransaction(ctx, signature); err != nil {
		return err
	}
	return nil
}

func (t *Service) processTransaction(ctx context.Context, signature string) error {
	txDetails, err := t.solanaClient.GetTransaction(ctx, signature)
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
