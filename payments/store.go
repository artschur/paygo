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

func (s *PaymentsStore) GetPaymentsByUserId(ctx context.Context, userId uuid.UUID) ([]models.PaymentWithNames, error) {
	var payments []models.PaymentWithNames

	query := `
		SELECT
			p.id,
			p.sender_id,
			p.receiver_id,
			p.amount,
			p.status,
			p.transaction_id,
			p.note,
			p.created_at,
			sender.name AS sender_name,
			receiver.name AS receiver_name
		FROM
			payments p
		JOIN
			users sender ON sender.id = p.sender_id
		JOIN
			users receiver ON receiver.id = p.receiver_id
		WHERE
			(p.sender_id = $1 OR p.receiver_id = $1) AND p.status != 'initialized'
		ORDER BY
			p.created_at DESC
	`

	rows, err := s.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("handler: error querying payments: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var payment models.PaymentWithNames
		err = rows.Scan(
			&payment.ID,
			&payment.SenderID,
			&payment.ReceiverID,
			&payment.Amount,
			&payment.Status,
			&payment.TransactionID,
			&payment.Note,
			&payment.CreatedAt,
			&payment.SenderName,
			&payment.ReceiverName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning payment row: %w", err)
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment rows: %w", err)
	}

	return payments, nil
}

func (s *PaymentsStore) GetAllPayments(ctx context.Context) (payments []models.Payment, err error) {
	var paymentsList []models.Payment

	wantCols := []string{"id", "sender_id", "receiver_id", "amount", "status", "transaction_id", "note", "created_at"}

	query := fmt.Sprintf(
		"SELECT %s FROM payments ORDER BY created_at DESC",
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
		SELECT id, sender_id, receiver_id, amount, status, transaction_id, note, created_at
		FROM payments
		WHERE sender_id = $1
		ORDER BY created_at DESC;
		`, userId)
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

func (s *PaymentsStore) ProcessDeposit(ctx context.Context, deposit *models.DepositInsert) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var walletId uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM wallets WHERE user_id = $1 FOR UPDATE
	`, deposit.UserID).Scan(&walletId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserIdNotFound
		}
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	var transactionId uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (
			from_wallet_id,
			to_wallet_id,
			 amount,
			 status,
			 type
		)
		VALUES (NULL, $1, $2, 'completed', 'deposit')
		RETURNING id
	`, walletId, deposit.Amount).Scan(&transactionId)

	if err != nil {
		return fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE wallets
		SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, deposit.Amount, walletId)

	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit deposit transaction: %w", err)
	}

	return nil
}
