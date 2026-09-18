package model

import "time"

// InstructorStatus represents an instructor's current status.
type InstructorStatus string

const (
	InstructorStatusActive    InstructorStatus = "active"
	InstructorStatusInactive  InstructorStatus = "inactive"
	InstructorStatusSuspended InstructorStatus = "suspended"
)

// Instructor represents an instructor in the enrollment system.
type Instructor struct {
	ID        int              `json:"id"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone,omitempty"`
	Bio       string           `json:"bio,omitempty"`
	Expertise string           `json:"expertise,omitempty"`
	Status    InstructorStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// CreateInstructorInput contains information required to create an instructor.
type CreateInstructorInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Bio       string `json:"bio"`
	Expertise string `json:"expertise"`
}

// UpdateInstructorInput contains editable instructor information.
type UpdateInstructorInput struct {
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
	Bio       string           `json:"bio"`
	Expertise string           `json:"expertise"`
	Status    InstructorStatus `json:"status"`
}

// FullName returns the instructor's full name.
func (i Instructor) FullName() string {
	if i.FirstName == "" {
		return i.LastName
	}

	if i.LastName == "" {
		return i.FirstName
	}

	return i.FirstName + " " + i.LastName
}

// IsActive reports whether the instructor can teach courses.
func (i Instructor) IsActive() bool {
	return i.Status == InstructorStatusActive
}
