package repository

import (
	"context"
	"errors"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

var (
	// ErrPaymentNotFound indicates that a payment does not exist.
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrPaymentExists indicates that an active payment already exists.
	ErrPaymentExists = errors.New(
		"an active payment already exists for this enrollment",
	)

	// ErrTransactionIDExists indicates that a transaction ID is duplicated.
	ErrTransactionIDExists = errors.New(
		"payment transaction ID already exists",
	)
)

// PaymentRepository defines payment data operations.
type PaymentRepository interface {
	List(
		ctx context.Context,
	) ([]model.Payment, error)

	GetByID(
		ctx context.Context,
		id int,
	) (model.Payment, error)

	ListByEnrollmentID(
		ctx context.Context,
		enrollmentID int,
	) ([]model.Payment, error)

	Create(
		ctx context.Context,
		payment model.Payment,
	) (model.Payment, error)

	Update(
		ctx context.Context,
		id int,
		payment model.Payment,
	) (model.Payment, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}
