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

// PostgresInstructorRepository stores instructors in PostgreSQL.
type PostgresInstructorRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresInstructorRepository creates a PostgreSQL instructor repository.
func NewPostgresInstructorRepository(
	pool *pgxpool.Pool,
) *PostgresInstructorRepository {
	return &PostgresInstructorRepository{
		pool: pool,
	}
}

// List returns all non-deleted instructors.
func (r *PostgresInstructorRepository) List(
	ctx context.Context,
) ([]model.Instructor, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			email,
			COALESCE(phone, ''),
			COALESCE(bio, ''),
			COALESCE(expertise, ''),
			status,
			created_at,
			updated_at
		FROM instructors
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"query instructors: %w",
			err,
		)
	}
	defer rows.Close()

	instructors := make([]model.Instructor, 0)

	for rows.Next() {
		instructor, err := scanInstructor(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan instructor: %w",
				err,
			)
		}

		instructors = append(instructors, instructor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate instructors: %w",
			err,
		)
	}

	return instructors, nil
}

// GetByID returns one non-deleted instructor by ID.
func (r *PostgresInstructorRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Instructor, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			email,
			COALESCE(phone, ''),
			COALESCE(bio, ''),
			COALESCE(expertise, ''),
			status,
			created_at,
			updated_at
		FROM instructors
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	instructor, err := scanInstructor(
		r.pool.QueryRow(ctx, query, id),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Instructor{}, ErrInstructorNotFound
		}

		return model.Instructor{}, fmt.Errorf(
			"query instructor by ID: %w",
			err,
		)
	}

	return instructor, nil
}

// Create stores a new instructor.
func (r *PostgresInstructorRepository) Create(
	ctx context.Context,
	instructor model.Instructor,
) (model.Instructor, error) {
	const query = `
		INSERT INTO instructors (
			first_name,
			last_name,
			email,
			phone,
			bio,
			expertise,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			NULLIF($5, ''),
			NULLIF($6, ''),
			'active'
		)
		RETURNING
			id,
			first_name,
			last_name,
			email,
			COALESCE(phone, ''),
			COALESCE(bio, ''),
			COALESCE(expertise, ''),
			status,
			created_at,
			updated_at;
	`

	createdInstructor, err := scanInstructor(
		r.pool.QueryRow(
			ctx,
			query,
			instructor.FirstName,
			instructor.LastName,
			instructor.Email,
			instructor.Phone,
			instructor.Bio,
			instructor.Expertise,
		),
	)
	if err != nil {
		return model.Instructor{},
			mapPostgresInstructorError(err)
	}

	return createdInstructor, nil
}

// Update replaces editable instructor information.
func (r *PostgresInstructorRepository) Update(
	ctx context.Context,
	id int,
	instructor model.Instructor,
) (model.Instructor, error) {
	const query = `
		UPDATE instructors
		SET
			first_name = $2,
			last_name = $3,
			email = $4,
			phone = NULLIF($5, ''),
			bio = NULLIF($6, ''),
			expertise = NULLIF($7, ''),
			status = $8
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			first_name,
			last_name,
			email,
			COALESCE(phone, ''),
			COALESCE(bio, ''),
			COALESCE(expertise, ''),
			status,
			created_at,
			updated_at;
	`

	updatedInstructor, err := scanInstructor(
		r.pool.QueryRow(
			ctx,
			query,
			id,
			instructor.FirstName,
			instructor.LastName,
			instructor.Email,
			instructor.Phone,
			instructor.Bio,
			instructor.Expertise,
			string(instructor.Status),
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Instructor{}, ErrInstructorNotFound
		}

		return model.Instructor{},
			mapPostgresInstructorError(err)
	}

	return updatedInstructor, nil
}

// Delete soft-deletes an instructor.
func (r *PostgresInstructorRepository) Delete(
	ctx context.Context,
	id int,
) error {
	const query = `
		UPDATE instructors
		SET
			status = 'inactive',
			deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"soft-delete instructor: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrInstructorNotFound
	}

	return nil
}

// rowScanner represents PostgreSQL rows that can scan values.
type rowScanner interface {
	Scan(destinations ...any) error
}

// scanInstructor scans a PostgreSQL row into an instructor model.
func scanInstructor(
	scanner rowScanner,
) (model.Instructor, error) {
	var instructor model.Instructor
	var status string

	err := scanner.Scan(
		&instructor.ID,
		&instructor.FirstName,
		&instructor.LastName,
		&instructor.Email,
		&instructor.Phone,
		&instructor.Bio,
		&instructor.Expertise,
		&status,
		&instructor.CreatedAt,
		&instructor.UpdatedAt,
	)
	if err != nil {
		return model.Instructor{}, err
	}

	instructor.Status = model.InstructorStatus(status)

	return instructor, nil
}

// mapPostgresInstructorError converts PostgreSQL errors to domain errors.
func mapPostgresInstructorError(err error) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch postgresError.ConstraintName {
		case "instructors_email_key":
			return ErrInstructorEmailExists
		}
	}

	return fmt.Errorf(
		"instructor repository operation: %w",
		err,
	)
}
