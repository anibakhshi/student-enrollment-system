package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// setupEnrollmentTestRouter creates an isolated enrollment router.
func setupEnrollmentTestRouter(
	t *testing.T,
) (
	http.Handler,
	*repository.MemoryCourseRepository,
) {
	t.Helper()

	studentRepository :=
		repository.NewMemoryStudentRepository()

	courseRepository :=
		repository.NewMemoryCourseRepository()

	enrollmentRepository :=
		repository.NewMemoryEnrollmentRepository()

	// Keep the test independent from the system date.
	for _, courseID := range []int{1, 2} {
		course, err := courseRepository.GetByID(
			context.Background(),
			courseID,
		)
		if err != nil {
			t.Fatalf("get course %d: %v", courseID, err)
		}

		course.Status = model.CourseStatusOpen
		course.StartDate = time.Now().
			UTC().
			Add(30 * 24 * time.Hour)

		course.EndDate = time.Now().
			UTC().
			Add(90 * 24 * time.Hour)

		_, err = courseRepository.Update(
			context.Background(),
			course.ID,
			course,
		)
		if err != nil {
			t.Fatalf(
				"prepare course %d: %v",
				courseID,
				err,
			)
		}
	}

	studentService :=
		service.NewStudentService(studentRepository)

	studentHandler :=
		NewStudentHandler(studentService)

	enrollmentService :=
		service.NewEnrollmentService(
			enrollmentRepository,
			studentRepository,
			courseRepository,
		)

	enrollmentHandler :=
		NewEnrollmentHandler(enrollmentService)

	router := NewRouter(
		studentHandler,
		RouterOptions{
			EnrollmentHandler: enrollmentHandler,
		},
	)

	return router, courseRepository
}

// performEnrollmentRequest sends a request to the test router.
func performEnrollmentRequest(
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

func TestListEnrollmentsEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"data":[]`,
	) {
		t.Fatalf(
			"expected empty enrollment list; body: %s",
			response.Body.String(),
		)
	}
}

func TestCreateEnrollmentEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1,
			"notes": "Data science enrollment"
		}`,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"student_id":1`) {
		t.Error("expected student ID 1")
	}

	if !strings.Contains(body, `"course_id":1`) {
		t.Error("expected course ID 1")
	}

	if !strings.Contains(body, `"status":"pending"`) {
		t.Error("expected pending enrollment status")
	}
}

func TestGetEnrollmentEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1,
			"notes": "Enrollment details test"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments/1",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"enrollment"`) {
		t.Error("expected enrollment details")
	}

	if !strings.Contains(body, `"student"`) {
		t.Error("expected student details")
	}

	if !strings.Contains(body, `"course"`) {
		t.Error("expected course details")
	}

	if !strings.Contains(body, `"first_name":"Anita"`) {
		t.Error("expected Anita student information")
	}

	if !strings.Contains(body, `"code":"DS-101"`) {
		t.Error("expected DS-101 course information")
	}
}

func TestGetMissingEnrollmentEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments/999",
		"",
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestGetEnrollmentWithInvalidIDEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments/abc",
		"",
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestCreateEnrollmentValidationEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 0,
			"course_id": 0,
			"notes": "Invalid enrollment"
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
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

func TestCreateEnrollmentMissingStudentEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 999,
			"course_id": 1
		}`,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Student not found"`,
	) {
		t.Error("expected student not found response")
	}
}

func TestCreateEnrollmentMissingCourseEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 999
		}`,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"message":"Course not found"`,
	) {
		t.Error("expected course not found response")
	}
}

func TestCreateDuplicateEnrollmentEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	requestBody := `{
		"student_id": 1,
		"course_id": 1
	}`

	firstResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		requestBody,
	)

	if firstResponse.Code != http.StatusCreated {
		t.Fatalf(
			"first enrollment failed: %s",
			firstResponse.Body.String(),
		)
	}

	secondResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		requestBody,
	)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusConflict,
			secondResponse.Code,
			secondResponse.Body.String(),
		)
	}
}

func TestCreateEnrollmentCourseFullEndpoint(
	t *testing.T,
) {
	router, courseRepository :=
		setupEnrollmentTestRouter(t)

	course, err := courseRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}

	course.Capacity = 1

	_, err = courseRepository.Update(
		context.Background(),
		course.ID,
		course,
	)
	if err != nil {
		t.Fatalf("update course capacity: %v", err)
	}

	firstResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if firstResponse.Code != http.StatusCreated {
		t.Fatalf(
			"first enrollment failed: %s",
			firstResponse.Body.String(),
		)
	}

	secondResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 2,
			"course_id": 1
		}`,
	)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusConflict,
			secondResponse.Code,
			secondResponse.Body.String(),
		)
	}

	if !strings.Contains(
		secondResponse.Body.String(),
		`"message":"Course capacity has been reached"`,
	) {
		t.Error("expected course capacity error")
	}
}

func TestUpdateEnrollmentEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/enrollments/1",
		`{
			"status": "confirmed",
			"notes": "Enrollment confirmed"
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	body := response.Body.String()

	if !strings.Contains(body, `"status":"confirmed"`) {
		t.Error("expected confirmed status")
	}

	if !strings.Contains(body, `"confirmed_at"`) {
		t.Error("expected confirmed_at timestamp")
	}
}

func TestUpdateEnrollmentInvalidTransitionEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/enrollments/1",
		`{
			"status": "completed"
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestListStudentEnrollmentsEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/students/1/enrollments",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"student_id":1`,
	) {
		t.Error("expected student enrollment")
	}
}

func TestListCourseEnrollmentsEndpoint(
	t *testing.T,
) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	response := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/courses/1/enrollments",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"course_id":1`,
	) {
		t.Error("expected course enrollment")
	}
}

func TestDeleteEnrollmentEndpoint(t *testing.T) {
	router, _ := setupEnrollmentTestRouter(t)

	createResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/enrollments",
		`{
			"student_id": 1,
			"course_id": 1
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"create enrollment failed: %s",
			createResponse.Body.String(),
		)
	}

	deleteResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodDelete,
		"/api/v1/enrollments/1",
		"",
	)

	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNoContent,
			deleteResponse.Code,
			deleteResponse.Body.String(),
		)
	}

	getResponse := performEnrollmentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments/1",
		"",
	)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			getResponse.Code,
			getResponse.Body.String(),
		)
	}
}
