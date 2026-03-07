# Project Management REST API

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-1.11-00B386)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-UI-85EA2D?logo=swagger&logoColor=black)

REST API for a project management system built with Go, Gin, GORM, and PostgreSQL. The service supports authentication, project and task management, nested comment resources, Swagger documentation, and automated test coverage for both unit and integration scenarios.

## Overview

The API exposes endpoints for:

- User registration and login with JWT-based authentication
- Project management with filtering, sorting, pagination, and search
- Task management with project-level nesting and optional due dates
- Comment management with task-level nesting
- Swagger UI for interactive API exploration

Base API path: `/api`

## Technology Stack

- Go 1.26
- Gin
- GORM
- PostgreSQL
- Swagger via `swaggo/swag` and `gin-swagger`

## Getting Started

### 1. Configure PostgreSQL

Create `docker/.env`:

```env
POSTGRES_DB=project_management
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
DB_PORT=5435
```

Start the database:

```bash
docker compose -f docker/docker-compose.yml --env-file docker/.env up -d
```

### 2. Configure the API

Create `.env` in the project root:

```env
APP_PORT=8080
CORS_ORIGIN=http://localhost:4200

DB_HOST=localhost
DB_PORT=5435
DB_NAME=project_management
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
```

### 3. Run the service

```bash
go mod tidy
go run .
```

The API starts on `http://localhost:8080` by default.

## Swagger

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

To regenerate the Swagger assets:

```bash
swag init -g main.go
```

Generated files are written to `docs/`.

## Authentication

Public endpoints:

- `POST /api/auth/register`
- `POST /api/auth/login`

Protected endpoints require a bearer token:

```text
Authorization: Bearer <token>
```

The authenticated user can be retrieved with `GET /api/auth/me`.

## Resource Summary

### Projects

- `GET /api/projects`
- `POST /api/projects`
- `GET /api/projects/{id}`
- `PUT /api/projects/{id}`
- `DELETE /api/projects/{id}`
- `GET /api/projects/{projectId}/tasks`
- `POST /api/projects/{projectId}/tasks`

### Tasks

- `GET /api/tasks`
- `POST /api/tasks`
- `GET /api/tasks/{id}`
- `PUT /api/tasks/{id}`
- `DELETE /api/tasks/{id}`
- `GET /api/tasks/{taskId}/comments`
- `POST /api/tasks/{taskId}/comments`

### Comments

- `GET /api/comments`
- `POST /api/comments`
- `GET /api/comments/{id}`
- `PUT /api/comments/{id}`
- `DELETE /api/comments/{id}`

### Users

- `GET /api/users`

## Query Capabilities

List endpoints support pagination through:

- `page`
- `pageSize`

Sorting is available through:

- `sort=createdAt`
- `sort=-createdAt`

Supported filters include:

- Projects: `status`, `q`
- Tasks: `projectId`, `status`, `assigneeId`, `dueFrom`, `dueTo`
- Comments: `taskId`, `author`

Optional eager loading:

- Projects: `include=tasks`
- Tasks: `include=comments`

## Testing

Run the standard Go test suite:

```bash
go test ./...
```

Unit tests are present across the main backend layers, including:

- `internal/auth` for JWT behavior
- `internal/config` for environment loading
- `internal/db` for connection setup and migration wiring
- `internal/handler` for HTTP handlers and error paths
- `internal/httpx` for query parsing and error helpers
- `internal/middleware` for auth middleware
- `internal/service` for business logic

Run the PowerShell helper for unit and integration coverage:

```powershell
.\scripts\test-all.ps1
```

Use a custom environment file if needed:

```powershell
.\scripts\test-all.ps1 -EnvFile .env.test
```

Generate coverage outputs:

```powershell
.\scripts\test-all.ps1 -Coverage
```

Integration tests are opt-in and use a real PostgreSQL database:

```bash
go test -tags=integration ./tests/integration/...
```

The integration suite exercises repository behavior against PostgreSQL, including:

- auth repository persistence and lookup
- project, task, and comment CRUD flows
- filtering, sorting, pagination, and include behavior
- user listing
- database reset and migration setup for test runs

The integration suite truncates application tables before each test run. Do not point it at a database containing important data.

## Project Structure

```text
.
|-- main.go
|-- go.mod
|-- docker/
|   `-- docker-compose.yml
|-- docs/
|   |-- swagger.json
|   |-- swagger.yaml
|   `-- docs.go
|-- internal/
|   |-- auth/
|   |-- config/
|   |-- db/
|   |-- handler/
|   |-- httpx/
|   |-- middleware/
|   |-- model/
|   |-- repository/
|   `-- service/
|-- scripts/
|   `-- test-all.ps1
`-- tests/
    `-- integration/
```
