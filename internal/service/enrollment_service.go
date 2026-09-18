package service

import (
	"context"
	"strings"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// EnrollmentService contains student enrollment business rules.
type EnrollmentService struct {
	enrollmentRepository repository.EnrollmentRepository
	studentRepository    repository.StudentRepository
	courseRepository     repository.CourseRepository
	now                  func() time.Time
}

// NewEnrollmentService creates an enrollment service.
func NewEnrollmentService(
	enrollmentRepository repository.EnrollmentRepository,
	studentRepository repository.StudentRepository,
	courseRepository repository.CourseRepository,
) *EnrollmentService {
	return &EnrollmentService{
		enrollmentRepository: enrollmentRepository,
		studentRepository:    studentRepository,
		courseRepository:     courseRepository,
		now:                  time.Now,
	}
}

// List returns all enrollments.
func (s *EnrollmentService) List(
	ctx context.Context,
) ([]model.Enrollment, error) {
	return s.enrollmentRepository.List(ctx)
}

// GetByID returns one enrollment.
func (s *EnrollmentService) GetByID(
	ctx context.Context,
	id int,
) (model.Enrollment, error) {
	if id <= 0 {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"id": "Enrollment ID must be a positive integer",
			},
		}
	}

	return s.enrollmentRepository.GetByID(ctx, id)
}

// GetDetails returns enrollment, student, and course information.
func (s *EnrollmentService) GetDetails(
	ctx context.Context,
	id int,
) (model.EnrollmentDetails, error) {
	enrollment, err := s.GetByID(ctx, id)
	if err != nil {
		return model.EnrollmentDetails{}, err
	}

	student, err := s.studentRepository.GetByID(
		ctx,
		enrollment.StudentID,
	)
	if err != nil {
		return model.EnrollmentDetails{}, err
	}

	course, err := s.courseRepository.GetByID(
		ctx,
		enrollment.CourseID,
	)
	if err != nil {
		return model.EnrollmentDetails{}, err
	}

	return model.EnrollmentDetails{
		Enrollment: enrollment,
		Student:    student,
		Course:     course,
	}, nil
}

// ListByStudentID returns all enrollments belonging to a student.
func (s *EnrollmentService) ListByStudentID(
	ctx context.Context,
	studentID int,
) ([]model.Enrollment, error) {
	if studentID <= 0 {
		return nil, &ValidationError{
			Fields: map[string]string{
				"student_id": "Student ID must be a positive integer",
			},
		}
	}

	if _, err := s.studentRepository.GetByID(
		ctx,
		studentID,
	); err != nil {
		return nil, err
	}

	return s.enrollmentRepository.ListByStudentID(
		ctx,
		studentID,
	)
}

// ListByCourseID returns all enrollments belonging to a course.
func (s *EnrollmentService) ListByCourseID(
	ctx context.Context,
	courseID int,
) ([]model.Enrollment, error) {
	if courseID <= 0 {
		return nil, &ValidationError{
			Fields: map[string]string{
				"course_id": "Course ID must be a positive integer",
			},
		}
	}

	if _, err := s.courseRepository.GetByID(
		ctx,
		courseID,
	); err != nil {
		return nil, err
	}

	return s.enrollmentRepository.ListByCourseID(
		ctx,
		courseID,
	)
}

// Create validates and creates a student enrollment.
func (s *EnrollmentService) Create(
	ctx context.Context,
	input model.CreateEnrollmentInput,
) (model.Enrollment, error) {
	input.Notes = strings.TrimSpace(input.Notes)

	validationErrors :=
		validateCreateEnrollmentInput(input)

	if len(validationErrors) > 0 {
		return model.Enrollment{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	if _, err := s.studentRepository.GetByID(
		ctx,
		input.StudentID,
	); err != nil {
		return model.Enrollment{}, err
	}

	course, err := s.courseRepository.GetByID(
		ctx,
		input.CourseID,
	)
	if err != nil {
		return model.Enrollment{}, err
	}

	if !course.IsOpen() {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"course_id": "Course must be open for enrollment",
			},
		}
	}

	currentTime := s.now().UTC()

	// An ended course has also started, so the end-date
	// validation must be checked first.
	if course.HasEnded(currentTime) {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"course_id": "Enrollment is closed because the course has ended",
			},
		}
	}

	if course.HasStarted(currentTime) {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"course_id": "Enrollment is closed because the course has started",
			},
		}
	}

	enrollment := model.Enrollment{
		StudentID: input.StudentID,
		CourseID:  input.CourseID,
		Status:    model.EnrollmentStatusPending,
		Notes:     input.Notes,
	}

	return s.enrollmentRepository.Create(
		ctx,
		enrollment,
		course.Capacity,
	)
}

