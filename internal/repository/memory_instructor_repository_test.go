package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
)

func TestMemoryInstructorRepositoryList(t *testing.T) {
	instructorRepository := NewMemoryInstructorRepository()

	instructors, err := instructorRepository.List(
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

func TestMemoryInstructorRepositoryGetByID(t *testing.T) {
	instructorRepository := NewMemoryInstructorRepository()

	instructor, err := instructorRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get instructor: %v", err)
	}

	if instructor.FirstName != "Parham" {
		t.Errorf(
			"expected instructor Parham, but got %q",
			instructor.FirstName,
		)
	}
}

func TestMemoryInstructorRepositoryCreate(t *testing.T) {
	instructorRepository := NewMemoryInstructorRepository()

	instructor, err := instructorRepository.Create(
		context.Background(),
		model.Instructor{
			FirstName: "Sara",
			LastName:  "Mohammadi",
			Email:     "sara.instructor@example.com",
			Phone:     "09125556666",
			Bio:       "Software instructor",
			Expertise: "Backend Development",
		},
	)
	if err != nil {
		t.Fatalf("create instructor: %v", err)
	}

	if instructor.ID != 3 {
		t.Errorf(
			"expected instructor ID 3, but got %d",
			instructor.ID,
		)
	}

	if instructor.Status != model.InstructorStatusActive {
		t.Errorf(
			"expected active status, but got %q",
			instructor.Status,
		)
	}
}

func TestMemoryInstructorRepositoryRejectsDuplicateEmail(
	t *testing.T,
) {
	instructorRepository := NewMemoryInstructorRepository()

	_, err := instructorRepository.Create(
		context.Background(),
		model.Instructor{
			FirstName: "Another",
			LastName:  "Instructor",
			Email:     "parham.darvishi@example.com",
		},
	)

	if !errors.Is(err, ErrInstructorEmailExists) {
		t.Fatalf(
			"expected ErrInstructorEmailExists, but got %v",
			err,
		)
	}
}

func TestMemoryInstructorRepositoryUpdate(t *testing.T) {
	instructorRepository := NewMemoryInstructorRepository()

	instructor, err := instructorRepository.Update(
		context.Background(),
		1,
		model.Instructor{
			FirstName: "Parham",
			LastName:  "Darvishi Updated",
			Email:     "parham.updated@example.com",
			Phone:     "09121112222",
			Bio:       "Updated biography",
			Expertise: "Software Engineering",
			Status:    model.InstructorStatusActive,
		},
	)
	if err != nil {
		t.Fatalf("update instructor: %v", err)
	}

	if instructor.LastName != "Darvishi Updated" {
		t.Errorf(
			"unexpected last name: %q",
			instructor.LastName,
		)
	}
}

func TestMemoryInstructorRepositoryDelete(t *testing.T) {
	instructorRepository := NewMemoryInstructorRepository()

	err := instructorRepository.Delete(
		context.Background(),
		2,
	)
	if err != nil {
		t.Fatalf("delete instructor: %v", err)
	}

	_, err = instructorRepository.GetByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, ErrInstructorNotFound) {
		t.Fatalf(
			"expected ErrInstructorNotFound, but got %v",
			err,
		)
	}
}
