package model

import "time"

// EnrollmentStatus represents the current enrollment lifecycle status.
type EnrollmentStatus string

const (
	EnrollmentStatusPending   EnrollmentStatus = "pending"
	EnrollmentStatusConfirmed EnrollmentStatus = "confirmed"
	EnrollmentStatusCancelled EnrollmentStatus = "cancelled"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
)

// Enrollment connects a student to a course.
type Enrollment struct {
	ID          int              `json:"id"`
	StudentID   int              `json:"student_id"`
	CourseID    int              `json:"course_id"`
	Status      EnrollmentStatus `json:"status"`
	EnrolledAt  time.Time        `json:"enrolled_at"`
	ConfirmedAt *time.Time       `json:"confirmed_at,omitempty"`
	CancelledAt *time.Time       `json:"cancelled_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Notes       string           `json:"notes,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// EnrollmentDetails contains enrollment, student, and course information.
type EnrollmentDetails struct {
	Enrollment Enrollment `json:"enrollment"`
	Student    Student    `json:"student"`
	Course     Course     `json:"course"`
}

// CreateEnrollmentInput contains information required to enroll a student.
type CreateEnrollmentInput struct {
	StudentID int    `json:"student_id"`
	CourseID  int    `json:"course_id"`
	Notes     string `json:"notes"`
}

// UpdateEnrollmentInput contains editable enrollment information.
type UpdateEnrollmentInput struct {
	Status EnrollmentStatus `json:"status"`
	Notes  string           `json:"notes"`
}

// IsPending reports whether the enrollment is awaiting confirmation.
func (e Enrollment) IsPending() bool {
	return e.Status == EnrollmentStatusPending
}

// IsConfirmed reports whether the enrollment is confirmed.
func (e Enrollment) IsConfirmed() bool {
	return e.Status == EnrollmentStatusConfirmed
}

// IsCancelled reports whether the enrollment is cancelled.
func (e Enrollment) IsCancelled() bool {
	return e.Status == EnrollmentStatusCancelled
}

// IsCompleted reports whether the enrollment is completed.
func (e Enrollment) IsCompleted() bool {
	return e.Status == EnrollmentStatusCompleted
}

// IsActive reports whether the enrollment is currently active.
func (e Enrollment) IsActive() bool {
	return e.Status == EnrollmentStatusPending ||
		e.Status == EnrollmentStatusConfirmed
}
