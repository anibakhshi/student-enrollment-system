package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// setupPaymentService creates isolated repositories and one pending enrollment.
func setupPaymentService(
	t *testing.T,
) (
	*PaymentService,
	*repository.MemoryPaymentRepository,
	*repository.MemoryEnrollmentRepository,
) {
	t.Helper()

	studentRepository :=
		repository.NewMemoryStudentRepository()

	courseRepository :=
		repository.NewMemoryCourseRepository()

	enrollmentRepository :=
		repository.NewMemoryEnrollmentRepository()

	paymentRepository :=
		repository.NewMemoryPaymentRepository()

	_, err := enrollmentRepository.Create(
		context.Background(),
		model.Enrollment{
			StudentID: 1,
			CourseID:  1,
			Status:    model.EnrollmentStatusPending,
			Notes:     "Payment service test enrollment",
		},
		25,
	)
	if err != nil {
		t.Fatalf(
			"failed to create test enrollment: %v",
			err,
		)
	}

	paymentService := NewPaymentService(
		paymentRepository,
		enrollmentRepository,
		studentRepository,
		courseRepository,
	)

	paymentService.now = func() time.Time {
		return time.Date(
			2027,
			time.January,
			10,
			12,
			30,
			0,
			0,
			time.UTC,
		)
	}

	return paymentService,
		paymentRepository,
		enrollmentRepository
}

func TestPaymentServiceList(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	_, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
			Description:   "Course payment",
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	payments, err := paymentService.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if len(payments) != 1 {
		t.Fatalf(
			"expected 1 payment, got %d",
			len(payments),
		)
	}
}

func TestPaymentServiceCreateUsesCoursePrice(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
			Description:   "Data science course payment",
		},
	)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if payment.ID != 1 {
		t.Fatalf(
			"expected payment ID 1, got %d",
			payment.ID,
		)
	}

	if payment.Amount != 15000000 {
		t.Errorf(
			"expected amount 15000000, got %.2f",
			payment.Amount,
		)
	}

	if payment.Currency != model.PaymentCurrencyIRR {
		t.Errorf(
			"expected IRR currency, got %q",
			payment.Currency,
		)
	}

	if payment.Status != model.PaymentStatusPending {
		t.Errorf(
			"expected pending status, got %q",
			payment.Status,
		)
	}
}

func TestPaymentServiceCreateValidation(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	_, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  0,
			PaymentMethod: model.PaymentMethod("crypto"),
			Description:   string(make([]byte, 501)),
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	expectedFields := []string{
		"enrollment_id",
		"payment_method",
		"description",
	}

	for _, field := range expectedFields {
		if _, exists := validationError.Fields[field]; !exists {
			t.Errorf(
				"expected validation error for %q",
				field,
			)
		}
	}
}

func TestPaymentServiceRejectsMissingEnrollment(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	_, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  999,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)

	if !errors.Is(err, repository.ErrEnrollmentNotFound) {
		t.Fatalf(
			"expected ErrEnrollmentNotFound, got %v",
			err,
		)
	}
}

func TestPaymentServiceRejectsCancelledEnrollment(t *testing.T) {
	paymentService, _, enrollmentRepository :=
		setupPaymentService(t)

	enrollment, err := enrollmentRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("failed to get enrollment: %v", err)
	}

	currentTime := time.Now().UTC()

	enrollment.Status = model.EnrollmentStatusCancelled
	enrollment.CancelledAt = &currentTime

	_, err = enrollmentRepository.Update(
		context.Background(),
		enrollment.ID,
		enrollment,
	)
	if err != nil {
		t.Fatalf("failed to cancel enrollment: %v", err)
	}

	_, err = paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["enrollment_id"]; !exists {
		t.Error(
			"expected enrollment_id validation error",
		)
	}
}

func TestPaymentServiceRejectsDuplicateActivePayment(
	t *testing.T,
) {
	paymentService, _, _ := setupPaymentService(t)

	input := model.CreatePaymentInput{
		EnrollmentID:  1,
		PaymentMethod: model.PaymentMethodOnline,
	}

	_, err := paymentService.Create(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	_, err = paymentService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrPaymentExists) {
		t.Fatalf(
			"expected ErrPaymentExists, got %v",
			err,
		)
	}
}

func TestPaymentServiceSuccessfulPaymentConfirmsEnrollment(
	t *testing.T,
) {
	paymentService, _, enrollmentRepository :=
		setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	updatedPayment, err := paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status:           model.PaymentStatusSucceeded,
			TransactionID:    "txn-success-001",
			GatewayReference: "gateway-001",
			CardLastFour:     "1234",
			Description:      "Payment verified",
		},
	)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if updatedPayment.Status != model.PaymentStatusSucceeded {
		t.Errorf(
			"expected succeeded status, got %q",
			updatedPayment.Status,
		)
	}

	if updatedPayment.TransactionID != "TXN-SUCCESS-001" {
		t.Errorf(
			"unexpected transaction ID: %q",
			updatedPayment.TransactionID,
		)
	}

	if updatedPayment.PaidAt == nil {
		t.Fatal("expected paid_at to be set")
	}

	enrollment, err := enrollmentRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("failed to get enrollment: %v", err)
	}

	if enrollment.Status != model.EnrollmentStatusConfirmed {
		t.Errorf(
			"expected confirmed enrollment, got %q",
			enrollment.Status,
		)
	}

	if enrollment.ConfirmedAt == nil {
		t.Error("expected enrollment confirmed_at to be set")
	}
}

