package routes

import (
	"context"
	"fmt"
	"net/http"
	"paygo/config"
	database "paygo/db"
	"paygo/md"
	"paygo/payments"
	"paygo/users"
)

func CreateRouter(ctx context.Context, mux *http.ServeMux, config config.Config) *http.ServeMux {

	db := database.Connect(ctx, config.DatabaseURL)

	paymentsStore := payments.NewPaymentsStore(db)
	paymentsService := payments.NewPaymentService(paymentsStore)
	paymentHandler := payments.NewPaymentsHandler(paymentsService)

	userStore := users.NewUserStore(db)
	userService := users.NewUserService(userStore)
	userHandler := users.NewUserHandler(userService)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to PayGo API!")
	})

	mux.HandleFunc("GET /payments", paymentHandler.GetAllPayments)
	mux.HandleFunc("GET /user/payments", paymentHandler.GetPaymentsByUserId)

	mux.Handle("POST /pay", md.AuthMiddleware(http.HandlerFunc(paymentHandler.InsertPayment)))
	mux.Handle("POST /deposit", md.AuthMiddleware(http.HandlerFunc(paymentHandler.Deposit)))

	mux.HandleFunc("GET /users", userHandler.GetAllUsers)
	mux.HandleFunc("GET /user", userHandler.GetUserById)
	mux.HandleFunc("POST /user", userHandler.CreateUser)

	return mux
}
