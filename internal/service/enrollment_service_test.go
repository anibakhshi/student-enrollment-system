package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

var fixedEnrollmentServiceTime = time.Date(
	2026,
	time.September,
	20,
	10,
	0,
	0,
	0,
	time.UTC,
)

// setupEnrollmentService creates isolated repositories and service.
func setupEnrollmentService() (
	*EnrollmentService,
	*repository.MemoryEnrollmentRepository,
	*repository.MemoryStudentRepository,
	*repository.MemoryCourseRepository,
) {
	enrollmentRepository :=
		repository.NewMemoryEnrollmentRepository()

	studentRepository :=
		repository.NewMemoryStudentRepository()

	courseRepository :=
		repository.NewMemoryCourseRepository()

	enrollmentService := NewEnrollmentService(
		enrollmentRepository,
		studentRepository,
		courseRepository,
	)

	enrollmentService.now = func() time.Time {
		return fixedEnrollmentServiceTime
	}

	return enrollmentService,
		enrollmentRepository,
		studentRepository,
		courseRepository
}

func validCreateEnrollmentInput() model.CreateEnrollmentInput {
	return model.CreateEnrollmentInput{
		StudentID: 1,
		CourseID:  1,
		Notes:     "  Data science registration  ",
	}
}

func TestEnrollmentServiceList(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	_, err := enrollmentService.Create(
		context.Background(),
		validCreateEnrollmentInput(),
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	enrollments, err := enrollmentService.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("list enrollments: %v", err)
	}

	if len(enrollments) != 1 {
		t.Fatalf(
			"expected 1 enrollment, got %d",
			len(enrollments),
		)
	}
}

func TestEnrollmentServiceCreate(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	enrollment, err := enrollmentService.Create(
		context.Background(),
		validCreateEnrollmentInput(),
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	if enrollment.ID != 1 {
		t.Fatalf(
			"expected enrollment ID 1, got %d",
			enrollment.ID,
		)
	}

	if enrollment.Status !=
		model.EnrollmentStatusPending {
		t.Fatalf(
			"expected pending status, got %q",
			enrollment.Status,
		)
	}

	if enrollment.Notes !=
		"Data science registration" {
		t.Fatalf(
			"unexpected normalized notes: %q",
			enrollment.Notes,
		)
	}
}

func TestEnrollmentServiceCreateValidation(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	_, err := enrollmentService.Create(
		context.Background(),
		model.CreateEnrollmentInput{
			StudentID: 0,
			CourseID:  -1,
			Notes:     strings.Repeat("a", 2001),
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	expectedFields := []string{
		"student_id",
		"course_id",
		"notes",
	}

	for _, field := range expectedFields {
		if _, exists :=
			validationError.Fields[field]; !exists {
			t.Errorf(
				"expected validation error for %q",
				field,
			)
		}
	}
}

func TestEnrollmentServiceRejectsMissingStudent(
	t *testing.T,
) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	input := validCreateEnrollmentInput()
	input.StudentID = 999

	_, err := enrollmentService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrStudentNotFound) {
		t.Fatalf(
			"expected ErrStudentNotFound, got %v",
			err,
		)
	}
}

func TestEnrollmentServiceRejectsMissingCourse(
	t *testing.T,
) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	input := validCreateEnrollmentInput()
	input.CourseID = 999

	_, err := enrollmentService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrCourseNotFound) {
		t.Fatalf(
			"expected ErrCourseNotFound, got %v",
			err,
		)
	}
}

