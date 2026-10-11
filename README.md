# Kanban Board API (`kanban-board-be`)

Go REST API for a workspace-based Kanban board. Workspaces own boards, boards own columns and labels, columns own tasks. Auth is Supabase Auth (JWT verified via JWKS); data lives in PostgreSQL (Supabase); the API deploys to Vercel.

[![Backend CI](https://github.com/LDukeH/kanban-board-be/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/LDukeH/kanban-board-be/actions/workflows/backend-ci.yml)

- Go version: see `go.mod` (`go 1.27.1` at time of writing)

## Frontend

Companion board UI: `https://kanvas-board.vercel.app`.
## Features

Verified against `internal/feature/*`, `internal/server/routes.go`, and `migrations/`.

- **Auth:** Supabase JWT (`Authorization: Bearer <token>`) verified via JWKS. `POST /auth/sync` upserts the user row, `GET /auth/me` returns the current user (no password hash), `DELETE /auth/me` deletes the account. No local register/login endpoints.
- **Workspaces:** create, list (with member/board counts), read, update (name + description), delete. Creator becomes `owner` in the same transaction.
- **Memberships / roles:** `owner`, `admin`, `member` only. Owner-only role changes/removals, leave-workspace, ownership transfer.
- **Invitations:** `pending -> accepted | declined | cancelled | expired`. Owner/admin invite existing users; `409` if already a member or pending.
- **Boards / columns / tasks:** full CRUD; columns and tasks have dense `position` values reordered transactionally.
- **Task fields:** title, Tiptap `description` (`{ "type": "doc", ... }`), `priority` (`none|low|medium|high`), `due_date` (`YYYY-MM-DD`), `completed`, `position`, `version`.
- **Labels:** board-scoped name + hex colour; attach/detach on tasks.
- **Assignees:** assign/remove workspace members on tasks.
- **Comments:** list/create on a task, delete by ID (author or owner/admin). No edit endpoint.
- **Notifications (polling):** list, unread count, mark one/all as read. Server-created only (`task_assigned`, `workspace_invited`) in the triggering transaction.
- **User search:** `GET /users/search` (ILIKE, `LIMIT 10`).
- **Optimistic concurrency:** `version` on tasks (`020`) and columns (`021`); stale edit/delete yields `409 Conflict`.
- **Realtime broadcast:** Postgres triggers on `tasks`/`columns` call `public.broadcast_board_event()` which sends via `realtime.send(...)` to private `board:{board_id}` channels (migration `023`). Go only stamps `app.actor_id`.

## Engineering highlights

- **Optimistic concurrency with `409`s** — `internal/feature/task/repository.go`, `internal/feature/column/repository.go`: guarded writes (`WHERE id=$1 AND version=$2`), `resolveVersionMismatch` maps no-match to `409` vs `404`; moves take a `FOR UPDATE` lock and bump neighbours so stale clients refetch. Handlers require `version` in body or `?version=`.
- **Realtime via DB trigger** — `migrations/023_realtime_board_broadcast.sql`: row trigger payload (`id`, `type`, `board_id`, `entity_id`, `column_id`, `source_column_id`, `version`, `actor_id`, `occurred_at`) sent inside the caller transaction, so rollback emits nothing. Go side is only `internal/dbtx/actor.go`.
- **One-query authorization + indexes** — `internal/feature/membership/access.go` joins each resource to `workspace_memberships` in one statement; `internal/feature/authz/service.go` maps no-row to `403`. `migrations/018_perf_indexes.sql` adds missing FK-side indexes with per-endpoint comments.
- **Transactions for multi-step writes** — workspace create + owner (`internal/feature/workspace/service.go`), task assign + notification (`internal/feature/task/service.go`), invitation + notification (`internal/feature/invitation/service.go`).
- **Aggregated reads, typed errors** — `GetByColumnID` uses CTEs for labels + assignees in one query; `internal/apperrors` maps sentinels to HTTP codes.

## Tech stack

| Layer | Choice | Where |
| --- | --- | --- |
| Language | Go (`go 1.27.1` in `go.mod`) | `go.mod` |
| Router | `go-chi/chi/v5 v5.3.1`, `go-chi/cors v1.2.2` | `go.mod`, `internal/server/routes.go` |
| DB driver | `jackc/pgx/v5 v5.10.0` (+ `pgxpool`) | `go.mod`, `internal/db/postgres.go` |
| Auth verify | `golang-jwt/jwt/v5 v5.3.1`, `MicahParks/keyfunc/v3` (JWKS) | `internal/feature/auth/service.go` |
| IDs | `google/uuid v1.6.0` | `go.mod` |
| Env | `joho/godotenv v1.5.1` | `cmd/api/main.go` |
| Migrations | `goose v3.28.0` (installed in CI) | `.github/workflows/backend-ci.yml`, `migrations/` |
| Data/Auth/Realtime | PostgreSQL + Supabase Auth + Realtime Broadcast | `migrations/022_*`, `023_*` |
| Tests | stdlib `testing` + `pashagolub/pgxmock/v4` | `tests/` |
| Deploy | Vercel, region `sin1` | `vercel.json` |
| Scripts | PowerShell smoke + latency probes | `scripts/` |
| Hot reload | Air | `.air.toml` |


## API overview

All routes require `Authorization: Bearer <Supabase access token>`. Roles: any = any workspace member; OA = owner/admin; O = owner only.

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| POST | `/auth/sync` | auth | Upsert user from JWT |
| GET | `/auth/me` | auth | Current user |
| DELETE | `/auth/me` | auth | Delete own account |
| GET | `/users/{userID}` | any | User by ID |
| GET | `/users/search` | any | Email search, max 10 |
| POST | `/workspaces/` | any | Create workspace |
| GET | `/workspaces/` | any | List my workspaces |
| GET | `/workspaces/{workspaceID}` | any | Workspace by ID |
| PATCH | `/workspaces/{workspaceID}` | O | Rename / describe |
| DELETE | `/workspaces/{workspaceID}` | O | Delete workspace |
| POST | `/workspaces/{workspaceID}/transfer-ownership` | O | Transfer ownership |
| DELETE | `/workspaces/{workspaceID}/memberships` | self | Leave workspace |
| POST | `/workspaces/{workspaceID}/boards/` | any | Create board |
| GET | `/workspaces/{workspaceID}/boards/` | any | Boards in workspace |
| GET | `/workspaces/{workspaceID}/members/` | any | Member list |
| PATCH | `/workspaces/{workspaceID}/members/{membershipID}` | O | Change role |
| DELETE | `/workspaces/{workspaceID}/members/{membershipID}` | O | Remove member |
| POST | `/workspaces/{workspaceID}/invitations/` | OA | Invite user |
| GET | `/workspaces/{workspaceID}/invitations/` | OA | List invitations |
| POST | `/workspaces/{workspaceID}/invitations/{invitationID}/cancel` | OA | Cancel invite |
| POST | `/invitations/{invitationID}/accept` | invitee | Accept invite |
| POST | `/invitations/{invitationID}/decline` | invitee | Decline invite |
| GET | `/boards/{boardID}` | any | Board by ID |
| PATCH | `/boards/{boardID}` | any | Rename board |
| DELETE | `/boards/{boardID}` | any | Delete board |
| POST | `/boards/{boardID}/columns/` | OA | Create column |
| GET | `/boards/{boardID}/columns/` | any | Columns by position |
| POST | `/boards/{boardID}/labels/` | any | Create label |
| GET | `/boards/{boardID}/labels/` | any | Labels for board |
| PATCH | `/columns/{columnID}` | OA | Rename column (`name` + `version`) |
| PATCH | `/columns/{columnID}/position` | OA | Move column (`position` + `version`) |
| DELETE | `/columns/{columnID}?version=N` | OA | Delete column |
| POST | `/columns/{columnID}/tasks/` | any | Create task |
| GET | `/columns/{columnID}/tasks/` | any | Tasks + labels + assignees |
| GET | `/labels/{labelID}` | any | Label by ID |
| PATCH | `/labels/{labelID}` | any | Update label |
| DELETE | `/labels/{labelID}` | any | Delete label |
| PATCH | `/tasks/{taskID}` | any | Partial update + `version` |
| PATCH | `/tasks/{taskID}/position` | any | Move task + `version` |
| DELETE | `/tasks/{taskID}?version=N` | OA | Delete task |
| GET | `/tasks/{taskID}/labels/` | any | Labels on task |
| POST | `/tasks/{taskID}/labels/` | any | Attach label |
| DELETE | `/tasks/{taskID}/labels/{labelID}` | any | Detach label |
| GET | `/tasks/{taskID}/users/` | any | Assignees |
| POST | `/tasks/{taskID}/users/` | any | Assign member |
| DELETE | `/tasks/{taskID}/users/` | self/OA | Remove assignee |
| GET | `/tasks/{taskID}/comments/` | any | Comments oldest-first |
| POST | `/tasks/{taskID}/comments/` | any | Add comment |
| DELETE | `/comments/{commentID}` | author/OA | Delete comment |
| GET | `/notifications/` | own | List mine (`limit` 1-50) |
| GET | `/notifications/unread-count` | own | Unread count |
| POST | `/notifications/{notificationID}/read` | own | Mark one read |
| POST | `/notifications/read-all` | own | Mark all read |

### Status codes

`200` GET/PATCH, `201` POST, `204` DELETE/mark-read/sync, `400` validation, `401` bad token, `403` forbidden, `404` not found, `409` version conflict or duplicate, `500` other. No `422` in this codebase. Task/column PATCH returns `200` with the object; DELETE needs `?version=N`.

## Database

Migrations are `goose`, Up-only, in `migrations/` (`001`-`024`). CI pins `goose v3.28.0` and runs `validate`, `up`, `status` on `postgres:16`.

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
goose -dir ./migrations validate
goose -dir ./migrations up
goose -dir ./migrations status
```

Tables: `users`, `workspaces`, `workspace_memberships`, `workspace_invitations`, `boards`, `columns`, `tasks`, `task_assignees`, `labels`, `task_labels`, `comments`, `notifications`. Notable constraints: one `owner` per workspace, unique `(user_id, workspace_id)`, one pending invite per workspace+user, `priority` CHECK, `version` defaults.

```mermaid
erDiagram
  users ||--o{ workspace_memberships : has
  workspaces ||--o{ workspace_memberships : has
  workspaces ||--o{ boards : has
  workspaces ||--o{ workspace_invitations : has
  boards ||--o{ columns : has
  boards ||--o{ labels : has
  columns ||--o{ tasks : has
  tasks ||--o{ task_assignees : has
  tasks ||--o{ task_labels : has
  tasks ||--o{ comments : has
  users ||--o{ notifications : receives
```

Seeds: `db/seeds/seed_1.sql`, `db/seeds/seed_2.sql` (apply manually with `psql`; not run by CI).

Realtime note: `023` creates triggers always, but the `realtime.messages` policy only when `realtime`/`auth` schemas exist (CI skips it). On Supabase, enable Realtime so the policy applies.

## Getting started

### Prerequisites

- Go toolchain matching `go.mod`
- `goose v3.28.0`
- PostgreSQL or a Supabase project

### 1. Supabase setup

1. Create a project; note `https://your-project.supabase.co`.
2. Backend verifies JWTs at `<SUPABASE_URL>/auth/v1/.well-known/jwks.json`, issuer `<SUPABASE_URL>/auth/v1` (`cmd/api/main.go`).
3. Use the pooler URL for `DATABASE_URL`, a direct URL for `GOOSE_DBSTRING`.
4. Enable Realtime, then apply migrations.

### 2. Env

```bash
cp .env.example .env
```

| Name | Required? | Description | Example |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | Runtime pool URL | `postgresql://USER:PASSWORD@HOST:6543/postgres` |
| `SUPABASE_URL` | Yes | Project URL; server exits without it | `https://your-project.supabase.co` |
| `PORT` | No | Listen port, default `3000` | `3000` |
| `GOOSE_DRIVER` | For CLI | goose driver | `postgres` |
| `GOOSE_DBSTRING` | For CLI | goose direct URL | `postgresql://USER:PASSWORD@HOST:5432/postgres` |
| `GOOSE_MIGRATION_DIR` | For CLI | Migrations path | `./migrations` |
| `FRONTEND_URL` | No | CORS origin, default `http://localhost:5173` | `http://localhost:5173` |

### 3. Migrate and run

```bash
goose -dir ./migrations up
go run ./cmd/api
```

Hot reload: `air` (config `.air.toml`). CORS allows one origin with credentials.

## Testing

```bash
go test ./tests/...
go test -race -count=1 ./...
```

Realtime tests need an empty disposable Postgres: set `KANBAN_REALTIME_TEST_DATABASE_URL` (skipped without it). PowerShell E2E (needs a running server + DB):

```powershell
./scripts/smoke_api.ps1
./scripts/smoke_api.ps1 -BaseUrl https://kanban-board-be.vercel.app
```

## CI and deployment

Workflow `.github/workflows/backend-ci.yml` (push/PR to `main`): gofmt check, `go vet`, `go test -race` with coverage artifact, static build artifact, goose `validate/up/status` on `postgres:16`. Deploy: Vercel, `vercel.json` pins `sin1` (Singapore) next to Supabase. Pool is small (`MaxConns=2` in `internal/db/postgres.go`) for serverless.

## Roadmap / Not included

Verified absent: viewer role (only owner/admin/member), activity log, rate limiting, caching/Redis, Docker/Compose, pagination on task/comment lists, comment editing, server-side task filters.

