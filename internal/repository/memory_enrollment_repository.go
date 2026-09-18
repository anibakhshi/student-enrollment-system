package repository

import (
	"context"
	"sync"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// MemoryEnrollmentRepository stores enrollments in memory.
type MemoryEnrollmentRepository struct {
	mu          sync.RWMutex
	enrollments []model.Enrollment
	nextID      int
}

// NewMemoryEnrollmentRepository creates an empty enrollment repository.
func NewMemoryEnrollmentRepository() *MemoryEnrollmentRepository {
	return &MemoryEnrollmentRepository{
		enrollments: make([]model.Enrollment, 0),
		nextID:      1,
	}
}

// List returns all enrollments.
func (r *MemoryEnrollmentRepository) List(
	_ context.Context,
) ([]model.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	enrollments := make(
		[]model.Enrollment,
		len(r.enrollments),
	)

	copy(enrollments, r.enrollments)

	return enrollments, nil
}

// GetByID returns an enrollment by ID.
func (r *MemoryEnrollmentRepository) GetByID(
	_ context.Context,
	id int,
) (model.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, enrollment := range r.enrollments {
		if enrollment.ID == id {
			return enrollment, nil
		}
	}

	return model.Enrollment{}, ErrEnrollmentNotFound
}

// ListByStudentID returns a student's enrollments.
func (r *MemoryEnrollmentRepository) ListByStudentID(
	_ context.Context,
	studentID int,
) ([]model.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	enrollments := make([]model.Enrollment, 0)

	for _, enrollment := range r.enrollments {
		if enrollment.StudentID == studentID {
			enrollments = append(
				enrollments,
				enrollment,
			)
		}
	}

	return enrollments, nil
}

// ListByCourseID returns a course's enrollments.
func (r *MemoryEnrollmentRepository) ListByCourseID(
	_ context.Context,
	courseID int,
) ([]model.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	enrollments := make([]model.Enrollment, 0)

	for _, enrollment := range r.enrollments {
		if enrollment.CourseID == courseID {
			enrollments = append(
				enrollments,
				enrollment,
			)
		}
	}

	return enrollments, nil
}

// CountActiveByCourseID returns the number of active enrollments.
func (r *MemoryEnrollmentRepository) CountActiveByCourseID(
	_ context.Context,
	courseID int,
) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.countActiveByCourseID(courseID), nil
}

// Create stores a new enrollment.
func (r *MemoryEnrollmentRepository) Create(
	_ context.Context,
	enrollment model.Enrollment,
	courseCapacity int,
) (model.Enrollment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existingEnrollment := range r.enrollments {
		if existingEnrollment.StudentID == enrollment.StudentID &&
			existingEnrollment.CourseID == enrollment.CourseID &&
			existingEnrollment.Status !=
				model.EnrollmentStatusCancelled {
			return model.Enrollment{}, ErrEnrollmentExists
		}
	}

	if r.countActiveByCourseID(enrollment.CourseID) >=
		courseCapacity {
		return model.Enrollment{}, ErrCourseFull
	}

	now := time.Now().UTC()

	enrollment.ID = r.nextID
	enrollment.Status = model.EnrollmentStatusPending
	enrollment.EnrolledAt = now
	enrollment.CreatedAt = now
	enrollment.UpdatedAt = now

	r.nextID++

	r.enrollments = append(
		r.enrollments,
		enrollment,
	)

	return enrollment, nil
}

// Update changes an existing enrollment.
func (r *MemoryEnrollmentRepository) Update(
	_ context.Context,
	id int,
	enrollment model.Enrollment,
) (model.Enrollment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, existingEnrollment := range r.enrollments {
		if existingEnrollment.ID != id {
			continue
		}

		enrollment.ID = existingEnrollment.ID
		enrollment.StudentID = existingEnrollment.StudentID
		enrollment.CourseID = existingEnrollment.CourseID
		enrollment.EnrolledAt = existingEnrollment.EnrolledAt
		enrollment.CreatedAt = existingEnrollment.CreatedAt
		enrollment.UpdatedAt = time.Now().UTC()

		r.enrollments[index] = enrollment

		return enrollment, nil
	}

	return model.Enrollment{}, ErrEnrollmentNotFound
}

// Delete removes an enrollment from the active collection.
func (r *MemoryEnrollmentRepository) Delete(
	_ context.Context,
	id int,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, enrollment := range r.enrollments {
		if enrollment.ID != id {
			continue
		}

		r.enrollments = append(
			r.enrollments[:index],
			r.enrollments[index+1:]...,
		)

		return nil
	}

	return ErrEnrollmentNotFound
}

// countActiveByCourseID counts pending and confirmed enrollments.
// The caller must hold at least a read lock.
func (r *MemoryEnrollmentRepository) countActiveByCourseID(
	courseID int,
) int {
	count := 0

	for _, enrollment := range r.enrollments {
		if enrollment.CourseID != courseID {
			continue
		}

		if enrollment.Status == model.EnrollmentStatusPending ||
			enrollment.Status ==
				model.EnrollmentStatusConfirmed {
			count++
		}
	}

	return count
}
