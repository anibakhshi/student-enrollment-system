package repository

import (
	"context"
	"errors"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

var (
	ErrCourseNotFound   = errors.New("course not found")
	ErrCourseCodeExists = errors.New("course code already exists")
)

// CourseRepository defines course data operations.
type CourseRepository interface {
	List(
		ctx context.Context,
	) ([]model.Course, error)

	GetByID(
		ctx context.Context,
		id int,
	) (model.Course, error)

	Create(
		ctx context.Context,
		course model.Course,
	) (model.Course, error)

	Update(
		ctx context.Context,
		id int,
		course model.Course,
	) (model.Course, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}
