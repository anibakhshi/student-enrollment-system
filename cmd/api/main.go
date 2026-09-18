package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/anita-bakhshi/student-enrollment-system/internal/config"
	"github.com/anita-bakhshi/student-enrollment-system/internal/database"
	"github.com/anita-bakhshi/student-enrollment-system/internal/handler"
	"github.com/anita-bakhshi/student-enrollment-system/internal/repository"
	"github.com/anita-bakhshi/student-enrollment-system/internal/service"
	"github.com/anita-bakhshi/student-enrollment-system/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"failed to load configuration",
			"error", err,
		)
		os.Exit(1)
	}

	startupContext, cancelStartup := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelStartup()

	databasePool, err := database.NewPostgresPool(
		startupContext,
		cfg.DatabaseURL,
	)
	if err != nil {
		slog.Error(
			"failed to connect to PostgreSQL",
			"error", err,
		)
		os.Exit(1)
	}
	defer databasePool.Close()

	slog.Info("connected to PostgreSQL")

	imageStorage, err := storage.NewLocalImageStorage(
		cfg.UploadDirectory,
	)
	if err != nil {
		slog.Error(
			"failed to initialize image storage",
			"upload_directory", cfg.UploadDirectory,
			"error", err,
		)
		os.Exit(1)
	}

	slog.Info(
		"image storage initialized",
		"upload_directory", cfg.UploadDirectory,
	)

	studentRepository :=
		repository.NewPostgresStudentRepository(databasePool)

	studentService := service.NewStudentService(studentRepository)

	studentHandler := handler.NewStudentHandler(studentService)

	studentPhotoHandler := handler.NewStudentPhotoHandler(
		studentService,
		imageStorage,
	)

	router := handler.NewRouter(
		studentHandler,
		handler.RouterOptions{
			StudentPhotoHandler: studentPhotoHandler,
			UploadDirectory:     cfg.UploadDirectory,
		},
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	slog.Info(
		"starting server",
		"environment", cfg.AppEnvironment,
		"address", cfg.HTTPAddress,
	)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error(
			"server failed",
			"error", err,
		)
		os.Exit(1)
	}
}
