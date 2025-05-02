package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"paygo/config"
	database "paygo/db"
	"paygo/payments"
	"syscall"
	"time"
)

type PaymentHandler struct {
	service *payments.PaymentService
}

func (p *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	p.service.ListPayments(r.Context())
}
func NewPaymentsHandler(s *payments.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: s,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := config.LoadConfig()

	if config.DatabaseURL == "" || config.Port == "" {
		log.Println("Missing required environment variables (DATABASE_URL, APP_PORT)")
		log.Println("Shutting down server...")
		os.Exit(1) // Exit the program with error code
	}
	fmt.Printf("db key %v", config.DatabaseURL)
	db := database.Connect(ctx, config.DatabaseURL)

	paymentsStore := payments.NewPaymentsStore(db)
	paymentsService := payments.NewPaymentService(paymentsStore)
	paymentHandler := NewPaymentsHandler(paymentsService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", paymentHandler.ListPayments)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("Server started on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}
