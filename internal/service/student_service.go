package service

import (
	"context"
	"net/mail"
	"strings"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// ValidationError contains validation errors for input fields.
type ValidationError struct {
	Fields map[string]string
}

// Error implements the standard error interface.
func (e *ValidationError) Error() string {
	return "validation failed"
}

// StudentService contains student business logic.
type StudentService struct {
	repository repository.StudentRepository
}

// NewStudentService creates a new student service.
func NewStudentService(
	studentRepository repository.StudentRepository,
) *StudentService {
	return &StudentService{
		repository: studentRepository,
	}
}

// List returns all students.
func (s *StudentService) List(
	ctx context.Context,
) ([]model.Student, error) {
	return s.repository.List(ctx)
}

// GetByID returns a student by ID.
func (s *StudentService) GetByID(
	ctx context.Context,
	id int,
) (model.Student, error) {
	return s.repository.GetByID(ctx, id)
}

// Create validates and creates a student.
func (s *StudentService) Create(
	ctx context.Context,
	input model.CreateStudentInput,
) (model.Student, error) {
	normalizeCreateStudentInput(&input)

	validationErrors := validateCreateStudentInput(input)
	if len(validationErrors) > 0 {
		return model.Student{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	student := model.Student{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Age:          input.Age,
		NationalCode: input.NationalCode,
		Email:        input.Email,
		Phone:        input.Phone,
	}

	return s.repository.Create(ctx, student)
}

// normalizeCreateStudentInput removes unnecessary whitespace.
func normalizeCreateStudentInput(input *model.CreateStudentInput) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.NationalCode = strings.TrimSpace(input.NationalCode)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
}

// validateCreateStudentInput validates student information.
func validateCreateStudentInput(
	input model.CreateStudentInput,
) map[string]string {
	validationErrors := make(map[string]string)

	if len(input.FirstName) < 2 {
		validationErrors["first_name"] =
			"First name must contain at least 2 characters"
	}

	if len(input.LastName) < 2 {
		validationErrors["last_name"] =
			"Last name must contain at least 2 characters"
	}

	if input.Age < 16 || input.Age > 100 {
		validationErrors["age"] =
			"Age must be between 16 and 100"
	}

	if !isExactlyDigits(input.NationalCode, 10) {
		validationErrors["national_code"] =
			"National code must contain exactly 10 digits"
	}

	if !isValidEmail(input.Email) {
		validationErrors["email"] =
			"Email address is invalid"
	}

	if input.Phone != "" && !isExactlyDigits(input.Phone, 11) {
		validationErrors["phone"] =
			"Phone number must contain exactly 11 digits"
	}

	return validationErrors
}

// isExactlyDigits checks the length and characters of a numeric string.
func isExactlyDigits(value string, expectedLength int) bool {
	if len(value) != expectedLength {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

// isValidEmail checks the basic validity of an email address.
func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}

	return address.Address == value
}
