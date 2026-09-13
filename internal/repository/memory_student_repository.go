package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

var (
	ErrStudentNotFound    = errors.New("student not found")
	ErrNationalCodeExists = errors.New("national code already exists")
	ErrEmailExists        = errors.New("email already exists")
)

// MemoryStudentRepository stores students safely in memory.
type MemoryStudentRepository struct {
	mu       sync.RWMutex
	students []model.Student
	nextID   int
}

// NewMemoryStudentRepository creates a repository with initial sample data.
func NewMemoryStudentRepository() *MemoryStudentRepository {
	now := time.Now().UTC()

	return &MemoryStudentRepository{
		students: []model.Student{
			{
				ID:           1,
				FirstName:    "Anita",
				LastName:     "Bakhshi",
				Age:          20,
				NationalCode: "0012345678",
				Email:        "anita@example.com",
				Phone:        "09121234567",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			{
				ID:           2,
				FirstName:    "Ali",
				LastName:     "Ahmadi",
				Age:          22,
				NationalCode: "0023456789",
				Email:        "ali@example.com",
				Phone:        "09129876543",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		nextID: 3,
	}
}

// List returns a copy of all students.
func (r *MemoryStudentRepository) List(
	ctx context.Context,
) ([]model.Student, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	students := make([]model.Student, len(r.students))
	copy(students, r.students)

	return students, nil
}

// GetByID finds a student by ID.
func (r *MemoryStudentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Student, error) {
	if err := ctx.Err(); err != nil {
		return model.Student{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, student := range r.students {
		if student.ID == id {
			return student, nil
		}
	}

	return model.Student{}, ErrStudentNotFound
}

// Create stores a new student.
func (r *MemoryStudentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	if err := ctx.Err(); err != nil {
		return model.Student{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existingStudent := range r.students {
		if existingStudent.NationalCode == student.NationalCode {
			return model.Student{}, ErrNationalCodeExists
		}

		if strings.EqualFold(existingStudent.Email, student.Email) {
			return model.Student{}, ErrEmailExists
		}
	}

	now := time.Now().UTC()

	student.ID = r.nextID
	student.CreatedAt = now
	student.UpdatedAt = now

	r.nextID++
	r.students = append(r.students, student)

	return student, nil
}
