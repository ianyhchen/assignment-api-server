# Task API

A small RESTful task management API written in Go.

This project is an interview assignment focused on explicit HTTP behavior,
separation of responsibilities, concurrency-safe in-memory storage, graceful
server lifecycle management, and automated tests.

## Features

- List, create, update, and delete tasks
- Strict JSON decoding with a 1 MiB request-body limit
- Consistent JSON error responses
- Concurrency-safe in-memory storage
- Structured request logging with `log/slog`
- Panic recovery without exposing internal details
- HTTP server timeouts and graceful shutdown
- Multi-stage Docker build with a minimal non-root runtime image
- Unit, HTTP integration, and race-detector tests

## Requirements

- Go 1.22 or later
- Docker, when running the container or the race detector through Docker

The project is developed and containerized with Go 1.26.3. The module remains
compatible with Go 1.22.

## Quick Start

### Run locally

```bash
go run ./cmd/api
```

The server listens on port `8080` by default. Set `PORT` to use another
port:

```bash
PORT=9000 go run ./cmd/api
```

PowerShell equivalent:

```powershell
$env:PORT = "9000"
go run ./cmd/api
```

### Run with Docker

```bash
docker build -t assignment-api-server:local .
docker run --rm -p 8080:8080 assignment-api-server:local
```

Verify that the API is running:

```bash
curl -i http://localhost:8080/tasks
```

An empty store returns:

```json
[]
```

## API

Tasks have the following fields:

| Field | Type | Description |
| --- | --- | --- |
| `id` | unsigned integer | Server-generated task identifier |
| `name` | string | Task name |
| `status` | integer | `0` for incomplete, `1` for completed |

### List tasks

```http
GET /tasks
```

Successful response: `200 OK`

```json
[
  {
    "id": 1,
    "name": "Prepare interview assignment",
    "status": 0
  }
]
```

When no tasks exist, the response is an empty JSON array rather than `null`.

### Create a task

```http
POST /tasks
Content-Type: application/json
```

```json
{
  "name": "Prepare interview assignment",
  "status": 0
}
```

Successful response: `201 Created`

```json
{
  "id": 1,
  "name": "Prepare interview assignment",
  "status": 0
}
```

The response also includes a `Location: /tasks/{id}` header.

### Update a task

```http
PUT /tasks/{id}
Content-Type: application/json
```

```json
{
  "name": "Prepare interview assignment",
  "status": 1
}
```

Successful response: `200 OK`

```json
{
  "id": 1,
  "name": "Prepare interview assignment",
  "status": 1
}
```

`PUT` uses full-replacement semantics for client-editable fields. Both
`name` and `status` are required.

### Delete a task

```http
DELETE /tasks/{id}
```

Successful response: `204 No Content` with no response body.

## Example CRUD Flow

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Prepare assignment","status":0}' \
  http://localhost:8080/tasks

curl -i http://localhost:8080/tasks

curl -i \
  -X PUT \
  -H "Content-Type: application/json" \
  -d '{"name":"Submit assignment","status":1}' \
  http://localhost:8080/tasks/1

curl -i -X DELETE http://localhost:8080/tasks/1
```

## Validation and Errors

- `name` must be present and non-empty after trimming surrounding whitespace.
- `status` must be present and equal to `0` or `1`.
- A path `id` must be a positive unsigned integer.
- Unknown JSON fields, malformed JSON, and multiple JSON values are rejected.
- JSON request bodies are limited to 1 MiB.

Errors use a consistent JSON representation:

```json
{
  "error": "task not found"
}
```

| Condition | Status |
| --- | --- |
| Invalid path parameter or request body | `400 Bad Request` |
| Task does not exist | `404 Not Found` |
| Request context deadline exceeded | `504 Gateway Timeout` |
| Unexpected server failure | `500 Internal Server Error` |

Unexpected errors and recovered panic details are logged but are not exposed in
HTTP responses.

## Project Structure

```text
.
|-- cmd/
|   `-- api/
|       |-- main.go
|       `-- main_test.go
|-- internal/
|   |-- task/
|   |   |-- model.go
|   |   |-- errors.go
|   |   |-- store.go
|   |   |-- service.go
|   |   `-- service_test.go
|   |-- memory/
|   |   |-- task_store.go
|   |   `-- task_store_test.go
|   `-- httpapi/
|       |-- handler.go
|       |-- handler_test.go
|       |-- middleware.go
|       |-- middleware_test.go
|       |-- request.go
|       |-- request_test.go
|       |-- response.go
|       |-- response_test.go
|       |-- router.go
|       `-- router_test.go
|-- Dockerfile
|-- .dockerignore
|-- go.mod
`-- README.md
```

