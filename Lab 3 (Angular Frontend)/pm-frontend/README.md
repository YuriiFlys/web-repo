
# Project Management Frontend
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)
![Angular](https://img.shields.io/badge/Angular-21-0F0F11?logo=angular&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4-0EA5E9?logo=tailwindcss&logoColor=white)
![RxJS](https://img.shields.io/badge/RxJS-7-B7178C?logo=reactivex&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)

Angular 21 + Tailwind CSS frontend for a simple project management app. It targets a REST API at `http://localhost:8080/api` and provides authentication, project listings, a kanban-style task board, and inline CRUD flows.

## Highlights

- Projects list with search, filtering, pagination, and create/delete.
- Project details board with task drag-and-drop, inline edit, comments, and delete.
- Authentication with token persistence and automatic logout on 401.
- Global toast notifications for success/error feedback.
- SSR-ready setup with hydration enabled.

## Tech Stack

- Angular 21 (standalone components)
- RxJS 7
- Tailwind CSS 4
- Angular SSR with hydration

## Quick Start

1. Install dependencies.
```
npm install
```

2. Start the dev server.
```
npm start
```

3. Open the app.
```
http://localhost:4200
```

The API base URL is hardcoded to `http://localhost:8080/api` in the feature components and `Auth` service. If your backend runs elsewhere, update those constants.

## Scripts

- `npm start` runs `ng serve`
- `npm run build` builds production output to `dist/`
- `npm run watch` builds in watch mode
- `npm test` runs unit tests
- `npm run serve:ssr:pm-frontend` serves the SSR build

## App Structure

```
src/
  app/
    core/
      auth/             Auth service and guard
      http/             HTTP interceptors (auth + 401 handling)
      toast/            Global toast service + component
      models.ts         Shared API types
    features/
      login/
      register/
      profile/
      projects/
      project-details/
    app.routes.ts       Route definitions
    app.config.ts       Providers (router, http, hydration)
  styles.css            Tailwind entrypoint
```

## Routing

- `/login` and `/register` are public.
- `/projects`, `/projects/:id`, and `/profile` require auth (guarded).
- Unknown routes redirect to `/projects`.

## Shared State

- `Auth` service (`providedIn: 'root'`) stores auth state and user data in `localStorage`.
- `Toasts` service (`providedIn: 'root'`) manages global notifications.
- Feature components keep their own UI state in local `BehaviorSubject`s.

## HTTP Behavior

- `authInterceptor` adds `Authorization: Bearer <token>` for authenticated calls.
- `errorInterceptor` logs the user out on 401 and redirects to `/login`.
- Feature components handle API errors with `catchError`, set local error messages, and emit toasts.

## UI Behavior Notes

- Projects list supports `q` and `status` filters plus pagination.
- Tasks can be dragged between columns and are persisted via `PUT`.
- Task details open in a modal with inline editing and comments.
- Delete actions prompt for confirmation.
