# Web Repository
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)
![Angular](https://img.shields.io/badge/Angular-21-0F0F11?logo=angular&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4-0EA5E9?logo=tailwindcss&logoColor=white)
![RxJS](https://img.shields.io/badge/RxJS-7-B7178C?logo=reactivex&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)

Monorepo containing:
- **Lab 3 (Patters)**: Contains implementation of **Builder**, **Bridge** and **Mediator** patterns.
- **Lab 2 (Rest Api)**: Project Management REST API (Go + Gin + GORM + PostgreSQL).
- **Lab 3 (Angular Frontend)**: Angular 21 frontend with Tailwind CSS.

# Quick Start

## Lab 1: Patterns


### 1) Go to Lab 1 directory:
```
cd "Lab 1 (Patterns)"
```

### 2) Start the API

```
go mod tidy
go run .
```

## Lab 2: Project Management API

### 1)  Run PostgreSQL (Docker)

- Go to Lab 2 directory:
```
cd "Lab 2 (Rest Api)"
```

- Create `.env` in the `/docker` directory:
```env
POSTGRES_DB=project_management
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
DB_PORT=5435
```

- If your `docker-compose.yml` is inside `/docker`:

```bash
docker compose docker/docker-compose.yml up -d
```


### 2) Configure .env

- Create `.env` in the repository root:

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

### 3) Start the API

```
go mod tidy
go run .
```

- The API starts at `http://localhost:8080` with base path `/api`.

Optional: if you want a local Postgres instance, use Docker from `Lab 2 (Rest Api)/docker`.

## Lab 3: Project Management frontend app

### 1) Start the Frontend

```
cd "Lab 3 (Angular Frontend)\pm-frontend"
npm install
npm start
```

- Open `http://localhost:4200`.

The frontend targets `http://localhost:8080/api` by default.

## Repository Layout

```
Lab 2 (Rest Api)/          Go REST API + Swagger docs
Lab 3 (Angular Frontend)/  Angular frontend (pm-frontend)
```

## Notes

- API uses JWT auth. Register or login first to get a token.
- Frontend stores the token in `localStorage` and adds it as a Bearer token on requests.