The dependency direction points inward toward the task package:

```text
cmd/api --> httpapi --> task
    |
    `----> memory ----> task
```

- `task` contains the domain model, application errors, store contract, and
  service-level validation.
- `memory` implements the store with a map protected by `sync.RWMutex` and
  owns ID generation.
- `httpapi` translates HTTP requests and responses, registers routes, maps
  application errors, and provides logging and recovery middleware.
- `cmd/api` constructs the concrete dependencies and manages the HTTP server
  lifecycle.

Dependencies are passed through constructors rather than stored in mutable
global variables.

## Server Behavior

The server uses the following defaults:

| Setting | Value |
| --- | --- |
| Listen port | `8080` |
| Read-header timeout | 5 seconds |
| Read timeout | 10 seconds |
| Write timeout | 10 seconds |
| Idle timeout | 60 seconds |
| Graceful-shutdown timeout | 10 seconds |

Request logs contain the HTTP method, path, response status, and duration.
Request and response bodies are not logged.

The process handles `SIGINT` and `SIGTERM`. During shutdown it stops
accepting new requests and waits up to 10 seconds for active requests.

To verify container shutdown behavior:

```bash
docker run -d --name assignment-api-server -p 8080:8080 assignment-api-server:local
docker stop --time 15 assignment-api-server
docker logs assignment-api-server
docker rm assignment-api-server
```

## Storage Behavior

Tasks are stored in process memory. Access to the task map and ID sequence is
synchronized for concurrent requests.

Intentional limitations:

- Data is lost when the process stops.
- Data is not shared between application instances.
- IDs restart from the initial value after a restart.

The task service depends on a small store interface rather than the memory
implementation. A future SQLite store can therefore be introduced in a
separate package without changing the domain service or HTTP handlers.

## Testing

Run formatting, static analysis, and tests:

```bash
gofmt -d .
go vet ./...
go test ./...
```

Run tests repeatedly:

```bash
go test -count=10 ./...
```

Run the race detector through the Go 1.26.3 Docker image:

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.26.3 go test -race ./...
```

Tests cover:

- Store CRUD behavior, ID generation, deterministic ordering, and concurrency
- Service validation, normalization, and error propagation
- JSON decoding, response encoding, handlers, and error mapping
- Request logging and panic recovery
- Full router CRUD flow
- Application assembly, server configuration, and routing behavior

## Design Decisions

### Standard library HTTP stack

Go's `http.ServeMux` supports method-aware patterns and path parameters and
is sufficient for the required routes. Avoiding a web framework keeps the
dependency surface small and makes HTTP behavior explicit.

### Explicit layers

HTTP transport, application rules, and storage are separated so they can be
tested and replaced independently. The project intentionally avoids a
dependency-injection framework, generic repository, or ORM because the current
scope does not justify those abstractions.

### In-memory storage

A synchronized map keeps the implementation aligned with the assignment while
still demonstrating correct concurrent access. Persistence would be the main
reason to replace it with SQLite or another database.

### Minimal container

The Docker build produces a statically linked Linux binary and copies it into
a `scratch` image. The final container does not contain the Go compiler,
source code, or a shell and runs under a numeric non-root user.

## License

This project is provided as an interview assignment.
