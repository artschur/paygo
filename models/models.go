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
	ID            uuid.UUID  `json:"id"`
	SenderID      uuid.UUID  `json:"sender_id"`
	ReceiverID    uuid.UUID  `json:"receiver_id"`
	Amount        int64      `json:"amount"`
	Status        string     `json:"status"`
	TransactionID *uuid.UUID `json:"transaction_id"`
	Note          string     `json:"note"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PaymentInsert struct {
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	Note       string    `json:"note"`
}
