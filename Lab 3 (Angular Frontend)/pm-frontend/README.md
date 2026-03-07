# Project Management Frontend

![Angular](https://img.shields.io/badge/Angular-21.2-0F0F11?logo=angular&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4-0EA5E9?logo=tailwindcss&logoColor=white)
![RxJS](https://img.shields.io/badge/RxJS-7.8-B7178C?logo=reactivex&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)
![Playwright](https://img.shields.io/badge/Playwright-1.58.2-2EAD33)

Angular frontend for the Project Management application. The client provides authentication, project browsing, project detail views, task management, inline comments, and shared notification handling on top of the REST API.

## Overview

The application includes:

- Login, registration, and profile flows
- Protected routes backed by JWT authentication
- Project listing with search, filtering, pagination, and CRUD actions
- Project detail workspace with task organization, editing, and comments
- Global toast notifications for user feedback
- SSR-ready Angular setup with hydration enabled

## Technology Stack

- Angular 21.2 with standalone components
- TypeScript 5.9
- RxJS 7.8
- Tailwind CSS 4
- Angular SSR

## Prerequisites

- Node.js 20 or newer
- npm 11 or newer
- Running backend API at `http://localhost:8080/api`

The API base URL is currently hardcoded in the frontend services and feature code. If the backend runs on a different host or port, update those references before starting the application.

## Getting Started

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm start
```

Open the application at:

```text
http://localhost:4200
```

## Available Scripts

- `npm start` starts the Angular development server
- `npm run build` creates a production build in `dist/`
- `npm run watch` builds in watch mode using the development configuration
- `npm test` runs the frontend test suite
- `npm run serve:ssr:pm-frontend` serves the SSR build output

## End-to-End Testing

Playwright is configured at the parent workspace level in `Lab 3 (Angular Frontend)`, not inside `pm-frontend`.

Relevant files:

- `../playwright.config.ts`
- `../tests/app.spec.ts`
- `../package.json`

Available e2e commands from `Lab 3 (Angular Frontend)`:

- `npm run test:e2e`
- `npm run test:e2e:ui`
- `npm run test:e2e:headed`
- `npm run test:e2e:report`

The Playwright configuration starts the Angular app automatically with:

```bash
npm start --prefix "./pm-frontend" -- --host 127.0.0.1 --port 4200
```

The current Playwright suite covers authentication flows, project filtering and CRUD, task and comment interactions, profile access, and logout behavior using mocked API responses.

## Application Structure

```text
src/
|-- app/
|   |-- core/
|   |   |-- auth/      Authentication service and route guard
|   |   |-- http/      HTTP interceptors
|   |   |-- toast/     Global toast state and UI
|   |   `-- models.ts  Shared API models
|   |-- features/
|   |   |-- login/
|   |   |-- register/
|   |   |-- profile/
|   |   |-- projects/
|   |   `-- project-details/
|   |-- app.config.ts
|   `-- app.routes.ts
|-- main.ts
|-- main.server.ts
|-- server.ts
`-- styles.css
|../../tests/
|    --/app.spec.ts    Playwright tests
```

## Routing

Public routes:

- `/login`
- `/register`

Authenticated routes:

- `/projects`
- `/projects/:id`
- `/profile`

Unknown routes redirect to `/projects`.

## HTTP and State Management

- The `Auth` service stores the token and current user in `localStorage`
- `authInterceptor` attaches the bearer token to authenticated requests
- `errorInterceptor` handles `401 Unauthorized` responses by logging the user out and redirecting to the login page
- Feature areas manage local UI state while shared services handle cross-cutting concerns such as authentication and notifications

## Notes

- The frontend expects the REST API contract exposed by the backend in `Lab 2 (Rest Api)`
- SSR support is present, but browser-only persistence is guarded with platform checks
- Toasts are used for both success and error feedback across the application
- Browser end-to-end coverage is maintained through the parent Playwright workspace rather than through files local to `pm-frontend`
