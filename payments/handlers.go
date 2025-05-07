package payments

import (
	"encoding/json"
	"fmt"
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

func (p *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := p.service.ListPayments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Payment created with ID: %s", newPaymentId)
}
