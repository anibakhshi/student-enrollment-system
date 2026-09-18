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

// NewMemoryStudentRepository creates a repository with sample data.
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

// List returns a safe copy of all students.
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

// GetByID returns a student by ID.
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

	if err := r.checkUniqueFields(student, 0); err != nil {
		return model.Student{}, err
	}

	now := time.Now().UTC()

	student.ID = r.nextID
	student.CreatedAt = now
	student.UpdatedAt = now

	r.nextID++
	r.students = append(r.students, student)

	return student, nil
}

// Update replaces an existing student's editable information.
func (r *MemoryStudentRepository) Update(
	ctx context.Context,
	id int,
	student model.Student,
) (model.Student, error) {
	if err := ctx.Err(); err != nil {
		return model.Student{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	studentIndex := -1

	for index, existingStudent := range r.students {
		if existingStudent.ID == id {
			studentIndex = index
			break
		}
	}

	if studentIndex == -1 {
		return model.Student{}, ErrStudentNotFound
	}

	if err := r.checkUniqueFields(student, id); err != nil {
		return model.Student{}, err
	}

	existingStudent := r.students[studentIndex]

	student.ID = id
	student.ProfileImagePath = existingStudent.ProfileImagePath
	student.CreatedAt = existingStudent.CreatedAt
	student.UpdatedAt = time.Now().UTC()

	r.students[studentIndex] = student

	return student, nil
}

// UpdateProfileImage updates a student's profile image path.
func (r *MemoryStudentRepository) UpdateProfileImage(
	ctx context.Context,
	id int,
	profileImagePath string,
) (model.Student, error) {
	if err := ctx.Err(); err != nil {
		return model.Student{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for index, student := range r.students {
		if student.ID != id {
			continue
		}

		student.ProfileImagePath = profileImagePath
		student.UpdatedAt = time.Now().UTC()

		r.students[index] = student

		return student, nil
	}

	return model.Student{}, ErrStudentNotFound
}

// Delete removes a student from the memory repository.
func (r *MemoryStudentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for index, student := range r.students {
		if student.ID == id {
			r.students = append(
				r.students[:index],
				r.students[index+1:]...,
			)

			return nil
		}
	}

	return ErrStudentNotFound
}

// checkUniqueFields verifies national-code and email uniqueness.
// The ignoredID is used when updating an existing student.
func (r *MemoryStudentRepository) checkUniqueFields(
	student model.Student,
	ignoredID int,
) error {
	for _, existingStudent := range r.students {
		if existingStudent.ID == ignoredID {
			continue
		}

		if existingStudent.NationalCode == student.NationalCode {
			return ErrNationalCodeExists
		}

		if strings.EqualFold(existingStudent.Email, student.Email) {
			return ErrEmailExists
		}
	}

	return nil
}
