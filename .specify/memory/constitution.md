<!--
SYNC IMPACT REPORT
==================
- Version Change: 1.1.0 -> 1.2.0
- List of modified principles:
  * None
- Added sections:
  * Core Principle: VII. Phased Implementation with Human Gates (NON-NEGOTIABLE)
- Removed sections:
  * None
- Templates requiring updates:
  * ✅ .specify/templates/plan-template.md (Updated Constitution Check with Phased Implementation gate)
  * ✅ .specify/templates/tasks-template.md (Updated within-story guidelines and implementation strategy to mandate backend -> frontend -> desktop sequence with human reviews)
- Follow-up TODOs:
  * None
-->

# GoBox Constitution

## Core Principles

### I. Strict Multi-Tenant Data Isolation (NON-NEGOTIABLE)
All data access, querying, and storage mechanisms must guarantee absolute tenant-level isolation. Files and folders stored on AWS S3 must be segregated under the `users/{user_id}/` key prefix. All database records (e.g., file metadata) must be filtered and queried explicitly using the unique Cognito-provided user identifier, which is extracted by the Go REST API from the `X-User-Id` HTTP header. Cross-tenant data leaks are treated as critical security failures.

### II. Direct-to-Cloud S3 Transfers
To ensure optimal server performance and backend bandwidth efficiency, files must be uploaded and downloaded directly between client applications (React Web and Go Sync Client) and AWS S3. The Go REST API backend acts solely as a metadata manager and URL generator, issuing secure AWS S3 Presigned URLs for direct transfers. Large file transfers must support AWS S3 Multipart Uploads and local resumption to handle connection drops gracefully.

### III. Event-Driven Background Synchronization
The Go Sync Client must run unobtrusively in the background, continuously monitoring the designated local sandbox directory using filesystem event listeners (`fsnotify`). It must track local modifications incrementally by computing SHA-256 file hashes and comparing them against a lightweight local SQLite database cache, minimizing network usage and API traffic.

### IV. Non-Destructive Conflict Resolution
To prevent accidental data loss due to simultaneous local and cloud changes, the Go Sync Client must not overwrite files blindly. If a file is modified locally and in the cloud concurrently, the client must preserve the cloud version and rename the local file using a conflict-naming scheme (e.g., `[filename]_conflict_copy.[ext]`), leaving the user to resolve the difference manually.

### V. Strict TDD-First Development (NON-NEGOTIABLE)
Every component in the codebase—including the Go REST API backend, the React Web Dashboard frontend, and the Go Sync Client desktop application—must be developed following a strict Test-Driven Development (TDD) lifecycle. Failing unit and/or integration test specifications must be written and verified to fail before any implementation code is written.

### VI. End-to-End (E2E) Testing for Client Applications
To guarantee seamless integration and end-to-end functionality under realistic usage scenarios, both the React Web Dashboard frontend and the Go Sync Client desktop application must have comprehensive automated end-to-end (E2E) tests. These E2E tests must execute the full system stack (including mock/test backend and S3 endpoints where appropriate) to validate user flows and directory synchronization behavior before delivery.

### VII. Phased Implementation with Human Gates (NON-NEGOTIABLE)
Any feature or user story implementation must progress through a strict sequential order: starting with the backend REST API, followed by the React Web Dashboard frontend, and concluding with the Go Sync Client desktop application. Each phase must be fully implemented, tested, and pass a mandatory manual human review and approval gate before work on the next component phase can begin.

## Technical Constraints & Security Standards
- **Token-Based Authentication**: All API routes require JWT authorization offloaded to AWS Cognito, with user identity extracted from the `X-User-Id` HTTP header.
- **Storage Quota Enforcement**: The backend must rigidly enforce a 2 GB storage capacity quota per user, rejecting any S3 presigned upload URL requests that would cause a tenant to exceed their designated limits.
- **Resource Constraints**: The Go Sync Client must maintain a low resource profile, ensuring that idle CPU is negligible and RAM consumption stays strictly below 50MB.
- **Cloud Budgeting**: The application must run entirely within the AWS Free Tier limits, hosting the database on Amazon RDS (PostgreSQL) and computing resources on a single `t3.micro` or `t4g.micro` instance. S3 buckets must enforce secure policies with public block access enabled.
- **Local Cache**: The Go Sync Client must utilize SQLite as its local state repository for local hashing and sync history.

## Development Workflow & Quality Gates
- **Ginkgo BDD Suite**: Go backend tests must be implemented using Ginkgo and run via `go test -v ./internal/...` or the Ginkgo CLI.
- **Client Testing Frameworks**: Frontend and desktop app must have unit/integration tests running under standard test suites (e.g., Jest/Vitest for frontend), and automated end-to-end suites (e.g., Cypress/Playwright for Web, native/driver tests for Desktop) verifying the full integration flows.
- **Linting Verification**: All Go files must pass `gofmt` formatting and standard `golangci-lint` verification before merge. Frontend files must pass ESLint and TypeScript compilation.
- **Complexity Tracking Check**: Any architectural complexity or deviation from core principles must be documented and justified in the feature's `plan.md` under the Complexity Tracking section.

## Governance
- **Supreme Authority**: This Constitution is the final arbiter for development standards, architectural boundaries, and coding patterns across the GoBox codebase.
- **Amendments**: Amendments require updating this document, incrementing the constitution version, updating dependent templates, and appending a Sync Impact Report detailing the changes.
- **PR Compliance**: Every Pull Request review must check the code against these principles. Non-compliant code will be rejected unless an explicit variance is approved and logged.

**Version**: 1.2.0 | **Ratified**: 2026-06-18 | **Last Amended**: 2026-06-18