func TestEnrollmentServiceRejectsClosedCourse(
	t *testing.T,
) {
	enrollmentService, _, _, courseRepository :=
		setupEnrollmentService()

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}

	course.Status = model.CourseStatusClosed

	_, err = courseRepository.Update(
		context.Background(),
		course.ID,
		course,
	)
	if err != nil {
		t.Fatalf("close course: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		validCreateEnrollmentInput(),
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["course_id"]; !exists {
		t.Fatal("expected course_id validation error")
	}
}

func TestEnrollmentServiceRejectsStartedCourse(
	t *testing.T,
) {
	enrollmentService, _, _, courseRepository :=
		setupEnrollmentService()

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}

	course.Status = model.CourseStatusOpen
	course.StartDate =
		fixedEnrollmentServiceTime.Add(-24 * time.Hour)
	course.EndDate =
		fixedEnrollmentServiceTime.Add(30 * 24 * time.Hour)

	_, err = courseRepository.Update(
		context.Background(),
		course.ID,
		course,
	)
	if err != nil {
		t.Fatalf("update course dates: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		validCreateEnrollmentInput(),
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["course_id"]; !exists {
		t.Fatal("expected course_id validation error")
	}
}

func TestEnrollmentServiceRejectsEndedCourse(
	t *testing.T,
) {
	enrollmentService, _, _, courseRepository :=
		setupEnrollmentService()

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}

	course.Status = model.CourseStatusOpen
	course.StartDate =
		fixedEnrollmentServiceTime.Add(-60 * 24 * time.Hour)
	course.EndDate =
		fixedEnrollmentServiceTime.Add(-24 * time.Hour)

	_, err = courseRepository.Update(
		context.Background(),
		course.ID,
		course,
	)
	if err != nil {
		t.Fatalf("update course dates: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		validCreateEnrollmentInput(),
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	message :=
		validationError.Fields["course_id"]

	if message !=
		"Enrollment is closed because the course has ended" {
		t.Fatalf(
			"unexpected validation message: %q",
			message,
		)
	}
}

func TestEnrollmentServiceRejectsDuplicate(
	t *testing.T,
) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	input := validCreateEnrollmentInput()

	_, err := enrollmentService.Create(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrEnrollmentExists) {
		t.Fatalf(
			"expected ErrEnrollmentExists, got %v",
			err,
		)
	}
}

func TestEnrollmentServiceRejectsFullCourse(
	t *testing.T,
) {
	enrollmentService, _, _, courseRepository :=
		setupEnrollmentService()

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}

	course.Capacity = 1

	_, err = courseRepository.Update(
		context.Background(),
		course.ID,
		course,
	)
	if err != nil {
		t.Fatalf("update course capacity: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		model.CreateEnrollmentInput{
			StudentID: 1,
			CourseID:  1,
		},
	)
	if err != nil {
		t.Fatalf("create first enrollment: %v", err)
	}

	_, err = enrollmentService.Create(
		context.Background(),
		model.CreateEnrollmentInput{
			StudentID: 2,
			CourseID:  1,
		},
	)

	if !errors.Is(err, repository.ErrCourseFull) {
		t.Fatalf(
			"expected ErrCourseFull, got %v",
			err,
		)
	}
}

func TestEnrollmentServiceGetDetails(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	details, err := enrollmentService.GetDetails(
		context.Background(),
		createdEnrollment.ID,
	)
	if err != nil {
		t.Fatalf("get enrollment details: %v", err)
	}

	if details.Enrollment.ID != createdEnrollment.ID {
		t.Fatalf("unexpected enrollment details")
	}

	if details.Student.ID != 1 {
		t.Fatalf(
			"expected student ID 1, got %d",
			details.Student.ID,
		)
	}

	if details.Course.ID != 1 {
		t.Fatalf(
			"expected course ID 1, got %d",
			details.Course.ID,
		)
	}
}

func TestEnrollmentServiceConfirm(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	updatedEnrollment, err :=
		enrollmentService.Update(
			context.Background(),
			createdEnrollment.ID,
			model.UpdateEnrollmentInput{
				Status: model.EnrollmentStatusConfirmed,
				Notes:  "Payment will be added later",
			},
		)
	if err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}

	if updatedEnrollment.Status !=
		model.EnrollmentStatusConfirmed {
		t.Fatalf(
			"expected confirmed status, got %q",
			updatedEnrollment.Status,
		)
	}

	if updatedEnrollment.ConfirmedAt == nil {
		t.Fatal("expected confirmed_at to be populated")
	}
}

func TestEnrollmentServiceComplete(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	confirmedEnrollment, err :=
		enrollmentService.Update(
			context.Background(),
			createdEnrollment.ID,
			model.UpdateEnrollmentInput{
				Status: model.EnrollmentStatusConfirmed,
			},
		)
	if err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}

	completedEnrollment, err :=
		enrollmentService.Update(
			context.Background(),
			confirmedEnrollment.ID,
			model.UpdateEnrollmentInput{
				Status: model.EnrollmentStatusCompleted,
			},
		)
	if err != nil {
		t.Fatalf("complete enrollment: %v", err)
	}

	if completedEnrollment.Status !=
		model.EnrollmentStatusCompleted {
		t.Fatalf(
			"expected completed status, got %q",
			completedEnrollment.Status,
		)
	}

	if completedEnrollment.CompletedAt == nil {
		t.Fatal("expected completed_at to be populated")
	}
}

func TestEnrollmentServiceRejectsInvalidTransition(
	t *testing.T,
) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	_, err = enrollmentService.Update(
		context.Background(),
		createdEnrollment.ID,
		model.UpdateEnrollmentInput{
			Status: model.EnrollmentStatusCompleted,
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["status"]; !exists {
		t.Fatal("expected status validation error")
	}
}

func TestEnrollmentServiceRejectsUnknownStatus(
	t *testing.T,
) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	_, err = enrollmentService.Update(
		context.Background(),
		createdEnrollment.ID,
		model.UpdateEnrollmentInput{
			Status: model.EnrollmentStatus("unknown"),
		},
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, got %v",
			err,
		)
	}

	if _, exists :=
		validationError.Fields["status"]; !exists {
		t.Fatal("expected status validation error")
	}
}

func TestEnrollmentServiceDelete(t *testing.T) {
	enrollmentService, _, _, _ :=
		setupEnrollmentService()

	createdEnrollment, err :=
		enrollmentService.Create(
			context.Background(),
			validCreateEnrollmentInput(),
		)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	err = enrollmentService.Delete(
		context.Background(),
		createdEnrollment.ID,
	)
	if err != nil {
		t.Fatalf("delete enrollment: %v", err)
	}

	_, err = enrollmentService.GetByID(
		context.Background(),
		createdEnrollment.ID,
	)

	if !errors.Is(err, repository.ErrEnrollmentNotFound) {
		t.Fatalf(
			"expected ErrEnrollmentNotFound, got %v",
			err,
		)
	}
}
