# Project Management REST API

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-00B386)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-UI-85EA2D?logo=swagger&logoColor=black)

A production-style REST API built with **Go + Gin + GORM** for managing **Projects**, **Tasks**, and **Comments**.
Includes **pagination**, **sorting**, **filtering**, nested routes, and **Swagger UI** powered by **swaggo**.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Data Model](#data-model)
- [Quick Start](#quick-start)
  - [1) Run PostgreSQL (Docker)](#1-run-postgresql-docker)
  - [2) Configure .env](#2-configure-env)
  - [3) Run API](#3-run-api)
- [Swagger](#swagger)
- [Authentication](#authentication)
- [API Overview](#api-overview)
  - [Auth](#auth)
  - [Projects](#projects)
  - [Tasks](#tasks)
  - [Comments](#comments)
  - [Users](#users)
- [Query Parameters](#query-parameters)
  - [Pagination](#pagination)
  - [Sorting](#sorting)
  - [Filtering](#filtering)
  - [Includes](#includes)
- [Examples (cURL)](#examples-curl)
- [HTTP Status Codes](#http-status-codes)
- [Project Structure](#project-structure)

---

## Features

- CRUD for **Projects**, **Tasks**, **Comments**
- Relations:
  - Project **has many** Tasks
  - Task **has many** Comments
- Nested endpoints:
  - /projects/{projectId}/tasks
  - /tasks/{taskId}/comments
- Pagination on all list endpoints (page, pageSize)
- Sorting on list endpoints (sort)
- Filtering:
  - Projects: status, q
  - Tasks: projectId, status, ssigneeId, dueFrom, dueTo
  - Comments: 	askId, uthor
- Swagger docs + Swagger UI
- JWT auth with protected routes

---
## Tech Stack

- **Go**
- **Gin** (HTTP framework)
- **GORM** (ORM)
- **PostgreSQL**
- **swaggo/swag** + **gin-swagger** (OpenAPI/Swagger docs + UI)

---

## Data Model

```text
Project (1) ──── (N) Task (1) ──── (N) Comment
````

* **Project**: title, description, status (`active|archived`)
* **Task**: projectId, title, status (`todo|in_progress|done`), assigneeId, dueDate
* **Comment**: taskId, author, text

---

## Quick Start

### 1) Run PostgreSQL (Docker)

Create `.env` in the `/docker` directory:
```env
POSTGRES_DB=project_management
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
DB_PORT=5435
```

If your `docker-compose.yml` is inside `/docker`:

```bash
docker compose docker/docker-compose.yml up -d
```


### 2) Configure .env

Create `.env` in the repository root:

```env
APP_PORT=8080

# API connection params
DB_HOST=example
DB_PORT=5435
DB_NAME=project_management
DB_USER=example
DB_PASSWORD=example
DB_SSLMODE=disable
```

### 3) Run API

```bash
go mod tidy
go run .
```

Server will start at:

* `http://localhost:8080` (default)
* or `http://localhost:<APP_PORT>`

---

## Swagger

Swagger UI:

* `http://localhost:<APP_PORT>/swagger/index.html`

Regenerate Swagger docs:

```bash
swag init -g main.go
```

Generated output:

* `docs/docs.go`
* `docs/swagger.json`
* `docs/swagger.yaml`

---

## Authentication

All API endpoints under `/api` are protected by JWT **except**:

- `POST /auth/register`
- `POST /auth/login`

Use the token from login/register responses and send it as a bearer token:

```
Authorization: Bearer <token>
```

You can validate a token with `GET /auth/me`.

---

## API Overview

Base path: `/api`

### Auth

* `POST /auth/register`
* `POST /auth/login`
* `GET  /auth/me` (protected)

### Projects

* `GET    /projects`
  Query: `page, pageSize, sort, status, q, include=tasks`
* `POST   /projects`
* `GET    /projects/{id}`
  Query: `include=tasks`
* `PUT    /projects/{id}`
* `DELETE /projects/{id}`
* `GET    /projects/{projectId}/tasks`
  Query: `page, pageSize, sort, status, assigneeId`
* `POST   /projects/{projectId}/tasks`

### Tasks

* `GET    /tasks`
  Query: `page, pageSize, sort, projectId, status, assigneeId, dueFrom, dueTo, include=comments`
* `POST   /tasks`
* `GET    /tasks/{id}`
  Query: `include=comments`
* `PUT    /tasks/{id}`
  Note: send `{"dueDate": null}` to clear due date
* `DELETE /tasks/{id}`
* `GET    /tasks/{taskId}/comments`
  Query: `page, pageSize, sort, author`
* `POST   /tasks/{taskId}/comments`

### Comments

* `GET    /comments`
  Query: `page, pageSize, sort, taskId, author`
* `POST   /comments`
* `GET    /comments/{id}`
* `PUT    /comments/{id}`
* `DELETE /comments/{id}`

### Users

* `GET /users` (protected)

---

## Query Parameters

### Pagination

Supported by all list endpoints:

* `page` (default: `1`)
* `pageSize` (default: `20`, max: `100`)

Example:

```bash
curl "http://localhost:8080/api/projects?page=2&pageSize=10"
```

### Sorting

Use `sort` with a comma-separated list:

* `createdAt` = ascending
* `-createdAt` = descending

Example:

```bash
curl "http://localhost:8080/api/tasks?sort=title,-createdAt"
```

### Filtering

**Projects**

* `status=active|archived`
* `q=<search>` (search in title/description)

**Tasks**

* `projectId=<id>`
* `status=todo|in_progress|done`
* `assigneeId=<userId>`
* `dueFrom=YYYY-MM-DD`
* `dueTo=YYYY-MM-DD`

**Comments**

* `taskId=<id>`
* `author=<name>`

### Includes

Optional eager-load:

* Projects: `include=tasks`
* Tasks: `include=comments`

Example:

```bash
curl "http://localhost:8080/api/projects/1?include=tasks"
```

---

## Examples (cURL)

### Register

```bash
curl -i -X POST "http://localhost:8080/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Demo User","email":"demo@example.com","password":"secret123"}'
```

### Login

```bash
curl -i -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"secret123"}'
```

### Create a Project (authenticated)

```bash
curl -i -X POST "http://localhost:8080/api/projects" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Demo Project","description":"for testing","status":"active"}'
```

### List Projects (filter + search + sort + paging)

```bash
curl -i "http://localhost:8080/api/projects?status=active&q=demo&sort=-createdAt&page=1&pageSize=5" \
  -H "Authorization: Bearer <token>"
```

### Create a Task under Project (nested)

```bash
curl -i -X POST "http://localhost:8080/api/projects/1/tasks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Setup auth","status":"todo","assigneeId":1}'
```

### List Tasks (filters)

```bash
curl -i "http://localhost:8080/api/tasks?projectId=1&status=todo&assigneeId=1&sort=-createdAt" \
  -H "Authorization: Bearer <token>"
```

### Create Comment under Task

```bash
curl -i -X POST "http://localhost:8080/api/tasks/1/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"author":"Mentor","text":"Looks good"}'
```

### Clear Task dueDate

```bash
curl -i -X PUT "http://localhost:8080/api/tasks/1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"dueDate": null}'
```

---

## HTTP Status Codes

* `200 OK` — successful GET/PUT
* `201 Created` — successful POST
* `204 No Content` — successful DELETE
* `400 Bad Request` — invalid JSON / validation errors
* `404 Not Found` — entity not found

Error response shape:

```json
{ "code": "BAD_REQUEST", "message": "..." }
```

---

## Project Structure

```text
.
├── main.go
├── go.mod
├── internal
│   ├── db          # DB connection + migrations
│   ├── model       # GORM models
│   ├── handler     # Gin handlers/controllers
│   └── httpx       # shared helpers (errors, query parsing)
├── docs            # generated swagger docs
└── docker
    └── docker-compose.yml
```

---

