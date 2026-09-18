package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalImageStorageSave(t *testing.T) {
	storageDirectory := t.TempDir()

	imageStorage, err := NewLocalImageStorage(storageDirectory)
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	fileName, err := imageStorage.Save(
		strings.NewReader("fake image content"),
		".png",
	)
	if err != nil {
		t.Fatalf("save image: %v", err)
	}

	if filepath.Ext(fileName) != ".png" {
		t.Fatalf(
			"expected .png extension, but got %q",
			filepath.Ext(fileName),
		)
	}

	filePath := filepath.Join(storageDirectory, fileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read saved image: %v", err)
	}

	if string(content) != "fake image content" {
		t.Errorf(
			"unexpected saved content: %q",
			string(content),
		)
	}
}

func TestLocalImageStorageGeneratesUniqueNames(t *testing.T) {
	imageStorage, err := NewLocalImageStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	firstFileName, err := imageStorage.Save(
		strings.NewReader("first"),
		".jpg",
	)
	if err != nil {
		t.Fatalf("save first image: %v", err)
	}

	secondFileName, err := imageStorage.Save(
		strings.NewReader("second"),
		".jpg",
	)
	if err != nil {
		t.Fatalf("save second image: %v", err)
	}

	if firstFileName == secondFileName {
		t.Fatal("expected unique generated file names")
	}
}

func TestLocalImageStorageRejectsInvalidExtension(t *testing.T) {
	imageStorage, err := NewLocalImageStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	_, err = imageStorage.Save(
		strings.NewReader("not an image"),
		".exe",
	)

	if !errors.Is(err, ErrInvalidImageExtension) {
		t.Fatalf(
			"expected ErrInvalidImageExtension, but got %v",
			err,
		)
	}
}

func TestLocalImageStorageDelete(t *testing.T) {
	storageDirectory := t.TempDir()

	imageStorage, err := NewLocalImageStorage(storageDirectory)
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	fileName, err := imageStorage.Save(
		strings.NewReader("image"),
		".webp",
	)
	if err != nil {
		t.Fatalf("save image: %v", err)
	}

	if err := imageStorage.Delete(fileName); err != nil {
		t.Fatalf("delete image: %v", err)
	}

	_, err = os.Stat(filepath.Join(storageDirectory, fileName))

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf(
			"expected deleted file not to exist, but got %v",
			err,
		)
	}
}

func TestLocalImageStorageRejectsUnsafeFileName(t *testing.T) {
	imageStorage, err := NewLocalImageStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create image storage: %v", err)
	}

	err = imageStorage.Delete("../outside.png")

	if !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf(
			"expected ErrInvalidFileName, but got %v",
			err,
		)
	}
}
