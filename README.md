# Task API

A small RESTful task management API written in Go.

This repository is an interview assignment. The API is intentionally kept
small so that the implementation can focus on clear HTTP behavior, separation
of responsibilities, concurrency-safe in-memory storage, and automated tests.

> Development status: API contract drafted; implementation has not started.

## Requirements

- Go 1.22 or later
- Four REST endpoints for listing, creating, updating, and deleting tasks
- Unit tests
- A Dockerfile for running the API in a container
- In-memory data storage

The project is developed and tested with Go 1.26.3.

## Planned Technical Approach

- Router: Go standard library `net/http` and `http.ServeMux`
- Storage: `map[uint64]Task` protected by `sync.RWMutex`
- Logging: Go standard library `log/slog`
- Tests: Go `testing` package and `net/http/httptest`
- Container build: multi-stage Docker build using Go 1.26.3

No third-party web framework or database is planned. Go's standard HTTP router
is sufficient for the four required routes, while an in-memory store keeps the
scope aligned with the assignment.

## Task Model

A task contains the following fields:

| Field | Type | Description |
| --- | --- | --- |
| `id` | unsigned integer | Server-generated task identifier |
| `name` | string | Task name |
| `status` | integer | `0` for incomplete, `1` for completed |

Example:

```json
{
  "id": 1,
  "name": "Prepare interview assignment",
  "status": 0
}
```

## API Contract

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

When no tasks exist, the API returns an empty JSON array:

```json
[]
```

### Create a task

```http
POST /tasks
Content-Type: application/json
```

Request body:

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

Both `name` and `status` are required.

### Update a task

```http
PUT /tasks/{id}
Content-Type: application/json
```

Request body:

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

`PUT` uses full-replacement semantics for the editable fields, so both `name`
and `status` are required.

### Delete a task

```http
DELETE /tasks/{id}
```

Successful response: `204 No Content`

The successful response has no body.

## Validation Rules

- `name` must be present and must not be empty after trimming surrounding
  whitespace.
- `status` must be present and must be either `0` or `1`.
- A path `id` must be a positive unsigned integer.
- Unknown JSON fields and malformed JSON request bodies will be rejected.

## Error Responses

Errors use a consistent JSON structure:

```json
{
  "error": "task not found"
}
```

Planned status codes:

| Condition | Status |
| --- | --- |
| Invalid path parameter or request body | `400 Bad Request` |
| Task does not exist | `404 Not Found` |
| Unexpected server failure | `500 Internal Server Error` |

Internal error details will not be exposed in API responses.

## Planned Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── task/
│   │   ├── model.go
│   │   ├── errors.go
│   │   ├── store.go
│   │   ├── service.go
│   │   └── service_test.go
│   ├── memory/
│   │   ├── task_store.go
│   │   └── task_store_test.go
│   └── httpapi/
│       ├── handler.go
│       ├── handler_test.go
│       ├── middleware.go
│       ├── response.go
│       └── router.go
├── .github/
│   └── workflows/
│       └── ci.yml
├── Dockerfile
├── .dockerignore
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

The intended dependency direction is inward toward the task package:

```text
                 +-> httpapi --+
cmd/api ---------+             +-> task
                 +-> memory ---+
```

- `task` contains the domain model, application errors, store contract, and
  service-level validation. It does not depend on HTTP or storage details.
- `memory` implements the task store using process memory and owns ID
  generation and concurrency control.
- `httpapi` translates HTTP requests and responses and maps domain errors to
  HTTP status codes.
- `cmd/api` constructs the concrete components and starts the server.

Dependencies will be passed through constructors rather than stored in global
variables. The application assembly is expected to follow this order:

```text
task store -> task service -> HTTP handler -> router -> HTTP server
```

This keeps each layer independently testable and allows infrastructure to be
replaced without changing the task rules.

## Component Boundaries

### Task domain and service

The `task` package owns concepts and rules that remain valid regardless of how
the API is exposed or where tasks are stored. It includes:

- `Task` and `Status` types
- validation of task names and status values
- application-level errors such as task-not-found
- the store interface required by the service
- create, list, update, and delete operations

The store interface accepts `context.Context` so a future database-backed
implementation can support cancellation and timeouts without changing the
service API.

### In-memory store

