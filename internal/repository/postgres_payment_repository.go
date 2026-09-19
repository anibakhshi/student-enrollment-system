package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPaymentRepository stores payments in PostgreSQL.
type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPaymentRepository creates a PostgreSQL payment repository.
func NewPostgresPaymentRepository(
	pool *pgxpool.Pool,
) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{
		pool: pool,
	}
}

// List returns all non-deleted payments.
func (r *PostgresPaymentRepository) List(
	ctx context.Context,
) ([]model.Payment, error) {
	const query = `
		SELECT
			id,
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			COALESCE(transaction_id, ''),
			COALESCE(gateway_reference, ''),
			COALESCE(card_last_four, ''),
			COALESCE(failure_reason, ''),
			COALESCE(description, ''),
			paid_at,
			failed_at,
			refunded_at,
			created_at,
			updated_at
		FROM payments
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()

	payments := make([]model.Payment, 0)

	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payments: %w", err)
	}

	return payments, nil
}

// GetByID returns one non-deleted payment by ID.
func (r *PostgresPaymentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Payment, error) {
	const query = `
		SELECT
			id,
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			COALESCE(transaction_id, ''),
			COALESCE(gateway_reference, ''),
			COALESCE(card_last_four, ''),
			COALESCE(failure_reason, ''),
			COALESCE(description, ''),
			paid_at,
			failed_at,
			refunded_at,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	payment, err := scanPayment(
		r.pool.QueryRow(ctx, query, id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, ErrPaymentNotFound
		}

		return model.Payment{}, fmt.Errorf(
			"query payment by ID: %w",
			err,
		)
	}

	return payment, nil
}

// ListByEnrollmentID returns all non-deleted payments for an enrollment.
func (r *PostgresPaymentRepository) ListByEnrollmentID(
	ctx context.Context,
	enrollmentID int,
) ([]model.Payment, error) {
	const query = `
		SELECT
			id,
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			COALESCE(transaction_id, ''),
			COALESCE(gateway_reference, ''),
			COALESCE(card_last_four, ''),
			COALESCE(failure_reason, ''),
			COALESCE(description, ''),
			paid_at,
			failed_at,
			refunded_at,
			created_at,
			updated_at
		FROM payments
		WHERE enrollment_id = $1
		  AND deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		enrollmentID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query enrollment payments: %w",
			err,
		)
	}
	defer rows.Close()

	payments := make([]model.Payment, 0)

	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate enrollment payments: %w",
			err,
		)
	}

	return payments, nil
}

// Create stores a new payment.
func (r *PostgresPaymentRepository) Create(
	ctx context.Context,
	payment model.Payment,
) (model.Payment, error) {
	const query = `
		INSERT INTO payments (
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			transaction_id,
			gateway_reference,
			card_last_four,
			failure_reason,
			description,
			paid_at,
			failed_at,
			refunded_at
		)
		VALUES (
			$1,
			$2,
			$3::VARCHAR(10),
			$4::VARCHAR(30),
			$5::VARCHAR(20),
			NULLIF($6::TEXT, ''),
			NULLIF($7::TEXT, ''),
			NULLIF($8::TEXT, ''),
			NULLIF($9::TEXT, ''),
			NULLIF($10::TEXT, ''),
			$11::TIMESTAMPTZ,
			$12::TIMESTAMPTZ,
			$13::TIMESTAMPTZ
		)
		RETURNING
			id,
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			COALESCE(transaction_id, ''),
			COALESCE(gateway_reference, ''),
			COALESCE(card_last_four, ''),
			COALESCE(failure_reason, ''),
			COALESCE(description, ''),
			paid_at,
			failed_at,
			refunded_at,
			created_at,
			updated_at;
	`

	createdPayment, err := scanPayment(
		r.pool.QueryRow(
			ctx,
			query,
			payment.EnrollmentID,
			payment.Amount,
			string(payment.Currency),
			string(payment.PaymentMethod),
			string(payment.Status),
			payment.TransactionID,
			payment.GatewayReference,
			payment.CardLastFour,
			payment.FailureReason,
			payment.Description,
			payment.PaidAt,
			payment.FailedAt,
			payment.RefundedAt,
		),
	)
	if err != nil {
		return model.Payment{}, mapPostgresPaymentError(err)
	}

	return createdPayment, nil
}

// Update changes a payment's status and transaction information.
func (r *PostgresPaymentRepository) Update(
	ctx context.Context,
	id int,
	payment model.Payment,
) (model.Payment, error) {
	const query = `
		UPDATE payments
		SET
			status = $2::VARCHAR(20),
			transaction_id = NULLIF($3::TEXT, ''),
			gateway_reference = NULLIF($4::TEXT, ''),
			card_last_four = NULLIF($5::TEXT, ''),
			failure_reason = NULLIF($6::TEXT, ''),
			description = NULLIF($7::TEXT, ''),
			paid_at = $8::TIMESTAMPTZ,
			failed_at = $9::TIMESTAMPTZ,
			refunded_at = $10::TIMESTAMPTZ
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			enrollment_id,
			amount,
			currency,
			payment_method,
			status,
			COALESCE(transaction_id, ''),
			COALESCE(gateway_reference, ''),
			COALESCE(card_last_four, ''),
			COALESCE(failure_reason, ''),
			COALESCE(description, ''),
			paid_at,
			failed_at,
			refunded_at,
			created_at,
			updated_at;
	`

	updatedPayment, err := scanPayment(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			string(payment.Status),
			payment.TransactionID,
			payment.GatewayReference,
			payment.CardLastFour,
			payment.FailureReason,
			payment.Description,
			payment.PaidAt,
			payment.FailedAt,
			payment.RefundedAt,
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, ErrPaymentNotFound
		}

		return model.Payment{}, mapPostgresPaymentError(err)
	}

	return updatedPayment, nil
}

// Delete soft-deletes a payment.
func (r *PostgresPaymentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	const query = `
		UPDATE payments
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-delete payment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

// paymentScanner is implemented by pgx.Row and pgx.Rows.
type paymentScanner interface {
	Scan(destinations ...any) error
}

// scanPayment scans a PostgreSQL row into a payment model.
func scanPayment(
	row paymentScanner,
) (model.Payment, error) {
	var payment model.Payment

	err := row.Scan(
		&payment.ID,
		&payment.EnrollmentID,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.TransactionID,
		&payment.GatewayReference,
		&payment.CardLastFour,
		&payment.FailureReason,
		&payment.Description,
		&payment.PaidAt,
		&payment.FailedAt,
		&payment.RefundedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return model.Payment{}, err
	}

	return payment, nil
}

// mapPostgresPaymentError converts PostgreSQL errors to domain errors.
func mapPostgresPaymentError(err error) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch postgresError.ConstraintName {
		case "idx_payments_one_active_per_enrollment":
			return ErrPaymentExists

		case "idx_payments_transaction_id_unique":
			return ErrTransactionIDExists
		}
	}

	return fmt.Errorf("payment repository operation: %w", err)
}
