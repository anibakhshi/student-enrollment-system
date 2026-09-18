package model

import "time"

// Student represents a student in the enrollment system.
// Student represents a student in the enrollment system.
type Student struct {
	ID               int       `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Age              int       `json:"age"`
	NationalCode     string    `json:"national_code"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	ProfileImagePath string    `json:"profile_image_path,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateStudentInput contains information required to create a student.
type CreateStudentInput struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	NationalCode string `json:"national_code"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// UpdateStudentInput contains information required to update a student.
type UpdateStudentInput struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	NationalCode string `json:"national_code"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}
