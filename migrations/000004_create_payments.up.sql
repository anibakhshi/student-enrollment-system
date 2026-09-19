BEGIN;

CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,

    enrollment_id BIGINT NOT NULL
        REFERENCES enrollments(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    amount NUMERIC(14, 2) NOT NULL
        CHECK (amount > 0),

    currency VARCHAR(10) NOT NULL DEFAULT 'IRR'
        CHECK (currency IN ('IRR', 'USD', 'EUR')),

    payment_method VARCHAR(30) NOT NULL
        CHECK (
            payment_method IN (
                'online',
                'card',
                'cash',
                'bank_transfer'
            )
        ),

    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (
            status IN (
                'pending',
                'succeeded',
                'failed',
                'refunded'
            )
        ),

    transaction_id VARCHAR(150),
    gateway_reference VARCHAR(150),
    card_last_four VARCHAR(4),

    failure_reason VARCHAR(500),
    description VARCHAR(500),

    paid_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT payments_transaction_id_not_blank
        CHECK (
            transaction_id IS NULL
            OR LENGTH(TRIM(transaction_id)) > 0
        ),

    CONSTRAINT payments_gateway_reference_not_blank
        CHECK (
            gateway_reference IS NULL
            OR LENGTH(TRIM(gateway_reference)) > 0
        ),

    CONSTRAINT payments_card_last_four_format
        CHECK (
            card_last_four IS NULL
            OR card_last_four ~ '^[0-9]{4}$'
        ),

    CONSTRAINT payments_status_timestamps
        CHECK (
            (status <> 'succeeded' OR paid_at IS NOT NULL)
            AND
            (status <> 'failed' OR failed_at IS NOT NULL)
            AND
            (status <> 'refunded' OR refunded_at IS NOT NULL)
        )
);

CREATE INDEX IF NOT EXISTS idx_payments_enrollment_id
    ON payments (enrollment_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payments_status
    ON payments (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payments_created_at
    ON payments (created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_transaction_id_unique
    ON payments (transaction_id)
    WHERE transaction_id IS NOT NULL
      AND deleted_at IS NULL;

-- Each enrollment can have several failed attempts, but only one
-- pending or successful payment at a time.
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_one_active_per_enrollment
    ON payments (enrollment_id)
    WHERE status IN ('pending', 'succeeded')
      AND deleted_at IS NULL;

DROP TRIGGER IF EXISTS payments_set_updated_at ON payments;

CREATE TRIGGER payments_set_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO schema_migrations (version)
VALUES ('000004')
ON CONFLICT (version) DO NOTHING;

COMMIT;