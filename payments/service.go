package payments

import (
	"context"
	"fmt"
	"paygo/models"

	"github.com/google/uuid"
)

type PaymentService struct {
	store *PaymentsStore
}

func NewPaymentService(store *PaymentsStore) *PaymentService {
	return &PaymentService{store: store}
}

func (s *PaymentService) GetAllPayments(ctx context.Context) ([]models.Payment, error) {
	payments, err := s.store.GetAllPayments(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: failed to fetch payments %v", err.Error())
	}

	if len(payments) == 0 {
		return nil, ErrNoPaymentsFound
	}

	return payments, nil
}

func (s *PaymentService) InsertNewPayment(ctx context.Context, newP *models.PaymentInsert) (newPaymentId uuid.UUID, err error) {
	newPaymentId, err = s.store.InsertNewPayment(ctx, newP)
	if err != nil {
		return uuid.Nil, err
	}

	return newPaymentId, nil
}

func (s *PaymentService) GetPaymentsByUserId(ctx context.Context, userId uuid.UUID) (
	payments []models.PaymentWithNames, err error) {

	payments, err = s.store.GetPaymentsByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("service: error querying payments: %v", err)
	}
	if len(payments) == 0 {
		return nil, ErrNoPaymentsFound
	}

	return payments, nil

}
