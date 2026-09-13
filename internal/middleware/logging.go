package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// responseWriter records the HTTP status code and response size.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

// WriteHeader records the status code before sending it.
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write records the number of response bytes.
func (w *responseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}

	count, err := w.ResponseWriter.Write(data)
	w.bytes += count

	return count, err
}

// RequestID adds a unique request ID to every HTTP request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			w.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(
				r.Context(),
				requestIDKey,
				requestID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

// Logging records information about every HTTP request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			writer := &responseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(writer, r)

			slog.Info(
				"http request completed",
				"request_id", RequestIDFromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", writer.statusCode,
				"bytes", writer.bytes,
				"duration_ms", time.Since(startedAt).Milliseconds(),
				"remote_address", r.RemoteAddr,
			)
		},
	)
}

// RequestIDFromContext returns the request ID stored in context.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

// generateRequestID creates a random request identifier.
func generateRequestID() string {
	randomBytes := make([]byte, 8)

	if _, err := rand.Read(randomBytes); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(randomBytes)
}
