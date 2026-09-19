package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// PaymentService implements payment business rules.
type PaymentService struct {
	paymentRepository    repository.PaymentRepository
	enrollmentRepository repository.EnrollmentRepository
	studentRepository    repository.StudentRepository
	courseRepository     repository.CourseRepository
	now                  func() time.Time
}

// NewPaymentService creates a payment service.
func NewPaymentService(
	paymentRepository repository.PaymentRepository,
	enrollmentRepository repository.EnrollmentRepository,
	studentRepository repository.StudentRepository,
	courseRepository repository.CourseRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepository:    paymentRepository,
		enrollmentRepository: enrollmentRepository,
		studentRepository:    studentRepository,
		courseRepository:     courseRepository,
		now:                  time.Now,
	}
}

// List returns all payments.
func (s *PaymentService) List(
	ctx context.Context,
) ([]model.Payment, error) {
	return s.paymentRepository.List(ctx)
}

// GetByID returns a payment by ID.
func (s *PaymentService) GetByID(
	ctx context.Context,
	id int,
) (model.Payment, error) {
	if id <= 0 {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"id": "Payment ID must be a positive integer",
			},
		}
	}

	return s.paymentRepository.GetByID(ctx, id)
}

// GetDetails returns payment, enrollment, student, and course data.
func (s *PaymentService) GetDetails(
	ctx context.Context,
	id int,
) (model.PaymentDetails, error) {
	payment, err := s.GetByID(ctx, id)
	if err != nil {
		return model.PaymentDetails{}, err
	}

	enrollment, err := s.enrollmentRepository.GetByID(
		ctx,
		payment.EnrollmentID,
	)
	if err != nil {
		return model.PaymentDetails{}, fmt.Errorf(
			"get payment enrollment: %w",
			err,
		)
	}

	student, err := s.studentRepository.GetByID(
		ctx,
		enrollment.StudentID,
	)
	if err != nil {
		return model.PaymentDetails{}, fmt.Errorf(
			"get payment student: %w",
			err,
		)
	}

	course, err := s.courseRepository.GetByID(
		ctx,
		enrollment.CourseID,
	)
	if err != nil {
		return model.PaymentDetails{}, fmt.Errorf(
			"get payment course: %w",
			err,
		)
	}

	return model.PaymentDetails{
		Payment:    payment,
		Enrollment: enrollment,
		Student:    student,
		Course:     course,
	}, nil
}

// ListByEnrollmentID returns all payments for an enrollment.
func (s *PaymentService) ListByEnrollmentID(
	ctx context.Context,
	enrollmentID int,
) ([]model.Payment, error) {
	if enrollmentID <= 0 {
		return nil, &ValidationError{
			Fields: map[string]string{
				"enrollment_id": "Enrollment ID must be a positive integer",
			},
		}
	}

	_, err := s.enrollmentRepository.GetByID(
		ctx,
		enrollmentID,
	)
	if err != nil {
		return nil, err
	}

	return s.paymentRepository.ListByEnrollmentID(
		ctx,
		enrollmentID,
	)
}

// Create starts a payment for an enrollment.
func (s *PaymentService) Create(
	ctx context.Context,
	input model.CreatePaymentInput,
) (model.Payment, error) {
	validationErrors := validateCreatePaymentInput(input)
	if len(validationErrors) > 0 {
		return model.Payment{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	enrollment, err := s.enrollmentRepository.GetByID(
		ctx,
		input.EnrollmentID,
	)
	if err != nil {
		return model.Payment{}, err
	}

	if enrollment.Status == model.EnrollmentStatusCancelled {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"enrollment_id": "Cancelled enrollment cannot be paid",
			},
		}
	}

	course, err := s.courseRepository.GetByID(
		ctx,
		enrollment.CourseID,
	)
	if err != nil {
		return model.Payment{}, err
	}

	if course.Price <= 0 {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"amount": "Course price must be greater than zero",
			},
		}
	}

	payment := model.Payment{
		EnrollmentID:  enrollment.ID,
		Amount:        float64(course.Price),
		Currency:      model.PaymentCurrencyIRR,
		PaymentMethod: input.PaymentMethod,
		Status:        model.PaymentStatusPending,
		Description:   strings.TrimSpace(input.Description),
	}

	return s.paymentRepository.Create(ctx, payment)
}

