package md

import (
	"context"
	"log"
	"net/http"
	"paygo/auth"
	"time"

	"github.com/google/uuid"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		authToken := authHeader[len("Bearer "):]
		if authToken == "" {
			http.Error(w, "Empty Token", http.StatusUnauthorized)
			return
		}

		token, err := auth.ValidateToken(authToken)

		if err != nil {
			log.Printf("Token validation error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userId, err := uuid.Parse(token.Subject)
		if err != nil {
			log.Printf("Invalid user ID format: %v", err)
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userId)
		ctx = context.WithValue(r.Context(), "username", token.Username)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		log.Printf(
			"%s | %s %s | %s |  %d",
			start.Format("15:04:05.000"), // Simple time format: HH:mm:ss.mmm
			r.Method,
			r.URL.Path,
			formatDuration(duration),
			wrappedWriter.statusCode,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func formatDuration(d time.Duration) string {
	return d.Truncate(time.Millisecond).String()
}
