# Student Enrollment System

A production-oriented REST API for managing students, instructors, courses, enrollments, profile images, and payments.

The project is implemented in Go using a layered architecture, PostgreSQL persistence, automated SQL migrations, structured logging, Docker Compose, validation, concurrency-safe repositories, and comprehensive tests.

## Project Information

- **Student:** Anita Bakhshi
- **Instructor:** Parham Darvishi
- **Language:** Go 1.22
- **Database:** PostgreSQL 16
- **API Port:** `8081`

## Features

### Student Management

- Create, list, retrieve, update, and soft-delete students
- Validate student information
- Prevent duplicate national codes and email addresses
- Upload and serve secure profile images
- Preserve uploaded files in a Docker volume

### Instructor Management

- Complete instructor CRUD operations
- Validate instructor status and contact information
- Prevent duplicate instructor email addresses
- Soft-delete instructor records

### Course Management

- Complete course CRUD operations
- Assign courses to instructors
- Validate course dates, capacity, duration, price, and status
- Prevent duplicate course codes
- Reject inactive or missing instructors

### Enrollment Management

- Enroll students in courses
- List enrollments globally, by student, or by course
- Prevent duplicate active enrollments
- Enforce course capacity
- Protect capacity checks against concurrent requests
- Support enrollment lifecycle transitions:

```text
pending → confirmed → completed
       ↘ cancelled
```

### Payment Management

- Create payments for enrollments
- Obtain the authoritative payment amount from the course price
- Prevent clients from manually changing the payment amount
- Prevent multiple active payments for one enrollment
- Prevent duplicate transaction IDs
- Support failed payment retries
- Automatically confirm pending enrollments after successful payment
- Support the following payment lifecycle:

```text
pending → succeeded → refunded
       ↘ failed
```

### Infrastructure

- PostgreSQL persistence using `pgx`
- Sequential SQL migrations
- Docker multi-stage build
- Docker Compose orchestration
- PostgreSQL health checks
- Persistent database and upload volumes
- Adminer database interface
- Structured JSON logs
- Request ID middleware
- Non-root API container
- Unit, integration, concurrency, and race-detector tests

## Technology Stack

| Component | Technology |
|---|---|
| Language | Go 1.22 |
| HTTP server | Go `net/http` |
| Database | PostgreSQL 16 |
| PostgreSQL driver | `pgx/v5` |
| Containers | Docker |
| Orchestration | Docker Compose |
| Database UI | Adminer |
| Testing | Go `testing`, `httptest`, Race Detector |
| Logging | Go `log/slog` |

## Architecture

The application follows a layered architecture:

```mermaid
flowchart TD
    Client["HTTP Client"] --> Middleware["Request ID and Logging"]
    Middleware --> Handler["HTTP Handlers"]
    Handler --> Service["Business Services"]
    Service --> Repository["Repository Interfaces"]
    Repository --> PostgreSQL["PostgreSQL Repositories"]
    Repository --> Memory["In-memory Repositories"]
    PostgreSQL --> Database["PostgreSQL"]
    Handler --> Storage["Local Image Storage"]
```

### Layers

- **Handler:** HTTP parsing, response generation, and status codes
- **Service:** validation and business rules
- **Repository:** persistence abstraction and domain errors
- **Model:** application entities and request structures
- **Database:** PostgreSQL connection pool
- **Storage:** secure local profile-image storage
- **Middleware:** structured request logging and request IDs

## Database Relationships

```mermaid
erDiagram
    INSTRUCTORS ||--o{ COURSES : teaches
    STUDENTS ||--o{ ENROLLMENTS : creates
    COURSES ||--o{ ENROLLMENTS : contains
    ENROLLMENTS ||--o{ PAYMENTS : receives
```

## Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── service/
│   └── storage/
├── migrations/
│   ├── 000001_create_students.up.sql
│   ├── 000001_create_students.down.sql
│   ├── 000002_create_instructors_courses.up.sql
│   ├── 000002_create_instructors_courses.down.sql
│   ├── 000003_create_enrollments.up.sql
│   ├── 000003_create_enrollments.down.sql
│   ├── 000004_create_payments.up.sql
│   └── 000004_create_payments.down.sql
├── scripts/
│   ├── build.sh
│   ├── dev.sh
│   ├── logs.sh
│   ├── migrate-up.sh
│   ├── migrate-down.sh
│   ├── start.sh
│   ├── stop.sh
│   └── test.sh
├── .env.example
├── .dockerignore
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Requirements

For the Docker-based setup:

- Docker Engine
- Docker Compose

For local development:

- Go 1.22 or newer
- Docker and Docker Compose for PostgreSQL

## Quick Start with Docker

### 1. Clone the repository

```bash
git clone https://github.com/anita-bakhshi/student-enrollment-system.git
cd student-enrollment-system
```

