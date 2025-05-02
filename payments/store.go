package payments

import (
	"context"
	"errors"
	"paygo/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentsStore struct {
	db *pgxpool.Pool
}

func NewPaymentsStore(db *pgxpool.Pool) *PaymentsStore {
	return &PaymentsStore{db}
}

func (r *PaymentsStore) GetAllPayments(ctx context.Context) (payments []models.Payment, err error) {
	var paymentsList []models.Payment
	rows, err := r.db.Query(ctx, `select * from payments;`)
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
