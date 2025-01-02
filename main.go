package main

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	monitoredTokens = map[string]bool{
		"NativeSOLMintAddress": true,
		"USDTMintAddress":      true,
	}
	monitoredWallets = map[string]bool{
		"WalletAddress1": true,
		"WalletAddress2": true,
	}
	db *mongo.Client
)

func main() {

	initMongoDB()

	c := client.NewClient(rpc.MainnetRPCEndpoint)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go monitorTransactions(c)

	<-shutdown
	log.Println("Shutting down...")
	if err := db.Disconnect(context.Background()); err != nil {
		log.Fatalf("Failed to disconnect MongoDB: %v", err)
	}
}

func initMongoDB() {
	uri := "mongodb://localhost:27017"
	var err error

	db, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB.")
}

func monitorTransactions(c *client.Client) {
	walletAddress := "YourMonitoredWalletAddress"

	for {

		response, err := c.RpcClient.GetSignaturesForAddressWithConfig(
			context.Background(),
			walletAddress,
			rpc.GetSignaturesForAddressConfig{
				Limit: 100,
			},
		)
		if err != nil {
			log.Printf("Error fetching transaction signatures: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if response.Result == nil || len(response.Result) == 0 {
			log.Println("No new transactions found.")
			time.Sleep(10 * time.Second)
			continue
		}

		for _, sig := range response.Result {
			go func(signature string) {
				err := processTransaction(c, signature)
				if err != nil {
					log.Printf("Failed to process transaction %s: %v", signature, err)
				}
			}(sig.Signature)
		}

		time.Sleep(10 * time.Second)
	}
}

// Process a single transaction
func processTransaction(c *client.Client, signature string) error {
	txDetails, err := c.GetTransaction(context.TODO(), signature)
	if err != nil {
		return fmt.Errorf("failed to get transaction details: %v", err)
	}
	if txDetails == nil {
		log.Printf("Transaction %s not found or not confirmed yet.", signature)
		return nil
	}

	if !filterTransaction(txDetails) {
		log.Printf("Transaction %s does not match monitored tokens or wallets.", signature)
		return nil
	}

	return storeTransaction(txDetails)
}

// Filter transaction based on monitored tokens and wallets
func filterTransaction(tx *client.Transaction) bool {
	// Check for monitored tokens
	for _, balance := range tx.Meta.PreTokenBalances {
		if monitoredTokens[balance.Mint] {
			return true
		}
	}

	// Check for monitored wallets
	for _, account := range tx.AccountKeys {
		if monitoredWallets[account.ToBase58()] {
			return true
		}
	}
	return false
}

// Store matched transaction in MongoDB
func storeTransaction(tx *client.Transaction) error {
	collection := db.Database("solsniffer").Collection("transactions")

	transaction := bson.M{
		"hash":        tx.Transaction.Signatures[0],
		"source":      tx.AccountKeys[0].ToBase58(),
		"destination": tx.AccountKeys[1].ToBase58(),
		"amount":      tx.Meta.PreTokenBalances[0].UITokenAmount,
		"token_mint":  tx.Meta.PreTokenBalances[0].Mint,
		"timestamp":   time.Now(),
	}

	_, err := collection.InsertOne(context.Background(), transaction)
	if err != nil {
		return fmt.Errorf("failed to insert transaction into MongoDB: %v", err)
	}
	log.Printf("Stored transaction %s successfully.", tx.Transaction.Signatures[0])
	return nil
}
