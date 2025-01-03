package backfillTransaction

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/contracts/repositories"
	log "github.com/delaram-gholampoor-sagha/SOLSniffer/internal/logger"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/services/tokenTransactionProcessor"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/transport/solanaClient"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/utils"
	"sync"
	"time"
)

type Service struct {
	solanaClient            *solanaClient.SolanaClient
	metadataRepo            repositories.BackfillTransactionRepository
	tokenTransactionService *tokenTransactionProcessor.Service
	backfillConfig          *configs.BackfillConfig
}

func New(solanaClient *solanaClient.SolanaClient, config *configs.BackfillConfig, metadataRepo repositories.BackfillTransactionRepository, transactionService *tokenTransactionProcessor.Service) *Service {
	return &Service{
		solanaClient:            solanaClient,
		metadataRepo:            metadataRepo,
		tokenTransactionService: transactionService,
		backfillConfig:          config,
	}
}

func (s *Service) BackfillMissedBlocks(ctx context.Context) error {
	// Get the latest block height and the last processed block height
	currentBlock, err := s.solanaClient.GetBlockHeight(ctx)
	if err != nil {
		return err
	}

	lastProcessedBlock, err := s.getLastProcessedBlock(ctx)
	if err != nil {
		return err
	}

	// Semaphore to control the concurrency
	sem := make(chan struct{}, s.backfillConfig.MaxConcurrency)
	var wg sync.WaitGroup

	// Iterate through the blocks to backfill
	for start := lastProcessedBlock + 1; start <= currentBlock; start += s.backfillConfig.ChunkSize {
		end := s.calculateEndBlock(start, currentBlock)

		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			s.processBlockRange(ctx, start, end, sem)
		}(start, end)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	return nil
}

// getLastProcessedBlock retrieves the last processed block from the metadata repository
func (s *Service) getLastProcessedBlock(ctx context.Context) (int64, error) {
	lastProcessedBlock, err := s.metadataRepo.GetLastProcessedBlock(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch last processed block: %w", err)
	}
	return lastProcessedBlock, nil
}

// calculateEndBlock calculates the end block based on the chunk size and the current block
func (s *Service) calculateEndBlock(start, currentBlock int64) int64 {
	end := start + s.backfillConfig.ChunkSize - 1
	if end > currentBlock {
		end = currentBlock
	}
	return end
}

// processBlockRange processes a range of blocks with retry logic
func (s *Service) processBlockRange(ctx context.Context, start, end int64, sem chan struct{}) {
	for block := start; block <= end; block++ {
		sem <- struct{}{} // Acquire a semaphore slot
		go func(block int64) {
			defer func() { <-sem }() // Release the semaphore slot

			// Retry logic for processing the block
			err := retry.Do(
				func() error {
					return s.processBlock(ctx, block)
				},
				retry.Attempts(3),
				retry.Delay(2*time.Second),
				retry.DelayType(retry.BackOffDelay),
				retry.OnRetry(func(n uint, err error) {
					log.Warnf("Retrying block %d (attempt %d): %v", block, n+1, err)
				}),
			)

			if err != nil {
				log.Errorf("Failed to process block %d after retries: %v", block, err)
			}
		}(block)
	}
}

func (s *Service) processBlock(ctx context.Context, block int64) error {
	// Convert block to uint64 for compatibility with the Solana client
	uint64Block := uint64(block)

	// Fetch block details using the client
	blockDetails, err := s.solanaClient.GetBlock(ctx, uint64Block)
	if err != nil {
		return fmt.Errorf("failed to fetch block %d: %w", block, err)
	}

	// Convert block time to int64 (Unix timestamp)
	var blockTimeUnix *int64
	if blockDetails.BlockTime != nil {
		timestamp := blockDetails.BlockTime.Unix()
		blockTimeUnix = &timestamp
	}

	// Process each transaction in the block
	for _, tx := range blockDetails.Transactions {
		clientTx := utils.ConvertToClientTransaction(
			&tx.Transaction,         // types.Transaction
			tx.Meta,                 // TransactionMeta (client.TransactionMeta)
			tx.AccountKeys,          // AccountKeys
			blockDetails.ParentSlot, // Slot
			blockTimeUnix,           // BlockTime as int64
		)

		// Process the transaction using the transaction processor
		if err := s.tokenTransactionService.ProcessTransaction(ctx, clientTx); err != nil {
			log.Errorf("Failed to process transaction in block %d: %v", block, err)
		}
	}

	// Update the last processed block
	if err := s.metadataRepo.UpdateLastProcessedBlock(ctx, block); err != nil {
		return fmt.Errorf("failed to update last processed block to %d: %w", block, err)
	}

	log.Infof("Successfully processed block %d", block)
	return nil
}
