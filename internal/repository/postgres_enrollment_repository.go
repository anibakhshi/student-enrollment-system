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

// PostgresEnrollmentRepository stores enrollments in PostgreSQL.
type PostgresEnrollmentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresEnrollmentRepository creates a PostgreSQL repository.
func NewPostgresEnrollmentRepository(
	pool *pgxpool.Pool,
) *PostgresEnrollmentRepository {
	return &PostgresEnrollmentRepository{
		pool: pool,
	}
}

// List returns all non-deleted enrollments.
func (r *PostgresEnrollmentRepository) List(
	ctx context.Context,
) ([]model.Enrollment, error) {
	const query = `
		SELECT
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at
		FROM enrollments
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := make([]model.Enrollment, 0)

	for rows.Next() {
		enrollment, err := scanEnrollment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}

		enrollments = append(enrollments, enrollment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enrollments: %w", err)
	}

	return enrollments, nil
}

// GetByID returns one non-deleted enrollment by ID.
func (r *PostgresEnrollmentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Enrollment, error) {
	const query = `
		SELECT
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at
		FROM enrollments
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	enrollment, err := scanEnrollment(
		r.pool.QueryRow(ctx, query, id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrEnrollmentNotFound
		}

		return model.Enrollment{}, fmt.Errorf(
			"query enrollment by ID: %w",
			err,
		)
	}

	return enrollment, nil
}

// ListByStudentID returns a student's non-deleted enrollments.
func (r *PostgresEnrollmentRepository) ListByStudentID(
	ctx context.Context,
	studentID int,
) ([]model.Enrollment, error) {
	const query = `
		SELECT
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at
		FROM enrollments
		WHERE student_id = $1
		  AND deleted_at IS NULL
		ORDER BY enrolled_at DESC, id DESC;
	`

	return r.listByArgument(ctx, query, studentID)
}

// ListByCourseID returns a course's non-deleted enrollments.
func (r *PostgresEnrollmentRepository) ListByCourseID(
	ctx context.Context,
	courseID int,
) ([]model.Enrollment, error) {
	const query = `
		SELECT
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at
		FROM enrollments
		WHERE course_id = $1
		  AND deleted_at IS NULL
		ORDER BY enrolled_at DESC, id DESC;
	`

	return r.listByArgument(ctx, query, courseID)
}

// listByArgument executes an enrollment list query with one argument.
func (r *PostgresEnrollmentRepository) listByArgument(
	ctx context.Context,
	query string,
	argument int,
) ([]model.Enrollment, error) {
	rows, err := r.pool.Query(ctx, query, argument)
	if err != nil {
		return nil, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := make([]model.Enrollment, 0)

	for rows.Next() {
		enrollment, err := scanEnrollment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}

		enrollments = append(enrollments, enrollment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enrollments: %w", err)
	}

	return enrollments, nil
}

// CountActiveByCourseID returns the number of pending or confirmed
// enrollments belonging to a course.
func (r *PostgresEnrollmentRepository) CountActiveByCourseID(
	ctx context.Context,
	courseID int,
) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1
		  AND status IN ('pending', 'confirmed')
		  AND deleted_at IS NULL;
	`

	var count int

	if err := r.pool.QueryRow(ctx, query, courseID).Scan(&count); err != nil {
		return 0, fmt.Errorf(
			"count active course enrollments: %w",
			err,
		)
	}

	return count, nil
}

// Create stores a new enrollment while protecting course capacity.
func (r *PostgresEnrollmentRepository) Create(
	ctx context.Context,
	enrollment model.Enrollment,
	courseCapacity int,
) (model.Enrollment, error) {
	transaction, err := r.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel: pgx.ReadCommitted,
		},
	)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf(
			"begin enrollment transaction: %w",
			err,
		)
	}

	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	// Locking the course row serializes enrollment creation for the
	// same course and prevents concurrent requests from exceeding capacity.
	const lockCourseQuery = `
		SELECT capacity
		FROM courses
		WHERE id = $1
		  AND deleted_at IS NULL
		FOR UPDATE;
	`

	var databaseCapacity int

	err = transaction.QueryRow(
		ctx,
		lockCourseQuery,
		enrollment.CourseID,
	).Scan(&databaseCapacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrCourseNotFound
		}

		return model.Enrollment{}, fmt.Errorf(
			"lock enrollment course: %w",
			err,
		)
	}

	// Prefer the current database capacity. The parameter remains part
	// of the repository interface for memory and PostgreSQL consistency.
	if databaseCapacity > 0 {
		courseCapacity = databaseCapacity
	}

	const duplicateQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM enrollments
			WHERE student_id = $1
			  AND course_id = $2
			  AND deleted_at IS NULL
		);
	`

	var alreadyEnrolled bool

	err = transaction.QueryRow(
		ctx,
		duplicateQuery,
		enrollment.StudentID,
		enrollment.CourseID,
	).Scan(&alreadyEnrolled)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf(
			"check duplicate enrollment: %w",
			err,
		)
	}

	if alreadyEnrolled {
		return model.Enrollment{}, ErrEnrollmentExists
	}

	const countQuery = `
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1
		  AND status IN ('pending', 'confirmed')
		  AND deleted_at IS NULL;
	`

	var activeEnrollmentCount int

	err = transaction.QueryRow(
		ctx,
		countQuery,
		enrollment.CourseID,
	).Scan(&activeEnrollmentCount)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf(
			"count active enrollments: %w",
			err,
		)
	}

	if activeEnrollmentCount >= courseCapacity {
		return model.Enrollment{}, ErrCourseFull
	}

	const insertQuery = `
		INSERT INTO enrollments (
			student_id,
			course_id,
			status,
			notes
		)
		VALUES (
			$1,
			$2,
			$3::VARCHAR(20),
			NULLIF($4::TEXT, '')
		)
		RETURNING
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at;
	`

	createdEnrollment, err := scanEnrollment(
		transaction.QueryRow(
			ctx,
			insertQuery,
			enrollment.StudentID,
			enrollment.CourseID,
			string(enrollment.Status),
			enrollment.Notes,
		),
	)
	if err != nil {
		return model.Enrollment{}, mapPostgresEnrollmentError(err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return model.Enrollment{}, fmt.Errorf(
			"commit enrollment transaction: %w",
			err,
		)
	}

	return createdEnrollment, nil
}

