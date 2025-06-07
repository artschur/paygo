package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

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
	UserID    uuid.UUID `json:"user_id,omitzero"`
	Balance   int64     `json:"balance"` // in cents
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"last_transaction"`
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
	ID uuid.UUID `json:"id"`
	PaymentInsert
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"` // Nullable field
	CreatedAt     time.Time  `json:"created_at"`
}

type PaymentInsert struct {
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	Note       string    `json:"note"`
}

type DepositInsert struct {
	UserID    uuid.UUID `json:"user_id"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentWithNames struct {
	Payment
	SenderName   string `json:"sender_name"`
	ReceiverName string `json:"receiver_name"`
}
