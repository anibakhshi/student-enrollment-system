package handler

import "net/http"

// Health handles the application health-check endpoint.
func Health(w http.ResponseWriter, r *http.Request) {
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
