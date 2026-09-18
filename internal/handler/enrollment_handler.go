package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// EnrollmentHandler handles enrollment HTTP requests.
type EnrollmentHandler struct {
	service *service.EnrollmentService
}

// NewEnrollmentHandler creates an enrollment handler.
func NewEnrollmentHandler(
	enrollmentService *service.EnrollmentService,
) *EnrollmentHandler {
	return &EnrollmentHandler{
		service: enrollmentService,
	}
}

// List handles GET /api/v1/enrollments.
func (h *EnrollmentHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	enrollments, err := h.service.List(r.Context())
	if err != nil {
		slog.Error(
			"failed to list enrollments",
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Enrollments retrieved successfully",
		Data:    enrollments,
	})
}

// GetByID handles GET /api/v1/enrollments/{id}.
func (h *EnrollmentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	enrollmentID, err := parsePositiveEnrollmentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Enrollment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid enrollment ID",
			},
		})
		return
	}

	enrollmentDetails, err := h.service.GetDetails(
		r.Context(),
		enrollmentID,
	)
	if err != nil {
		h.handleError(
			w,
			"get",
			err,
		)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Enrollment retrieved successfully",
		Data:    enrollmentDetails,
	})
}

// ListByStudentID handles
// GET /api/v1/students/{id}/enrollments.
func (h *EnrollmentHandler) ListByStudentID(
	w http.ResponseWriter,
	r *http.Request,
) {
	studentID, err := parsePositiveRelatedID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Student ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid student ID",
			},
		})
		return
	}

	enrollments, err := h.service.ListByStudentID(
		r.Context(),
		studentID,
	)
	if err != nil {
		h.handleError(
			w,
			"list_by_student",
			err,
		)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Student enrollments retrieved successfully",
		Data:    enrollments,
	})
}

// ListByCourseID handles
// GET /api/v1/courses/{id}/enrollments.
func (h *EnrollmentHandler) ListByCourseID(
	w http.ResponseWriter,
	r *http.Request,
) {
	courseID, err := parsePositiveRelatedID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Course ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid course ID",
			},
		})
		return
	}

	enrollments, err := h.service.ListByCourseID(
		r.Context(),
		courseID,
	)
	if err != nil {
		h.handleError(
			w,
			"list_by_course",
			err,
		)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Course enrollments retrieved successfully",
		Data:    enrollments,
	})
}

// Create handles POST /api/v1/enrollments.
func (h *EnrollmentHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	var input model.CreateEnrollmentInput

	if err := decodeSingleJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	enrollment, err := h.service.Create(
		r.Context(),
		input,
	)
	if err != nil {
		h.handleError(
			w,
			"create",
			err,
		)
		return
	}

	slog.Info(
		"student enrolled in course",
		"enrollment_id", enrollment.ID,
		"student_id", enrollment.StudentID,
		"course_id", enrollment.CourseID,
		"status", enrollment.Status,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Student enrolled successfully",
		Data:    enrollment,
	})
}

// Update handles PUT /api/v1/enrollments/{id}.
func (h *EnrollmentHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	enrollmentID, err := parsePositiveEnrollmentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Enrollment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid enrollment ID",
			},
		})
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	var input model.UpdateEnrollmentInput

	if err := decodeSingleJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	enrollment, err := h.service.Update(
		r.Context(),
		enrollmentID,
		input,
	)
	if err != nil {
		h.handleError(
			w,
			"update",
			err,
		)
		return
	}

	slog.Info(
		"enrollment updated",
		"enrollment_id", enrollment.ID,
		"student_id", enrollment.StudentID,
		"course_id", enrollment.CourseID,
		"status", enrollment.Status,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Enrollment updated successfully",
		Data:    enrollment,
	})
}

// Delete handles DELETE /api/v1/enrollments/{id}.
func (h *EnrollmentHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	enrollmentID, err := parsePositiveEnrollmentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Enrollment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid enrollment ID",
			},
		})
		return
	}

	err = h.service.Delete(
		r.Context(),
		enrollmentID,
	)
	if err != nil {
		h.handleError(
			w,
			"delete",
			err,
		)
		return
	}

	slog.Info(
		"enrollment deleted",
		"enrollment_id", enrollmentID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleError converts enrollment errors to HTTP responses.
func (h *EnrollmentHandler) handleError(
	w http.ResponseWriter,
	operation string,
	err error,
) {
	var validationError *service.ValidationError

	switch {
	case errors.As(err, &validationError):
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationError.Fields,
		})

	case errors.Is(
		err,
		repository.ErrEnrollmentNotFound,
	):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Enrollment not found",
		})

	case errors.Is(err, repository.ErrStudentNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Student not found",
			Errors: map[string]string{
				"student_id": "Student does not exist",
			},
		})

	case errors.Is(err, repository.ErrCourseNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Course not found",
			Errors: map[string]string{
				"course_id": "Course does not exist",
			},
		})

	case errors.Is(err, repository.ErrEnrollmentExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "Student is already enrolled in this course",
			Errors: map[string]string{
				"enrollment": "Active enrollment already exists",
			},
		})

	case errors.Is(err, repository.ErrCourseFull):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "Course capacity has been reached",
			Errors: map[string]string{
				"course_id": "Course is full",
			},
		})

	default:
		slog.Error(
			"failed to process enrollment",
			"operation", operation,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}

// parsePositiveEnrollmentID validates an enrollment ID.
func parsePositiveEnrollmentID(
	value string,
) (int, error) {
	enrollmentID, err := strconv.Atoi(value)
	if err != nil || enrollmentID <= 0 {
		return 0, errors.New("invalid enrollment ID")
	}

	return enrollmentID, nil
}

// parsePositiveRelatedID validates a student or course ID.
func parsePositiveRelatedID(
	value string,
) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid related ID")
	}

	return id, nil
}
