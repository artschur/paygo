package payments

import (
	"context"
	"paygo/models"
)

type PaymentService struct {
	store *PaymentsStore
}

func NewPaymentService(store *PaymentsStore) *PaymentService {
	return &PaymentService{store: store}
}

func (s *PaymentService) ListPayments(ctx context.Context) ([]models.Payment, error) {
	return s.store.GetAllPayments(ctx)
}
