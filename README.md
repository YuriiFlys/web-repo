# Web Repository
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)
![Angular](https://img.shields.io/badge/Angular-21-0F0F11?logo=angular&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4-0EA5E9?logo=tailwindcss&logoColor=white)
![RxJS](https://img.shields.io/badge/RxJS-7-B7178C?logo=reactivex&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)
![Digital Ocean](https://img.shields.io/badge/DigitalOcean-0080FF?style=for-the-badge&logo=digitalocean&logoColor=white)
![Firebase](https://img.shields.io/badge/Firebase-FFCA28?style=flat-square&logo=firebase&logoColor=white%22/%3E)

A monorepo that contains multiple university/lab projects focused on backend development, frontend development, design patterns, containerization, and cloud deployment.

## Links

- [Project management app](https://project-management-489517.web.app)
- [Backend API](https://pm-api-fy-hmvaq.ondigitalocean.app/)
- [Swagger UI](https://pm-api-fy-hmvaq.ondigitalocean.app/swagger/index.html)

## Projects Included

### 1. Lab 1 — Patterns
A Go project that demonstrates several classic design patterns:

- Builder
- Bridge
- Mediator

### 2. Lab 2 — Project Management REST API
A backend application built with:

- Go
- Gin
- GORM
- PostgreSQL
- JWT Authentication
- Swagger

This API provides project management functionality with authentication, database integration, and documented endpoints.

### 3. Lab 3 — Angular Frontend
A frontend application built with:

- Angular
- TypeScript
- Tailwind CSS
- RxJS

This frontend works with the REST API from Lab 2 and provides a user interface for authentication and project/task management.

---

## Repository Structure

```text
web-repo/
│
├── Lab 1 (Patterns)/
│   └── Go implementation of design patterns
│
├── Lab 2 (Rest Api)/
│   └── Project Management REST API (Go + Gin + GORM + PostgreSQL)
│
├── frontend/
│   └── Angular frontend application
│       └── pm-frontend/
│
└── .github/
    └── workflows/
````

---

## Tech Stack

### Backend

* Go
* Gin
* GORM
* PostgreSQL
* JWT
* Swagger
* Docker

### Frontend

* Angular
* TypeScript
* Tailwind CSS
* RxJS
* Playwright

### DevOps / Deployment

* Docker
* GitHub Actions
* Google Cloud Run
* Google Cloud SQL
* Google Artifact Registry
* Firebase Hosting

---

## Quick Start

## Lab 1 — Patterns

```bash
cd "Lab 1 (Patterns)"
go mod tidy
go run .
```

---

## Lab 2 — REST API

### 1. Navigate to the backend directory

```bash
cd "Lab 2 (Rest Api)"
```

### 2. Configure environment variables

Create a `.env` file in the backend root:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5435
DB_NAME=project_management
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

CORS_ORIGIN=http://localhost:4200
```

### 3. Run PostgreSQL with Docker

If you use the provided Docker setup for PostgreSQL:

```bash
docker compose -f docker/docker-compose.yml up -d
```

### 4. Start the backend

```bash
go mod tidy
go run .
```

### 5. API access

* Base URL: `http://localhost:8080/api`
* Swagger UI: `http://localhost:8080/swagger/index.html`

---

## Lab 3 — Angular Frontend

### 1. Navigate to the frontend directory

```bash
cd "frontend/pm-frontend"
```

### 2. Install dependencies

```bash
npm install
```

### 3. Start the frontend

```bash
npm start
```

### 4. Open in browser

```text
http://localhost:4200
```

The frontend communicates with the backend API and should be configured to use the backend base URL.

---

## Authentication

The backend uses JWT-based authentication.

Typical flow:

1. Register a new user
2. Login with credentials
3. Receive JWT token
4. Use the token for authorized requests

The frontend stores the token in `localStorage` and sends it in the `Authorization: Bearer <token>` header when required. This behavior is also reflected in the current repository notes. ([GitHub][1])

---

## Docker

The backend is containerized with Docker.

Typical usage:

```bash
docker build -t pm-api .
docker run -p 8080:8080 pm-api
```

For local database development, PostgreSQL can be started with Docker Compose from the backend project. The current repository already documents a Docker-based PostgreSQL setup under `Lab 2 (Rest Api)/docker`. ([GitHub][1])

---

## CI/CD

The backend uses GitHub Actions for CI/CD.

Pipeline overview:

1. Run automated Go tests
2. Build Docker image
3. Push image to Google Artifact Registry
4. Deploy the new version to Google Cloud Run

The frontend can be tested with Playwright and deployed separately to Firebase Hosting.

---

## Deployment Architecture

### Backend

* Hosted on **Google Cloud Run**
* Container image stored in **Google Artifact Registry**
* Database hosted on **Google Cloud SQL for PostgreSQL**

### Frontend

* Hosted on **Firebase Hosting**

This setup provides a clean separation between:

* frontend hosting
* backend API hosting
* managed database infrastructure

---

## Environment Configuration

### Backend

Example production/backend variables:

```env
APP_PORT=8080
DB_HOST=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME
DB_PORT=5432
DB_NAME=project_management
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSLMODE=disable
CORS_ORIGIN=https://your-frontend-domain.web.app
```

### Frontend

Example production API configuration:

```ts
apiUrl: 'https://your-cloud-run-service-url/api'
```

---

## Testing

### Backend

```bash
go test ./...
```

### Frontend

```bash
npm test
```

### End-to-End

Playwright is used for frontend E2E testing.

```bash
npx playwright test
```

---

## Goals of the Repository

This repository demonstrates practical skills in:

* Go backend development
* REST API design
* Angular frontend development
* PostgreSQL integration
* JWT authentication
* Docker containerization
* Cloud deployment on GCP
* CI/CD automation with GitHub Actions
* Frontend hosting with Firebase

---

## Future Improvements

Possible future enhancements:

* add role-based access control
* improve validation and error handling
* add refresh token support
* extend automated integration testing
* add frontend deployment workflow
* add custom domain configuration

---
