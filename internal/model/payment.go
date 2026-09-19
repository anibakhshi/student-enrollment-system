package model

import "time"

// PaymentStatus represents the current payment lifecycle status.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod represents the method used to pay.
type PaymentMethod string

const (
	PaymentMethodOnline       PaymentMethod = "online"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// PaymentCurrency represents the payment currency.
type PaymentCurrency string

const (
	PaymentCurrencyIRR PaymentCurrency = "IRR"
	PaymentCurrencyUSD PaymentCurrency = "USD"
	PaymentCurrencyEUR PaymentCurrency = "EUR"
)

// Payment contains information about an enrollment payment.
type Payment struct {
	ID               int             `json:"id"`
	EnrollmentID     int             `json:"enrollment_id"`
	Amount           float64         `json:"amount"`
	Currency         PaymentCurrency `json:"currency"`
	PaymentMethod    PaymentMethod   `json:"payment_method"`
	Status           PaymentStatus   `json:"status"`
	TransactionID    string          `json:"transaction_id,omitempty"`
	GatewayReference string          `json:"gateway_reference,omitempty"`
	CardLastFour     string          `json:"card_last_four,omitempty"`
	FailureReason    string          `json:"failure_reason,omitempty"`
	Description      string          `json:"description,omitempty"`
	PaidAt           *time.Time      `json:"paid_at,omitempty"`
	FailedAt         *time.Time      `json:"failed_at,omitempty"`
	RefundedAt       *time.Time      `json:"refunded_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// PaymentDetails contains payment and enrollment information.
type PaymentDetails struct {
	Payment    Payment    `json:"payment"`
	Enrollment Enrollment `json:"enrollment"`
	Student    Student    `json:"student"`
	Course     Course     `json:"course"`
}

// CreatePaymentInput contains information required to create a payment.
//
// Amount and currency are not accepted from the client. The service obtains
// the authoritative price from the enrollment's course and uses IRR.
type CreatePaymentInput struct {
	EnrollmentID  int           `json:"enrollment_id"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Description   string        `json:"description"`
}

// UpdatePaymentInput contains information used to change payment status.
type UpdatePaymentInput struct {
	Status           PaymentStatus `json:"status"`
	TransactionID    string        `json:"transaction_id"`
	GatewayReference string        `json:"gateway_reference"`
	CardLastFour     string        `json:"card_last_four"`
	FailureReason    string        `json:"failure_reason"`
	Description      string        `json:"description"`
}

// IsPending reports whether the payment is waiting for a result.
func (p Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsSucceeded reports whether the payment was completed successfully.
func (p Payment) IsSucceeded() bool {
	return p.Status == PaymentStatusSucceeded
}

// IsFailed reports whether the payment failed.
func (p Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsRefunded reports whether the payment was refunded.
func (p Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// CanTransitionTo reports whether a payment status transition is allowed.
func (p Payment) CanTransitionTo(
	nextStatus PaymentStatus,
) bool {
	switch p.Status {
	case PaymentStatusPending:
		return nextStatus == PaymentStatusSucceeded ||
			nextStatus == PaymentStatusFailed

	case PaymentStatusSucceeded:
		return nextStatus == PaymentStatusRefunded

	case PaymentStatusFailed,
		PaymentStatusRefunded:
		return false

	default:
		return false
	}
}
