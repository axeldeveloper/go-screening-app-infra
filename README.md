# my-tristack-app

A Go project scaffolded with [TriStack](https://tristack.space).

## Stack

- **Framework:** gin
- **ORM / DB layer:** sqlc
- **Database:** postgres
- **Migrations:** golang-migrate
- **Package manager:** go

## Getting Started

```bash
# 1. Install dependencies

go mod tidy

# 2. Generate query code (sqlc)

go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate

# 3. Run database migrations

go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database 'postgres://postgres:postgres@localhost:5432/my_tristack_app?sslmode=disable' up

# 4. Start the dev server

go run ./cmd/api
```

## Common commands

```bash
go test ./...                          # run all tests
go vet ./...                           # static analysis
gofmt -l .                             # check formatting
go build -o ./bin/my_tristack_app ./cmd/api  # compile the binary
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate  # regenerate query code
```

## API

- `GET /health` — liveness probe.
- `GET /api/v1/items` / `POST /api/v1/items` — example resource; replace with your own models, repositories, and handlers.

## Project layout

```
cmd/api/             entrypoint: config -> db -> repository -> service -> handler
internal/config/     environment configuration (PORT, APP_NAME)
internal/handler/    HTTP handlers (framework-specific)
internal/service/    business logic
internal/repository/ data access per ORM
internal/model/      domain models
internal/db/         connection + connection pool
Makefile             convenience targets (build, run, test, vet, fmt, generate, migrate-up/down)
```

## Environment

Defaults are baked in, so the app runs with no setup. Override them with environment variables:

- `PORT` — HTTP port (default `3000`)
- `APP_NAME` — application name
- `DATABASE_URL` — database DSN; see `env.example` for per-database values

For a local `.env`, load it into your shell before running (Go does not read `.env` on its own):

```bash
# macOS / Linux
set -a && . ./env.example && set +a

# Windows PowerShell
Get-Content ./env.example | Where-Object { $_ -match '=' -and $_ -notmatch '^\s*#' } | ForEach-Object { $n, $v = $_ -split '=', 2; Set-Item -Path "env:$n" -Value $v }
```
