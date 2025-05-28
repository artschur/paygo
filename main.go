package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"paygo/config"
	"paygo/md"
	"paygo/routes"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := config.LoadConfig()

	mux := http.NewServeMux()
	mux = routes.CreateRouter(ctx, mux, config)

	wrappedMux := md.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: wrappedMux,
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
