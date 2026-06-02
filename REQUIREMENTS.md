# Product Requirements Document (PRD) - GoBox

This document outlines the Minimum Viable Product (MVP) business and functional requirements for **GoBox**, a self-hosted cloud storage and directory synchronization platform. The goal of this project is to build a lightweight, highly performant Dropbox clone while mastering the Go programming language, React frontend development, and AWS cloud architecture.

---

## 1. System Architecture Overview

The system consists of three primary components interacting with AWS Cloud Infrastructure:
1. **Go REST API (Backend):** Manages user sessions, authenticates requests, stores file metadata, and generates secure AWS S3 Presigned URLs.
2. **React Web Dashboard (Frontend):** A web-based file explorer interface allowing manual management via browsers.
3. **Go Sync Client (Desktop):** A background application that watches a localized filesystem directory and bi-directionally synchronizes changes with the cloud.

```mermaid
graph TD
    %% Components
    Web[React Web Dashboard]
    Desktop[Go Desktop Sync Client]
    Backend[Go REST API Backend]
    
    %% AWS Infrastructure
    RDS[(Amazon RDS PostgreSQL)]
    S3[(Amazon S3 Object Storage)]

    %% Connections
    Web -->|HTTPS / JWT Auth| Backend
    Web -.->|Direct Upload/Download via Presigned URL| S3
    
    Desktop -->|HTTPS / JWT Auth| Backend
    Desktop -.->|Background Sync via Presigned URL| S3
    
    Backend -->|Read/Write Metadata| RDS
    Backend -->|Generate Upload URLs| S3
