package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Exclude from JSON for security
	CreatedAt    time.Time `json:"created_at"`
	Wallet       Wallet    `json:"wallet,omitzero"`
}

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   int64     `json:"balance"` // in cents
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID           uuid.UUID  `json:"id"`
	FromWalletID *uuid.UUID `json:"from_wallet_id,omitempty"` // Nullable field
	ToWalletID   *uuid.UUID `json:"to_wallet_id,omitempty"`   // Nullable field
	Amount       int64      `json:"amount"`                   // in cents
	Status       string     `json:"status"`                   // 'pending', 'completed', etc
	Type         string     `json:"type"`                     // 'payment', 'refund', 'adjustment'
	ReferenceID  *uuid.UUID `json:"reference_id,omitempty"`   // Nullable field
	CreatedAt    time.Time  `json:"created_at"`
}

type Payment struct {
	ID            uuid.UUID  `json:"id"`
	SenderID      uuid.UUID  `json:"sender_id"`
	ReceiverID    uuid.UUID  `json:"receiver_id"`
	Amount        int64      `json:"amount"`
	Status        string     `json:"status"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"` // Nullable field
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
