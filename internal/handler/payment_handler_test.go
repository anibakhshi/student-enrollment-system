package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
)

// setupPaymentTestRouter creates an isolated router for payment tests.
func setupPaymentTestRouter(
	t *testing.T,
) (
	http.Handler,
	*repository.MemoryEnrollmentRepository,
) {
	t.Helper()

	studentRepository :=
		repository.NewMemoryStudentRepository()

	courseRepository :=
		repository.NewMemoryCourseRepository()

	enrollmentRepository :=
		repository.NewMemoryEnrollmentRepository()

	paymentRepository :=
		repository.NewMemoryPaymentRepository()

	_, err := enrollmentRepository.Create(
		context.Background(),
		model.Enrollment{
			StudentID: 1,
			CourseID:  1,
			Status:    model.EnrollmentStatusPending,
			Notes:     "Payment handler test",
		},
		25,
	)
	if err != nil {
		t.Fatalf(
			"failed to create test enrollment: %v",
			err,
		)
	}

	studentService :=
		service.NewStudentService(studentRepository)

	studentHandler :=
		NewStudentHandler(studentService)

	paymentService :=
		service.NewPaymentService(
			paymentRepository,
			enrollmentRepository,
			studentRepository,
			courseRepository,
		)

	paymentHandler :=
		NewPaymentHandler(paymentService)

	router := NewRouter(
		studentHandler,
		RouterOptions{
			PaymentHandler: paymentHandler,
		},
	)

	return router, enrollmentRepository
}

// performPaymentRequest executes an HTTP request against the test router.
func performPaymentRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody io.Reader

	if body != "" {
		requestBody = bytes.NewBufferString(body)
	}

	request := httptest.NewRequest(
		method,
		path,
		requestBody,
	)

	if body != "" {
		request.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func TestListPaymentsEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/payments",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		"Payments retrieved successfully",
	) {
		t.Fatalf(
			"unexpected response: %s",
			response.Body.String(),
		)
	}
}

func TestCreatePaymentEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "online",
			"description": "Course payment"
		}`,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	responseBody := response.Body.String()

	expectedValues := []string{
		`"id":1`,
		`"enrollment_id":1`,
		`"amount":15000000`,
		`"currency":"IRR"`,
		`"payment_method":"online"`,
		`"status":"pending"`,
	}

	for _, expectedValue := range expectedValues {
		if !strings.Contains(responseBody, expectedValue) {
			t.Errorf(
				"expected response to contain %s: %s",
				expectedValue,
				responseBody,
			)
		}
	}
}

func TestCreatePaymentValidationEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 0,
			"payment_method": "crypto"
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"enrollment_id"`,
	) {
		t.Fatalf(
			"expected enrollment validation error: %s",
			response.Body.String(),
		)
	}
}

func TestCreatePaymentMissingEnrollmentEndpoint(
	t *testing.T,
) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 999,
			"payment_method": "online"
		}`,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		"Enrollment not found",
	) {
		t.Fatalf(
			"unexpected response: %s",
			response.Body.String(),
		)
	}
}

func TestCreateDuplicatePaymentEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	body := `{
		"enrollment_id": 1,
		"payment_method": "online"
	}`

	firstResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		body,
	)

	if firstResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create first payment: %s",
			firstResponse.Body.String(),
		)
	}

	secondResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		body,
	)

	if secondResponse.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			secondResponse.Code,
			secondResponse.Body.String(),
		)
	}
}

func TestGetPaymentEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	createResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "cash"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create payment: %s",
			createResponse.Body.String(),
		)
	}

	response := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/payments/1",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	responseBody := response.Body.String()

	expectedValues := []string{
		`"payment"`,
		`"enrollment"`,
		`"student"`,
		`"course"`,
	}

	for _, expectedValue := range expectedValues {
		if !strings.Contains(responseBody, expectedValue) {
			t.Errorf(
				"expected response to contain %s: %s",
				expectedValue,
				responseBody,
			)
		}
	}
}

func TestGetMissingPaymentEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/payments/999",
		"",
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestGetPaymentWithInvalidIDEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	response := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/payments/abc",
		"",
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestUpdatePaymentEndpoint(t *testing.T) {
	router, enrollmentRepository :=
		setupPaymentTestRouter(t)

	createResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "online"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create payment: %s",
			createResponse.Body.String(),
		)
	}

	response := performPaymentRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/payments/1",
		`{
			"status": "succeeded",
			"transaction_id": "handler-txn-001",
			"gateway_reference": "gateway-001",
			"card_last_four": "1234",
			"description": "Payment verified"
		}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	responseBody := response.Body.String()

	if !strings.Contains(
		responseBody,
		`"status":"succeeded"`,
	) {
		t.Fatalf(
			"unexpected payment response: %s",
			responseBody,
		)
	}

	if !strings.Contains(
		responseBody,
		`"transaction_id":"HANDLER-TXN-001"`,
	) {
		t.Fatalf(
			"unexpected transaction ID: %s",
			responseBody,
		)
	}

	enrollment, err := enrollmentRepository.GetByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("failed to get enrollment: %v", err)
	}

	if enrollment.Status != model.EnrollmentStatusConfirmed {
		t.Fatalf(
			"expected confirmed enrollment, got %q",
			enrollment.Status,
		)
	}
}

func TestUpdatePaymentRequiresTransactionID(
	t *testing.T,
) {
	router, _ := setupPaymentTestRouter(t)

	createResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "online"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create payment: %s",
			createResponse.Body.String(),
		)
	}

	response := performPaymentRequest(
		t,
		router,
		http.MethodPut,
		"/api/v1/payments/1",
		`{
			"status": "succeeded"
		}`,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"transaction_id"`,
	) {
		t.Fatalf(
			"expected transaction validation error: %s",
			response.Body.String(),
		)
	}
}

func TestListEnrollmentPaymentsEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	createResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "card"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create payment: %s",
			createResponse.Body.String(),
		)
	}

	response := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/enrollments/1/payments",
		"",
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(
		response.Body.String(),
		`"enrollment_id":1`,
	) {
		t.Fatalf(
			"unexpected response: %s",
			response.Body.String(),
		)
	}
}

func TestDeletePaymentEndpoint(t *testing.T) {
	router, _ := setupPaymentTestRouter(t)

	createResponse := performPaymentRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/payments",
		`{
			"enrollment_id": 1,
			"payment_method": "cash"
		}`,
	)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf(
			"failed to create payment: %s",
			createResponse.Body.String(),
		)
	}

	deleteResponse := performPaymentRequest(
		t,
		router,
		http.MethodDelete,
		"/api/v1/payments/1",
		"",
	)

	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNoContent,
			deleteResponse.Code,
			deleteResponse.Body.String(),
		)
	}

	getResponse := performPaymentRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/payments/1",
		"",
	)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf(
			"expected deleted payment to return 404, got %d",
			getResponse.Code,
		)
	}
}
