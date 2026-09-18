package repository

import (
	"context"
	"errors"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

var (
	// ErrEnrollmentNotFound indicates that an enrollment does not exist.
	ErrEnrollmentNotFound = errors.New("enrollment not found")

	// ErrEnrollmentExists indicates that a student is already enrolled.
	ErrEnrollmentExists = errors.New(
		"student is already enrolled in this course",
	)

	// ErrCourseFull indicates that the course capacity has been reached.
	ErrCourseFull = errors.New("course capacity has been reached")
)

// EnrollmentRepository defines enrollment data operations.
type EnrollmentRepository interface {
	List(
		ctx context.Context,
	) ([]model.Enrollment, error)

	GetByID(
		ctx context.Context,
		id int,
	) (model.Enrollment, error)

	ListByStudentID(
		ctx context.Context,
		studentID int,
	) ([]model.Enrollment, error)

	ListByCourseID(
		ctx context.Context,
		courseID int,
	) ([]model.Enrollment, error)

	CountActiveByCourseID(
		ctx context.Context,
		courseID int,
	) (int, error)

	Create(
		ctx context.Context,
		enrollment model.Enrollment,
		courseCapacity int,
	) (model.Enrollment, error)

	Update(
		ctx context.Context,
		id int,
		enrollment model.Enrollment,
	) (model.Enrollment, error)

	Delete(
		ctx context.Context,
		id int,
	) error
}
