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

// PaymentHandler handles payment HTTP requests.
type PaymentHandler struct {
	service *service.PaymentService
}

// NewPaymentHandler creates a payment handler.
func NewPaymentHandler(
	paymentService *service.PaymentService,
) *PaymentHandler {
	return &PaymentHandler{
		service: paymentService,
	}
}

// List handles GET /api/v1/payments.
func (h *PaymentHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	payments, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list payments", "error", err)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Payments retrieved successfully",
		Data:    payments,
	})
}

// GetByID handles GET /api/v1/payments/{id}.
func (h *PaymentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	paymentID, err := parsePositivePaymentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Payment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid payment ID",
			},
		})
		return
	}

	details, err := h.service.GetDetails(
		r.Context(),
		paymentID,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment retrieved successfully",
		Data:    details,
	})
}

// ListByEnrollmentID handles GET /api/v1/enrollments/{id}/payments.
func (h *PaymentHandler) ListByEnrollmentID(
	w http.ResponseWriter,
	r *http.Request,
) {
	enrollmentID, err := strconv.Atoi(
		r.PathValue("id"),
	)
	if err != nil || enrollmentID <= 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Enrollment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid enrollment ID",
			},
		})
		return
	}

	payments, err := h.service.ListByEnrollmentID(
		r.Context(),
		enrollmentID,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Enrollment payments retrieved successfully",
		Data:    payments,
	})
}

// Create handles POST /api/v1/payments.
func (h *PaymentHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.CreatePaymentInput

	if err := decodePaymentJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	payment, err := h.service.Create(
		r.Context(),
		input,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	slog.Info(
		"payment created",
		"payment_id", payment.ID,
		"enrollment_id", payment.EnrollmentID,
		"amount", payment.Amount,
		"status", payment.Status,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Payment created successfully",
		Data:    payment,
	})
}

// Update handles PUT /api/v1/payments/{id}.
func (h *PaymentHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	paymentID, err := parsePositivePaymentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Payment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid payment ID",
			},
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input model.UpdatePaymentInput

	if err := decodePaymentJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	payment, err := h.service.Update(
		r.Context(),
		paymentID,
		input,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	slog.Info(
		"payment updated",
		"payment_id", payment.ID,
		"enrollment_id", payment.EnrollmentID,
		"status", payment.Status,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment updated successfully",
		Data:    payment,
	})
}

// Delete handles DELETE /api/v1/payments/{id}.
func (h *PaymentHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	paymentID, err := parsePositivePaymentID(
		r.PathValue("id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Payment ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid payment ID",
			},
		})
		return
	}

	if err := h.service.Delete(
		r.Context(),
		paymentID,
	); err != nil {
		h.handleError(w, err)
		return
	}

	slog.Info(
		"payment deleted",
		"payment_id", paymentID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// handleError converts payment errors to HTTP responses.
func (h *PaymentHandler) handleError(
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

	case errors.Is(err, repository.ErrPaymentNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Payment not found",
		})

	case errors.Is(err, repository.ErrEnrollmentNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Enrollment not found",
		})

	case errors.Is(err, repository.ErrStudentNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Student not found",
		})

	case errors.Is(err, repository.ErrCourseNotFound):
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Course not found",
		})

	case errors.Is(err, repository.ErrPaymentExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "An active payment already exists for this enrollment",
			Errors: map[string]string{
				"enrollment_id": "Enrollment already has an active payment",
			},
		})

	case errors.Is(err, repository.ErrTransactionIDExists):
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "Transaction ID already exists",
			Errors: map[string]string{
				"transaction_id": "Transaction ID must be unique",
			},
		})

	default:
		slog.Error(
			"payment operation failed",
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
	}
}

// decodePaymentJSON decodes exactly one JSON request object.
func decodePaymentJSON(
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

// parsePositivePaymentID validates a payment ID.
func parsePositivePaymentID(
	value string,
) (int, error) {
	paymentID, err := strconv.Atoi(value)
	if err != nil || paymentID <= 0 {
		return 0, errors.New("invalid payment ID")
	}

	return paymentID, nil
}
