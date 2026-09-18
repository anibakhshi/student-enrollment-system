package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

// MemoryCourseRepository stores courses in memory.
type MemoryCourseRepository struct {
	mutex   sync.RWMutex
	courses map[int]model.Course
	nextID  int
}

// NewMemoryCourseRepository creates a repository with sample courses.
func NewMemoryCourseRepository() *MemoryCourseRepository {
	createdAt := time.Now().UTC()

	return &MemoryCourseRepository{
		courses: map[int]model.Course{
			1: {
				ID:            1,
				InstructorID:  1,
				Code:          "DS-101",
				Title:         "Data Science Fundamentals",
				Description:   "Introduction to practical data science concepts",
				Price:         15000000,
				Capacity:      25,
				DurationHours: 60,
				StartDate: time.Date(
					2026,
					time.October,
					1,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				EndDate: time.Date(
					2026,
					time.December,
					15,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				Status:    model.CourseStatusOpen,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
			2: {
				ID:            2,
				InstructorID:  2,
				Code:          "GO-201",
				Title:         "REST API Development with Go",
				Description:   "Building production-ready REST APIs with Go and PostgreSQL",
				Price:         18000000,
				Capacity:      20,
				DurationHours: 72,
				StartDate: time.Date(
					2026,
					time.October,
					10,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				EndDate: time.Date(
					2027,
					time.January,
					10,
					0,
					0,
					0,
					0,
					time.UTC,
				),
				Status:    model.CourseStatusOpen,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
		},
		nextID: 3,
	}
}

// List returns all courses ordered by ID.
func (r *MemoryCourseRepository) List(
	_ context.Context,
) ([]model.Course, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	courses := make([]model.Course, 0, len(r.courses))

	for _, course := range r.courses {
		courses = append(courses, course)
	}

	sort.Slice(courses, func(i int, j int) bool {
		return courses[i].ID < courses[j].ID
	})

	return courses, nil
}

// GetByID returns one course by ID.
func (r *MemoryCourseRepository) GetByID(
	_ context.Context,
	id int,
) (model.Course, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	course, exists := r.courses[id]
	if !exists {
		return model.Course{}, ErrCourseNotFound
	}

	return course, nil
}

// Create stores a new course.
func (r *MemoryCourseRepository) Create(
	_ context.Context,
	course model.Course,
) (model.Course, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.codeExists(course.Code, 0) {
		return model.Course{}, ErrCourseCodeExists
	}

	now := time.Now().UTC()

	course.ID = r.nextID
	course.Code = strings.ToUpper(strings.TrimSpace(course.Code))

	if course.Status == "" {
		course.Status = model.CourseStatusDraft
	}

	course.CreatedAt = now
	course.UpdatedAt = now

	r.courses[course.ID] = course
	r.nextID++

	return course, nil
}

// Update replaces editable course information.
func (r *MemoryCourseRepository) Update(
	_ context.Context,
	id int,
	course model.Course,
) (model.Course, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingCourse, exists := r.courses[id]
	if !exists {
		return model.Course{}, ErrCourseNotFound
	}

	if r.codeExists(course.Code, id) {
		return model.Course{}, ErrCourseCodeExists
	}

	course.ID = id
	course.Code = strings.ToUpper(strings.TrimSpace(course.Code))
	course.CreatedAt = existingCourse.CreatedAt
	course.UpdatedAt = time.Now().UTC()

	r.courses[id] = course

	return course, nil
}

// Delete removes a course from the memory repository.
func (r *MemoryCourseRepository) Delete(
	_ context.Context,
	id int,
) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.courses[id]; !exists {
		return ErrCourseNotFound
	}

	delete(r.courses, id)

	return nil
}

// codeExists reports whether a course code is already stored.
func (r *MemoryCourseRepository) codeExists(
	code string,
	excludedID int,
) bool {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))

	for id, course := range r.courses {
		if id == excludedID {
			continue
		}

		existingCode := strings.ToUpper(
			strings.TrimSpace(course.Code),
		)

		if existingCode == normalizedCode {
			return true
		}
	}

	return false
}
