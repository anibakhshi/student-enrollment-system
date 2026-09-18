package handler

import (
	"net/http"
	"strings"

	"github.com/anita-bakhshi/student-enrollment-system/internal/middleware"
)

// RouterOptions contains optional router dependencies.
type RouterOptions struct {
	StudentPhotoHandler *StudentPhotoHandler
	UploadDirectory     string
}

// NewRouter creates and configures all application routes.
func NewRouter(
	studentHandler *StudentHandler,
	options ...RouterOptions,
) http.Handler {
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

	mux.HandleFunc(
		"PUT /api/v1/students/{id}",
		studentHandler.Update,
	)

	mux.HandleFunc(
		"DELETE /api/v1/students/{id}",
		studentHandler.Delete,
	)

	if len(options) > 0 {
		registerPhotoRoutes(mux, options[0])
	}

	handler := middleware.Logging(mux)
	handler = middleware.RequestID(handler)

	return handler
}

// registerPhotoRoutes registers profile-image routes when configured.
func registerPhotoRoutes(
	mux *http.ServeMux,
	options RouterOptions,
) {
	if options.StudentPhotoHandler == nil {
		return
	}

	mux.HandleFunc(
		"POST /api/v1/students/{id}/photo",
		options.StudentPhotoHandler.Upload,
	)

	uploadDirectory := strings.TrimSpace(options.UploadDirectory)
	if uploadDirectory == "" {
		return
	}

	fileServer := http.FileServer(
		http.Dir(uploadDirectory),
	)

	mux.Handle(
		"GET /uploads/",
		http.StripPrefix(
			"/uploads/",
			fileServer,
		),
	)
}
