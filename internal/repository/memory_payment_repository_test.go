package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// newTestPayment creates valid payment information for repository tests.
func newTestPayment(
	enrollmentID int,
	transactionID string,
) model.Payment {
	return model.Payment{
		EnrollmentID:  enrollmentID,
		Amount:        15000000,
		Currency:      model.PaymentCurrencyIRR,
		PaymentMethod: model.PaymentMethodOnline,
		Status:        model.PaymentStatusPending,
		TransactionID: transactionID,
		Description:   "Test payment",
	}
}

func TestMemoryPaymentRepositoryCreate(t *testing.T) {
	paymentRepository := NewMemoryPaymentRepository()

	createdPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdPayment.ID != 1 {
		t.Fatalf(
			"expected payment ID 1, got %d",
			createdPayment.ID,
		)
	}

	if createdPayment.EnrollmentID != 1 {
		t.Errorf(
			"expected enrollment ID 1, got %d",
			createdPayment.EnrollmentID,
		)
	}

	if createdPayment.Status != model.PaymentStatusPending {
		t.Errorf(
			"expected pending status, got %q",
			createdPayment.Status,
		)
	}

	if createdPayment.Currency != model.PaymentCurrencyIRR {
		t.Errorf(
			"expected IRR currency, got %q",
			createdPayment.Currency,
		)
	}

	if createdPayment.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}

	if createdPayment.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestMemoryPaymentRepositoryListAndGetByID(t *testing.T) {
	paymentRepository := NewMemoryPaymentRepository()

	firstPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf(
			"failed to create first payment: %v",
			err,
		)
	}

	secondPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(2, ""),
	)
	if err != nil {
		t.Fatalf(
			"failed to create second payment: %v",
			err,
		)
	}

	payments, err := paymentRepository.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if len(payments) != 2 {
		t.Fatalf(
			"expected 2 payments, got %d",
			len(payments),
		)
	}

	if payments[0].ID != firstPayment.ID {
		t.Errorf(
			"expected first payment ID %d, got %d",
			firstPayment.ID,
			payments[0].ID,
		)
	}

	if payments[1].ID != secondPayment.ID {
		t.Errorf(
			"expected second payment ID %d, got %d",
			secondPayment.ID,
			payments[1].ID,
		)
	}

	foundPayment, err := paymentRepository.GetByID(
		context.Background(),
		firstPayment.ID,
	)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	if foundPayment.ID != firstPayment.ID {
		t.Errorf(
			"expected payment ID %d, got %d",
			firstPayment.ID,
			foundPayment.ID,
		)
	}
}

func TestMemoryPaymentRepositoryReturnsNotFound(t *testing.T) {
	paymentRepository := NewMemoryPaymentRepository()

	_, err := paymentRepository.GetByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf(
			"expected ErrPaymentNotFound, got %v",
			err,
		)
	}
}

func TestMemoryPaymentRepositoryListByEnrollmentID(
	t *testing.T,
) {
	paymentRepository := NewMemoryPaymentRepository()

	firstPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	failedPayment := firstPayment
	failedPayment.Status = model.PaymentStatusFailed

	currentTime := time.Now().UTC()
	failedPayment.FailedAt = &currentTime

	_, err = paymentRepository.Update(
		context.Background(),
		firstPayment.ID,
		failedPayment,
	)
	if err != nil {
		t.Fatalf("failed to mark payment as failed: %v", err)
	}

	_, err = paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create retry payment: %v", err)
	}

	_, err = paymentRepository.Create(
		context.Background(),
		newTestPayment(2, ""),
	)
	if err != nil {
		t.Fatalf(
			"failed to create another enrollment payment: %v",
			err,
		)
	}

	payments, err := paymentRepository.ListByEnrollmentID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if len(payments) != 2 {
		t.Fatalf(
			"expected 2 payments for enrollment 1, got %d",
			len(payments),
		)
	}

	for _, payment := range payments {
		if payment.EnrollmentID != 1 {
			t.Errorf(
				"unexpected enrollment ID: %d",
				payment.EnrollmentID,
			)
		}
	}
}

func TestMemoryPaymentRepositoryRejectsActiveDuplicate(
	t *testing.T,
) {
	paymentRepository := NewMemoryPaymentRepository()

	_, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	_, err = paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)

	if !errors.Is(err, ErrPaymentExists) {
		t.Fatalf(
			"expected ErrPaymentExists, got %v",
			err,
		)
	}
}

