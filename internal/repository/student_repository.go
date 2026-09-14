package repository

import (
	"context"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// StudentRepository defines required student data operations.
type StudentRepository interface {
	List(ctx context.Context) ([]model.Student, error)

	GetByID(
		ctx context.Context,
		id int,
	) (model.Student, error)

	Create(
		ctx context.Context,
		student model.Student,
	) (model.Student, error)

	Update(
		ctx context.Context,
		id int,
		student model.Student,
	) (model.Student, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}
