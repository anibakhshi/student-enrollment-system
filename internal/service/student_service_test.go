package service

import (
	"context"
	"errors"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

func TestStudentServiceList(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	students, err := studentService.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if len(students) != 2 {
		t.Errorf("expected 2 students, but got %d", len(students))
	}
}

func TestStudentServiceGetByID(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	student, err := studentService.GetByID(context.Background(), 1)
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

func TestStudentServiceCreate(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	input := model.CreateStudentInput{
		FirstName:    "  Sara  ",
		LastName:     "  Mohammadi  ",
		Age:          21,
		NationalCode: "1234567890",
		Email:        "  SARA@EXAMPLE.COM  ",
		Phone:        "09123456789",
	}

	student, err := studentService.Create(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if student.ID != 3 {
		t.Errorf("expected ID 3, but got %d", student.ID)
	}

	if student.FirstName != "Sara" {
		t.Errorf(
			"expected normalized first name Sara, but got %q",
			student.FirstName,
		)
	}

	if student.Email != "sara@example.com" {
		t.Errorf(
			"expected normalized email, but got %q",
			student.Email,
		)
	}
}

func TestStudentServiceValidation(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	input := model.CreateStudentInput{
		FirstName:    "A",
		LastName:     "",
		Age:          10,
		NationalCode: "123",
		Email:        "wrong-email",
		Phone:        "123",
	}

	_, err := studentService.Create(context.Background(), input)

	var validationError *ValidationError

	if !errors.As(err, &validationError) {
		t.Fatalf(
			"expected ValidationError, but got %v",
			err,
		)
	}

	expectedFields := []string{
		"first_name",
		"last_name",
		"age",
		"national_code",
		"email",
		"phone",
	}

	for _, field := range expectedFields {
		if _, exists := validationError.Fields[field]; !exists {
			t.Errorf(
				"expected validation error for field %q",
				field,
			)
		}
	}
}

func TestStudentServiceRejectsDuplicateNationalCode(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	input := model.CreateStudentInput{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          25,
		NationalCode: "0012345678",
		Email:        "new@example.com",
		Phone:        "09120000000",
	}

	_, err := studentService.Create(context.Background(), input)

	if !errors.Is(err, repository.ErrNationalCodeExists) {
		t.Errorf(
			"expected ErrNationalCodeExists, but got %v",
			err,
		)
	}
}

func TestStudentServiceRejectsDuplicateEmail(t *testing.T) {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := NewStudentService(studentRepository)

	input := model.CreateStudentInput{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          25,
		NationalCode: "9999999999",
		Email:        "ANITA@EXAMPLE.COM",
		Phone:        "09120000000",
	}

	_, err := studentService.Create(context.Background(), input)

	if !errors.Is(err, repository.ErrEmailExists) {
		t.Errorf(
			"expected ErrEmailExists, but got %v",
			err,
		)
	}
}
