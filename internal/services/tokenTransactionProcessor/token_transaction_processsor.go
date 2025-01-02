package tokenTransactionProcessor

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/contracts/repositories"
	log "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"time"
)

type Service struct {
	repo             repositories.Transaction
	monitoredTokens  map[string]bool
	monitoredWallets map[string]bool
}

func New(repo repositories.Transaction, tokens, wallets []string) *Service {
	tokenSet := make(map[string]bool)

	// Add Native SOL explicitly
	const nativeSOLMint = "NativeSOL"
	tokenSet[nativeSOLMint] = true

	// Add other tokens
	for _, token := range tokens {
		tokenSet[token] = true
	}

	walletSet := make(map[string]bool)
	for _, wallet := range wallets {
		walletSet[wallet] = true
	}

	return &Service{
		repo:             repo,
		monitoredTokens:  tokenSet,
		monitoredWallets: walletSet,
	}
}

func (s *Service) ProcessTransaction(ctx context.Context, txDetails *client.Transaction) error {
	// Check for signatures
	if len(txDetails.Transaction.Signatures) == 0 {
		return fmt.Errorf("no signatures found in transaction")
	}

	// Convert signature to string
	hash := hex.EncodeToString(txDetails.Transaction.Signatures[0])

	// Ensure there are enough account keys
	if len(txDetails.Transaction.Message.Accounts) < 2 {
		return fmt.Errorf("not enough accounts in transaction message")
	}

	source := txDetails.Transaction.Message.Accounts[0].ToBase58()
	destination := txDetails.Transaction.Message.Accounts[1].ToBase58()

	// Handle token balances
	if len(txDetails.Meta.PreTokenBalances) == 0 {
		log.Warnf("Transaction %s has no token balances; skipping", hash)
		return nil
	}

	for _, balance := range txDetails.Meta.PreTokenBalances {
		amountStr := balance.UITokenAmount.Amount
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.WithError(err).Errorf("Failed to parse amount: %s", amountStr)
			continue
		}

		// Normalize the token amount
		amount /= math.Pow10(int(balance.UITokenAmount.Decimals))
		token := balance.Mint

		// Determine if the token is Native SOL or SPL Token
		isNativeSOL := token == "NativeSOL"
		if !isNativeSOL && !s.monitoredTokens[token] {
			log.Infof("Transaction %s with token %s does not match token filters", hash, token)
			continue
		}

		// Filter by destination wallet
		if !s.monitoredWallets[destination] {
			log.Infof("Transaction %s with destination %s does not match wallet filters", hash, destination)
			continue
		}

		// Save to the database

		err = s.repo.Save(ctx, hash, source, destination, amount, token, time.Now())
		if err != nil {
			log.WithError(err).Errorf("Failed to save transaction %s", hash)
			continue
		}

		log.Infof("Transaction %s with token %s processed successfully", hash, token)
	}

	return nil
}