The `memory` package owns the `map[uint64]task.Task`, `sync.RWMutex`, and ID
sequence. It does not perform HTTP-specific validation or select HTTP status
codes.

### HTTP transport

The `httpapi` package owns:

- route registration
- path and JSON decoding
- request and response data transfer objects
- consistent JSON responses
- mapping application errors to HTTP status codes
- HTTP middleware

HTTP request types are kept separate from the domain model. This allows the
API to detect omitted fields, prevents clients from assigning server-owned
fields such as `id`, and avoids coupling the domain to JSON input behavior.

## Storage Behavior

Tasks will be stored in process memory. Access to the task map and ID sequence
will be synchronized so that concurrent HTTP requests do not cause data races.

The following limitations are intentional for this assignment:

- Data is lost when the process stops.
- Data is not shared between multiple application instances.
- No external database is required.

The storage behavior will be exposed through an interface so that another
implementation can be introduced without changing the HTTP layer.

## Extensibility

The application is designed around a small store interface rather than a
specific storage implementation. A future SQLite implementation could be
introduced as a separate package:

```text
internal/
└── sqlite/
    ├── task_store.go
    ├── task_store_test.go
    └── migrations.go
```

The application entry point would replace the memory store constructor with a
SQLite store constructor. The task service and HTTP handlers would keep the
same contracts.

Other changes have similarly bounded locations:

- New task rules belong in the task service.
- New HTTP representations belong in `httpapi` data transfer objects.
- Authentication or request tracing can be added as middleware.
- A CLI or another transport can call the same task service.

These extension points are intentionally limited. The project does not use a
generic repository, dependency injection framework, ORM, or additional storage
implementation because those would add complexity without being required by
the assignment.

## Planned Server Behavior

The server will include:

- configurable listen port with a documented default
- read, write, idle, and header timeouts
- structured request logging with `log/slog`
- panic recovery with a generic `500 Internal Server Error` response
- a limit on JSON request body size
- graceful shutdown with a bounded timeout

Request logs will contain method, path, response status, and duration. Request
bodies and internal error details will not be logged or returned by default.

## Testing Strategy

Tests are planned at several boundaries:

| Test level | Responsibility |
| --- | --- |
| Store tests | CRUD behavior, ID generation, ordering, and not-found errors |
| Service tests | Validation, normalization, and error propagation |
| Handler tests | JSON handling, status codes, headers, and error responses |
| Router flow test | Complete create, list, update, and delete sequence |
| Race detector | Concurrent safety of the in-memory store |

The store behavior will be expressed as reusable contract tests where
practical. A future SQLite store can then be checked against the same behavior
as the memory store.

The planned CI workflow will run formatting checks, `go vet`, unit tests, the
race detector, and a build on pushes and pull requests.

## Design Decisions and Trade-offs

### Standard library router

Go's `http.ServeMux` supports method-aware patterns and path parameters and is
sufficient for the four required endpoints. Avoiding a third-party framework
keeps the dependency surface small and makes HTTP behavior explicit.

### In-memory storage

The assignment explicitly permits in-memory storage. A synchronized map keeps
the implementation focused on the API contract, concurrency, and tests. The
trade-off is that data is process-local and non-persistent.

### Explicit layers without framework abstractions

Separating HTTP, application rules, and storage makes replacement and testing
possible. The project intentionally stops short of generic repositories,
dependency injection containers, or duplicated implementations. This provides
useful boundaries without introducing architecture that the current scope does
not need.

### Full update semantics

`PUT /tasks/{id}` replaces all client-editable fields. Both `name` and `status`
are therefore required. A future partial update would be exposed separately,
for example through `PATCH`, rather than making `PUT` behavior ambiguous.

## Development Plan

1. Define the task model, application errors, and store contract.
2. Implement and test the concurrency-safe in-memory store.
3. Implement and test service-level validation.
4. Implement HTTP response helpers, handlers, and route-level tests.
5. Add focused logging, recovery, and body-limit middleware.
6. Configure server timeouts and graceful shutdown.
7. Add the Docker build and usage instructions.
8. Add the GitHub Actions verification workflow.
9. Run formatting, tests, the race detector, static analysis, and a Docker
   smoke test.

## License

This project is provided as an interview assignment.