func TestMemoryPaymentRepositoryAllowsRetryAfterFailure(
	t *testing.T,
) {
	paymentRepository := NewMemoryPaymentRepository()

	createdPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	failedPayment := createdPayment
	failedPayment.Status = model.PaymentStatusFailed
	failedPayment.FailureReason = "Gateway timeout"

	currentTime := time.Now().UTC()
	failedPayment.FailedAt = &currentTime

	updatedPayment, err := paymentRepository.Update(
		context.Background(),
		createdPayment.ID,
		failedPayment,
	)
	if err != nil {
		t.Fatalf("failed to update payment: %v", err)
	}

	if updatedPayment.Status != model.PaymentStatusFailed {
		t.Fatalf(
			"expected failed status, got %q",
			updatedPayment.Status,
		)
	}

	retryPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf(
			"expected retry payment to succeed, got %v",
			err,
		)
	}

	if retryPayment.ID == createdPayment.ID {
		t.Error("expected retry payment to have a new ID")
	}
}

func TestMemoryPaymentRepositoryRejectsDuplicateTransactionID(
	t *testing.T,
) {
	paymentRepository := NewMemoryPaymentRepository()

	firstPayment := newTestPayment(
		1,
		"TXN-10001",
	)

	_, err := paymentRepository.Create(
		context.Background(),
		firstPayment,
	)
	if err != nil {
		t.Fatalf("failed to create first payment: %v", err)
	}

	secondPayment := newTestPayment(
		2,
		"txn-10001",
	)

	_, err = paymentRepository.Create(
		context.Background(),
		secondPayment,
	)

	if !errors.Is(err, ErrTransactionIDExists) {
		t.Fatalf(
			"expected ErrTransactionIDExists, got %v",
			err,
		)
	}
}

func TestMemoryPaymentRepositoryUpdate(t *testing.T) {
	paymentRepository := NewMemoryPaymentRepository()

	createdPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	currentTime := time.Now().UTC()

	paymentToUpdate := createdPayment
	paymentToUpdate.Status = model.PaymentStatusSucceeded
	paymentToUpdate.TransactionID = "TXN-SUCCESS-001"
	paymentToUpdate.GatewayReference = "GATEWAY-001"
	paymentToUpdate.CardLastFour = "1234"
	paymentToUpdate.Description = "Payment completed"
	paymentToUpdate.PaidAt = &currentTime

	updatedPayment, err := paymentRepository.Update(
		context.Background(),
		createdPayment.ID,
		paymentToUpdate,
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
		t.Error("expected paid_at to be set")
	}

	if updatedPayment.EnrollmentID != createdPayment.EnrollmentID {
		t.Error("update changed the enrollment ID")
	}

	if updatedPayment.Amount != createdPayment.Amount {
		t.Error("update changed the payment amount")
	}
}

func TestMemoryPaymentRepositoryDelete(t *testing.T) {
	paymentRepository := NewMemoryPaymentRepository()

	createdPayment, err := paymentRepository.Create(
		context.Background(),
		newTestPayment(1, ""),
	)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	err = paymentRepository.Delete(
		context.Background(),
		createdPayment.ID,
	)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	_, err = paymentRepository.GetByID(
		context.Background(),
		createdPayment.ID,
	)

	if !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf(
			"expected ErrPaymentNotFound, got %v",
			err,
		)
	}

	err = paymentRepository.Delete(
		context.Background(),
		createdPayment.ID,
	)

	if !errors.Is(err, ErrPaymentNotFound) {
		t.Fatalf(
			"expected ErrPaymentNotFound on second delete, got %v",
			err,
		)
	}
}

func TestMemoryPaymentRepositoryConcurrentActivePayment(
	t *testing.T,
) {
	paymentRepository := NewMemoryPaymentRepository()

	const requestCount = 20

	results := make(chan error, requestCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(requestCount)

	for index := 0; index < requestCount; index++ {
		go func(requestNumber int) {
			defer waitGroup.Done()

			payment := newTestPayment(
				1,
				fmt.Sprintf(
					"CONCURRENT-%02d",
					requestNumber,
				),
			)

			_, err := paymentRepository.Create(
				context.Background(),
				payment,
			)

			results <- err
		}(index)
	}

	waitGroup.Wait()
	close(results)

	successCount := 0
	duplicateCount := 0

	for err := range results {
		switch {
		case err == nil:
			successCount++

		case errors.Is(err, ErrPaymentExists):
			duplicateCount++

		default:
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}

	if successCount != 1 {
		t.Fatalf(
			"expected exactly 1 successful payment, got %d",
			successCount,
		)
	}

	if duplicateCount != requestCount-1 {
		t.Fatalf(
			"expected %d duplicate errors, got %d",
			requestCount-1,
			duplicateCount,
		)
	}
}
