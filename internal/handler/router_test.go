package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// setupTestRouter creates an isolated router for every test.
func setupTestRouter() http.Handler {
	studentRepository := repository.NewMemoryStudentRepository()
	studentService := service.NewStudentService(studentRepository)
	studentHandler := NewStudentHandler(studentService)

	return NewRouter(studentHandler)
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusOK,
			response.Code,
		)
	}

	contentType := response.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf(
			"expected JSON content type, but got %q",
			contentType,
		)
	}

	if response.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID response header")
	}
}

func TestListStudentsEndpoint(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/students",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusOK,
			response.Code,
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"first_name":"Anita"`) {
		t.Error("expected response to contain Anita")
	}

	if !strings.Contains(body, `"first_name":"Ali"`) {
		t.Error("expected response to contain Ali")
	}
}

func TestGetStudentByIDEndpoint(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/students/1",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusOK,
			response.Code,
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"first_name":"Anita"`,
	) {
		t.Error("expected response to contain Anita")
	}
}

func TestGetMissingStudentEndpoint(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/students/999",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusNotFound,
			response.Code,
		)
	}
}

func TestGetStudentWithInvalidIDEndpoint(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/students/abc",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestCreateStudentEndpoint(t *testing.T) {
	router := setupTestRouter()

	requestBody := `{
		"first_name": "Sara",
		"last_name": "Mohammadi",
		"age": 21,
		"national_code": "1234567890",
		"email": "sara@example.com",
		"phone": "09123456789"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/students",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Errorf(
			"expected status %d, but got %d; body: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"first_name":"Sara"`,
	) {
		t.Error("expected response to contain created student")
	}
}

func TestCreateStudentValidationEndpoint(t *testing.T) {
	router := setupTestRouter()

	requestBody := `{
		"first_name": "A",
		"last_name": "",
		"age": 10,
		"national_code": "123",
		"email": "wrong-email",
		"phone": "123"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/students",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Validation failed"`,
	) {
		t.Error("expected validation failure response")
	}
}

func TestCreateDuplicateStudentEndpoint(t *testing.T) {
	router := setupTestRouter()

	requestBody := `{
		"first_name": "Another",
		"last_name": "Student",
		"age": 25,
		"national_code": "0012345678",
		"email": "another@example.com",
		"phone": "09120000000"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/students",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Errorf(
			"expected status %d, but got %d",
			http.StatusConflict,
			response.Code,
		)
	}
}

func TestCustomRequestID(t *testing.T) {
	router := setupTestRouter()

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)
	request.Header.Set("X-Request-ID", "anita-test-001")

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	requestID := response.Header().Get("X-Request-ID")

	if requestID != "anita-test-001" {
		t.Errorf(
			"expected request ID %q, but got %q",
			"anita-test-001",
			requestID,
		)
	}
}