// Update changes a payment status and transaction information.
func (s *PaymentService) Update(
	ctx context.Context,
	id int,
	input model.UpdatePaymentInput,
) (model.Payment, error) {
	if id <= 0 {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"id": "Payment ID must be a positive integer",
			},
		}
	}

	input = normalizeUpdatePaymentInput(input)

	validationErrors := validateUpdatePaymentInput(input)
	if len(validationErrors) > 0 {
		return model.Payment{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	existingPayment, err := s.paymentRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return model.Payment{}, err
	}

	if input.Status == existingPayment.Status {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"status": "Payment already has the requested status",
			},
		}
	}

	if !existingPayment.CanTransitionTo(input.Status) {
		return model.Payment{}, &ValidationError{
			Fields: map[string]string{
				"status": fmt.Sprintf(
					"Payment status cannot change from %s to %s",
					existingPayment.Status,
					input.Status,
				),
			},
		}
	}

	if err := validatePaymentStatusInformation(
		existingPayment,
		input,
	); err != nil {
		return model.Payment{}, err
	}

	currentTime := s.now().UTC()

	existingPayment.Status = input.Status
	existingPayment.TransactionID = input.TransactionID
	existingPayment.GatewayReference = input.GatewayReference
	existingPayment.CardLastFour = input.CardLastFour
	existingPayment.Description = input.Description

	switch input.Status {
	case model.PaymentStatusSucceeded:
		existingPayment.PaidAt = &currentTime
		existingPayment.FailureReason = ""

	case model.PaymentStatusFailed:
		existingPayment.FailedAt = &currentTime
		existingPayment.FailureReason = input.FailureReason

	case model.PaymentStatusRefunded:
		existingPayment.RefundedAt = &currentTime
		existingPayment.FailureReason = ""
	}

	updatedPayment, err := s.paymentRepository.Update(
		ctx,
		id,
		existingPayment,
	)
	if err != nil {
		return model.Payment{}, err
	}

	if updatedPayment.Status == model.PaymentStatusSucceeded {
		if err := s.confirmPendingEnrollment(
			ctx,
			updatedPayment.EnrollmentID,
			currentTime,
		); err != nil {
			return model.Payment{}, fmt.Errorf(
				"confirm enrollment after payment: %w",
				err,
			)
		}
	}

	return updatedPayment, nil
}

// Delete removes a pending or failed payment.
func (s *PaymentService) Delete(
	ctx context.Context,
	id int,
) error {
	if id <= 0 {
		return &ValidationError{
			Fields: map[string]string{
				"id": "Payment ID must be a positive integer",
			},
		}
	}

	payment, err := s.paymentRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if payment.Status == model.PaymentStatusSucceeded ||
		payment.Status == model.PaymentStatusRefunded {
		return &ValidationError{
			Fields: map[string]string{
				"status": "Successful or refunded payment cannot be deleted",
			},
		}
	}

	return s.paymentRepository.Delete(ctx, id)
}

// confirmPendingEnrollment confirms an enrollment after successful payment.
func (s *PaymentService) confirmPendingEnrollment(
	ctx context.Context,
	enrollmentID int,
	confirmedAt time.Time,
) error {
	enrollment, err := s.enrollmentRepository.GetByID(
		ctx,
		enrollmentID,
	)
	if err != nil {
		return err
	}

	// Completed and already confirmed enrollments need no further update.
	if enrollment.Status == model.EnrollmentStatusConfirmed ||
		enrollment.Status == model.EnrollmentStatusCompleted {
		return nil
	}

	if enrollment.Status != model.EnrollmentStatusPending {
		return &ValidationError{
			Fields: map[string]string{
				"enrollment_id": "Enrollment cannot be confirmed after payment",
			},
		}
	}

	enrollment.Status = model.EnrollmentStatusConfirmed

	if enrollment.ConfirmedAt == nil {
		enrollment.ConfirmedAt = &confirmedAt
	}

	_, err = s.enrollmentRepository.Update(
		ctx,
		enrollmentID,
		enrollment,
	)

	return err
}

