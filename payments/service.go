package payments

import (
	"context"
	"errors"
	"paygo/models"

	"github.com/google/uuid"
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

func (s *PaymentService) InsertNewPayment(ctx context.Context, newP *models.PaymentInsert) (newPaymentId uuid.UUID, err error) {
	newPaymentId, err = s.store.InsertNewPayment(ctx, newP)
	if err != nil {
		return uuid.Nil, errors.New("Failed inserting new payment:" + err.Error())
	}

	return newPaymentId, nil

}