// Update validates and changes enrollment status or notes.
func (s *EnrollmentService) Update(
	ctx context.Context,
	id int,
	input model.UpdateEnrollmentInput,
) (model.Enrollment, error) {
	if id <= 0 {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"id": "Enrollment ID must be a positive integer",
			},
		}
	}

	input.Notes = strings.TrimSpace(input.Notes)

	validationErrors :=
		validateUpdateEnrollmentInput(input)

	if len(validationErrors) > 0 {
		return model.Enrollment{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	existingEnrollment, err :=
		s.enrollmentRepository.GetByID(ctx, id)
	if err != nil {
		return model.Enrollment{}, err
	}

	if !isValidEnrollmentTransition(
		existingEnrollment.Status,
		input.Status,
	) {
		return model.Enrollment{}, &ValidationError{
			Fields: map[string]string{
				"status": enrollmentTransitionErrorMessage(
					existingEnrollment.Status,
					input.Status,
				),
			},
		}
	}

	existingEnrollment.Status = input.Status
	existingEnrollment.Notes = input.Notes

	currentTime := s.now().UTC()

	switch input.Status {
	case model.EnrollmentStatusConfirmed:
		if existingEnrollment.ConfirmedAt == nil {
			existingEnrollment.ConfirmedAt = &currentTime
		}

	case model.EnrollmentStatusCancelled:
		if existingEnrollment.CancelledAt == nil {
			existingEnrollment.CancelledAt = &currentTime
		}

	case model.EnrollmentStatusCompleted:
		if existingEnrollment.CompletedAt == nil {
			existingEnrollment.CompletedAt = &currentTime
		}
	}

	return s.enrollmentRepository.Update(
		ctx,
		id,
		existingEnrollment,
	)
}

// Delete cancels and soft-deletes an enrollment.
func (s *EnrollmentService) Delete(
	ctx context.Context,
	id int,
) error {
	if id <= 0 {
		return &ValidationError{
			Fields: map[string]string{
				"id": "Enrollment ID must be a positive integer",
			},
		}
	}

	return s.enrollmentRepository.Delete(ctx, id)
}

// CountActiveByCourseID returns the number of active enrollments.
func (s *EnrollmentService) CountActiveByCourseID(
	ctx context.Context,
	courseID int,
) (int, error) {
	if courseID <= 0 {
		return 0, &ValidationError{
			Fields: map[string]string{
				"course_id": "Course ID must be a positive integer",
			},
		}
	}

	if _, err := s.courseRepository.GetByID(
		ctx,
		courseID,
	); err != nil {
		return 0, err
	}

	return s.enrollmentRepository.CountActiveByCourseID(
		ctx,
		courseID,
	)
}

// validateCreateEnrollmentInput validates enrollment creation fields.
func validateCreateEnrollmentInput(
	input model.CreateEnrollmentInput,
) map[string]string {
	validationErrors := make(map[string]string)

	if input.StudentID <= 0 {
		validationErrors["student_id"] =
			"Student ID must be a positive integer"
	}

	if input.CourseID <= 0 {
		validationErrors["course_id"] =
			"Course ID must be a positive integer"
	}

	if len(input.Notes) > 2000 {
		validationErrors["notes"] =
			"Enrollment notes must not exceed 2000 characters"
	}

	return validationErrors
}

// validateUpdateEnrollmentInput validates enrollment update fields.
func validateUpdateEnrollmentInput(
	input model.UpdateEnrollmentInput,
) map[string]string {
	validationErrors := make(map[string]string)

	if !isValidEnrollmentStatus(input.Status) {
		validationErrors["status"] =
			"Enrollment status must be pending, confirmed, cancelled, or completed"
	}

	if len(input.Notes) > 2000 {
		validationErrors["notes"] =
			"Enrollment notes must not exceed 2000 characters"
	}

	return validationErrors
}

// isValidEnrollmentStatus reports whether a status is supported.
func isValidEnrollmentStatus(
	status model.EnrollmentStatus,
) bool {
	switch status {
	case model.EnrollmentStatusPending,
		model.EnrollmentStatusConfirmed,
		model.EnrollmentStatusCancelled,
		model.EnrollmentStatusCompleted:
		return true

	default:
		return false
	}
}

// isValidEnrollmentTransition validates enrollment lifecycle changes.
func isValidEnrollmentTransition(
	currentStatus model.EnrollmentStatus,
	newStatus model.EnrollmentStatus,
) bool {
	if currentStatus == newStatus {
		return currentStatus ==
			model.EnrollmentStatusPending ||
			currentStatus ==
				model.EnrollmentStatusConfirmed
	}

	switch currentStatus {
	case model.EnrollmentStatusPending:
		return newStatus ==
			model.EnrollmentStatusConfirmed ||
			newStatus ==
				model.EnrollmentStatusCancelled

	case model.EnrollmentStatusConfirmed:
		return newStatus ==
			model.EnrollmentStatusCompleted ||
			newStatus ==
				model.EnrollmentStatusCancelled

	case model.EnrollmentStatusCancelled,
		model.EnrollmentStatusCompleted:
		return false

	default:
		return false
	}
}

// enrollmentTransitionErrorMessage creates a readable status error.
func enrollmentTransitionErrorMessage(
	currentStatus model.EnrollmentStatus,
	newStatus model.EnrollmentStatus,
) string {
	return "Enrollment status cannot change from " +
		string(currentStatus) +
		" to " +
		string(newStatus)
}
