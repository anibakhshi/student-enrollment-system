package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

func TestMemoryCourseRepositoryList(t *testing.T) {
	courseRepository := NewMemoryCourseRepository()

	courses, err := courseRepository.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(courses) != 2 {
		t.Fatalf(
			"expected 2 courses, but got %d",
			len(courses),
		)
	}

	if courses[0].Code != "DS-101" {
		t.Errorf(
			"expected first course code DS-101, but got %q",
			courses[0].Code,
		)
	}

	if courses[1].Code != "GO-201" {
		t.Errorf(
			"expected second course code GO-201, but got %q",
			courses[1].Code,
		)
	}
}

func TestMemoryCourseRepositoryGetByID(t *testing.T) {
	courseRepository := NewMemoryCourseRepository()

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if course.Title != "Data Science Fundamentals" {
		t.Errorf(
			"unexpected course title: %q",
			course.Title,
		)
	}
}

func TestMemoryCourseRepositoryReturnsNotFound(
	t *testing.T,
) {
	courseRepository := NewMemoryCourseRepository()

	_, err := courseRepository.GetByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf(
			"expected ErrCourseNotFound, but got %v",
			err,
		)
	}
}

func TestMemoryCourseRepositoryCreate(t *testing.T) {
	courseRepository := NewMemoryCourseRepository()

	course := model.Course{
		InstructorID:  1,
		Code:          "ai-301",
		Title:         "Applied Artificial Intelligence",
		Description:   "Practical artificial intelligence course",
		Price:         20000000,
		Capacity:      30,
		DurationHours: 80,
		StartDate: time.Date(
			2027,
			time.February,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		EndDate: time.Date(
			2027,
			time.May,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		Status: model.CourseStatusDraft,
	}

	createdCourse, err := courseRepository.Create(
		context.Background(),
		course,
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

	if createdCourse.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}
}

func TestMemoryCourseRepositoryRejectsDuplicateCode(
	t *testing.T,
) {
	courseRepository := NewMemoryCourseRepository()

	course := model.Course{
		InstructorID:  1,
		Code:          "ds-101",
		Title:         "Duplicate Course",
		Price:         1000000,
		Capacity:      10,
		DurationHours: 20,
		StartDate:     time.Now().UTC(),
		EndDate:       time.Now().UTC().AddDate(0, 1, 0),
		Status:        model.CourseStatusDraft,
	}

	_, err := courseRepository.Create(
		context.Background(),
		course,
	)

	if !errors.Is(err, ErrCourseCodeExists) {
		t.Fatalf(
			"expected ErrCourseCodeExists, but got %v",
			err,
		)
	}
}

func TestMemoryCourseRepositoryUpdate(t *testing.T) {
	courseRepository := NewMemoryCourseRepository()

	existingCourse, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	existingCourse.Title = "Advanced Data Science"
	existingCourse.Price = 25000000
	existingCourse.Status = model.CourseStatusClosed

	updatedCourse, err := courseRepository.Update(
		context.Background(),
		1,
		existingCourse,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedCourse.Title != "Advanced Data Science" {
		t.Errorf(
			"unexpected updated title: %q",
			updatedCourse.Title,
		)
	}

	if updatedCourse.Price != 25000000 {
		t.Errorf(
			"unexpected updated price: %d",
			updatedCourse.Price,
		)
	}

	if updatedCourse.Status != model.CourseStatusClosed {
		t.Errorf(
			"unexpected updated status: %q",
			updatedCourse.Status,
		)
	}
}

func TestMemoryCourseRepositoryDelete(t *testing.T) {
	courseRepository := NewMemoryCourseRepository()

	err := courseRepository.Delete(
		context.Background(),
		2,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = courseRepository.GetByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf(
			"expected deleted course to return ErrCourseNotFound, got %v",
			err,
		)
	}
}
