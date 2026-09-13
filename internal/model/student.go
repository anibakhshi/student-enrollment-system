package model

import "time"

// Student represents a student in the enrollment system.
type Student struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Age          int       `json:"age"`
	NationalCode string    `json:"national_code"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateStudentInput contains the information required to create a student.
type CreateStudentInput struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	NationalCode string `json:"national_code"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}
