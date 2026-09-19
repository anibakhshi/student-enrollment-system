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
	// Configure structured JSON logging.
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	slog.SetDefault(logger)

	// Load application configuration.
	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"failed to load configuration",
			"error", err,
		)
		os.Exit(1)
	}

	// Create a startup context for the database connection.
	startupContext, cancelStartup := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelStartup()

	// Connect to PostgreSQL.
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

	// Initialize local profile-image storage.
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

	// Student dependencies.
	studentRepository :=
		repository.NewPostgresStudentRepository(databasePool)

	studentService :=
		service.NewStudentService(studentRepository)

	studentHandler :=
		handler.NewStudentHandler(studentService)

	studentPhotoHandler :=
		handler.NewStudentPhotoHandler(
			studentService,
			imageStorage,
		)

	// Instructor dependencies.
	instructorRepository :=
		repository.NewPostgresInstructorRepository(databasePool)

	instructorService :=
		service.NewInstructorService(instructorRepository)

	instructorHandler :=
		handler.NewInstructorHandler(instructorService)

	// Course dependencies.
	courseRepository :=
		repository.NewPostgresCourseRepository(databasePool)

	courseService :=
		service.NewCourseService(
			courseRepository,
			instructorRepository,
		)

	courseHandler :=
		handler.NewCourseHandler(courseService)

	// Enrollment dependencies.
	enrollmentRepository :=
		repository.NewPostgresEnrollmentRepository(databasePool)

	enrollmentService :=
		service.NewEnrollmentService(
			enrollmentRepository,
			studentRepository,
			courseRepository,
		)

	enrollmentHandler :=
		handler.NewEnrollmentHandler(enrollmentService)

	// Payment dependencies.
	paymentRepository :=
		repository.NewPostgresPaymentRepository(databasePool)

	paymentService :=
		service.NewPaymentService(
			paymentRepository,
			enrollmentRepository,
			studentRepository,
			courseRepository,
		)

	paymentHandler :=
		handler.NewPaymentHandler(paymentService)

	// Create the application router.
	router := handler.NewRouter(
		studentHandler,
		handler.RouterOptions{
			StudentPhotoHandler: studentPhotoHandler,
			InstructorHandler:   instructorHandler,
			CourseHandler:       courseHandler,
			EnrollmentHandler:   enrollmentHandler,
			PaymentHandler:      paymentHandler,
			UploadDirectory:     cfg.UploadDirectory,
		},
	)

	// Configure the HTTP server.
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

	// Start the HTTP server.
	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error(
			"server failed",
			"error", err,
		)
		os.Exit(1)
	}
}
