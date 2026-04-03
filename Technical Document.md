# **Technical Document**

### **Table of Contents**

- [1. Backend Technology Implementation](#1-backend-technology-implementation)
    - [1_1. System Architecture Overview](#1_1-system-architecture-overview)
    - [1_2. Technology Stack Selection](#1_2-technology-stack-selection)
    - [1_3. Core Function Implementation Plan](#1_3-core-function-implementation-plan)
    - [1_4. API Module Design](#1_4-api-module-design)
- [2. Frontend Technology Implementation](#2-frontend-technology-implementation)

---

## **1. Backend Technology Implementation**

- **Language**: Go 1.26.1

### **1_1. System Architecture Overview**
The backend is built using the **Gin** framework to construct RESTful APIs, operating on underlying storage directly through the **go-git** library. Since it involves large file processing and external tool integration (SonarQube), the system adopts an asynchronous task queue to handle time-consuming operations.

### **1_2. Technology Stack Selection**

<center>

| **Component** | **Recommended Technology** | **Rationale** |
|:---:|:---:|---|
| Web Framework | Gin | Most mature ecosystem, rich middleware, excellent performance and easy to maintain |
| Authentication | JWT (golang-jwt) | Used for stateless session management as required in documentation |
| Database (RDBMS) | PostgreSQL | Store user, project metadata and file attributes (such as MIME, size) |
| Cache | Redis | Store chunked upload status, JWT blacklist and frequently accessed project tree data |
| Git Library | go-git | Pure Go implementation, supports in-memory operations, no need to install Git binary on server |
| Object Storage | MinIO | S3 protocol compatible, used to store large file chunks and Office converted PDFs |
| Task Queue | Asynq | High-performance task queue based on Redis, handling code quality analysis and file compression |

</center>

### **1_3. Core Function Implementation Plan**

- **Chunked Upload and Resumable Transfer**
    - **Flow**: Frontend splits file into chunks and sends `upload_id` and `chunk_index`. Backend temporarily stores chunks in MinIO.
    - **Metadata Recording**: Maintain upload task table in PostgreSQL, recording completed chunk indices.
    - **Merge Trigger**: When the last chunk arrives, trigger an asynchronous task to merge chunks and calculate file Hash (e.g., SHA-256) to verify integrity.

- **Git Repository Management**
    - **Directory Structure**: Each Repository maps to a bare `.git` repository on server disk.
    - **Branches and Commits**: Use `go-git` to implement branch creation, switching and commit history queries.
    - **Directory Tree**: Backend recursively scans Git Tree objects, builds and caches directory tree in JSON structure for frontend rendering.

- **SonarQube Deep Integration**
    - **Configuration Management**: Store `project_key` and `auth_token` in Repository table.
    - **Analysis Trigger**: Call `sonar-scanner` command-line tool through `os/exec`, or trigger remote scan via API.
    - **Status Synchronization**: Implement an Endpoint that receives **Webhook**. When SonarQube completes analysis, send status confirmation. Backend updates Quality Gate status (Pass/Fail) in database.

### **1_4. API Module Design**

- **Authentication Middleware**

    All protected APIs must go through `AuthMiddleware`.
    - Parse `Authorization: Bearer <Token>`.
    - Verify JWT validity and user permissions (whether they are repository collaborators).

    - **Core Endpoint Overview**
    - `POST /api/v1/repos` : Create repository.
    - `GET /api/v1/repos/:id/tree` : Get left-side directory tree.
    - `POST /api/v1/files/upload` : Trigger chunked upload.
    - `POST /api/v1/analysis/trigger` : Manually trigger SonarQube analysis.

## **2. Frontend Technology Implementation**

---

<center>
 Document Revision History

| **Number** | **Reason for revision** | **Author** | **Revision Date** |
|:---:|---|:---:|:---:|
| 1 | Document Creation | Vito Lonius | 2026/04/03 |
| 2 | Improve backend technology implementation | Vito Lonius | 2026/04/03 |

</center>