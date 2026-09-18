package service

import (
	"context"
	"errors"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// setupCourseService creates isolated repositories and a course service.
func setupCourseService() (
	*CourseService,
	*repository.MemoryCourseRepository,
	*repository.MemoryInstructorRepository,
) {
	courseRepository :=
		repository.NewMemoryCourseRepository()

	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	courseService := NewCourseService(
		courseRepository,
		instructorRepository,
	)

	return courseService,
		courseRepository,
		instructorRepository
}

// validCreateCourseInput returns valid sample course data.
func validCreateCourseInput() model.CreateCourseInput {
	return model.CreateCourseInput{
		InstructorID:  1,
		Code:          "ai-301",
		Title:         "Applied Artificial Intelligence",
		Description:   "Practical artificial intelligence course",
		Price:         20000000,
		Capacity:      30,
		DurationHours: 80,
		StartDate:     "2027-02-01",
		EndDate:       "2027-05-01",
	}
}

func TestCourseServiceList(t *testing.T) {
	courseService, _, _ := setupCourseService()

	courses, err := courseService.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(courses) != 2 {
		t.Fatalf(
			"expected 2 courses, but got %d",
			len(courses),
		)
	}
}

func TestCourseServiceGetByID(t *testing.T) {
	courseService, _, _ := setupCourseService()

	course, err := courseService.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if course.Code != "DS-101" {
		t.Errorf(
			"expected course code DS-101, but got %q",
			course.Code,
		)
	}
}

func TestCourseServiceCreate(t *testing.T) {
	courseService, _, _ := setupCourseService()

	createdCourse, err := courseService.Create(
		context.Background(),
		validCreateCourseInput(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdCourse.ID != 3 {
		t.Errorf(
			"expected course ID 3, but got %d",
			createdCourse.ID,
		)
	}

	if createdCourse.Code != "AI-301" {
		t.Errorf(
			"expected normalized code AI-301, but got %q",
			createdCourse.Code,
		)
	}

	if createdCourse.Status != model.CourseStatusDraft {
		t.Errorf(
			"expected draft status, but got %q",
			createdCourse.Status,
		)
	}
}

func TestCourseServiceValidation(t *testing.T) {
	courseService, _, _ := setupCourseService()

	input := model.CreateCourseInput{
		InstructorID:  0,
		Code:          "A",
		Title:         "",
		Description:   "",
		Price:         -1,
		Capacity:      0,
		DurationHours: 0,
		StartDate:     "wrong-date",
		EndDate:       "invalid-date",
	}

	_, err := courseService.Create(
		context.Background(),
		input,
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, but got %v",
			err,
		)
	}

	expectedFields := []string{
		"instructor_id",
		"code",
		"title",
		"price",
		"capacity",
		"duration_hours",
		"start_date",
		"end_date",
	}

	for _, field := range expectedFields {
		if _, exists := validationError.Fields[field]; !exists {
			t.Errorf(
				"expected validation error for %q",
				field,
			)
		}
	}
}

func TestCourseServiceRejectsInvalidDateRange(
	t *testing.T,
) {
	courseService, _, _ := setupCourseService()

	input := validCreateCourseInput()
	input.StartDate = "2027-06-01"
	input.EndDate = "2027-05-01"

	_, err := courseService.Create(
		context.Background(),
		input,
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, but got %v",
			err,
		)
	}

	if _, exists := validationError.Fields["end_date"]; !exists {
		t.Error("expected end_date validation error")
	}
}

func TestCourseServiceRejectsMissingInstructor(
	t *testing.T,
) {
	courseService, _, _ := setupCourseService()

	input := validCreateCourseInput()
	input.InstructorID = 999

	_, err := courseService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrInstructorNotFound) {
		t.Fatalf(
			"expected ErrInstructorNotFound, but got %v",
			err,
		)
	}
}

func TestCourseServiceRejectsInactiveInstructor(
	t *testing.T,
) {
	courseService, _, instructorRepository :=
		setupCourseService()

	instructor, err := instructorRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	instructor.Status = model.InstructorStatusInactive

	_, err = instructorRepository.Update(
		context.Background(),
		instructor.ID,
		instructor,
	)
	if err != nil {
		t.Fatalf(
			"failed to deactivate instructor: %v",
			err,
		)
	}

	_, err = courseService.Create(
		context.Background(),
		validCreateCourseInput(),
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, but got %v",
			err,
		)
	}

	message, exists :=
		validationError.Fields["instructor_id"]

	if !exists {
		t.Fatal(
			"expected instructor_id validation error",
		)
	}

	if message != "Instructor must be active" {
		t.Errorf(
			"unexpected validation message: %q",
			message,
		)
	}
}

func TestCourseServiceRejectsDuplicateCode(
	t *testing.T,
) {
	courseService, _, _ := setupCourseService()

	input := validCreateCourseInput()
	input.Code = "ds-101"

	_, err := courseService.Create(
		context.Background(),
		input,
	)

	if !errors.Is(err, repository.ErrCourseCodeExists) {
		t.Fatalf(
			"expected ErrCourseCodeExists, but got %v",
			err,
		)
	}
}

func TestCourseServiceUpdate(t *testing.T) {
	courseService, _, _ := setupCourseService()

	input := model.UpdateCourseInput{
		InstructorID:  1,
		Code:          "ds-101",
		Title:         "Advanced Data Science",
		Description:   "Advanced practical data science concepts",
		Price:         25000000,
		Capacity:      35,
		DurationHours: 90,
		StartDate:     "2027-03-01",
		EndDate:       "2027-06-01",
		Status:        model.CourseStatusOpen,
	}

	updatedCourse, err := courseService.Update(
		context.Background(),
		1,
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedCourse.Code != "DS-101" {
		t.Errorf(
			"expected normalized code DS-101, but got %q",
			updatedCourse.Code,
		)
	}

	if updatedCourse.Title != "Advanced Data Science" {
		t.Errorf(
			"unexpected updated title: %q",
			updatedCourse.Title,
		)
	}

	if updatedCourse.Status != model.CourseStatusOpen {
		t.Errorf(
			"unexpected updated status: %q",
			updatedCourse.Status,
		)
	}
}

func TestCourseServiceRejectsInvalidStatus(
	t *testing.T,
) {
	courseService, _, _ := setupCourseService()

	input := model.UpdateCourseInput{
		InstructorID:  1,
		Code:          "DS-101",
		Title:         "Data Science Fundamentals",
		Description:   "Data science course",
		Price:         15000000,
		Capacity:      25,
		DurationHours: 60,
		StartDate:     "2027-02-01",
		EndDate:       "2027-05-01",
		Status:        model.CourseStatus("unknown"),
	}

	_, err := courseService.Update(
		context.Background(),
		1,
		input,
	)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, but got %v",
			err,
		)
	}

	if _, exists := validationError.Fields["status"]; !exists {
		t.Error("expected status validation error")
	}
}

func TestCourseServiceDelete(t *testing.T) {
	courseService, _, _ := setupCourseService()

	err := courseService.Delete(
		context.Background(),
		2,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = courseService.GetByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, repository.ErrCourseNotFound) {
		t.Fatalf(
			"expected ErrCourseNotFound, but got %v",
			err,
		)
	}
}
