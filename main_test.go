package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	responseRecorder := httptest.NewRecorder()

	healthHandler(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status code %d, but got %d",
			http.StatusOK,
			responseRecorder.Code,
		)
	}

	var response APIResponse

	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "Student Enrollment System is healthy" {
		t.Errorf(
			"unexpected response message: %s",
			response.Message,
		)
	}
}

func TestValidateStudentRequestWithValidData(t *testing.T) {
	request := CreateStudentRequest{
		FirstName:    "Sara",
		LastName:     "Mohammadi",
		Age:          21,
		NationalCode: "1234567890",
		Email:        "sara@example.com",
		Phone:        "09123456789",
	}

	validationErrors := validateStudentRequest(request)

	if len(validationErrors) != 0 {
		t.Errorf(
			"expected no validation errors, but got %v",
			validationErrors,
		)
	}
}

func TestValidateStudentRequestWithInvalidData(t *testing.T) {
	request := CreateStudentRequest{
		FirstName:    "A",
		LastName:     "",
		Age:          10,
		NationalCode: "123",
		Email:        "wrong-email",
		Phone:        "123",
	}

	validationErrors := validateStudentRequest(request)

	expectedFields := []string{
		"first_name",
		"last_name",
		"age",
		"national_code",
		"email",
		"phone",
	}

	for _, field := range expectedFields {
		if _, exists := validationErrors[field]; !exists {
			t.Errorf("expected validation error for field %q", field)
		}
	}
}

func TestStudentStoreCreate(t *testing.T) {
	store := NewStudentStore()

	student := Student{
		FirstName:    "Reza",
		LastName:     "Karimi",
		Age:          25,
		NationalCode: "9876543210",
		Email:        "reza@example.com",
		Phone:        "09121112233",
	}

	createdStudent, err := store.Create(student)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if createdStudent.ID != 3 {
		t.Errorf(
			"expected student ID 3, but got %d",
			createdStudent.ID,
		)
	}

	students := store.List()

	if len(students) != 3 {
		t.Errorf(
			"expected 3 students, but got %d",
			len(students),
		)
	}
}

func TestStudentStoreRejectsDuplicateNationalCode(t *testing.T) {
	store := NewStudentStore()

	student := Student{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          24,
		NationalCode: "0012345678",
		Email:        "new-email@example.com",
		Phone:        "09120000000",
	}

	_, err := store.Create(student)

	if !errors.Is(err, ErrNationalCodeExists) {
		t.Errorf(
			"expected ErrNationalCodeExists, but got %v",
			err,
		)
	}
}

func TestStudentStoreRejectsDuplicateEmail(t *testing.T) {
	store := NewStudentStore()

	student := Student{
		FirstName:    "Test",
		LastName:     "Student",
		Age:          24,
		NationalCode: "9999999999",
		Email:        "ANITA@EXAMPLE.COM",
		Phone:        "09120000000",
	}

	_, err := store.Create(student)

	if !errors.Is(err, ErrEmailExists) {
		t.Errorf(
			"expected ErrEmailExists, but got %v",
			err,
		)
	}
}
