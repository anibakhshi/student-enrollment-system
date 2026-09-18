package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

const courseDateLayout = "2006-01-02"

// CourseService contains course business rules.
type CourseService struct {
	courseRepository     repository.CourseRepository
	instructorRepository repository.InstructorRepository
}

// NewCourseService creates a course service.
func NewCourseService(
	courseRepository repository.CourseRepository,
	instructorRepository repository.InstructorRepository,
) *CourseService {
	return &CourseService{
		courseRepository:     courseRepository,
		instructorRepository: instructorRepository,
	}
}

// List returns all available courses.
func (s *CourseService) List(
	ctx context.Context,
) ([]model.Course, error) {
	return s.courseRepository.List(ctx)
}

// GetByID returns one course.
func (s *CourseService) GetByID(
	ctx context.Context,
	id int,
) (model.Course, error) {
	return s.courseRepository.GetByID(ctx, id)
}

// Create validates and creates a course.
func (s *CourseService) Create(
	ctx context.Context,
	input model.CreateCourseInput,
) (model.Course, error) {
	course, err := buildCourse(
		input.InstructorID,
		input.Code,
		input.Title,
		input.Description,
		input.Price,
		input.Capacity,
		input.DurationHours,
		input.StartDate,
		input.EndDate,
		model.CourseStatusDraft,
		false,
	)
	if err != nil {
		return model.Course{}, err
	}

	if err := s.ensureActiveInstructor(
		ctx,
		course.InstructorID,
	); err != nil {
		return model.Course{}, err
	}

	return s.courseRepository.Create(ctx, course)
}

// Update validates and updates a course.
func (s *CourseService) Update(
	ctx context.Context,
	id int,
	input model.UpdateCourseInput,
) (model.Course, error) {
	_, err := s.courseRepository.GetByID(ctx, id)
	if err != nil {
		return model.Course{}, err
	}

	course, err := buildCourse(
		input.InstructorID,
		input.Code,
		input.Title,
		input.Description,
		input.Price,
		input.Capacity,
		input.DurationHours,
		input.StartDate,
		input.EndDate,
		input.Status,
		true,
	)
	if err != nil {
		return model.Course{}, err
	}

	if err := s.ensureActiveInstructor(
		ctx,
		course.InstructorID,
	); err != nil {
		return model.Course{}, err
	}

	return s.courseRepository.Update(
		ctx,
		id,
		course,
	)
}

// Delete soft-deletes a course.
func (s *CourseService) Delete(
	ctx context.Context,
	id int,
) error {
	return s.courseRepository.Delete(ctx, id)
}

// ensureActiveInstructor verifies that the instructor exists and is active.
func (s *CourseService) ensureActiveInstructor(
	ctx context.Context,
	instructorID int,
) error {
	instructor, err := s.instructorRepository.GetByID(
		ctx,
		instructorID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrInstructorNotFound) {
			return repository.ErrInstructorNotFound
		}

		return err
	}

	if !instructor.IsActive() {
		return &ValidationError{
			Fields: map[string]string{
				"instructor_id": "Instructor must be active",
			},
		}
	}

	return nil
}

// buildCourse normalizes and validates course information.
func buildCourse(
	instructorID int,
	code string,
	title string,
	description string,
	price int64,
	capacity int,
	durationHours int,
	startDateValue string,
	endDateValue string,
	status model.CourseStatus,
	validateStatus bool,
) (model.Course, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	startDateValue = strings.TrimSpace(startDateValue)
	endDateValue = strings.TrimSpace(endDateValue)

	validationErrors := make(map[string]string)

	if instructorID <= 0 {
		validationErrors["instructor_id"] =
			"Instructor ID must be a positive integer"
	}

	if len(code) < 2 {
		validationErrors["code"] =
			"Course code must contain at least 2 characters"
	} else if len(code) > 30 {
		validationErrors["code"] =
			"Course code must not exceed 30 characters"
	}

	if len(title) < 2 {
		validationErrors["title"] =
			"Course title must contain at least 2 characters"
	} else if len(title) > 200 {
		validationErrors["title"] =
			"Course title must not exceed 200 characters"
	}

	if len(description) > 5000 {
		validationErrors["description"] =
			"Course description must not exceed 5000 characters"
	}

	if price < 0 {
		validationErrors["price"] =
			"Course price must be zero or greater"
	} else if price > 999999999999 {
		validationErrors["price"] =
			"Course price is too large"
	}

	if capacity < 1 || capacity > 1000 {
		validationErrors["capacity"] =
			"Course capacity must be between 1 and 1000"
	}

	if durationHours < 1 || durationHours > 5000 {
		validationErrors["duration_hours"] =
			"Course duration must be between 1 and 5000 hours"
	}

	startDate, startDateError := time.Parse(
		courseDateLayout,
		startDateValue,
	)
	if startDateError != nil {
		validationErrors["start_date"] =
			"Start date must use YYYY-MM-DD format"
	}

	endDate, endDateError := time.Parse(
		courseDateLayout,
		endDateValue,
	)
	if endDateError != nil {
		validationErrors["end_date"] =
			"End date must use YYYY-MM-DD format"
	}

	if startDateError == nil &&
		endDateError == nil &&
		endDate.Before(startDate) {
		validationErrors["end_date"] =
			"End date must be equal to or after start date"
	}

	if validateStatus && !isValidCourseStatus(status) {
		validationErrors["status"] =
			"Course status must be draft, open, closed, completed, or cancelled"
	}

	if len(validationErrors) > 0 {
		return model.Course{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	return model.Course{
		InstructorID:  instructorID,
		Code:          code,
		Title:         title,
		Description:   description,
		Price:         price,
		Capacity:      capacity,
		DurationHours: durationHours,
		StartDate:     startDate,
		EndDate:       endDate,
		Status:        status,
	}, nil
}

// isValidCourseStatus reports whether a course status is supported.
func isValidCourseStatus(
	status model.CourseStatus,
) bool {
	switch status {
	case model.CourseStatusDraft,
		model.CourseStatusOpen,
		model.CourseStatusClosed,
		model.CourseStatusCompleted,
		model.CourseStatusCancelled:
		return true

	default:
		return false
	}
}