### 2. Create the environment file

```bash
cp .env.example .env
```

Change passwords and configuration values when using the project outside a local development environment.

### 3. Start the application

```bash
./scripts/start.sh
```

Alternatively:

```bash
docker compose up -d --build
```

The migration container automatically applies pending migrations before the API starts.

### 4. Check the containers

```bash
docker compose ps -a
```

### 5. Check API health

```bash
curl -i http://localhost:8081/health
```

Expected status:

```text
HTTP/1.1 200 OK
```

### Available Services

| Service | Address |
|---|---|
| REST API | http://localhost:8081 |
| Health endpoint | http://localhost:8081/health |
| Adminer | http://localhost:18080 |
| PostgreSQL host port | `5433` |

### Adminer Connection

Use these values with the default development configuration:

| Field | Value |
|---|---|
| System | PostgreSQL |
| Server | `postgres` |
| Username | `student_app` |
| Password | Value of `POSTGRES_PASSWORD` |
| Database | `student_enrollment` |

## Local Development

Copy the environment configuration:

```bash
cp .env.example .env
```

Start PostgreSQL and run the API locally:

```bash
./scripts/dev.sh
```

This command:

1. Stops the containerized API
2. Starts PostgreSQL and Adminer
3. Applies pending migrations
4. Loads variables from `.env`
5. Executes `go run ./cmd/api`

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Application environment |
| `APP_PORT` | `8081` | HTTP server port |
| `HTTP_READ_TIMEOUT` | `10s` | HTTP read timeout |
| `HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `HTTP_IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `POSTGRES_DB` | `student_enrollment` | PostgreSQL database |
| `POSTGRES_USER` | `student_app` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `student_secret` | Development password |
| `POSTGRES_HOST_PORT` | `5433` | PostgreSQL host port |
| `DATABASE_URL` | PostgreSQL URL | API database connection string |
| `UPLOAD_DIR` | `./uploads` | Profile-image storage directory |

Do not commit the local `.env` file. It is excluded through `.gitignore`.

## API Response Format

Successful responses use this structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {}
}
```

Validation errors use this structure:

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "field": "Validation message"
  }
}
```

## API Endpoints

### Health

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check API health |

### Students

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/students` | List students |
| GET | `/api/v1/students/{id}` | Get one student |
| POST | `/api/v1/students` | Create a student |
| PUT | `/api/v1/students/{id}` | Update a student |
| DELETE | `/api/v1/students/{id}` | Soft-delete a student |
| POST | `/api/v1/students/{id}/photo` | Upload profile image |
| GET | `/uploads/{filename}` | Retrieve uploaded image |
| GET | `/api/v1/students/{id}/enrollments` | List student enrollments |

### Instructors

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/instructors` | List instructors |
| GET | `/api/v1/instructors/{id}` | Get one instructor |
| POST | `/api/v1/instructors` | Create an instructor |
| PUT | `/api/v1/instructors/{id}` | Update an instructor |
| DELETE | `/api/v1/instructors/{id}` | Soft-delete an instructor |

### Courses

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/courses` | List courses |
| GET | `/api/v1/courses/{id}` | Get one course |
| POST | `/api/v1/courses` | Create a course |
| PUT | `/api/v1/courses/{id}` | Update a course |
| DELETE | `/api/v1/courses/{id}` | Soft-delete a course |
| GET | `/api/v1/courses/{id}/enrollments` | List course enrollments |

### Enrollments

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/enrollments` | List enrollments |
| GET | `/api/v1/enrollments/{id}` | Get enrollment details |
| POST | `/api/v1/enrollments` | Enroll a student |
| PUT | `/api/v1/enrollments/{id}` | Change enrollment status |
| DELETE | `/api/v1/enrollments/{id}` | Soft-delete an enrollment |
| GET | `/api/v1/enrollments/{id}/payments` | List enrollment payments |

### Payments

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/payments` | List payments |
| GET | `/api/v1/payments/{id}` | Get complete payment details |
| POST | `/api/v1/payments` | Create a pending payment |
| PUT | `/api/v1/payments/{id}` | Change payment status |
| DELETE | `/api/v1/payments/{id}` | Delete pending or failed payment |

## API Examples

### Create a Student

```bash
curl -i -X POST \
  http://localhost:8081/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Sara",
    "last_name": "Mohammadi",
    "age": 21,
    "national_code": "1234567890",
    "email": "sara@example.com",
    "phone": "09123456789"
  }'
```

### Upload a Student Profile Image

```bash
curl -i -X POST \
  http://localhost:8081/api/v1/students/1/photo \
  -F "photo=@/path/to/profile.png"
```

### Create an Enrollment

```bash
curl -i -X POST \
  http://localhost:8081/api/v1/enrollments \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": 1,
    "course_id": 1,
    "notes": "Online registration"
  }'
