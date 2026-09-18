package repository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

func TestMemoryEnrollmentRepositoryCreate(t *testing.T) {
	enrollmentRepository := NewMemoryEnrollmentRepository()

	createdEnrollment, err := enrollmentRepository.Create(
		context.Background(),
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
			Notes:     "First enrollment",
		},
		20,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	if createdEnrollment.ID != 1 {
		t.Fatalf(
			"expected enrollment ID 1, got %d",
			createdEnrollment.ID,
		)
	}

	if createdEnrollment.StudentID != 1 {
		t.Fatalf(
			"expected student ID 1, got %d",
			createdEnrollment.StudentID,
		)
	}

	if createdEnrollment.CourseID != 10 {
		t.Fatalf(
			"expected course ID 10, got %d",
			createdEnrollment.CourseID,
		)
	}

	if createdEnrollment.Status !=
		model.EnrollmentStatusPending {
		t.Fatalf(
			"expected pending status, got %q",
			createdEnrollment.Status,
		)
	}

	if createdEnrollment.EnrolledAt.IsZero() {
		t.Fatal("expected enrolled_at to be populated")
	}

	if createdEnrollment.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}

	if createdEnrollment.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}
}

func TestMemoryEnrollmentRepositoryList(t *testing.T) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	_, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create first enrollment: %v", err)
	}

	_, err = enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 2,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create second enrollment: %v", err)
	}

	enrollments, err := enrollmentRepository.List(ctx)
	if err != nil {
		t.Fatalf("list enrollments: %v", err)
	}

	if len(enrollments) != 2 {
		t.Fatalf(
			"expected 2 enrollments, got %d",
			len(enrollments),
		)
	}
}

func TestMemoryEnrollmentRepositoryGetByID(t *testing.T) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	createdEnrollment, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	foundEnrollment, err := enrollmentRepository.GetByID(
		ctx,
		createdEnrollment.ID,
	)
	if err != nil {
		t.Fatalf("get enrollment: %v", err)
	}

	if foundEnrollment.ID != createdEnrollment.ID {
		t.Fatalf(
			"expected enrollment ID %d, got %d",
			createdEnrollment.ID,
			foundEnrollment.ID,
		)
	}
}

func TestMemoryEnrollmentRepositoryReturnsNotFound(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()

	_, err := enrollmentRepository.GetByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrEnrollmentNotFound) {
		t.Fatalf(
			"expected ErrEnrollmentNotFound, got %v",
			err,
		)
	}
}

func TestMemoryEnrollmentRepositoryListByStudentID(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	testEnrollments := []model.Enrollment{
		{
			StudentID: 1,
			CourseID:  10,
		},
		{
			StudentID: 1,
			CourseID:  20,
		},
		{
			StudentID: 2,
			CourseID:  10,
		},
	}

	for _, enrollment := range testEnrollments {
		_, err := enrollmentRepository.Create(
			ctx,
			enrollment,
			20,
		)
		if err != nil {
			t.Fatalf("create enrollment: %v", err)
		}
	}

	enrollments, err :=
		enrollmentRepository.ListByStudentID(ctx, 1)
	if err != nil {
		t.Fatalf("list student enrollments: %v", err)
	}

	if len(enrollments) != 2 {
		t.Fatalf(
			"expected 2 enrollments, got %d",
			len(enrollments),
		)
	}
}

func TestMemoryEnrollmentRepositoryListByCourseID(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	testEnrollments := []model.Enrollment{
		{
			StudentID: 1,
			CourseID:  10,
		},
		{
			StudentID: 2,
			CourseID:  10,
		},
		{
			StudentID: 3,
			CourseID:  20,
		},
	}

	for _, enrollment := range testEnrollments {
		_, err := enrollmentRepository.Create(
			ctx,
			enrollment,
			20,
		)
		if err != nil {
			t.Fatalf("create enrollment: %v", err)
		}
	}

	enrollments, err :=
		enrollmentRepository.ListByCourseID(ctx, 10)
	if err != nil {
		t.Fatalf("list course enrollments: %v", err)
	}

	if len(enrollments) != 2 {
		t.Fatalf(
			"expected 2 enrollments, got %d",
			len(enrollments),
		)
	}
}