// Update changes an enrollment's status and notes.
func (r *PostgresEnrollmentRepository) Update(
	ctx context.Context,
	id int,
	enrollment model.Enrollment,
) (model.Enrollment, error) {
	const query = `
		UPDATE enrollments
		SET
			status = $2::VARCHAR(20),
			notes = NULLIF($3::TEXT, ''),

			confirmed_at = CASE
				WHEN $2::VARCHAR(20) = 'confirmed'
					AND confirmed_at IS NULL
				THEN COALESCE($4, CURRENT_TIMESTAMP)
				ELSE confirmed_at
			END,

			cancelled_at = CASE
				WHEN $2::VARCHAR(20) = 'cancelled'
					AND cancelled_at IS NULL
				THEN COALESCE($5, CURRENT_TIMESTAMP)
				ELSE cancelled_at
			END,

			completed_at = CASE
				WHEN $2::VARCHAR(20) = 'completed'
					AND completed_at IS NULL
				THEN COALESCE($6, CURRENT_TIMESTAMP)
				ELSE completed_at
			END
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			student_id,
			course_id,
			status,
			enrolled_at,
			confirmed_at,
			cancelled_at,
			completed_at,
			COALESCE(notes, ''),
			created_at,
			updated_at;
	`

	updatedEnrollment, err := scanEnrollment(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			string(enrollment.Status),
			enrollment.Notes,
			enrollment.ConfirmedAt,
			enrollment.CancelledAt,
			enrollment.CompletedAt,
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrEnrollmentNotFound
		}

		return model.Enrollment{}, mapPostgresEnrollmentError(err)
	}

	return updatedEnrollment, nil
}

// Delete soft-deletes an enrollment.
func (r *PostgresEnrollmentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	const query = `
		UPDATE enrollments
		SET
			status = 'cancelled',
			cancelled_at = COALESCE(
				cancelled_at,
				CURRENT_TIMESTAMP
			),
			deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-delete enrollment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrEnrollmentNotFound
	}

	return nil
}

// enrollmentScanner is implemented by pgx.Row and pgx.Rows.
type enrollmentScanner interface {
	Scan(destinations ...any) error
}

// scanEnrollment scans a PostgreSQL result into an enrollment model.
func scanEnrollment(
	row enrollmentScanner,
) (model.Enrollment, error) {
	var enrollment model.Enrollment

	err := row.Scan(
		&enrollment.ID,
		&enrollment.StudentID,
		&enrollment.CourseID,
		&enrollment.Status,
		&enrollment.EnrolledAt,
		&enrollment.ConfirmedAt,
		&enrollment.CancelledAt,
		&enrollment.CompletedAt,
		&enrollment.Notes,
		&enrollment.CreatedAt,
		&enrollment.UpdatedAt,
	)
	if err != nil {
		return model.Enrollment{}, err
	}

	return enrollment, nil
}

// mapPostgresEnrollmentError converts PostgreSQL errors to domain errors.
func mapPostgresEnrollmentError(err error) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch postgresError.ConstraintName {
		case "idx_enrollments_unique_active_student_course":
			return ErrEnrollmentExists
		}
	}

	return fmt.Errorf("enrollment repository operation: %w", err)
}
