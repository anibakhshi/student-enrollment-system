package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

func TestMemoryStudentRepositoryList(t *testing.T) {
	repository := NewMemoryStudentRepository()

	students, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if len(students) != 2 {
		t.Errorf("expected 2 students, but got %d", len(students))
	}
}

func TestMemoryStudentRepositoryGetByID(t *testing.T) {
	repository := NewMemoryStudentRepository()

	student, err := repository.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if student.FirstName != "Anita" {
		t.Errorf(
			"expected first name Anita, but got %s",
			student.FirstName,
		)
	}
}

func TestMemoryStudentRepositoryReturnsNotFound(t *testing.T) {
	repository := NewMemoryStudentRepository()

	_, err := repository.GetByID(context.Background(), 999)

	if !errors.Is(err, ErrStudentNotFound) {
		t.Errorf(
			"expected ErrStudentNotFound, but got %v",
			err,
		)
	}
}

func TestMemoryStudentRepositoryCreate(t *testing.T) {
	repository := NewMemoryStudentRepository()

	student := model.Student{
		FirstName:    "Sara",
		LastName:     "Mohammadi",
		Age:          21,
		NationalCode: "1234567890",
		Email:        "sara@example.com",
		Phone:        "09123456789",
	}

	createdStudent, err := repository.Create(
		context.Background(),
		student,
	)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if createdStudent.ID != 3 {
		t.Errorf(
			"expected student ID 3, but got %d",
			createdStudent.ID,
		)
	}

	students, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list students: %v", err)
	}

	if len(students) != 3 {
		t.Errorf("expected 3 students, but got %d", len(students))
	}
}

func TestMemoryStudentRepositoryRejectsDuplicateNationalCode(
	t *testing.T,
) {
	repository := NewMemoryStudentRepository()

	student := model.Student{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          24,
		NationalCode: "0012345678",
		Email:        "new@example.com",
		Phone:        "09120000000",
	}

	_, err := repository.Create(context.Background(), student)

	if !errors.Is(err, ErrNationalCodeExists) {
		t.Errorf(
			"expected ErrNationalCodeExists, but got %v",
			err,
		)
	}
}

func TestMemoryStudentRepositoryRejectsDuplicateEmail(t *testing.T) {
	repository := NewMemoryStudentRepository()

	student := model.Student{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          24,
		NationalCode: "9999999999",
		Email:        "ANITA@EXAMPLE.COM",
		Phone:        "09120000000",
	}

	_, err := repository.Create(context.Background(), student)

	if !errors.Is(err, ErrEmailExists) {
		t.Errorf(
			"expected ErrEmailExists, but got %v",
			err,
		)
	}
}
