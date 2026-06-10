# GoBox Backend REST API

This is the backend REST API for **GoBox**, a lightweight, self-hosted cloud storage and directory synchronization platform. The backend is designed with Go, leveraging a clean package structure, strict multi-tenancy, and test-driven development (TDD).

---

## 1. Architectural Highlights

* **Upstream Authentication**: User login, registration, and token validation are offloaded to AWS Cognito. The backend trusts requests pre-authorized by AWS API Gateway or an ALB, extracting the unique user identifier (Cognito `sub`) from the incoming `X-User-Id` HTTP header.
* **Strict Multi-Tenancy**: Tenant data is logically isolated in the database by `user_id`. File objects in S3 are physically segregated using the prefix path `users/{user_id}/`.
* **Direct-to-S3 Uploads/Downloads**: To optimize bandwidth and server performance, the backend generates AWS S3 Presigned URLs, allowing clients to upload/download files directly to/from S3.
* **Storage Quota Enforcement**: Enforces a strict 2 GB storage capacity threshold per user, denying S3 upload URL generation if the file exceeds the tenant's remaining quota.

---

## 2. Directory Structure

```text
backend/
├── internal/
│   ├── files/         # File & Folder metadata, S3 URL integration
│   │   ├── file.go             # File core domain models
│   │   ├── repository.go       # File database query interfaces
│   │   ├── storage.go          # AWS S3 integration contracts
│   │   ├── service.go          # Core FileService business logic
│   │   ├── files_test.go       # Ginkgo BDD unit tests
│   │   └── files_suite_test.go # Test runner configuration
│   └── users/         # Quota limit and usage management
│       ├── user.go             # User domain struct & constraints
│       ├── repository.go       # User database query interfaces
│       ├── quota.go            # QuotaService business logic
│       ├── users_test.go       # Ginkgo BDD unit tests
│       └── users_suite_test.go # Test runner configuration
├── migrations/            # Database schema migrations
├── go.mod                 # Go dependencies
└── README.md              # This file
```

---

## 3. Test-Driven Development (TDD)

The codebase is built following a strict TDD approach. All business logic contracts and requirements are defined as failing test specifications before implementation code is written.

### Prerequisites
Make sure you have Go installed (v1.25+ recommended) and Ginkgo v2.

To install Ginkgo CLI (optional):
```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

### Running Tests

You can run the BDD test suites using the standard Go test command from the `backend` directory:

```bash
# Run all domain tests
go test -v ./internal/...

# Or run tests using Ginkgo CLI
ginkgo -r internal/
```

Currently, all tests in both the `users` and `files` domains are passing successfully!

---

## 4. Key Contracts & API Interfaces

### Users Domain
* **[User](internal/users/user.go)**: Domain model mapping a tenant's Cognito UUID and their `UsedQuota` in bytes.
* **[Repository](internal/users/repository.go)**: Handles persistence updates to the PostgreSQL database for quota tracking.
* **[QuotaService](internal/users/quota.go)**: Orchestrates quota validation and reservations.

### Files Domain
* **[File](internal/files/file.go)**: Represents metadata for files/directories.
* **[Repository](internal/files/repository.go)**: Interacts with database records for files.
* **[StorageService](internal/files/storage.go)**: Interfaces with AWS S3 SDK to generate presigned upload and download URLs.
