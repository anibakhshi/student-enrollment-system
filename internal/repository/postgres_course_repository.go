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

// PostgresCourseRepository stores courses in PostgreSQL.
type PostgresCourseRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCourseRepository creates a PostgreSQL course repository.
func NewPostgresCourseRepository(
	pool *pgxpool.Pool,
) *PostgresCourseRepository {
	return &PostgresCourseRepository{
		pool: pool,
	}
}

// List returns all non-deleted courses.
func (r *PostgresCourseRepository) List(
	ctx context.Context,
) ([]model.Course, error) {
	const query = `
		SELECT
			id,
			instructor_id,
			code,
			title,
			COALESCE(description, ''),
			price,
			capacity,
			duration_hours,
			start_date,
			end_date,
			status,
			created_at,
			updated_at
		FROM courses
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"query courses: %w",
			err,
		)
	}
	defer rows.Close()

	courses := make([]model.Course, 0)

	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan course: %w",
				err,
			)
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate courses: %w",
			err,
		)
	}

	return courses, nil
}

// GetByID returns one non-deleted course by ID.
func (r *PostgresCourseRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Course, error) {
	const query = `
		SELECT
			id,
			instructor_id,
			code,
			title,
			COALESCE(description, ''),
			price,
			capacity,
			duration_hours,
			start_date,
			end_date,
			status,
			created_at,
			updated_at
		FROM courses
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	course, err := scanCourse(
		r.pool.QueryRow(ctx, query, id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrCourseNotFound
		}

		return model.Course{}, fmt.Errorf(
			"query course by ID: %w",
			err,
		)
	}

	return course, nil
}

// Create stores a new course.
func (r *PostgresCourseRepository) Create(
	ctx context.Context,
	course model.Course,
) (model.Course, error) {
	const query = `
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
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			$5,
			$6,
			$7,
			$8,
			$9,
			'draft'
		)
		RETURNING
			id,
			instructor_id,
			code,
			title,
			COALESCE(description, ''),
			price,
			capacity,
			duration_hours,
			start_date,
			end_date,
			status,
			created_at,
			updated_at;
	`

	createdCourse, err := scanCourse(
		r.pool.QueryRow(
			ctx,
			query,
			course.InstructorID,
			course.Code,
			course.Title,
			course.Description,
			course.Price,
			course.Capacity,
			course.DurationHours,
			course.StartDate,
			course.EndDate,
		),
	)
	if err != nil {
		return model.Course{},
			mapPostgresCourseError(err)
	}

	return createdCourse, nil
}

// Update replaces editable course information.
func (r *PostgresCourseRepository) Update(
	ctx context.Context,
	id int,
	course model.Course,
) (model.Course, error) {
	const query = `
		UPDATE courses
		SET
			instructor_id = $2,
			code = $3,
			title = $4,
			description = NULLIF($5, ''),
			price = $6,
			capacity = $7,
			duration_hours = $8,
			start_date = $9,
			end_date = $10,
			status = $11
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			instructor_id,
			code,
			title,
			COALESCE(description, ''),
			price,
			capacity,
			duration_hours,
			start_date,
			end_date,
			status,
			created_at,
			updated_at;
	`

	updatedCourse, err := scanCourse(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			course.InstructorID,
			course.Code,
			course.Title,
			course.Description,
			course.Price,
			course.Capacity,
			course.DurationHours,
			course.StartDate,
			course.EndDate,
			string(course.Status),
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrCourseNotFound
		}

		return model.Course{},
			mapPostgresCourseError(err)
	}

	return updatedCourse, nil
}

// Delete soft-deletes a course.
func (r *PostgresCourseRepository) Delete(
	ctx context.Context,
	id int,
) error {
	const query = `
		UPDATE courses
		SET
			status = 'cancelled',
			deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"soft-delete course: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCourseNotFound
	}

	return nil
}

// scanCourse scans a PostgreSQL row into a course model.
func scanCourse(
	scanner rowScanner,
) (model.Course, error) {
	var course model.Course
	var status string

	err := scanner.Scan(
		&course.ID,
		&course.InstructorID,
		&course.Code,
		&course.Title,
		&course.Description,
		&course.Price,
		&course.Capacity,
		&course.DurationHours,
		&course.StartDate,
		&course.EndDate,
		&status,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	if err != nil {
		return model.Course{}, err
	}

	course.Status = model.CourseStatus(status)

	return course, nil
}

// mapPostgresCourseError converts PostgreSQL errors to domain errors.
func mapPostgresCourseError(err error) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch postgresError.ConstraintName {
		case "courses_code_key":
			return ErrCourseCodeExists

		case "courses_instructor_id_fkey":
			return ErrInstructorNotFound
		}
	}

	return fmt.Errorf(
		"course repository operation: %w",
		err,
	)
}
