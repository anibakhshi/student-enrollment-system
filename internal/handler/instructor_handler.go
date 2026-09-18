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

// InstructorHandler handles instructor HTTP requests.
type InstructorHandler struct {
	service *service.InstructorService
}

// NewInstructorHandler creates an instructor handler.
func NewInstructorHandler(
	instructorService *service.InstructorService,
) *InstructorHandler {
	return &InstructorHandler{
		service: instructorService,
	}
}

// List handles GET /api/v1/instructors.
func (h *InstructorHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	instructors, err := h.service.List(r.Context())
	if err != nil {
		slog.Error(
			"failed to list instructors",
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
		Message: "Instructors retrieved successfully",
		Data:    instructors,
	})
}

// GetByID handles GET /api/v1/instructors/{id}.
func (h *InstructorHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	instructorID, err := parsePositiveInstructorID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Instructor ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid instructor ID",
			},
		})
		return
	}

	instructor, err := h.service.GetByID(
		r.Context(),
		instructorID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrInstructorNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Instructor not found",
			})
			return
		}

		slog.Error(
			"failed to get instructor",
			"instructor_id", instructorID,
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
		Message: "Instructor retrieved successfully",
		Data:    instructor,
	})
}

// Create handles POST /api/v1/instructors.
func (h *InstructorHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.CreateInstructorInput

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

	instructor, err := h.service.Create(
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
		"instructor created",
		"instructor_id", instructor.ID,
		"email", instructor.Email,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Instructor created successfully",
		Data:    instructor,
	})
}

// Update handles PUT /api/v1/instructors/{id}.
func (h *InstructorHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	instructorID, err := parsePositiveInstructorID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Instructor ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid instructor ID",
			},
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.UpdateInstructorInput

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

	instructor, err := h.service.Update(
		r.Context(),
		instructorID,
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
		"instructor updated",
		"instructor_id", instructor.ID,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Instructor updated successfully",
		Data:    instructor,
	})
}

// Delete handles DELETE /api/v1/instructors/{id}.
func (h *InstructorHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	instructorID, err := parsePositiveInstructorID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Instructor ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid instructor ID",
			},
		})
		return
	}

	err = h.service.Delete(
		r.Context(),
		instructorID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrInstructorNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Instructor not found",
			})
			return
		}

		slog.Error(
			"failed to delete instructor",
			"instructor_id", instructorID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	slog.Info(
		"instructor deleted",
		"instructor_id", instructorID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleMutationError converts instructor errors to HTTP responses.
func (h *InstructorHandler) handleMutationError(
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

	case errors.Is(err, repository.ErrInstructorNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Instructor not found",
		})

	case errors.Is(err, repository.ErrInstructorEmailExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "An instructor with this email already exists",
			Errors: map[string]string{
				"email": "Email must be unique",
			},
		})

	default:
		slog.Error(
			"failed to mutate instructor",
			"operation", operation,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}

// parsePositiveInstructorID validates an instructor ID.
func parsePositiveInstructorID(
	value string,
) (int, error) {
	instructorID, err := strconv.Atoi(value)
	if err != nil || instructorID <= 0 {
		return 0, errors.New("invalid instructor ID")
	}

	return instructorID, nil
}

// decodeSingleJSON decodes exactly one JSON object.
func decodeSingleJSON(
	r *http.Request,
	destination any,
) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New(
				"request body must contain only one JSON object",
			)
		}

		return err
	}

	return nil
}
