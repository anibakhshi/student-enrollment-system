BEGIN;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(20) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,

    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,

    age SMALLINT NOT NULL
        CHECK (age BETWEEN 16 AND 100),

    national_code VARCHAR(10) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(11),

    profile_image_path VARCHAR(500),

    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'suspended')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT students_first_name_not_blank
        CHECK (LENGTH(TRIM(first_name)) >= 2),

    CONSTRAINT students_last_name_not_blank
        CHECK (LENGTH(TRIM(last_name)) >= 2),

    CONSTRAINT students_national_code_format
        CHECK (national_code ~ '^[0-9]{10}$'),

    CONSTRAINT students_phone_format
        CHECK (
            phone IS NULL
            OR phone ~ '^[0-9]{11}$'
        )
);

CREATE INDEX IF NOT EXISTS idx_students_last_name
    ON students (last_name);

CREATE INDEX IF NOT EXISTS idx_students_status
    ON students (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_students_created_at
    ON students (created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS students_set_updated_at ON students;

CREATE TRIGGER students_set_updated_at
BEFORE UPDATE ON students
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO students (
    first_name,
    last_name,
    age,
    national_code,
    email,
    phone
)
VALUES
    (
        'Anita',
        'Bakhshi',
        20,
        '0012345678',
        'anita@example.com',
        '09121234567'
    ),
    (
        'Ali',
        'Ahmadi',
        22,
        '0023456789',
        'ali@example.com',
        '09129876543'
    )
ON CONFLICT DO NOTHING;

INSERT INTO schema_migrations (version)
VALUES ('000001')
ON CONFLICT (version) DO NOTHING;

COMMIT;