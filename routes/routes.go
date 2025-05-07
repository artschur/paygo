package routes

import (
	"context"
	"fmt"
	"net/http"
	"paygo/config"
	database "paygo/db"
	"paygo/payments"
)

func CreateRouter(ctx context.Context, mux *http.ServeMux, config config.Config) *http.ServeMux {

	db := database.Connect(ctx, config.DatabaseURL)

	paymentsStore := payments.NewPaymentsStore(db)
	paymentsService := payments.NewPaymentService(paymentsStore)
	paymentHandler := payments.NewPaymentsHandler(paymentsService)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to PayGo API!")
	})

	mux.HandleFunc("GET /payments", paymentHandler.ListPayments)
	mux.HandleFunc("POST /payments", paymentHandler.InsertPayment)

	return mux
}
