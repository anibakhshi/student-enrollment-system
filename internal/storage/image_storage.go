package storage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidImageExtension = errors.New("invalid image extension")
	ErrInvalidFileName       = errors.New("invalid file name")
)

// LocalImageStorage stores uploaded images on the local filesystem.
type LocalImageStorage struct {
	baseDirectory string
}

// NewLocalImageStorage creates a local image storage.
func NewLocalImageStorage(
	baseDirectory string,
) (*LocalImageStorage, error) {
	baseDirectory = strings.TrimSpace(baseDirectory)
	if baseDirectory == "" {
		return nil, errors.New("storage directory is required")
	}

	if err := os.MkdirAll(baseDirectory, 0o750); err != nil {
		return nil, fmt.Errorf(
			"create storage directory: %w",
			err,
		)
	}

	return &LocalImageStorage{
		baseDirectory: baseDirectory,
	}, nil
}

// Save stores an image and returns its generated file name.
func (s *LocalImageStorage) Save(
	source io.Reader,
	extension string,
) (string, error) {
	normalizedExtension, err := normalizeImageExtension(extension)
	if err != nil {
		return "", err
	}

	randomName, err := generateRandomName()
	if err != nil {
		return "", fmt.Errorf(
			"generate image file name: %w",
			err,
		)
	}

	fileName := randomName + normalizedExtension
	filePath := filepath.Join(s.baseDirectory, fileName)

	destination, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o640,
	)
	if err != nil {
		return "", fmt.Errorf(
			"create image file: %w",
			err,
		)
	}

	copySucceeded := false

	defer func() {
		_ = destination.Close()

		if !copySucceeded {
			_ = os.Remove(filePath)
		}
	}()

	if _, err := io.Copy(destination, source); err != nil {
		return "", fmt.Errorf(
			"write image file: %w",
			err,
		)
	}

	if err := destination.Sync(); err != nil {
		return "", fmt.Errorf(
			"sync image file: %w",
			err,
		)
	}

	if err := destination.Close(); err != nil {
		return "", fmt.Errorf(
			"close image file: %w",
			err,
		)
	}

	copySucceeded = true

	return fileName, nil
}

// Delete removes an image from local storage.
func (s *LocalImageStorage) Delete(fileName string) error {
	fileName = strings.TrimSpace(fileName)

	if fileName == "" {
		return nil
	}

	if filepath.Base(fileName) != fileName {
		return ErrInvalidFileName
	}

	filePath := filepath.Join(s.baseDirectory, fileName)

	err := os.Remove(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(
			"delete image file: %w",
			err,
		)
	}

	return nil
}

// normalizeImageExtension validates and normalizes image extensions.
func normalizeImageExtension(extension string) (string, error) {
	extension = strings.ToLower(strings.TrimSpace(extension))

	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	switch extension {
	case ".jpg", ".jpeg":
		return ".jpg", nil

	case ".png":
		return ".png", nil

	case ".webp":
		return ".webp", nil

	default:
		return "", ErrInvalidImageExtension
	}
}

// generateRandomName generates a cryptographically secure file name.
func generateRandomName() (string, error) {
	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
