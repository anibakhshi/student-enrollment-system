package handler

import (
	"net/http"
	"strings"

	"github.com/anita-bakhshi/student-enrollment-system/internal/middleware"
)

// RouterOptions contains optional router dependencies.
type RouterOptions struct {
	StudentPhotoHandler *StudentPhotoHandler
	InstructorHandler   *InstructorHandler
	UploadDirectory     string
}

// NewRouter creates and configures all application routes.
func NewRouter(
	studentHandler *StudentHandler,
	options ...RouterOptions,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", Health)

	registerStudentRoutes(mux, studentHandler)

	if len(options) > 0 {
		registerInstructorRoutes(
			mux,
			options[0].InstructorHandler,
		)

		registerPhotoRoutes(mux, options[0])
	}

	router := middleware.Logging(mux)
	router = middleware.RequestID(router)

	return router
}

// registerStudentRoutes registers student CRUD routes.
func registerStudentRoutes(
	mux *http.ServeMux,
	studentHandler *StudentHandler,
) {
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
}

// registerInstructorRoutes registers instructor CRUD routes.
func registerInstructorRoutes(
	mux *http.ServeMux,
	instructorHandler *InstructorHandler,
) {
	if instructorHandler == nil {
		return
	}

	mux.HandleFunc(
		"GET /api/v1/instructors",
		instructorHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/instructors/{id}",
		instructorHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /api/v1/instructors",
		instructorHandler.Create,
	)

	mux.HandleFunc(
		"PUT /api/v1/instructors/{id}",
		instructorHandler.Update,
	)

	mux.HandleFunc(
		"DELETE /api/v1/instructors/{id}",
		instructorHandler.Delete,
	)
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
