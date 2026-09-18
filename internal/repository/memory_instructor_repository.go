package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// MemoryInstructorRepository stores instructors safely in memory.
type MemoryInstructorRepository struct {
	mu          sync.RWMutex
	instructors []model.Instructor
	nextID      int
}

// NewMemoryInstructorRepository creates a repository with sample data.
func NewMemoryInstructorRepository() *MemoryInstructorRepository {
	now := time.Now().UTC()

	return &MemoryInstructorRepository{
		instructors: []model.Instructor{
			{
				ID:        1,
				FirstName: "Parham",
				LastName:  "Darvishi",
				Email:     "parham.darvishi@example.com",
				Phone:     "09121112222",
				Bio:       "Software and data science instructor",
				Expertise: "Data Science and Software Engineering",
				Status:    model.InstructorStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        2,
				FirstName: "Maryam",
				LastName:  "Rahimi",
				Email:     "maryam.rahimi@example.com",
				Phone:     "09123334444",
				Bio:       "Backend development instructor",
				Expertise: "Go and PostgreSQL",
				Status:    model.InstructorStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		nextID: 3,
	}
}

// List returns a safe copy of all instructors.
func (r *MemoryInstructorRepository) List(
	ctx context.Context,
) ([]model.Instructor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	instructors := make(
		[]model.Instructor,
		len(r.instructors),
	)

	copy(instructors, r.instructors)

	return instructors, nil
}

// GetByID returns an instructor by ID.
func (r *MemoryInstructorRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Instructor, error) {
	if err := ctx.Err(); err != nil {
		return model.Instructor{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, instructor := range r.instructors {
		if instructor.ID == id {
			return instructor, nil
		}
	}

	return model.Instructor{}, ErrInstructorNotFound
}

// Create stores a new instructor.
func (r *MemoryInstructorRepository) Create(
	ctx context.Context,
	instructor model.Instructor,
) (model.Instructor, error) {
	if err := ctx.Err(); err != nil {
		return model.Instructor{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.emailExists(instructor.Email, 0) {
		return model.Instructor{}, ErrInstructorEmailExists
	}

	now := time.Now().UTC()

	instructor.ID = r.nextID
	instructor.Status = model.InstructorStatusActive
	instructor.CreatedAt = now
	instructor.UpdatedAt = now

	r.nextID++
	r.instructors = append(r.instructors, instructor)

	return instructor, nil
}

// Update replaces editable instructor information.
func (r *MemoryInstructorRepository) Update(
	ctx context.Context,
	id int,
	instructor model.Instructor,
) (model.Instructor, error) {
	if err := ctx.Err(); err != nil {
		return model.Instructor{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	instructorIndex := -1

	for index, existingInstructor := range r.instructors {
		if existingInstructor.ID == id {
			instructorIndex = index
			break
		}
	}

	if instructorIndex == -1 {
		return model.Instructor{}, ErrInstructorNotFound
	}

	if r.emailExists(instructor.Email, id) {
		return model.Instructor{}, ErrInstructorEmailExists
	}

	existingInstructor := r.instructors[instructorIndex]

	instructor.ID = id
	instructor.CreatedAt = existingInstructor.CreatedAt
	instructor.UpdatedAt = time.Now().UTC()

	r.instructors[instructorIndex] = instructor

	return instructor, nil
}

// Delete removes an instructor from the memory repository.
func (r *MemoryInstructorRepository) Delete(
	ctx context.Context,
	id int,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for index, instructor := range r.instructors {
		if instructor.ID != id {
			continue
		}

		r.instructors = append(
			r.instructors[:index],
			r.instructors[index+1:]...,
		)

		return nil
	}

	return ErrInstructorNotFound
}

// emailExists checks instructor email uniqueness.
func (r *MemoryInstructorRepository) emailExists(
	email string,
	ignoredID int,
) bool {
	for _, instructor := range r.instructors {
		if instructor.ID == ignoredID {
			continue
		}

		if strings.EqualFold(instructor.Email, email) {
			return true
		}
	}

	return false
}