// validateCreatePaymentInput validates payment creation information.
func validateCreatePaymentInput(
	input model.CreatePaymentInput,
) map[string]string {
	validationErrors := make(map[string]string)

	if input.EnrollmentID <= 0 {
		validationErrors["enrollment_id"] =
			"Enrollment ID must be a positive integer"
	}

	if !isValidPaymentMethod(input.PaymentMethod) {
		validationErrors["payment_method"] =
			"Payment method must be online, card, cash, or bank_transfer"
	}

	if len(strings.TrimSpace(input.Description)) > 500 {
		validationErrors["description"] =
			"Description must not exceed 500 characters"
	}

	return validationErrors
}

// validateUpdatePaymentInput validates payment update information.
func validateUpdatePaymentInput(
	input model.UpdatePaymentInput,
) map[string]string {
	validationErrors := make(map[string]string)

	if !isValidPaymentStatus(input.Status) {
		validationErrors["status"] =
			"Status must be succeeded, failed, or refunded"
	}

	if len(input.TransactionID) > 150 {
		validationErrors["transaction_id"] =
			"Transaction ID must not exceed 150 characters"
	}

	if len(input.GatewayReference) > 150 {
		validationErrors["gateway_reference"] =
			"Gateway reference must not exceed 150 characters"
	}

	if input.CardLastFour != "" &&
		!isFourDigits(input.CardLastFour) {
		validationErrors["card_last_four"] =
			"Card last four must contain exactly 4 digits"
	}

	if len(input.FailureReason) > 500 {
		validationErrors["failure_reason"] =
			"Failure reason must not exceed 500 characters"
	}

	if len(input.Description) > 500 {
		validationErrors["description"] =
			"Description must not exceed 500 characters"
	}

	return validationErrors
}

// validatePaymentStatusInformation validates status-specific fields.
func validatePaymentStatusInformation(
	existingPayment model.Payment,
	input model.UpdatePaymentInput,
) error {
	switch input.Status {
	case model.PaymentStatusSucceeded:
		if existingPayment.PaymentMethod != model.PaymentMethodCash &&
			input.TransactionID == "" {
			return &ValidationError{
				Fields: map[string]string{
					"transaction_id": "Transaction ID is required for successful payment",
				},
			}
		}

	case model.PaymentStatusFailed:
		if input.FailureReason == "" {
			return &ValidationError{
				Fields: map[string]string{
					"failure_reason": "Failure reason is required for failed payment",
				},
			}
		}

	case model.PaymentStatusRefunded:
		if existingPayment.Status != model.PaymentStatusSucceeded {
			return &ValidationError{
				Fields: map[string]string{
					"status": "Only a successful payment can be refunded",
				},
			}
		}
	}

	return nil
}

// normalizeUpdatePaymentInput normalizes optional payment fields.
func normalizeUpdatePaymentInput(
	input model.UpdatePaymentInput,
) model.UpdatePaymentInput {
	input.TransactionID = strings.ToUpper(
		strings.TrimSpace(input.TransactionID),
	)
	input.GatewayReference = strings.TrimSpace(
		input.GatewayReference,
	)
	input.CardLastFour = strings.TrimSpace(
		input.CardLastFour,
	)
	input.FailureReason = strings.TrimSpace(
		input.FailureReason,
	)
	input.Description = strings.TrimSpace(
		input.Description,
	)

	return input
}

// isValidPaymentMethod reports whether a payment method is supported.
func isValidPaymentMethod(
	method model.PaymentMethod,
) bool {
	switch method {
	case model.PaymentMethodOnline,
		model.PaymentMethodCard,
		model.PaymentMethodCash,
		model.PaymentMethodBankTransfer:
		return true

	default:
		return false
	}
}

// isValidPaymentStatus reports whether a mutable status is supported.
func isValidPaymentStatus(
	status model.PaymentStatus,
) bool {
	switch status {
	case model.PaymentStatusSucceeded,
		model.PaymentStatusFailed,
		model.PaymentStatusRefunded:
		return true

	default:
		return false
	}
}

// isFourDigits reports whether a string contains exactly four digits.
func isFourDigits(value string) bool {
	if len(value) != 4 {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}
