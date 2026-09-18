package repository

import (
	"context"
	"errors"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

var (
	ErrInstructorNotFound    = errors.New("instructor not found")
	ErrInstructorEmailExists = errors.New("instructor email already exists")
)

// InstructorRepository defines instructor data operations.
type InstructorRepository interface {
	List(
		ctx context.Context,
	) ([]model.Instructor, error)

	GetByID(
		ctx context.Context,
		id int,
	) (model.Instructor, error)

	Create(
		ctx context.Context,
		instructor model.Instructor,
	) (model.Instructor, error)

	Update(
		ctx context.Context,
		id int,
		instructor model.Instructor,
	) (model.Instructor, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}
