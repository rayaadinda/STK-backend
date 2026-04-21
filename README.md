# STK Backend

Backend service for menu tree CRUD API, built with Go, Gin, GORM, and PostgreSQL.

## Prerequisites

- Go 1.25+
- PostgreSQL 16+ (if running without Docker)
- Docker + Docker Compose (optional, for containerized run)

## Setup Instructions

1. Move to backend folder.

```bash
cd backend
```

2. Copy environment template.

```bash
cp .env.example .env
```

3. Install dependencies.

```bash
go mod download
```

4. Make sure PostgreSQL is running and matches variables in `.env`.

## Run in Development Mode

From backend folder:

```bash
go run ./cmd/server
```

Or with Makefile:

```bash
make run
```

The API will run on `http://localhost:8080` by default.

## Run in Production Mode

1. Set production environment (for example in `.env`):

```env
APP_ENV=production
```

2. Build binary:

```bash
go build -o bin/stk-backend ./cmd/server
```

3. Run binary:

```bash
./bin/stk-backend
```

## Run with Docker

This backend repository includes local Docker setup:

- `db` (PostgreSQL)
- `backend` (this service)

From backend folder:

```bash
cp .env.example .env
docker compose up -d --build
```

Check logs:

```bash
docker compose logs -f backend
```

Stop containers:

```bash
docker compose down
```

## API Documentation

- Swagger UI: `http://localhost:8080/api/docs`
- OpenAPI YAML: `http://localhost:8080/openapi.yaml`
- OpenAPI source file: `docs/openapi.yaml`

## Technology Choices and Architecture Decisions

- Framework: Gin for fast HTTP routing and middleware.
- ORM: GORM for PostgreSQL access and transaction handling.
- Database: PostgreSQL with self-referencing menu hierarchy.
- API style: RESTful endpoints under `/api`.
- Layered architecture:
  - `internal/http/handlers` for transport and request validation
  - `internal/menu/service.go` for business logic
  - `internal/menu/repository.go` for persistence operations
- Runtime migrations: executed on startup through `internal/database/migrate.go`.
- Menu isolation by module scope (`module_key`) so each frontend route can have separate menu tree data.
- Swagger integration: Swaggo UI mounted on `/api/docs` and configured to use `/openapi.yaml`.
