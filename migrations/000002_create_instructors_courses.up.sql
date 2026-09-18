BEGIN;

CREATE TABLE IF NOT EXISTS instructors (
    id BIGSERIAL PRIMARY KEY,

    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,

    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(11),

    bio TEXT,
    expertise VARCHAR(150),

    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'suspended')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT instructors_first_name_not_blank
        CHECK (LENGTH(TRIM(first_name)) >= 2),

    CONSTRAINT instructors_last_name_not_blank
        CHECK (LENGTH(TRIM(last_name)) >= 2),

    CONSTRAINT instructors_phone_format
        CHECK (
            phone IS NULL
            OR phone ~ '^[0-9]{11}$'
        )
);

CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,

    instructor_id BIGINT NOT NULL
        REFERENCES instructors(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,

    code VARCHAR(30) NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    description TEXT,

    price NUMERIC(14, 2) NOT NULL
        CHECK (price >= 0),

    capacity SMALLINT NOT NULL
        CHECK (capacity BETWEEN 1 AND 1000),

    duration_hours SMALLINT NOT NULL
        CHECK (duration_hours BETWEEN 1 AND 5000),

    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (
            status IN (
                'draft',
                'open',
                'closed',
                'completed',
                'cancelled'
            )
        ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT courses_code_not_blank
        CHECK (LENGTH(TRIM(code)) >= 2),

    CONSTRAINT courses_title_not_blank
        CHECK (LENGTH(TRIM(title)) >= 2),

    CONSTRAINT courses_valid_date_range
        CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_instructors_last_name
    ON instructors (last_name);

CREATE INDEX IF NOT EXISTS idx_instructors_status
    ON instructors (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_courses_instructor_id
    ON courses (instructor_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_courses_status
    ON courses (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_courses_start_date
    ON courses (start_date)
    WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS instructors_set_updated_at
    ON instructors;

CREATE TRIGGER instructors_set_updated_at
BEFORE UPDATE ON instructors
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS courses_set_updated_at
    ON courses;

CREATE TRIGGER courses_set_updated_at
BEFORE UPDATE ON courses
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO instructors (
    first_name,
    last_name,
    email,
    phone,
    bio,
    expertise
)
VALUES
    (
        'Parham',
        'Darvishi',
        'parham.darvishi@example.com',
        '09121112222',
        'Software and data science instructor',
        'Data Science and Software Engineering'
    ),
    (
        'Maryam',
        'Rahimi',
        'maryam.rahimi@example.com',
        '09123334444',
        'Backend development instructor',
        'Go and PostgreSQL'
    )
ON CONFLICT (email) DO NOTHING;

INSERT INTO courses (
    instructor_id,
    code,
    title,
    description,
    price,
    capacity,
    duration_hours,
    start_date,
    end_date,
    status
)
SELECT
    instructor.id,
    course_data.code,
    course_data.title,
    course_data.description,
    course_data.price,
    course_data.capacity,
    course_data.duration_hours,
    course_data.start_date,
    course_data.end_date,
    course_data.status
FROM (
    VALUES
        (
            'parham.darvishi@example.com',
            'DS-101',
            'Data Science Fundamentals',
            'Introduction to practical data science concepts',
            15000000.00,
            25::SMALLINT,
            60::SMALLINT,
            DATE '2026-10-01',
            DATE '2026-12-15',
            'open'
        ),
        (
            'maryam.rahimi@example.com',
            'GO-201',
            'REST API Development with Go',
            'Building production-ready REST APIs with Go and PostgreSQL',
            18000000.00,
            20::SMALLINT,
            72::SMALLINT,
            DATE '2026-10-10',
            DATE '2027-01-10',
            'open'
        )
) AS course_data (
    instructor_email,
    code,
    title,
    description,
    price,
    capacity,
    duration_hours,
    start_date,
    end_date,
    status
)
JOIN instructors AS instructor
    ON instructor.email = course_data.instructor_email
ON CONFLICT (code) DO NOTHING;

INSERT INTO schema_migrations (version)
VALUES ('000002')
ON CONFLICT (version) DO NOTHING;

COMMIT;