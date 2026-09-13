package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrNationalCodeExists = errors.New("national code already exists")
	ErrEmailExists        = errors.New("email already exists")
)

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
}

// CreateStudentRequest contains the expected fields for creating a student.
type CreateStudentRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Age          int    `json:"age"`
	NationalCode string `json:"national_code"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// APIResponse defines the standard format of API responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// StudentStore is an in-memory, thread-safe student collection.
type StudentStore struct {
	mu       sync.RWMutex
	students []Student
	nextID   int
}

// NewStudentStore creates and initializes a student collection.
func NewStudentStore() *StudentStore {
	return &StudentStore{
		students: []Student{
			{
				ID:           1,
				FirstName:    "Anita",
				LastName:     "Bakhshi",
				Age:          20,
				NationalCode: "0012345678",
				Email:        "anita@example.com",
				Phone:        "09121234567",
				CreatedAt:    time.Now(),
			},
			{
				ID:           2,
				FirstName:    "Ali",
				LastName:     "Ahmadi",
				Age:          22,
				NationalCode: "0023456789",
				Email:        "ali@example.com",
				Phone:        "09129876543",
				CreatedAt:    time.Now(),
			},
		},
		nextID: 3,
	}
}

// List returns a safe copy of all students.
func (s *StudentStore) List() []Student {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Student, len(s.students))
	copy(result, s.students)

	return result
}

// Create adds a new student to the collection.
func (s *StudentStore) Create(student Student) (Student, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existingStudent := range s.students {
		if existingStudent.NationalCode == student.NationalCode {
			return Student{}, ErrNationalCodeExists
		}

		if strings.EqualFold(existingStudent.Email, student.Email) {
			return Student{}, ErrEmailExists
		}
	}

	student.ID = s.nextID
	student.CreatedAt = time.Now().UTC()

	s.nextID++
	s.students = append(s.students, student)

	return student, nil
}

var studentStore = NewStudentStore()

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/v1/students", getStudentsHandler)
	mux.HandleFunc("POST /api/v1/students", createStudentHandler)

	server := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info(
		"starting server",
		"address", "http://localhost:8081",
	)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

// healthHandler reports whether the application is running.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Message: "Student Enrollment System is healthy",
		Data: map[string]string{
			"status":  "healthy",
			"service": "student-enrollment-api",
			"version": "1.0.0",
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// getStudentsHandler returns all students.
func getStudentsHandler(w http.ResponseWriter, r *http.Request) {
	students := studentStore.List()

	response := APIResponse{
		Success: true,
		Message: "Students retrieved successfully",
		Data:    students,
	}

	writeJSON(w, http.StatusOK, response)
}

// createStudentHandler creates a new student.
func createStudentHandler(w http.ResponseWriter, r *http.Request) {
	// Limit the request body to one megabyte.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request CreateStudentRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON request",
			Errors: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	// Make sure the body contains only one JSON object.
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Request body must contain only one JSON object",
		})
		return
	}

	normalizeStudentRequest(&request)

	validationErrors := validateStudentRequest(request)

	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}

	student := Student{
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Age:          request.Age,
		NationalCode: request.NationalCode,
		Email:        request.Email,
		Phone:        request.Phone,
	}

	createdStudent, err := studentStore.Create(student)
	if err != nil {
		switch {
		case errors.Is(err, ErrNationalCodeExists):
			writeJSON(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "A student with this national code already exists",
				Errors: map[string]string{
					"national_code": "National code must be unique",
				},
			})

		case errors.Is(err, ErrEmailExists):
			writeJSON(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "A student with this email already exists",
				Errors: map[string]string{
					"email": "Email must be unique",
				},
			})

		default:
			slog.Error("failed to create student", "error", err)

			writeJSON(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
		}

		return
	}

	slog.Info(
		"student created",
		"student_id", createdStudent.ID,
		"national_code", createdStudent.NationalCode,
	)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Student created successfully",
		Data:    createdStudent,
	})
}

// normalizeStudentRequest removes unnecessary spaces and normalizes email.
func normalizeStudentRequest(request *CreateStudentRequest) {
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.NationalCode = strings.TrimSpace(request.NationalCode)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Phone = strings.TrimSpace(request.Phone)
}

// validateStudentRequest validates student input fields.
func validateStudentRequest(request CreateStudentRequest) map[string]string {
	validationErrors := make(map[string]string)

	if len(request.FirstName) < 2 {
		validationErrors["first_name"] = "First name must contain at least 2 characters"
	}

	if len(request.LastName) < 2 {
		validationErrors["last_name"] = "Last name must contain at least 2 characters"
	}

	if request.Age < 16 || request.Age > 100 {
		validationErrors["age"] = "Age must be between 16 and 100"
	}

	if !isExactlyDigits(request.NationalCode, 10) {
		validationErrors["national_code"] = "National code must contain exactly 10 digits"
	}

	if !isValidEmail(request.Email) {
		validationErrors["email"] = "Email address is invalid"
	}

	if request.Phone != "" && !isExactlyDigits(request.Phone, 11) {
		validationErrors["phone"] = "Phone number must contain exactly 11 digits"
	}

	return validationErrors
}

// isExactlyDigits checks that a string has the expected length and only digits.
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

// writeJSON sends a JSON response to the client.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
