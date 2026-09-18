package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
	"github.com/anita-bakhshi/student-enrollment-system/internal/storage"
)

const (
	maxProfileImageSize = 5 << 20
	maxMultipartSize    = 6 << 20
)

// StudentPhotoHandler handles student profile-image requests.
type StudentPhotoHandler struct {
	studentService *service.StudentService
	imageStorage   *storage.LocalImageStorage
}

// NewStudentPhotoHandler creates a student photo handler.
func NewStudentPhotoHandler(
	studentService *service.StudentService,
	imageStorage *storage.LocalImageStorage,
) *StudentPhotoHandler {
	return &StudentPhotoHandler{
		studentService: studentService,
		imageStorage:   imageStorage,
	}
}

// Upload handles POST /api/v1/students/{id}/photo.
func (h *StudentPhotoHandler) Upload(
	w http.ResponseWriter,
	r *http.Request,
) {
	studentID, err := parsePositiveStudentID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Student ID must be a positive integer",
			Errors: map[string]string{
				"id": "Invalid student ID",
			},
		})
		return
	}

	existingStudent, err := h.studentService.GetByID(
		r.Context(),
		studentID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Student not found",
			})
			return
		}

		slog.Error(
			"failed to get student before photo upload",
			"student_id", studentID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxMultipartSize,
	)

	if err := r.ParseMultipartForm(maxProfileImageSize); err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, APIResponse{
			Success: false,
			Message: "Uploaded image is too large",
			Errors: map[string]string{
				"photo": "Image size must not exceed 5 MB",
			},
		})
		return
	}

	uploadedFile, fileHeader, err := r.FormFile("photo")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Profile image is required",
			Errors: map[string]string{
				"photo": "Send an image using the photo field",
			},
		})
		return
	}
	defer uploadedFile.Close()

	if fileHeader.Size <= 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Uploaded image is empty",
			Errors: map[string]string{
				"photo": "Image file must not be empty",
			},
		})
		return
	}

	if fileHeader.Size > maxProfileImageSize {
		writeJSON(w, http.StatusRequestEntityTooLarge, APIResponse{
			Success: false,
			Message: "Uploaded image is too large",
			Errors: map[string]string{
				"photo": "Image size must not exceed 5 MB",
			},
		})
		return
	}

	contentBuffer := make([]byte, 512)

	bytesRead, readErr := io.ReadFull(
		uploadedFile,
		contentBuffer,
	)
	if readErr != nil &&
		!errors.Is(readErr, io.EOF) &&
		!errors.Is(readErr, io.ErrUnexpectedEOF) {
		slog.Error(
			"failed to inspect uploaded image",
			"student_id", studentID,
			"error", readErr,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Could not process uploaded image",
		})
		return
	}

	contentType := http.DetectContentType(
		contentBuffer[:bytesRead],
	)

	extension, valid := imageExtensionForContentType(contentType)
	if !valid {
		writeJSON(w, http.StatusUnsupportedMediaType, APIResponse{
			Success: false,
			Message: "Unsupported image format",
			Errors: map[string]string{
				"photo": "Only JPEG, PNG and WebP images are allowed",
			},
		})
		return
	}

	if _, err := uploadedFile.Seek(0, io.SeekStart); err != nil {
		slog.Error(
			"failed to reset uploaded image",
			"student_id", studentID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Could not process uploaded image",
		})
		return
	}

	fileName, err := h.imageStorage.Save(
		uploadedFile,
		extension,
	)
	if err != nil {
		slog.Error(
			"failed to store profile image",
			"student_id", studentID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Could not store uploaded image",
		})
		return
	}

	profileImagePath := "/uploads/" + fileName

	updatedStudent, err := h.studentService.UpdateProfileImage(
		r.Context(),
		studentID,
		profileImagePath,
	)
	if err != nil {
		_ = h.imageStorage.Delete(fileName)

		if errors.Is(err, repository.ErrStudentNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Student not found",
			})
			return
		}

		slog.Error(
			"failed to save student profile-image path",
			"student_id", studentID,
			"error", err,
		)

		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	deletePreviousProfileImage(
		h.imageStorage,
		existingStudent.ProfileImagePath,
	)

	slog.Info(
		"student profile image uploaded",
		"student_id", studentID,
		"content_type", contentType,
		"size_bytes", fileHeader.Size,
		"file_name", fileName,
	)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Student profile image uploaded successfully",
		Data:    updatedStudent,
	})
}

// parsePositiveStudentID validates a student ID path value.
func parsePositiveStudentID(value string) (int, error) {
	studentID, err := strconv.Atoi(value)
	if err != nil || studentID <= 0 {
		return 0, errors.New("invalid student ID")
	}

	return studentID, nil
}

// imageExtensionForContentType maps an allowed MIME type to an extension.
func imageExtensionForContentType(
	contentType string,
) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true

	case "image/png":
		return ".png", true

	case "image/webp":
		return ".webp", true

	default:
		return "", false
	}
}

// deletePreviousProfileImage removes a replaced profile image.
func deletePreviousProfileImage(
	imageStorage *storage.LocalImageStorage,
	profileImagePath string,
) {
	profileImagePath = strings.TrimSpace(profileImagePath)
	if profileImagePath == "" {
		return
	}

	fileName := filepath.Base(profileImagePath)
	if fileName == "." || fileName == "/" || fileName == "" {
		return
	}

	if err := imageStorage.Delete(fileName); err != nil {
		slog.Warn(
			"failed to delete previous profile image",
			"file_name", fileName,
			"error", err,
		)
	}
}
