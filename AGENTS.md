# AGENTS.md — Call Booking Project

## Project Overview

Call Booking service (monorepo) inspired by Cal.com. Users publish availability schedules, control visibility with a **public profile toggle**, **per-schedule links to fixed groups** (Family, Work, Friends), and book **30-minute** slots. JWT authentication, no external calendar integrations.

## Tech Stack

- **API Contract:** TypeSpec (generates OpenAPI 3.0) — core resource shapes; some read-only range endpoints exist in the Go API before they appear in TypeSpec.
- **Backend:** Go 1.22+, `chi` router, `pgx` database driver, JWT auth
- **Frontend:** Next.js (App Router), Mantine UI (`@mantine/core`, `@mantine/dates`, `@mantine/hooks`, `@mantine/notifications`, `@mantine/schedule`)
- **E2E:** Playwright (`web/e2e`, `web/playwright.config.ts`)
- **Database:** PostgreSQL 16
- **Infrastructure:** Docker, docker-compose, optional root `deploy.sh` + `nginx.conf` for reverse proxy in compose

## Project Structure

```
typespec/                 # TypeSpec API contract
  main.tsp                # Entry point
  models.tsp              # Data models (User, Schedule, Booking, Group, etc.)
  auth.tsp                # Auth endpoints (/api/auth/*)
  users.tsp               # User endpoints (/api/users/*)
  schedules.tsp           # Schedule endpoints (/api/my/schedules)
  groups.tsp              # Visibility group endpoints (/api/my/groups)
  bookings.tsp            # Booking endpoints (/api/my/bookings)

cmd/server/               # Go entry point
internal/
  api/                    # HTTP handlers & chi router
    router.go             # Main router with middleware
    handlers_auth.go      # Auth (register, login, me)
    handlers_users.go     # Users list, get, slots, slots-range, available-dates(-range), updateMe, availableUsers
    handlers_schedules.go # Schedule CRUD + group visibility on schedules
    handlers_groups.go    # Fixed groups — list, members add/remove
    handlers_bookings.go  # Bookings list, create, cancel
  auth/                   # JWT (jwt.go, middleware.go)
  db/                     # DB connection; Migrate() applies repo `migrations/*.up.sql` in lexicographic order
  models/                 # Go structs
migrations/               # SQL migrations (idempotent / IF NOT EXISTS)
  001_initial.up.sql
  002_add_is_public.up.sql
  002_add_slot_date.up.sql
  003_add_slot_start_time.up.sql
  003_schedule_visibility.up.sql
  (+ matching *.down.sql where present)

web/                      # Next.js frontend
  app/
    layout.tsx, globals.css
    (auth)/login, register
    (app)/layout.tsx      # Shell + auth routes
    (app)/page.tsx        # Redirects to /my/bookings
    (app)/my/bookings     # «Мои бронирования» — календарь Mantine Schedule
    (app)/my/schedule     # Расписания + вкладка «Настройки видимости» (профиль, группы, участники)
    (app)/users           # «Запись на встречу» — выбор пользователя, календарь слотов, модалка подтверждения
    (app)/users/[id]      # Упрощённый профиль и слоты по дате (legacy-страница)
    admin/page.tsx        # Служебная страница (по необходимости)
  components/             # AuthProvider, AppShell, BookingConfirmationModal, ThemeToggle, …
  lib/api.ts              # API client (fetch)
  e2e/                    # Playwright: fixtures, pages/, specs/
  playwright.config.ts
```

## Development Commands

### TypeSpec

```bash
cd typespec && npm install
cd typespec && npx tsp compile .
```

### Go Backend

```bash
go build ./cmd/server
go test ./... -v
go test ./internal/api/... -v
go test ./internal/auth/... -v
```

### Next.js Frontend

```bash
cd web && npm install
cd web && npm run dev
cd web && npm run build
```

### Playwright (E2E)

```bash
cd web && npx playwright install chromium   # once per machine
cd web && npm run test:e2e
cd web && npm run test:e2e:smoke
```

### Docker

```bash
docker compose up -d    # db (host port 5434→5432), api, web, nginx
docker compose down
```

## Architecture Rules

### API

- All endpoints under `/api/*` prefix
- Public routes: `/api/auth/*` (no JWT required)
- Protected routes: all other `/api/*` (JWT via `Authorization: Bearer <token>`)
- POST endpoints return `201 Created`
- DELETE endpoints return `204 No Content`
- PUT endpoints return `200 OK`
- Errors return `{ "error": "message" }` with appropriate status code

### Groups API (Fixed Groups)

Groups are created on registration. Users cannot create/delete groups — only manage members.

- `GET /api/my/groups` — List 3 fixed groups (Family, Work, Friends)
- `GET /api/my/groups/{id}/members` — List members
- `POST /api/my/groups/{id}/members` — Add member by email
- `DELETE /api/my/groups/{id}/members/{memberId}` — Remove member

### Available Users API

- `GET /api/my/available-users` — All registered users except current (for adding to groups). No visibility filtering.

### Users API

- `GET /api/users` — Users visible to the current user (public profiles or shared group membership with the owner)
- `GET /api/users/{id}` — Profile
- `GET /api/users/{id}/slots?date=` — Slots for one day (`YYYY-MM-DD`)
- `GET /api/users/{id}/slots-range?start=&end=` — Slots across an inclusive range (`YYYY-MM-DD`)
- `GET /api/users/{id}/available-dates?month=` — Dates with availability in a month (`month=YYYY-MM`)
- `GET /api/users/{id}/available-dates-range?start=&end=` — Same for an inclusive day range
- `PUT /api/users/me` — Update `isPublic`, `name` (also mounted at `/api/users/me` in router)

### Database Schema

- **users**: id, email, password_hash, name, **is_public**, created_at, updated_at
- **schedules**: id, user_id, type (recurring|one-time), day_of_week, date, start_time, end_time, is_blocked, …
- **schedule_visibility_groups**: many-to-many **schedule_id** ↔ **group_id** (empty = «общее» расписание; видимость на фронте зависит от `is_public` и групп)
- **visibility_groups**: id, owner_id, name, visibility_level (**family|work|friends**) — три фиксированные группы на пользователя
- **group_members**: id, group_id, member_id, added_by, added_at
- **bookings**: id, schedule_id, booker_id, owner_id, status, **slot_date**, **slot_start_time**, created_at, cancelled_at, cancelled_by (и пр.)

Migrations: `cmd/server` calls `db.Migrate(ctx, pool, "migrations")` — **lexicographic order** of `*.up.sql` in the `migrations/` directory (working directory must be repo root or path adjusted). No migration version table; SQL is idempotent where possible.

### Visibility Model

- **Public profile** (`is_public = true`): пользователь попадает в каталог для всех авторизованных; «общие» расписания (без привязки к группам) видны всем, расписания с группами — только участникам этих групп.
- **Private profile** (`is_public = false`): в каталоге видны только те, кто состоит в группах владельца; «общие» расписания без групп не показываются в каталоге.
- **Fixed groups**: «Семья», «Работа», «Друзья» — создаются при регистрации.

### Go Conventions

- Handlers use `pgxpool.Pool` directly (no repository layer — YAGNI)
- JWT middleware puts user ID in context
- `jsonResponse()` / `jsonError()` for responses
- Test helpers `ptrInt32()`, `strPtr()` where used

### Frontend Conventions

- Mantine only for UI (no Tailwind utility layout)
- `@mantine/schedule` for календари на `/users` и `/my/bookings`
- `web/lib/api.ts` for HTTP
- Client components (`"use client"`) where needed
- Russian UI copy
- `data-testid` on stable elements used by Playwright (см. `web/e2e`)

## Testing

- TDD: write tests before implementation (Go)
- Go API tests: `httptest`, real DB when available; `t.Skipf` if DB missing
- Test DB: `call_booking_test` on localhost (see project test docs / env)
- Run `go test ./... -v` before committing Go changes
- E2E: `cd web && npm run test:e2e` (requires dev server or `playwright.config` `webServer`)

## Git Workflow

- Commit after each plan completion when applicable
- Commit messages: `feat: …` or `fix: …`
- Do not commit: `node_modules/`, `.next/`, `tsp-output/`, `vendor/`, `.env`, `web/playwright-report/`, `web/test-results/`
