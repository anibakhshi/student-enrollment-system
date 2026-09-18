package service

import (
	"context"
	"strings"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
)

// InstructorService contains instructor business logic.
type InstructorService struct {
	repository repository.InstructorRepository
}

// NewInstructorService creates an instructor service.
func NewInstructorService(
	instructorRepository repository.InstructorRepository,
) *InstructorService {
	return &InstructorService{
		repository: instructorRepository,
	}
}

// List returns all instructors.
func (s *InstructorService) List(
	ctx context.Context,
) ([]model.Instructor, error) {
	return s.repository.List(ctx)
}

// GetByID returns an instructor by ID.
func (s *InstructorService) GetByID(
	ctx context.Context,
	id int,
) (model.Instructor, error) {
	return s.repository.GetByID(ctx, id)
}

// Create validates and creates an instructor.
func (s *InstructorService) Create(
	ctx context.Context,
	input model.CreateInstructorInput,
) (model.Instructor, error) {
	normalizeCreateInstructorInput(&input)

	validationErrors := validateInstructorFields(
		input.FirstName,
		input.LastName,
		input.Email,
		input.Phone,
		input.Bio,
		input.Expertise,
		model.InstructorStatusActive,
	)

	if len(validationErrors) > 0 {
		return model.Instructor{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	instructor := model.Instructor{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Bio:       input.Bio,
		Expertise: input.Expertise,
		Status:    model.InstructorStatusActive,
	}

	return s.repository.Create(ctx, instructor)
}

// Update validates and updates an instructor.
func (s *InstructorService) Update(
	ctx context.Context,
	id int,
	input model.UpdateInstructorInput,
) (model.Instructor, error) {
	normalizeUpdateInstructorInput(&input)

	validationErrors := validateInstructorFields(
		input.FirstName,
		input.LastName,
		input.Email,
		input.Phone,
		input.Bio,
		input.Expertise,
		input.Status,
	)

	if len(validationErrors) > 0 {
		return model.Instructor{}, &ValidationError{
			Fields: validationErrors,
		}
	}

	instructor := model.Instructor{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Bio:       input.Bio,
		Expertise: input.Expertise,
		Status:    input.Status,
	}

	return s.repository.Update(
		ctx,
		id,
		instructor,
	)
}

// Delete removes or soft-deletes an instructor.
func (s *InstructorService) Delete(
	ctx context.Context,
	id int,
) error {
	return s.repository.Delete(ctx, id)
}

// normalizeCreateInstructorInput normalizes creation input.
func normalizeCreateInstructorInput(
	input *model.CreateInstructorInput,
) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.ToLower(
		strings.TrimSpace(input.Email),
	)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Bio = strings.TrimSpace(input.Bio)
	input.Expertise = strings.TrimSpace(input.Expertise)
}

// normalizeUpdateInstructorInput normalizes update input.
func normalizeUpdateInstructorInput(
	input *model.UpdateInstructorInput,
) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.ToLower(
		strings.TrimSpace(input.Email),
	)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Bio = strings.TrimSpace(input.Bio)
	input.Expertise = strings.TrimSpace(input.Expertise)

	input.Status = model.InstructorStatus(
		strings.ToLower(
			strings.TrimSpace(string(input.Status)),
		),
	)
}

// validateInstructorFields validates instructor information.
func validateInstructorFields(
	firstName string,
	lastName string,
	email string,
	phone string,
	bio string,
	expertise string,
	status model.InstructorStatus,
) map[string]string {
	validationErrors := make(map[string]string)

	if len(firstName) < 2 {
		validationErrors["first_name"] =
			"First name must contain at least 2 characters"
	}

	if len(lastName) < 2 {
		validationErrors["last_name"] =
			"Last name must contain at least 2 characters"
	}

	if !isValidEmail(email) {
		validationErrors["email"] =
			"Email address is invalid"
	}

	if phone != "" && !isExactlyDigits(phone, 11) {
		validationErrors["phone"] =
			"Phone number must contain exactly 11 digits"
	}

	if len(bio) > 2000 {
		validationErrors["bio"] =
			"Biography must not exceed 2000 characters"
	}

	if len(expertise) < 2 {
		validationErrors["expertise"] =
			"Expertise must contain at least 2 characters"
	}

	if !isValidInstructorStatus(status) {
		validationErrors["status"] =
			"Instructor status is invalid"
	}

	return validationErrors
}

// isValidInstructorStatus checks allowed instructor statuses.
func isValidInstructorStatus(
	status model.InstructorStatus,
) bool {
	switch status {
	case model.InstructorStatusActive,
		model.InstructorStatusInactive,
		model.InstructorStatusSuspended:
		return true

	default:
		return false
	}
}