func TestPaymentServiceRejectsSuccessWithoutTransactionID(
	t *testing.T,
) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	_, err = paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status: model.PaymentStatusSucceeded,
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["transaction_id"]; !exists {
		t.Error(
			"expected transaction_id validation error",
		)
	}
}

func TestPaymentServiceFailedPaymentAllowsRetry(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	failedPayment, err := paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status:        model.PaymentStatusFailed,
			FailureReason: "Gateway timeout",
		},
	)
	if err != nil {
		t.Fatalf("failed to update payment: %v", err)
	}

	if failedPayment.Status != model.PaymentStatusFailed {
		t.Errorf(
			"expected failed status, got %q",
			failedPayment.Status,
		)
	}

	if failedPayment.FailedAt == nil {
		t.Error("expected failed_at to be set")
	}

	retryPayment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodCard,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected retry payment to succeed, got %v",
			err,
		)
	}

	if retryPayment.ID == payment.ID {
		t.Error("expected retry payment to have a new ID")
	}
}

func TestPaymentServiceRejectsFailureWithoutReason(
	t *testing.T,
) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	_, err = paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status: model.PaymentStatusFailed,
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["failure_reason"]; !exists {
		t.Error(
			"expected failure_reason validation error",
		)
	}
}

func TestPaymentServiceRefund(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	successfulPayment, err := paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status:        model.PaymentStatusSucceeded,
			TransactionID: "refund-test-001",
		},
	)
	if err != nil {
		t.Fatalf("failed to succeed payment: %v", err)
	}

	refundedPayment, err := paymentService.Update(
		context.Background(),
		successfulPayment.ID,
		model.UpdatePaymentInput{
			Status:           model.PaymentStatusRefunded,
			TransactionID:    successfulPayment.TransactionID,
			GatewayReference: "refund-reference-001",
			Description:      "Customer refund",
		},
	)
	if err != nil {
		t.Fatalf("failed to refund payment: %v", err)
	}

	if refundedPayment.Status != model.PaymentStatusRefunded {
		t.Errorf(
			"expected refunded status, got %q",
			refundedPayment.Status,
		)
	}

	if refundedPayment.RefundedAt == nil {
		t.Error("expected refunded_at to be set")
	}
}

func TestPaymentServiceRejectsInvalidTransition(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodOnline,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	failedPayment, err := paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status:        model.PaymentStatusFailed,
			FailureReason: "Payment rejected",
		},
	)
	if err != nil {
		t.Fatalf("failed to mark payment failed: %v", err)
	}

	_, err = paymentService.Update(
		context.Background(),
		failedPayment.ID,
		model.UpdatePaymentInput{
			Status:        model.PaymentStatusSucceeded,
			TransactionID: "invalid-transition",
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists := validationError.Fields["status"]; !exists {
		t.Error("expected status validation error")
	}
}

func TestPaymentServiceGetDetails(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodCash,
			Description:   "Cash payment",
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	details, err := paymentService.GetDetails(
		context.Background(),
		payment.ID,
	)
	if err != nil {
		t.Fatalf("unexpected details error: %v", err)
	}

	if details.Payment.ID != payment.ID {
		t.Errorf(
			"expected payment ID %d, got %d",
			payment.ID,
			details.Payment.ID,
		)
	}

	if details.Enrollment.ID != payment.EnrollmentID {
		t.Errorf(
			"expected enrollment ID %d, got %d",
			payment.EnrollmentID,
			details.Enrollment.ID,
		)
	}

	if details.Student.ID != 1 {
		t.Errorf(
			"expected student ID 1, got %d",
			details.Student.ID,
		)
	}

	if details.Course.ID != 1 {
		t.Errorf(
			"expected course ID 1, got %d",
			details.Course.ID,
		)
	}
}

func TestPaymentServiceListByEnrollmentID(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	_, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodCash,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	payments, err := paymentService.ListByEnrollmentID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if len(payments) != 1 {
		t.Fatalf(
			"expected 1 payment, got %d",
			len(payments),
		)
	}
}

func TestPaymentServiceDelete(t *testing.T) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodCash,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	err = paymentService.Delete(
		context.Background(),
		payment.ID,
	)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	_, err = paymentService.GetByID(
		context.Background(),
		payment.ID,
	)

	if !errors.Is(err, repository.ErrPaymentNotFound) {
		t.Fatalf(
			"expected ErrPaymentNotFound, got %v",
			err,
		)
	}
}

func TestPaymentServiceRejectsDeletingSuccessfulPayment(
	t *testing.T,
) {
	paymentService, _, _ := setupPaymentService(t)

	payment, err := paymentService.Create(
		context.Background(),
		model.CreatePaymentInput{
			EnrollmentID:  1,
			PaymentMethod: model.PaymentMethodCash,
		},
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	successfulPayment, err := paymentService.Update(
		context.Background(),
		payment.ID,
		model.UpdatePaymentInput{
			Status:      model.PaymentStatusSucceeded,
			Description: "Cash received",
		},
	)
	if err != nil {
		t.Fatalf("failed to succeed payment: %v", err)
	}

	err = paymentService.Delete(
		context.Background(),
		successfulPayment.ID,
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists := validationError.Fields["status"]; !exists {
		t.Error("expected status validation error")
	}
}
