package handler

import (
	"net/http"

	"github.com/anita-bakhshi/student-enrollment-system/internal/middleware"
)

// NewRouter creates and configures all application routes.
func NewRouter(studentHandler *StudentHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", Health)

	mux.HandleFunc(
		"GET /api/v1/students",
		studentHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/students/{id}",
		studentHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /api/v1/students",
		studentHandler.Create,
	)

	handler := middleware.Logging(mux)
	handler = middleware.RequestID(handler)

	return handler
}
