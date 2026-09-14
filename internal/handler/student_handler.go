package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// StudentHandler handles student HTTP requests.
type StudentHandler struct {
	service *service.StudentService
}

// NewStudentHandler creates a new student handler.
func NewStudentHandler(
	studentService *service.StudentService,
) *StudentHandler {
	return &StudentHandler{
		service: studentService,
	}
}

// List handles GET /api/v1/students.
func (h *StudentHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	students, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list students", "error", err)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Students retrieved successfully",
		Data:    students,
	})
}

// GetByID handles GET /api/v1/students/{id}.
func (h *StudentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	idValue := r.PathValue("id")

	studentID, err := strconv.Atoi(idValue)
	if err != nil || studentID <= 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Student ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid student ID",
			},
		})
		return
	}

	student, err := h.service.GetByID(r.Context(), studentID)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Student not found",
			})
			return
		}

		slog.Error(
			"failed to get student",
			"student_id", studentID,
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
		Message: "Student retrieved successfully",
		Data:    student,
	})
}

// Create handles POST /api/v1/students.
func (h *StudentHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.CreateStudentInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Request body must contain only one JSON object",
		})
		return
	}

	student, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	slog.Info(
		"student created",
		"student_id", student.ID,
		"national_code", student.NationalCode,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Student created successfully",
		Data:    student,
	})
}

// handleCreateError converts service errors to HTTP responses.
func (h *StudentHandler) handleCreateError(
	w http.ResponseWriter,
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

	case errors.Is(err, repository.ErrNationalCodeExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "A student with this national code already exists",
			Errors: map[string]string{
				"national_code": "National code must be unique",
			},
		})

	case errors.Is(err, repository.ErrEmailExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "A student with this email already exists",
			Errors: map[string]string{
				"email": "Email must be unique",
			},
		})

	default:
		slog.Error("failed to create student", "error", err)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}

// Update handles PUT /api/v1/students/{id}.
func (h *StudentHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	idValue := r.PathValue("id")

	studentID, err := strconv.Atoi(idValue)
	if err != nil || studentID <= 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Student ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid student ID",
			},
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.UpdateStudentInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Request body must contain only one JSON object",
		})
		return
	}

	student, err := h.service.Update(
		r.Context(),
		studentID,
		input,
	)
	if err != nil {
		h.handleUpdateError(w, err)
		return
	}

	slog.Info(
		"student updated",
		"student_id", student.ID,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Student updated successfully",
		Data:    student,
	})
}

// Delete handles DELETE /api/v1/students/{id}.
func (h *StudentHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	idValue := r.PathValue("id")

	studentID, err := strconv.Atoi(idValue)
	if err != nil || studentID <= 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Student ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid student ID",
			},
		})
		return
	}

	err = h.service.Delete(r.Context(), studentID)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Student not found",
			})
			return
		}

		slog.Error(
			"failed to delete student",
			"student_id", studentID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	slog.Info(
		"student deleted",
		"student_id", studentID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateError converts update errors to HTTP responses.
func (h *StudentHandler) handleUpdateError(
	w http.ResponseWriter,
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

	case errors.Is(err, repository.ErrStudentNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Student not found",
		})

	case errors.Is(err, repository.ErrNationalCodeExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "A student with this national code already exists",
			Errors: map[string]string{
				"national_code": "National code must be unique",
			},
		})

	case errors.Is(err, repository.ErrEmailExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "A student with this email already exists",
			Errors: map[string]string{
				"email": "Email must be unique",
			},
		})

	default:
		slog.Error("failed to update student", "error", err)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}
