# Marketplace Go

A learning project to build a marketplace backend from scratch while practicing software architecture: modular layout, layered design, documented decisions (ADRs) and end-to-end tests against a real database.

## Stack

* Go 1.25+
* [Gin](https://github.com/gin-gonic/gin) (HTTP)
* PostgreSQL
* [pgx](https://github.com/jackc/pgx) / pgxpool (driver, see [ADR-001](docs/decisions/ADR-001.md))
* [Goose](https://github.com/pressly/goose) (migrations)
* JWT (auth)
* [Testcontainers](https://golang.testcontainers.org) (e2e tests)
* [Air](https://github.com/air-verse/air) (optional, hot reload)

## Prerequisites

* [Go](https://go.dev/dl/) 1.25 or newer
* [Docker](https://www.docker.com) (Docker Desktop or the daemon running). It is used for the local database and for the e2e tests.
* [Air](https://github.com/air-verse/air) (optional): `go install github.com/air-verse/air@latest`

## How to run

### 1. Environment variables

Copy the [template](.env.template) to a `.env` file and fill it in:

```bash
cp .env.template .env
```

| Variable | Description | Example |
|---|---|---|
| `ADDR` | Address the server listens on | `:8080` |
| `POSTGRES_HOST` | Database host | `127.0.0.1` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_DB` | Database name | `marketplace` |
| `POSTGRES_USER` | Database user | `username` |
| `POSTGRES_PASSWORD` | Database password | `strong_password` |
| `GOOSE_DRIVER` | Goose driver (only for the goose CLI) | `postgres` |
| `GOOSE_DBSTRING` | Goose connection string (only for the goose CLI) | `postgres://username:strong_password@127.0.0.1:5432/marketplace?sslmode=disable` |
| `GOOSE_MIGRATION_DIR` | Migrations folder (only for the goose CLI) | `./migrations` |
| `JWT_SECRET` | Secret used to sign tokens | any long random string |

`POSTGRES_SSL` is optional and defaults to `disable`.

### 2. Start PostgreSQL

The easiest way is the included [compose.yml](compose.yml), which creates a Postgres 17 database with the same credentials as the template:

```bash
docker compose up -d
```

If you change the credentials in `compose.yml`, keep your `.env` in sync.

### 3. Apply migrations

Nothing to do: the migrations in [migrations/](migrations) are embedded in the binary and applied automatically every time the server starts.

If you want to manage them by hand, use the goose CLI (it reads the `GOOSE_*` variables from your environment):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export $(grep -v '^#' .env | xargs)

goose status                    # see applied migrations
goose create add_something sql  # new migration file
goose up                        # apply pending migrations
goose down                      # roll back the last one
```

### 4. Start the server

**Using the codebase**

```bash
go run ./cmd/api
```

**Using Air (hot reload, recommended for development)**

```bash
air
```

**Building the binary**

```bash
go build -o marketplace ./cmd/api
./marketplace
```

Check that it is up:

```bash
curl http://localhost:8080/healthz
```

## Project structure

```text
.
├── cmd/api/            # Entry point: wiring, routes, graceful shutdown
├── config/             # Env loading (.env) and helpers
├── internal/           # Business modules (not importable from outside)
│   ├── product/        # Products CRUD
│   ├── users/          # Users
│   ├── security/       # Register / login
│   └── token/          # JWT creation and auth middleware
├── pkg/                # Reusable code
│   ├── database/       # pgx pool and migration runner
│   └── httpx/          # JSON responses and request binding/validation
├── migrations/         # Goose SQL migrations (embedded in the binary)
├── test/e2e/           # End-to-end tests with Testcontainers
├── docs/decisions/     # Architecture Decision Records (ADRs)
├── compose.yml         # Local PostgreSQL
└── .air.toml           # Hot reload config
```

## Architecture

The project is a **modular monolith**: each feature lives in its own package under `internal/` and exposes a single `Register` function that wires its dependencies and mounts its routes, so `main.go` only composes modules.

Inside each module the code is layered:

```text
handler (HTTP)  ->  service (business rules)  ->  repo (PostgreSQL)
```

* **Handler**: parses and validates the request, calls the service and writes the response with `httpx`. It knows nothing about SQL.
* **Service**: holds the business rules and depends on a repository interface, which makes it easy to test and swap.
* **Repository**: the only layer that talks to PostgreSQL through pgx.

Other decisions worth knowing:

* **Errors are wrapped in the service** and raw pgx errors are never exposed to the client.
* **Modules talk through small interfaces** instead of importing each other's internals (for example `security` receives a credentials finder and a registrar implemented by `users`).
* **Startup is bounded by timeouts** (connect, ping, migrations) and the server shuts down gracefully on `SIGINT`/`SIGTERM`.
* **Decisions are documented** as ADRs in [docs/decisions](docs/decisions), e.g. [ADR-001: pgx over database/sql](docs/decisions/ADR-001.md).

## API endpoints

All routes are mounted under `/api`, except the health check.

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/healthz` | No | Health check |
| POST | `/api/auth/register` | No | Register a new user |
| POST | `/api/auth/login` | No | Log in and get a JWT |
| GET | `/api/products` | No | List products |
| GET | `/api/products/:id` | No | Get a product by id |
| POST | `/api/products` | No | Create a product |
| PUT | `/api/products/:id` | No | Replace a product's editable fields (`state` is optional) |
| DELETE | `/api/products/:id` | No | Deactivate a product (soft delete, state `I`) |
| POST | `/api/users` | Bearer JWT | Create a user |
| GET | `/api/users/:id` | Bearer JWT | Get a user by id |
| PUT | `/api/users/:id` | Bearer JWT | Update a user |
| DELETE | `/api/users/:id` | Bearer JWT | Delete a user |

Protected routes expect the header `Authorization: Bearer <token>`.

Product body (create and update):

```json
{
  "name": "Keyboard",
  "price": 49.9,
  "stock": 10,
  "attributes": { "color": "black" },
  "state": "A"
}
```

`state` is only accepted on update and must be `A` (active) or `I` (inactive).

## Tests

The e2e tests in [test/e2e](test/e2e) start a throwaway PostgreSQL container with Testcontainers, apply the real migrations and exercise the HTTP handlers. Docker must be running.

```bash
go test ./...          # everything
go test ./test/e2e -v  # only e2e, verbose
```

## Migrations

Files live in [migrations](migrations) and follow goose naming (`00001_name.sql`) with `-- +goose Up` and `-- +goose Down` sections. They are embedded in the binary and applied on startup (see [How to run](#3-apply-migrations)).

Create a new one:

```bash
goose create add_something sql
```

Never edit a migration that was already applied; add a new one instead.

## Roadmap

- [x] Month 1: product CRUD, repository/service layers, error wrapping, e2e tests, ADR-001
- [x] Modules for users and security (register / login with JWT)
- [x] Product state (active / inactive)
- [ ] Protect product write routes with the auth middleware
- [ ] Test coverage for users and security
- [ ] Next milestones: _to be defined_

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
