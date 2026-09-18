package model

import "time"

// CourseStatus represents a course's current lifecycle status.
type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusOpen      CourseStatus = "open"
	CourseStatusClosed    CourseStatus = "closed"
	CourseStatusCompleted CourseStatus = "completed"
	CourseStatusCancelled CourseStatus = "cancelled"
)

// Course represents an educational course.
type Course struct {
	ID            int          `json:"id"`
	InstructorID  int          `json:"instructor_id"`
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Description   string       `json:"description,omitempty"`
	Price         int64        `json:"price"`
	Capacity      int          `json:"capacity"`
	DurationHours int          `json:"duration_hours"`
	StartDate     time.Time    `json:"start_date"`
	EndDate       time.Time    `json:"end_date"`
	Status        CourseStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// CourseDetails represents a course with its instructor information.
type CourseDetails struct {
	Course     Course     `json:"course"`
	Instructor Instructor `json:"instructor"`
}

// CreateCourseInput contains information required to create a course.
type CreateCourseInput struct {
	InstructorID  int    `json:"instructor_id"`
	Code          string `json:"code"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Price         int64  `json:"price"`
	Capacity      int    `json:"capacity"`
	DurationHours int    `json:"duration_hours"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
}

// UpdateCourseInput contains editable course information.
type UpdateCourseInput struct {
	InstructorID  int          `json:"instructor_id"`
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Price         int64        `json:"price"`
	Capacity      int          `json:"capacity"`
	DurationHours int          `json:"duration_hours"`
	StartDate     string       `json:"start_date"`
	EndDate       string       `json:"end_date"`
	Status        CourseStatus `json:"status"`
}

// IsOpen reports whether the course accepts registrations.
func (c Course) IsOpen() bool {
	return c.Status == CourseStatusOpen
}

// HasStarted reports whether the course has started.
func (c Course) HasStarted(now time.Time) bool {
	return !now.Before(c.StartDate)
}

// HasEnded reports whether the course has ended.
func (c Course) HasEnded(now time.Time) bool {
	return now.After(c.EndDate)
}
