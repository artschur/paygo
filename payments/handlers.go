package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"paygo/models"

	"github.com/google/uuid"
)

type PaymentServiceInterface interface {
	GetAllPayments(context.Context) ([]models.Payment, error)
	GetPaymentsByUserId(ctx context.Context, userId uuid.UUID) ([]models.PaymentWithNames, error)
	InsertNewPayment(ctx context.Context, newP *models.PaymentInsert) (uuid.UUID, error)
	ProcessDeposit(ctx context.Context, deposit *models.DepositInsert) error
}

type PaymentHandler struct {
	service PaymentServiceInterface
}

func NewPaymentsHandler(s PaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{
		service: s,
	}
}

func (p *PaymentHandler) GetAllPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := p.service.GetAllPayments(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, ErrNoPaymentsFound):
			http.Error(w, "No payments found", http.StatusNotFound)
		default:
			log.Printf("error listing all payments, %v", err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func (p *PaymentHandler) GetPaymentsByUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "Invalid user UUID format", http.StatusBadRequest)
		return
	}

	payments, err := p.service.GetPaymentsByUserId(r.Context(), userUUID)
	if err != nil {
		if errors.Is(err, ErrNoPaymentsFound) {
			http.Error(w, "No payments found for user.", http.StatusNotFound)
			return
		}
		log.Printf("failed in retrieving payments: %v", err)
		http.Error(w, "Failed to retrieve payments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payments); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (p *PaymentHandler) InsertPayment(w http.ResponseWriter, r *http.Request) {
	var newPayment models.PaymentInsert
	newPayment.Status = "pending"

	if err := json.NewDecoder(r.Body).Decode(&newPayment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newPayment.ReceiverID == uuid.Nil || newPayment.SenderID == uuid.Nil || newPayment.Amount <= 0 {
		http.Error(w, "Invalid receiver or sender ID or amount", http.StatusBadRequest)
		return
	}
	if newPayment.SenderID == newPayment.ReceiverID {
		http.Error(w, "Sender and receiver cannot be the same", http.StatusBadRequest)
		return
	}
	newPaymentId, err := p.service.InsertNewPayment(r.Context(), &newPayment)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserIdNotFound):
			http.Error(w, "User ID passed does not exist in our DB.", http.StatusBadRequest)
		case errors.Is(err, ErrNoPaymentsFound):
			http.Error(w, "No payments found in DB", http.StatusNotFound)
		default:
			log.Printf("handler: error inserting payment: %v", err.Error())
			http.Error(w, "Error inserting payment", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Payment created with ID: %s", newPaymentId)
}

func (p *PaymentHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user_id").(uuid.UUID) //need to use auth first
	if !ok {
		http.Error(w, "Unauthorized: user not authenticated", http.StatusUnauthorized)
		return
	}

	var depositRequest models.DepositInsert
	err := json.NewDecoder(r.Body).Decode(&depositRequest)
	if err != nil {
		http.Error(w, "failed parsing new deposits", http.StatusBadRequest)
	}

	depositRequest.UserID = userId

	err = p.service.ProcessDeposit(r.Context(), &depositRequest)

	if err != nil {
		switch {
		case errors.Is(err, ErrUserIdNotFound):
			http.Error(w, "User ID does not exist in our DB.", http.StatusBadRequest)
		case errors.Is(err, ErrDepositAmountInvalid):
			http.Error(w, "Deposit amount must be greater than zero", http.StatusBadRequest)
		default:
			log.Printf("handler: error processing deposit: %v", err.Error())
			http.Error(w, "Error processing deposit", http.StatusInternalServerError)
		}
		return
	}
}
