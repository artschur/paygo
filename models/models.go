package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	Wallet       Wallet
}

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   int64 // in cents
	Currency  string
	UpdatedAt time.Time
}

type Transaction struct {
	ID           uuid.UUID
	FromWalletID *uuid.UUID // nullable for system credits, etc
	ToWalletID   *uuid.UUID
	Amount       int64      // in cents
	Status       string     // 'pending', 'completed', etc
	Type         string     // 'payment', 'refund', 'adjustment'
	ReferenceID  *uuid.UUID // links to payments, refunds, etc (optional)
	CreatedAt    time.Time
}

type Payment struct {
	ID            uuid.UUID
	SenderID      uuid.UUID
	ReceiverID    uuid.UUID
	Amount        int64  // in cents
	Status        string // 'initiated', 'completed', etc
	TransactionID *uuid.UUID
	Note          string
	CreatedAt     time.Time
}
