package md

import (
	"log"
	"net/http"
	"time"
)

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
