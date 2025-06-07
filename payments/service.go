package payments

import (
	"context"
	"fmt"
	"paygo/models"

	"github.com/google/uuid"
)

type PaymentStoreInterface interface {
	GetAllPayments(ctx context.Context) ([]models.Payment, error)
	GetPaymentsByUserId(ctx context.Context, userId uuid.UUID) (payments []models.PaymentWithNames, err error)
	InsertNewPayment(ctx context.Context, newP *models.PaymentInsert) (uuid.UUID, error)
	ProcessDeposit(ctx context.Context, deposit *models.DepositInsert) error
}

type PaymentService struct {
	store PaymentStoreInterface
}

func NewPaymentService(store PaymentStoreInterface) *PaymentService {
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

func (s *PaymentService) ProcessDeposit(ctx context.Context, deposit *models.DepositInsert) error {
	if deposit.Amount <= 0 {
		return ErrDepositAmountInvalid
	}

	if deposit.UserID == uuid.Nil {
		return ErrIllegalUserId
	}
	err := s.store.ProcessDeposit(ctx, deposit)
	if err != nil {
		return fmt.Errorf("processing deposit: %v", err)
	}
	return nil
}
