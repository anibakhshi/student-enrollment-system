package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// setupCourseTestRouter creates an isolated router with course routes.
func setupCourseTestRouter() http.Handler {
	studentRepository :=
		repository.NewMemoryStudentRepository()

	studentService :=
		service.NewStudentService(studentRepository)

	studentHandler :=
		NewStudentHandler(studentService)

	instructorRepository :=
		repository.NewMemoryInstructorRepository()

	courseRepository :=
		repository.NewMemoryCourseRepository()

	courseService :=
		service.NewCourseService(
			courseRepository,
			instructorRepository,
		)

	courseHandler :=
		NewCourseHandler(courseService)

	return NewRouter(
		studentHandler,
		RouterOptions{
			CourseHandler: courseHandler,
		},
	)
}

// performCourseRequest sends a request to the course test router.
func performCourseRequest(
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

func TestListCoursesEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	response := performCourseRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses",
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

	if !strings.Contains(body, `"code":"DS-101"`) {
		t.Error("expected response to contain DS-101")
	}

	if !strings.Contains(body, `"code":"GO-201"`) {
		t.Error("expected response to contain GO-201")
	}
}

func TestGetCourseByIDEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	response := performCourseRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses/1",
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
		`"code":"DS-101"`,
	) {
		t.Error("expected response to contain DS-101")
	}
}

func TestGetMissingCourseEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	response := performCourseRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses/999",
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

func TestGetCourseWithInvalidIDEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	response := performCourseRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses/abc",
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

func TestCreateCourseEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 1,
		"code": "ai-301",
		"title": "Applied Artificial Intelligence",
		"description": "Practical artificial intelligence course",
		"price": 20000000,
		"capacity": 30,
		"duration_hours": 80,
		"start_date": "2027-02-01",
		"end_date": "2027-05-01"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/courses",
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

	if !strings.Contains(body, `"code":"AI-301"`) {
		t.Error("expected normalized course code AI-301")
	}

	if !strings.Contains(body, `"status":"draft"`) {
		t.Error("expected created course status to be draft")
	}
}

func TestCreateCourseValidationEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 0,
		"code": "A",
		"title": "",
		"description": "",
		"price": -1,
		"capacity": 0,
		"duration_hours": 0,
		"start_date": "wrong-date",
		"end_date": "invalid-date"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/courses",
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

func TestCreateCourseWithMissingInstructorEndpoint(
	t *testing.T,
) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 999,
		"code": "AI-302",
		"title": "Machine Learning",
		"description": "Machine learning course",
		"price": 22000000,
		"capacity": 25,
		"duration_hours": 90,
		"start_date": "2027-02-01",
		"end_date": "2027-05-01"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/courses",
		requestBody,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Instructor not found"`,
	) {
		t.Error("expected instructor not found response")
	}
}

func TestCreateDuplicateCourseEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 1,
		"code": "ds-101",
		"title": "Duplicate Data Science",
		"description": "Duplicate course",
		"price": 10000000,
		"capacity": 20,
		"duration_hours": 40,
		"start_date": "2027-02-01",
		"end_date": "2027-04-01"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/courses",
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

func TestUpdateCourseEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 1,
		"code": "ds-101",
		"title": "Advanced Data Science",
		"description": "Advanced practical data science concepts",
		"price": 25000000,
		"capacity": 35,
		"duration_hours": 90,
		"start_date": "2027-03-01",
		"end_date": "2027-06-01",
		"status": "open"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/courses/1",
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
		`"title":"Advanced Data Science"`,
	) {
		t.Error("expected response to contain updated title")
	}

	if !strings.Contains(body, `"status":"open"`) {
		t.Error("expected updated course status to be open")
	}
}

func TestUpdateCourseWithInvalidStatusEndpoint(
	t *testing.T,
) {
	router := setupCourseTestRouter()

	requestBody := `{
		"instructor_id": 1,
		"code": "DS-101",
		"title": "Data Science Fundamentals",
		"description": "Data science course",
		"price": 15000000,
		"capacity": 25,
		"duration_hours": 60,
		"start_date": "2027-02-01",
		"end_date": "2027-05-01",
		"status": "unknown"
	}`

	response := performCourseRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/courses/1",
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

func TestDeleteCourseEndpoint(t *testing.T) {
	router := setupCourseTestRouter()

	deleteResponse := performCourseRequest(
		t,
		router,
		http.MethodDelete,
		"/api/v1/courses/2",
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

	getResponse := performCourseRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses/2",
		"",
	)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected deleted course status %d, but got %d; body: %s",
			http.StatusNotFound,
			getResponse.Code,
			getResponse.Body.String(),
		)
	}
}
