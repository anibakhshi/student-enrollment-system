package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// MemoryPaymentRepository stores payments in memory.
type MemoryPaymentRepository struct {
	mutex    sync.RWMutex
	payments map[int]model.Payment
	nextID   int
}

// NewMemoryPaymentRepository creates an empty payment repository.
func NewMemoryPaymentRepository() *MemoryPaymentRepository {
	return &MemoryPaymentRepository{
		payments: make(map[int]model.Payment),
		nextID:   1,
	}
}

// List returns all payments ordered by ID.
func (r *MemoryPaymentRepository) List(
	ctx context.Context,
) ([]model.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	payments := make([]model.Payment, 0, len(r.payments))

	for _, payment := range r.payments {
		payments = append(
			payments,
			clonePayment(payment),
		)
	}

	sort.Slice(
		payments,
		func(firstIndex, secondIndex int) bool {
			return payments[firstIndex].ID <
				payments[secondIndex].ID
		},
	)

	return payments, nil
}

// GetByID returns a payment by ID.
func (r *MemoryPaymentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Payment, error) {
	if err := ctx.Err(); err != nil {
		return model.Payment{}, err
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	payment, exists := r.payments[id]
	if !exists {
		return model.Payment{}, ErrPaymentNotFound
	}

	return clonePayment(payment), nil
}

// ListByEnrollmentID returns all payments for an enrollment.
func (r *MemoryPaymentRepository) ListByEnrollmentID(
	ctx context.Context,
	enrollmentID int,
) ([]model.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	payments := make([]model.Payment, 0)

	for _, payment := range r.payments {
		if payment.EnrollmentID == enrollmentID {
			payments = append(
				payments,
				clonePayment(payment),
			)
		}
	}

	sort.Slice(
		payments,
		func(firstIndex, secondIndex int) bool {
			return payments[firstIndex].ID <
				payments[secondIndex].ID
		},
	)

	return payments, nil
}

// Create stores a new payment.
func (r *MemoryPaymentRepository) Create(
	ctx context.Context,
	payment model.Payment,
) (model.Payment, error) {
	if err := ctx.Err(); err != nil {
		return model.Payment{}, err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.hasActivePaymentForEnrollment(
		payment.EnrollmentID,
		0,
	) {
		return model.Payment{}, ErrPaymentExists
	}

	if r.transactionIDExists(
		payment.TransactionID,
		0,
	) {
		return model.Payment{}, ErrTransactionIDExists
	}

	currentTime := time.Now().UTC()

	if payment.Status == "" {
		payment.Status = model.PaymentStatusPending
	}

	if payment.Currency == "" {
		payment.Currency = model.PaymentCurrencyIRR
	}

	payment.ID = r.nextID
	payment.CreatedAt = currentTime
	payment.UpdatedAt = currentTime

	r.payments[payment.ID] = clonePayment(payment)
	r.nextID++

	return clonePayment(payment), nil
}

// Update changes an existing payment.
func (r *MemoryPaymentRepository) Update(
	ctx context.Context,
	id int,
	payment model.Payment,
) (model.Payment, error) {
	if err := ctx.Err(); err != nil {
		return model.Payment{}, err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingPayment, exists := r.payments[id]
	if !exists {
		return model.Payment{}, ErrPaymentNotFound
	}

	if r.transactionIDExists(
		payment.TransactionID,
		id,
	) {
		return model.Payment{}, ErrTransactionIDExists
	}

	if isActivePaymentStatus(payment.Status) &&
		r.hasActivePaymentForEnrollment(
			existingPayment.EnrollmentID,
			id,
		) {
		return model.Payment{}, ErrPaymentExists
	}

	payment.ID = existingPayment.ID
	payment.EnrollmentID = existingPayment.EnrollmentID
	payment.Amount = existingPayment.Amount
	payment.Currency = existingPayment.Currency
	payment.PaymentMethod = existingPayment.PaymentMethod
	payment.CreatedAt = existingPayment.CreatedAt
	payment.UpdatedAt = time.Now().UTC()

	r.payments[id] = clonePayment(payment)

	return clonePayment(payment), nil
}

// Delete removes a payment from the memory repository.
func (r *MemoryPaymentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.payments[id]; !exists {
		return ErrPaymentNotFound
	}

	delete(r.payments, id)

	return nil
}

// hasActivePaymentForEnrollment checks for another pending or
// successful payment belonging to the enrollment.
func (r *MemoryPaymentRepository) hasActivePaymentForEnrollment(
	enrollmentID int,
	excludedPaymentID int,
) bool {
	for paymentID, payment := range r.payments {
		if paymentID == excludedPaymentID {
			continue
		}

		if payment.EnrollmentID == enrollmentID &&
			isActivePaymentStatus(payment.Status) {
			return true
		}
	}

	return false
}

// transactionIDExists checks for a duplicate non-empty transaction ID.
func (r *MemoryPaymentRepository) transactionIDExists(
	transactionID string,
	excludedPaymentID int,
) bool {
	normalizedTransactionID := strings.TrimSpace(transactionID)

	if normalizedTransactionID == "" {
		return false
	}

	for paymentID, payment := range r.payments {
		if paymentID == excludedPaymentID {
			continue
		}

		if strings.EqualFold(
			strings.TrimSpace(payment.TransactionID),
			normalizedTransactionID,
		) {
			return true
		}
	}

	return false
}

// isActivePaymentStatus reports whether the status occupies the
// enrollment's single active payment slot.
func isActivePaymentStatus(
	status model.PaymentStatus,
) bool {
	return status == model.PaymentStatusPending ||
		status == model.PaymentStatusSucceeded
}

// clonePayment returns a payment with independent timestamp pointers.
func clonePayment(
	payment model.Payment,
) model.Payment {
	payment.PaidAt = cloneTimePointer(payment.PaidAt)
	payment.FailedAt = cloneTimePointer(payment.FailedAt)
	payment.RefundedAt = cloneTimePointer(payment.RefundedAt)

	return payment
}

// cloneTimePointer returns an independent copy of a time pointer.
func cloneTimePointer(
	value *time.Time,
) *time.Time {
	if value == nil {
		return nil
	}

	copiedValue := *value

	return &copiedValue
}
