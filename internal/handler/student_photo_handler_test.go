package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/anita-bakhshi/student-enrollment-system/internal/model"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
	"github.com/anita-bakhshi/student-enrollment-system/internal/storage"
)

// setupPhotoTestRouter creates a router with isolated image storage.
func setupPhotoTestRouter(
	t *testing.T,
) (http.Handler, string) {
	t.Helper()

	uploadDirectory := t.TempDir()

	imageStorage, err := storage.NewLocalImageStorage(
		uploadDirectory,
	)
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	studentRepository := repository.NewMemoryStudentRepository()
	studentService := service.NewStudentService(studentRepository)

	studentHandler := NewStudentHandler(studentService)

	studentPhotoHandler := NewStudentPhotoHandler(
		studentService,
		imageStorage,
	)

	router := NewRouter(
		studentHandler,
		RouterOptions{
			StudentPhotoHandler: studentPhotoHandler,
			UploadDirectory:     uploadDirectory,
		},
	)

	return router, uploadDirectory
}

// createMultipartPhotoRequest creates a multipart upload request.
func createMultipartPhotoRequest(
	t *testing.T,
	targetURL string,
	fileName string,
	content []byte,
) *http.Request {
	t.Helper()

	var requestBody bytes.Buffer

	multipartWriter := multipart.NewWriter(&requestBody)

	filePart, err := multipartWriter.CreateFormFile(
		"photo",
		fileName,
	)
	if err != nil {
		t.Fatalf("create multipart file field: %v", err)
	}

	if _, err := filePart.Write(content); err != nil {
		t.Fatalf("write multipart file content: %v", err)
	}

	if err := multipartWriter.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		targetURL,
		&requestBody,
	)

	request.Header.Set(
		"Content-Type",
		multipartWriter.FormDataContentType(),
	)

	return request
}

// onePixelPNG returns a valid one-pixel PNG image.
func onePixelPNG(t *testing.T) []byte {
	t.Helper()

	const encodedPNG = "" +
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB" +
		"CAQAAAC1HAwCAAAAC0lEQVR42mNk+A8A" +
		"AQUBAScY42YAAAAASUVORK5CYII="

	imageBytes, err := base64.StdEncoding.DecodeString(
		encodedPNG,
	)
	if err != nil {
		t.Fatalf("decode test PNG: %v", err)
	}

	return imageBytes
}

func TestUploadStudentProfileImage(t *testing.T) {
	router, uploadDirectory := setupPhotoTestRouter(t)

	uploadRequest := createMultipartPhotoRequest(
		t,
		"/api/v1/students/1/photo",
		"profile.png",
		onePixelPNG(t),
	)

	uploadResponse := httptest.NewRecorder()

	router.ServeHTTP(uploadResponse, uploadRequest)

	if uploadResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusOK,
			uploadResponse.Code,
			uploadResponse.Body.String(),
		)
	}

	var response struct {
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Data    model.Student `json:"data"`
	}

	if err := json.NewDecoder(uploadResponse.Body).Decode(
		&response,
	); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}

	if !response.Success {
		t.Fatal("expected successful upload response")
	}

	if !strings.HasPrefix(
		response.Data.ProfileImagePath,
		"/uploads/",
	) {
		t.Fatalf(
			"expected upload path, but got %q",
			response.Data.ProfileImagePath,
		)
	}

	storedFiles, err := os.ReadDir(uploadDirectory)
	if err != nil {
		t.Fatalf("read upload directory: %v", err)
	}

	if len(storedFiles) != 1 {
		t.Fatalf(
			"expected one stored image, but got %d",
			len(storedFiles),
		)
	}

	imageRequest := httptest.NewRequest(
		http.MethodGet,
		response.Data.ProfileImagePath,
		nil,
	)
	imageResponse := httptest.NewRecorder()

	router.ServeHTTP(imageResponse, imageRequest)

	if imageResponse.Code != http.StatusOK {
		t.Fatalf(
			"expected image status %d, but got %d",
			http.StatusOK,
			imageResponse.Code,
		)
	}

	if imageResponse.Header().Get("Content-Type") != "image/png" {
		t.Errorf(
			"expected image/png content type, but got %q",
			imageResponse.Header().Get("Content-Type"),
		)
	}

	if !bytes.Equal(
		imageResponse.Body.Bytes(),
		onePixelPNG(t),
	) {
		t.Error("served image content does not match uploaded image")
	}
}

func TestUploadStudentProfileImageRejectsTextFile(
	t *testing.T,
) {
	router, _ := setupPhotoTestRouter(t)

	request := createMultipartPhotoRequest(
		t,
		"/api/v1/students/1/photo",
		"not-an-image.txt",
		[]byte("this is not an image"),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusUnsupportedMediaType,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestUploadStudentProfileImageMissingStudent(
	t *testing.T,
) {
	router, _ := setupPhotoTestRouter(t)

	request := createMultipartPhotoRequest(
		t,
		"/api/v1/students/999/photo",
		"profile.png",
		onePixelPNG(t),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, but got %d; body: %s",
			http.StatusNotFound,
			response.Code,
			response.Body.String(),
		)
	}
}
