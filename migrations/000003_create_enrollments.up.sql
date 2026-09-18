BEGIN;

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,

    student_id BIGINT NOT NULL
        REFERENCES students(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    course_id BIGINT NOT NULL
        REFERENCES courses(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (
            status IN (
                'pending',
                'confirmed',
                'cancelled',
                'completed'
            )
        ),

    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT enrollments_confirmed_at_required
        CHECK (
            status <> 'confirmed'
            OR confirmed_at IS NOT NULL
        ),

    CONSTRAINT enrollments_cancelled_at_required
        CHECK (
            status <> 'cancelled'
            OR cancelled_at IS NOT NULL
        ),

    CONSTRAINT enrollments_completed_at_required
        CHECK (
            status <> 'completed'
            OR completed_at IS NOT NULL
        )
);

-- A student cannot have more than one active registration
-- for the same course.
CREATE UNIQUE INDEX IF NOT EXISTS
    idx_enrollments_unique_active_student_course
ON enrollments (
    student_id,
    course_id
)
WHERE deleted_at IS NULL
  AND status <> 'cancelled';

CREATE INDEX IF NOT EXISTS idx_enrollments_student_id
    ON enrollments (student_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_enrollments_course_id
    ON enrollments (course_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_enrollments_status
    ON enrollments (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_enrollments_enrolled_at
    ON enrollments (enrolled_at DESC)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS enrollments_set_updated_at
    ON enrollments;

CREATE TRIGGER enrollments_set_updated_at
BEFORE UPDATE ON enrollments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO schema_migrations (version)
VALUES ('000003')
ON CONFLICT (version) DO NOTHING;

COMMIT;