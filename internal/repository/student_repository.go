package repository

import (
	"context"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// StudentRepository defines the data operations required by the student service.
type StudentRepository interface {
	List(ctx context.Context) ([]model.Student, error)
	GetByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
}
