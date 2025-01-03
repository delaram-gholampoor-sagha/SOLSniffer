package entity

import "time"

type Transaction struct {
	Hash        string    `bson:"hash"`
	Source      string    `bson:"source"`
	Destination string    `bson:"destination"`
	Amount      float64   `bson:"amount"`
	TokenMint   string    `bson:"token_mint"`
	Timestamp   time.Time `bson:"timestamp"`
}
