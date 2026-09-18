package service

import (
	"context"
	"errors"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

func TestInstructorServiceList(t *testing.T) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	instructors, err := instructorService.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("list instructors: %v", err)
	}

	if len(instructors) != 2 {
		t.Fatalf(
			"expected 2 instructors, but got %d",
			len(instructors),
		)
	}
}

func TestInstructorServiceCreate(t *testing.T) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	instructor, err := instructorService.Create(
		context.Background(),
		model.CreateInstructorInput{
			FirstName: "  Sara ",
			LastName:  " Mohammadi ",
			Email:     " SARA.INSTRUCTOR@EXAMPLE.COM ",
			Phone:     "09125556666",
			Bio:       " Backend development instructor ",
			Expertise: " Go Programming ",
		},
	)
	if err != nil {
		t.Fatalf("create instructor: %v", err)
	}

	if instructor.FirstName != "Sara" {
		t.Errorf(
			"expected normalized first name, but got %q",
			instructor.FirstName,
		)
	}

	if instructor.Email != "sara.instructor@example.com" {
		t.Errorf(
			"expected normalized email, but got %q",
			instructor.Email,
		)
	}

	if instructor.Status != model.InstructorStatusActive {
		t.Errorf(
			"expected active instructor, but got %q",
			instructor.Status,
		)
	}
}

func TestInstructorServiceValidation(t *testing.T) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	_, err := instructorService.Create(
		context.Background(),
		model.CreateInstructorInput{
			FirstName: "A",
			LastName:  "",
			Email:     "invalid-email",
			Phone:     "123",
			Expertise: "",
		},
	)

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
		"email",
		"phone",
		"expertise",
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

func TestInstructorServiceRejectsDuplicateEmail(
	t *testing.T,
) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	_, err := instructorService.Create(
		context.Background(),
		model.CreateInstructorInput{
			FirstName: "Another",
			LastName:  "Instructor",
			Email:     "parham.darvishi@example.com",
			Phone:     "09120000000",
			Expertise: "Software Engineering",
		},
	)

	if !errors.Is(
		err,
		repository.ErrInstructorEmailExists,
	) {
		t.Fatalf(
			"expected ErrInstructorEmailExists, but got %v",
			err,
		)
	}
}

func TestInstructorServiceUpdate(t *testing.T) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	instructor, err := instructorService.Update(
		context.Background(),
		1,
		model.UpdateInstructorInput{
			FirstName: "Parham",
			LastName:  "Darvishi Updated",
			Email:     "parham.updated@example.com",
			Phone:     "09121112222",
			Bio:       "Updated biography",
			Expertise: "Data Engineering",
			Status:    model.InstructorStatusActive,
		},
	)
	if err != nil {
		t.Fatalf("update instructor: %v", err)
	}

	if instructor.Expertise != "Data Engineering" {
		t.Errorf(
			"unexpected expertise: %q",
			instructor.Expertise,
		)
	}
}

func TestInstructorServiceRejectsInvalidStatus(
	t *testing.T,
) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	_, err := instructorService.Update(
		context.Background(),
		1,
		model.UpdateInstructorInput{
			FirstName: "Parham",
			LastName:  "Darvishi",
			Email:     "parham.darvishi@example.com",
			Phone:     "09121112222",
			Expertise: "Data Science",
			Status:    model.InstructorStatus("unknown"),
		},
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

func TestInstructorServiceDelete(t *testing.T) {
	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		NewInstructorService(instructorRepository)

	if err := instructorService.Delete(
		context.Background(),
		2,
	); err != nil {
		t.Fatalf("delete instructor: %v", err)
	}

	_, err := instructorService.GetByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, repository.ErrInstructorNotFound) {
		t.Fatalf(
			"expected ErrInstructorNotFound, but got %v",
			err,
		)
	}
}
