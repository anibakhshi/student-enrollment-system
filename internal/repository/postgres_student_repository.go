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

// PostgresStudentRepository stores students in PostgreSQL.
type PostgresStudentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresStudentRepository creates a PostgreSQL repository.
func NewPostgresStudentRepository(
	pool *pgxpool.Pool,
) *PostgresStudentRepository {
	return &PostgresStudentRepository{
		pool: pool,
	}
}

// List returns all non-deleted students.
func (r *PostgresStudentRepository) List(
	ctx context.Context,
) ([]model.Student, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			age,
			national_code,
			email,
			COALESCE(phone, ''),
			COALESCE(profile_image_path, ''),
			created_at,
			updated_at
		FROM students
		WHERE deleted_at IS NULL
		ORDER BY id;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query students: %w", err)
	}
	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Age,
			&student.NationalCode,
			&student.Email,
			&student.Phone,
			&student.ProfileImagePath,
			&student.CreatedAt,
			&student.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan student: %w", err)
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate students: %w", err)
	}

	return students, nil
}

// GetByID returns one non-deleted student by ID.
func (r *PostgresStudentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Student, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			age,
			national_code,
			email,
			COALESCE(phone, ''),
			COALESCE(profile_image_path, ''),
			created_at,
			updated_at
		FROM students
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	var student model.Student

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Age,
		&student.NationalCode,
		&student.Email,
		&student.Phone,
		&student.ProfileImagePath,
		&student.CreatedAt,
		&student.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrStudentNotFound
		}

		return model.Student{}, fmt.Errorf(
			"query student by ID: %w",
			err,
		)
	}

	return student, nil
}

// Create stores a new student.
func (r *PostgresStudentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	const query = `
		INSERT INTO students (
			first_name,
			last_name,
			age,
			national_code,
			email,
			phone
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			NULLIF($6, '')
		)
		RETURNING
			id,
			first_name,
			last_name,
			age,
			national_code,
			email,
			COALESCE(phone, ''),
			COALESCE(profile_image_path, ''),
			created_at,
			updated_at;
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		student.FirstName,
		student.LastName,
		student.Age,
		student.NationalCode,
		student.Email,
		student.Phone,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Age,
		&student.NationalCode,
		&student.Email,
		&student.Phone,
		&student.ProfileImagePath,
		&student.CreatedAt,
		&student.UpdatedAt,
	)

	if err != nil {
		return model.Student{}, mapPostgresStudentError(err)
	}

	return student, nil
}

// Update replaces a student's editable information.
func (r *PostgresStudentRepository) Update(
	ctx context.Context,
	id int,
	student model.Student,
) (model.Student, error) {
	const query = `
		UPDATE students
		SET
			first_name = $2,
			last_name = $3,
			age = $4,
			national_code = $5,
			email = $6,
			phone = NULLIF($7, '')
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			first_name,
			last_name,
			age,
			national_code,
			email,
			COALESCE(phone, ''),
			COALESCE(profile_image_path, ''),
			created_at,
			updated_at;
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		student.FirstName,
		student.LastName,
		student.Age,
		student.NationalCode,
		student.Email,
		student.Phone,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Age,
		&student.NationalCode,
		&student.Email,
		&student.Phone,
		&student.ProfileImagePath,
		&student.CreatedAt,
		&student.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrStudentNotFound
		}

		return model.Student{}, mapPostgresStudentError(err)
	}

	return student, nil
}

// UpdateProfileImage updates a student's profile image path.
func (r *PostgresStudentRepository) UpdateProfileImage(
	ctx context.Context,
	id int,
	profileImagePath string,
) (model.Student, error) {
	const query = `
		UPDATE students
		SET profile_image_path = NULLIF($2, '')
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			first_name,
			last_name,
			age,
			national_code,
			email,
			COALESCE(phone, ''),
			COALESCE(profile_image_path, ''),
			created_at,
			updated_at;
	`

	var student model.Student

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		profileImagePath,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Age,
		&student.NationalCode,
		&student.Email,
		&student.Phone,
		&student.ProfileImagePath,
		&student.CreatedAt,
		&student.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrStudentNotFound
		}

		return model.Student{}, fmt.Errorf(
			"update student profile image: %w",
			err,
		)
	}

	return student, nil
}

// Delete soft-deletes a student.
func (r *PostgresStudentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	const query = `
		UPDATE students
		SET
			deleted_at = CURRENT_TIMESTAMP,
			status = 'inactive'
		WHERE id = $1
		  AND deleted_at IS NULL;
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft-delete student: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrStudentNotFound
	}

	return nil
}

// mapPostgresStudentError converts PostgreSQL errors to domain errors.
func mapPostgresStudentError(err error) error {
	var postgresError *pgconn.PgError

	if errors.As(err, &postgresError) {
		switch postgresError.ConstraintName {
		case "students_national_code_key":
			return ErrNationalCodeExists

		case "students_email_key":
			return ErrEmailExists
		}
	}

	return fmt.Errorf("student repository operation: %w", err)
}