```

### Confirm an Enrollment

```bash
curl -i -X PUT \
  http://localhost:8081/api/v1/enrollments/1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "confirmed",
    "notes": "Enrollment confirmed"
  }'
```

### Create a Payment

The amount and currency are obtained from the course and cannot be supplied by the client.

```bash
curl -i -X POST \
  http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "enrollment_id": 1,
    "payment_method": "online",
    "description": "Course payment"
  }'
```

### Complete a Payment

```bash
curl -i -X PUT \
  http://localhost:8081/api/v1/payments/1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "succeeded",
    "transaction_id": "txn-20260919-001",
    "gateway_reference": "gateway-ref-001",
    "card_last_four": "1234",
    "description": "Payment verified"
  }'
```

## Profile Image Upload Security

The image upload subsystem includes:

- File-size restrictions
- Extension validation
- Content-type validation
- Generated random filenames
- Path traversal protection
- Non-executable upload storage
- Non-root container access
- Persistent Docker volume storage

Only the generated path is stored in PostgreSQL.

## Payment Security

The Payment API intentionally does not store complete card information.

Only the following limited reference information can be stored:

- Transaction identifier
- Gateway reference
- Last four card digits
- Payment result and timestamps

The payment amount is obtained from the course price on the server and is not accepted from the client.

## Database Migrations

Apply pending migrations:

```bash
./scripts/migrate-up.sh
```

Review applied migrations:

```bash
docker exec student_enrollment_postgres \
  psql -U student_app -d student_enrollment \
  -c "SELECT version, applied_at
      FROM schema_migrations
      ORDER BY version;"
```

Rollback operations are destructive and require explicit confirmation:

```bash
CONFIRM_MIGRATION_DOWN=yes ./scripts/migrate-down.sh
```

## Testing

Run all project checks:

```bash
./scripts/test.sh
```

The test script performs:

```bash
gofmt
go vet ./...
go test -v ./...
go test -race ./...
```

Individual commands can also be executed:

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
```

The test suite covers:

- Models and validation
- Memory repositories
- Service business rules
- HTTP handlers
- Integration routes
- Duplicate protection
- Course capacity
- Concurrent enrollment and payment operations
- Profile-image upload validation
- Payment lifecycle transitions
- Race detection

## Project Scripts

| Script | Description |
|---|---|
| `./scripts/start.sh` | Build and start the complete application |
| `./scripts/stop.sh` | Stop containers and preserve volumes |
| `./scripts/dev.sh` | Run PostgreSQL in Docker and API locally |
| `./scripts/build.sh` | Build the Docker API image |
| `./scripts/test.sh` | Run formatting, vet, tests, and Race Detector |
| `./scripts/logs.sh api` | Follow API logs |
| `./scripts/logs.sh postgres` | Follow PostgreSQL logs |
| `./scripts/migrate-up.sh` | Apply pending migrations |
| `./scripts/migrate-down.sh` | Roll back the database after confirmation |

## Useful Commands

Start services:

```bash
docker compose up -d
```

View status:

```bash
docker compose ps -a
```

Follow API logs:

```bash
./scripts/logs.sh api
```

Stop services while preserving data:

```bash
./scripts/stop.sh
```

Stop services and remove volumes:

```bash
docker compose down -v
```

> Warning: removing volumes permanently deletes database records and uploaded images.

## HTTP Status Codes

| Code | Meaning |
|---|---|
| `200 OK` | Successful retrieval or update |
| `201 Created` | Resource created |
| `204 No Content` | Resource deleted |
| `400 Bad Request` | Invalid JSON, ID, or validation |
| `404 Not Found` | Resource does not exist |
| `409 Conflict` | Duplicate or capacity conflict |
| `415 Unsupported Media Type` | Invalid image type |
| `500 Internal Server Error` | Unexpected server error |

## Logging and Request IDs

Every HTTP request is logged with information such as:

- Request ID
- HTTP method
- Request path
- Response status
- Response size
- Duration
- Remote address

A custom request ID can be supplied:

```bash
curl -i \
  http://localhost:8081/health \
  -H "X-Request-ID: anita-test-001"
```

The same request ID is returned in the response headers and included in structured logs.

## Data Persistence

Docker Compose uses two named volumes:

- `student_enrollment_postgres_data`
- `student_enrollment_uploads`

Running the following command preserves data:

```bash
docker compose down
```

Running this command removes persistent data:

```bash
docker compose down -v
```

## Current Version

```text
v1.0.0
```

## Author

**Anita Bakhshi**

Student Enrollment System — implemented with Go, PostgreSQL, Docker, layered architecture, automated migrations, secure uploads, enrollment capacity control, and payment lifecycle management.