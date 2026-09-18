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

// CourseHandler handles course HTTP requests.
type CourseHandler struct {
	service *service.CourseService
}

// NewCourseHandler creates a course handler.
func NewCourseHandler(
	courseService *service.CourseService,
) *CourseHandler {
	return &CourseHandler{
		service: courseService,
	}
}

// List handles GET /api/v1/courses.
func (h *CourseHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	courses, err := h.service.List(r.Context())
	if err != nil {
		slog.Error(
			"failed to list courses",
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
		Message: "Courses retrieved successfully",
		Data:    courses,
	})
}

// GetByID handles GET /api/v1/courses/{id}.
func (h *CourseHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	courseID, err := parsePositiveCourseID(
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

	course, err := h.service.GetByID(
		r.Context(),
		courseID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Course not found",
			})
			return
		}

		slog.Error(
			"failed to get course",
			"course_id", courseID,
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
		Message: "Course retrieved successfully",
		Data:    course,
	})
}

// Create handles POST /api/v1/courses.
func (h *CourseHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.CreateCourseInput

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

	course, err := h.service.Create(
		r.Context(),
		input,
	)
	if err != nil {
		h.handleMutationError(
			w,
			"create",
			err,
		)
		return
	}

	slog.Info(
		"course created",
		"course_id", course.ID,
		"course_code", course.Code,
		"instructor_id", course.InstructorID,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Course created successfully",
		Data:    course,
	})
}

// Update handles PUT /api/v1/courses/{id}.
func (h *CourseHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	courseID, err := parsePositiveCourseID(
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.UpdateCourseInput

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

	course, err := h.service.Update(
		r.Context(),
		courseID,
		input,
	)
	if err != nil {
		h.handleMutationError(
			w,
			"update",
			err,
		)
		return
	}

	slog.Info(
		"course updated",
		"course_id", course.ID,
		"course_code", course.Code,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Course updated successfully",
		Data:    course,
	})
}

// Delete handles DELETE /api/v1/courses/{id}.
func (h *CourseHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	courseID, err := parsePositiveCourseID(
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

	err = h.service.Delete(
		r.Context(),
		courseID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Course not found",
			})
			return
		}

		slog.Error(
			"failed to delete course",
			"course_id", courseID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	slog.Info(
		"course deleted",
		"course_id", courseID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleMutationError converts course errors to HTTP responses.
func (h *CourseHandler) handleMutationError(
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

	case errors.Is(err, repository.ErrCourseNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Course not found",
		})

	case errors.Is(err, repository.ErrInstructorNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Instructor not found",
			Errors: map[string]string{
				"instructor_id": "Instructor does not exist",
			},
		})

	case errors.Is(err, repository.ErrCourseCodeExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "A course with this code already exists",
			Errors: map[string]string{
				"code": "Course code must be unique",
			},
		})

	default:
		slog.Error(
			"failed to mutate course",
			"operation", operation,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}

// parsePositiveCourseID validates a course ID.
func parsePositiveCourseID(
	value string,
) (int, error) {
	courseID, err := strconv.Atoi(value)
	if err != nil || courseID <= 0 {
		return 0, errors.New("invalid course ID")
	}

	return courseID, nil
}
