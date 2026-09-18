package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// setupInstructorTestRouter creates an isolated router
// containing student and instructor routes.
func setupInstructorTestRouter() http.Handler {
	studentRepository :=
		repository.NewMemoryStudentRepository()

	studentService :=
		service.NewStudentService(studentRepository)

	studentHandler :=
		NewStudentHandler(studentService)

	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	instructorService :=
		service.NewInstructorService(instructorRepository)

	instructorHandler :=
		NewInstructorHandler(instructorService)

	return NewRouter(
		studentHandler,
		RouterOptions{
			InstructorHandler: instructorHandler,
		},
	)
}

// performInstructorRequest sends a request to the test router.
func performInstructorRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	var request *http.Request

	if body == "" {
		request = httptest.NewRequest(
			method,
			path,
			nil,
		)
	} else {
		request = httptest.NewRequest(
			method,
			path,
			strings.NewReader(body),
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func TestListInstructorsEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	response := performInstructorRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/instructors",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"first_name":"Parham"`) {
		t.Error("expected response to contain Parham")
	}

	if !strings.Contains(body, `"first_name":"Maryam"`) {
		t.Error("expected response to contain Maryam")
	}
}

func TestGetInstructorByIDEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	response := performInstructorRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/instructors/1",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"first_name":"Parham"`,
	) {
		t.Error("expected response to contain Parham")
	}
}

func TestGetMissingInstructorEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	response := performInstructorRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/instructors/999",
		"",
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestGetInstructorWithInvalidIDEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	response := performInstructorRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/instructors/abc",
		"",
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestCreateInstructorEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	requestBody := `{
		"first_name": "Reza",
		"last_name": "Karimi",
		"email": "reza.karimi@example.com",
		"phone": "09121112222",
		"bio": "Backend software engineer and instructor",
		"expertise": "Go and Distributed Systems"
	}`

	response := performInstructorRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/instructors",
		requestBody,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"first_name":"Reza"`) {
		t.Error("expected response to contain created instructor")
	}

	if !strings.Contains(
		body,
		`"email":"reza.karimi@example.com"`,
	) {
		t.Error("expected response to contain instructor email")
	}
}

func TestCreateInstructorValidationEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	requestBody := `{
		"first_name": "R",
		"last_name": "",
		"email": "wrong-email",
		"phone": "123",
		"bio": "Invalid instructor",
		"expertise": ""
	}`

	response := performInstructorRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/instructors",
		requestBody,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Validation failed"`,
	) {
		t.Error("expected validation failure response")
	}
}

func TestCreateDuplicateInstructorEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	requestBody := `{
		"first_name": "Another",
		"last_name": "Instructor",
		"email": "parham.darvishi@example.com",
		"phone": "09123334444",
		"bio": "Duplicate instructor",
		"expertise": "Software Engineering"
	}`

	response := performInstructorRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/instructors",
		requestBody,
	)

	if response.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusConflict,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestUpdateInstructorEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	requestBody := `{
		"first_name": "Parham",
		"last_name": "Darvishi Updated",
		"email": "parham.updated@example.com",
		"phone": "09125556666",
		"bio": "Senior data science instructor",
		"expertise": "Data Science and Go",
		"status": "active"
	}`

	response := performInstructorRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/instructors/1",
		requestBody,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(
		body,
		`"last_name":"Darvishi Updated"`,
	) {
		t.Error("expected response to contain updated last name")
	}

	if !strings.Contains(
		body,
		`"email":"parham.updated@example.com"`,
	) {
		t.Error("expected response to contain updated email")
	}
}

func TestUpdateInstructorWithInvalidStatusEndpoint(
	t *testing.T,
) {
	router := setupInstructorTestRouter()

	requestBody := `{
		"first_name": "Parham",
		"last_name": "Darvishi",
		"email": "parham.darvishi@example.com",
		"phone": "09121111111",
		"bio": "Instructor",
		"expertise": "Data Science",
		"status": "unknown"
	}`

	response := performInstructorRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/instructors/1",
		requestBody,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Validation failed"`,
	) {
		t.Error("expected validation failure response")
	}
}

func TestDeleteInstructorEndpoint(t *testing.T) {
	router := setupInstructorTestRouter()

	deleteResponse := performInstructorRequest(
		t,
		router,
		http.MethodDelete,
		"/api/v1/instructors/2",
		"",
	)

	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusNoContent,
			deleteResponse.Code,
			deleteResponse.Body.String(),
		)
	}

	getResponse := performInstructorRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/instructors/2",
		"",
	)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected deleted instructor status %d, but got %d; body: %s",
			http.StatusNotFound,
			getResponse.Code,
			getResponse.Body.String(),
		)
	}
}
