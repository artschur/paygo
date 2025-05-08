package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"paygo/models"

	"github.com/google/uuid"
)

type PaymentHandler struct {
	service *PaymentService
}

func NewPaymentsHandler(s *PaymentService) *PaymentHandler {
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
