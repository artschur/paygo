package payments

import (
	"context"
	"errors"
	"fmt"
	"paygo/models"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentsStore struct {
	db *pgxpool.Pool
}

func NewPaymentsStore(db *pgxpool.Pool) *PaymentsStore {
	return &PaymentsStore{db}
}

func (s *PaymentsStore) GetAllPayments(ctx context.Context) (payments []models.Payment, err error) {
	var paymentsList []models.Payment

	wantCols := []string{"id", "sender_id", "receiver_id", "amount", "status", "transaction_id", "note", "created_at"}

	query := fmt.Sprintf(
		"SELECT %s FROM payments",
		strings.Join(wantCols, ", "),
	)

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, errors.New("Error querying payments: " + err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var payment models.Payment
		err = rows.Scan(
			&payment.ID,
			&payment.SenderID,
			&payment.ReceiverID,
			&payment.Amount,
			&payment.Status,
			&payment.TransactionID,
			&payment.Note,
			&payment.CreatedAt,
		)

		if err != nil {
			return nil, err
		}
		paymentsList = append(paymentsList, payment)
	}

	if rows.Err() != nil {
		return nil, errors.New("Error scanning payments: " + rows.Err().Error())
	}

	return paymentsList, nil

}

func (s *PaymentsStore) GetPaymentsUserHasPaid(ctx context.Context, userId uuid.UUID) (payments []models.Payment, err error) {
	var paymentsList []models.Payment
	rows, err := s.db.Query(ctx, `
		select * from payments where sender_id = $1 order by created_at desc;
		`, userId)
	if err != nil {
		return nil, errors.New("Error querying payments: " + err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var payment models.Payment
		err = rows.Scan(&payment)
		if err != nil {
			return nil, err
		}
		paymentsList = append(paymentsList, payment)
	}

	if rows.Err() != nil {
		return nil, errors.New("Error scanning payments: " + rows.Err().Error())
	}

	return payments, nil
}

func (s *PaymentsStore) InsertNewPayment(ctx context.Context, newPayment *models.PaymentInsert) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, errors.New("Failed to begin transaction: " + err.Error())
	}

	defer tx.Rollback(ctx) // Will be ignored if tx.Commit() is called

	var (
		senderWalletId   uuid.UUID
		receiverWalletId uuid.UUID
		senderBalance    int64
	)

	err = tx.QueryRow(ctx, `
        SELECT w1.id, w1.balance, w2.id
        FROM wallets w1, wallets w2
        WHERE w1.user_id = $1 AND w2.user_id = $2
        FOR UPDATE OF w1, w2
    `, newPayment.SenderID, newPayment.ReceiverID).Scan(&senderWalletId, &senderBalance, &receiverWalletId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrUserIdNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	if senderBalance < newPayment.Amount {
		return uuid.Nil, errors.New("insufficient funds")
	}

	// create transaction itself
	var newTransactionId uuid.UUID
	err = tx.QueryRow(ctx,
		`
		insert into transactions (from_wallet_id, to_wallet_id, amount, status, type)
		values ($1, $2, $3, 'pending', 'payment') RETURNING id;
		`, senderWalletId, receiverWalletId, newPayment.Amount).Scan(&newTransactionId)

	if err != nil {
		return uuid.Nil, errors.New("Could not create transaction: " + err.Error())
	}

	var paymentId uuid.UUID
	err = tx.QueryRow(ctx, `
        INSERT INTO payments (sender_id, receiver_id, amount, status, transaction_id, note)
        VALUES ($1, $2, $3, 'initiated', $4, $5)
        RETURNING id;
    `, newPayment.SenderID, newPayment.ReceiverID, newPayment.Amount,
		newTransactionId, newPayment.Note).Scan(&paymentId)

	if err != nil {
		return uuid.Nil, errors.New("Could not create payment: " + err.Error())
	}

	batch := &pgx.Batch{}

	batch.Queue(`
    UPDATE wallets SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP
    WHERE id = $2
`, newPayment.Amount, senderWalletId)

	batch.Queue(`
        UPDATE wallets SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, newPayment.Amount, receiverWalletId)

	batch.Queue(`
        UPDATE transactions SET status = 'completed' WHERE id = $1
    `, newTransactionId)

	batch.Queue(`
        UPDATE payments SET status = 'completed' WHERE id = $1
    `, paymentId)

	results := tx.SendBatch(ctx, batch)

	for i := range 4 {
		_, err := results.Exec()
		if err != nil {
			results.Close()
			return uuid.Nil, fmt.Errorf("batch operation %d failed: %w", i, err)
		}
	}
	err = results.Close()
	if err != nil {
		return uuid.Nil, errors.New("Could not close batch results: " + err.Error())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return uuid.Nil, errors.New("Could not commit transaction: " + err.Error())
	}

	return paymentId, nil

}
