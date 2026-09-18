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
	CourseHandler       *CourseHandler
	EnrollmentHandler   *EnrollmentHandler
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

		registerCourseRoutes(
			mux,
			options[0].CourseHandler,
		)

		registerEnrollmentRoutes(
			mux,
			options[0].EnrollmentHandler,
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

// registerCourseRoutes registers course CRUD routes.
func registerCourseRoutes(
	mux *http.ServeMux,
	courseHandler *CourseHandler,
) {
	if courseHandler == nil {
		return
	}

	mux.HandleFunc(
		"GET /api/v1/courses",
		courseHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/courses/{id}",
		courseHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /api/v1/courses",
		courseHandler.Create,
	)

	mux.HandleFunc(
		"PUT /api/v1/courses/{id}",
		courseHandler.Update,
	)

	mux.HandleFunc(
		"DELETE /api/v1/courses/{id}",
		courseHandler.Delete,
	)
}

// registerEnrollmentRoutes registers enrollment routes.
func registerEnrollmentRoutes(
	mux *http.ServeMux,
	enrollmentHandler *EnrollmentHandler,
) {
	if enrollmentHandler == nil {
		return
	}

	mux.HandleFunc(
		"GET /api/v1/enrollments",
		enrollmentHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/enrollments/{id}",
		enrollmentHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /api/v1/enrollments",
		enrollmentHandler.Create,
	)

	mux.HandleFunc(
		"PUT /api/v1/enrollments/{id}",
		enrollmentHandler.Update,
	)

	mux.HandleFunc(
		"DELETE /api/v1/enrollments/{id}",
		enrollmentHandler.Delete,
	)

	mux.HandleFunc(
		"GET /api/v1/students/{id}/enrollments",
		enrollmentHandler.ListByStudentID,
	)

	mux.HandleFunc(
		"GET /api/v1/courses/{id}/enrollments",
		enrollmentHandler.ListByCourseID,
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