func TestMemoryEnrollmentRepositoryRejectsDuplicate(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	enrollment := model.Enrollment{
		StudentID: 1,
		CourseID:  10,
	}

	_, err := enrollmentRepository.Create(
		ctx,
		enrollment,
		20,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	_, err = enrollmentRepository.Create(
		ctx,
		enrollment,
		20,
	)

	if !errors.Is(err, ErrEnrollmentExists) {
		t.Fatalf(
			"expected ErrEnrollmentExists, got %v",
			err,
		)
	}
}

func TestMemoryEnrollmentRepositoryRejectsFullCourse(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	_, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
		},
		1,
	)
	if err != nil {
		t.Fatalf("create first enrollment: %v", err)
	}

	_, err = enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 2,
			CourseID:  10,
		},
		1,
	)

	if !errors.Is(err, ErrCourseFull) {
		t.Fatalf(
			"expected ErrCourseFull, got %v",
			err,
		)
	}
}

func TestMemoryEnrollmentRepositoryCountActiveByCourseID(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	firstEnrollment, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create first enrollment: %v", err)
	}

	_, err = enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 2,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create second enrollment: %v", err)
	}

	cancelledAt := time.Now().UTC()

	firstEnrollment.Status =
		model.EnrollmentStatusCancelled
	firstEnrollment.CancelledAt = &cancelledAt

	_, err = enrollmentRepository.Update(
		ctx,
		firstEnrollment.ID,
		firstEnrollment,
	)
	if err != nil {
		t.Fatalf("cancel enrollment: %v", err)
	}

	count, err :=
		enrollmentRepository.CountActiveByCourseID(ctx, 10)
	if err != nil {
		t.Fatalf("count active enrollments: %v", err)
	}

	if count != 1 {
		t.Fatalf(
			"expected 1 active enrollment, got %d",
			count,
		)
	}
}

func TestMemoryEnrollmentRepositoryUpdate(t *testing.T) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	createdEnrollment, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
			Notes:     "Waiting",
		},
		20,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	confirmedAt := time.Now().UTC()

	createdEnrollment.Status =
		model.EnrollmentStatusConfirmed
	createdEnrollment.ConfirmedAt = &confirmedAt
	createdEnrollment.Notes = "Enrollment confirmed"

	updatedEnrollment, err := enrollmentRepository.Update(
		ctx,
		createdEnrollment.ID,
		createdEnrollment,
	)
	if err != nil {
		t.Fatalf("update enrollment: %v", err)
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

	if updatedEnrollment.Notes != "Enrollment confirmed" {
		t.Fatalf(
			"unexpected notes: %q",
			updatedEnrollment.Notes,
		)
	}
}

func TestMemoryEnrollmentRepositoryDelete(t *testing.T) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	createdEnrollment, err := enrollmentRepository.Create(
		ctx,
		model.Enrollment{
			StudentID: 1,
			CourseID:  10,
		},
		20,
	)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	err = enrollmentRepository.Delete(
		ctx,
		createdEnrollment.ID,
	)
	if err != nil {
		t.Fatalf("delete enrollment: %v", err)
	}

	_, err = enrollmentRepository.GetByID(
		ctx,
		createdEnrollment.ID,
	)

	if !errors.Is(err, ErrEnrollmentNotFound) {
		t.Fatalf(
			"expected ErrEnrollmentNotFound, got %v",
			err,
		)
	}
}

func TestMemoryEnrollmentRepositoryConcurrentCapacity(
	t *testing.T,
) {
	enrollmentRepository := NewMemoryEnrollmentRepository()
	ctx := context.Background()

	const (
		requestCount   = 20
		courseCapacity = 5
	)

	var (
		waitGroup      sync.WaitGroup
		successfulAdds atomic.Int32
	)

	for index := 1; index <= requestCount; index++ {
		waitGroup.Add(1)

		go func(studentID int) {
			defer waitGroup.Done()

			_, err := enrollmentRepository.Create(
				ctx,
				model.Enrollment{
					StudentID: studentID,
					CourseID:  10,
				},
				courseCapacity,
			)

			switch {
			case err == nil:
				successfulAdds.Add(1)

			case errors.Is(err, ErrCourseFull):
				return

			default:
				t.Errorf(
					"unexpected create error: %v",
					err,
				)
			}
		}(index)
	}

	waitGroup.Wait()

	if successfulAdds.Load() != courseCapacity {
		t.Fatalf(
			"expected %d successful enrollments, got %d",
			courseCapacity,
			successfulAdds.Load(),
		)
	}

	count, err :=
		enrollmentRepository.CountActiveByCourseID(ctx, 10)
	if err != nil {
		t.Fatalf("count active enrollments: %v", err)
	}

	if count != courseCapacity {
		t.Fatalf(
			"expected active count %d, got %d",
			courseCapacity,
			count,
		)
	}
}
