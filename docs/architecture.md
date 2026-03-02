# Architecture Guide

Structural rules for how the Cove backend is organized and how layers interact.

---

## Project Layout

```
backend/          — Go module (github.com/gosusnp/cove/backend)
  cmd/mcp/        — standalone MCP server entry point
  internal/
    db/           — database connection and migrations/
    handlers/     — HTTP layer; helpers.go provides jsonOK, jsonError, jsonResponse, pathID
    service/      — business logic and ValidationError
    store/        — SQL queries, sentinel errors (ErrNotFound, ErrDuplicate), domain types in types.go
    middleware/   — APIKey and OAuth auth, request logging
    mcp/          — MCP tool registration grouped by resource
    testdb/       — shared test infrastructure (testcontainers + pgtestdb)
  main.go         — HTTP server entry point
frontend/         — Preact + Vite + Tailwind v4
  src/
    __mocks__/    — jsdom-compatible mocks for Radix packages used in tests
    components/
      ui/         — gold standard primitives (Button, Dialog, Switch, NavigationMenu, NavigationMenuBrand, TopBar, Avatar)
    hooks/        — signal-based hooks (useDialog, …)
    lib/          — shared utilities (cn)
    pages/        — route-level page components
docs/             — architecture and style guides
```

---

## Layer Hierarchy

```
Handlers  →  Services  →  Stores  →  Database
```

This is a strict, one-direction dependency chain.

- **DO** inject dependencies downward only: handlers depend on services, services depend on stores.
- **DON'T** let a store call a service, or a handler query the database directly.
- **DON'T** skip layers — a handler must not instantiate a store or call SQL.
- **DON'T** let services share state or call each other. Coordination belongs at the handler or a dedicated service.

---

## Constructors & Dependency Injection

Every layer is wired via constructors. No globals, no `init()` wiring.

```go
// Correct wiring order in main:
database := db.Open(dsn)
exStore   := store.NewExerciseStore(database)
exSvc     := service.NewExerciseService(exStore)
exHandler := handlers.NewExerciseHandler(exSvc)
```

- **DO** write a `New[Type](deps)` constructor for every handler, service, and store.
- **DO** store injected dependencies as unexported struct fields: `svc`, `db`, `store`.
- **DON'T** use package-level variables to share dependencies.
- **DON'T** pass `*sql.DB` to a handler — only to stores (and services that manage transactions).

---

## Handlers

Handlers own HTTP: routing, decoding requests, encoding responses. Nothing else.

```go
type ExerciseHandler struct {
    svc *service.ExerciseService
}

func NewExerciseHandler(s *service.ExerciseService) *ExerciseHandler {
    return &ExerciseHandler{svc: s}
}

func (h *ExerciseHandler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /exercises", h.list)
    mux.HandleFunc("POST /exercises", h.create)
    mux.HandleFunc("GET /exercises/{id}", h.get)
    mux.HandleFunc("PUT /exercises/{id}", h.update)
    mux.HandleFunc("DELETE /exercises/{id}", h.delete)
}
```

- **DO** register all routes in a `RegisterRoutes(*http.ServeMux)` method.
- **DO** use Go 1.22 method-path syntax in `HandleFunc`: `"GET /exercises"`, `"POST /exercises/{id}"`.
- **DO** use `jsonOK`, `jsonError`, and `jsonResponse` from `handlers/helpers.go` for all responses.
- **DO** use `pathID(r, "id")` from `handlers/helpers.go` to parse path parameters.
- **DON'T** perform business logic or validation in a handler — delegate to the service.
- **DON'T** write inline `json.Marshal` or `w.Write` calls — use the helpers.

### Request Decoding

```go
var req exerciseRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    jsonError(w, "invalid request body", http.StatusBadRequest)
    return
}
```

- **DO** define request DTOs as unexported structs within the handler file.
- **DO** return `400 Bad Request` immediately on decode failure.
- **DON'T** reuse domain types (from `store`) as request DTOs.

### HTTP Status Codes

| Situation | Status |
|---|---|
| Successful read or update | `200 OK` |
| Successful create | `201 Created` |
| Successful delete | `204 No Content` |
| Invalid request body / validation failure | `400 Bad Request` |
| Missing or invalid auth token | `401 Unauthorized` |
| Valid token but access denied | `403 Forbidden` |
| Resource not found | `404 Not Found` |
| Database or system failure | `500 Internal Server Error` |

---

## Services

Services own business logic: validation, normalization, orchestration, and transactions.

```go
type ExerciseService struct {
    store *store.ExerciseStore
}

func NewExerciseService(s *store.ExerciseStore) *ExerciseService {
    return &ExerciseService{store: s}
}

func (s *ExerciseService) Create(name string, progression *string) (*store.ExerciseDetail, error) {
    name = normalizeName(name)
    if name == "" {
        return nil, &ValidationError{Msg: "name is required"}
    }
    e, err := s.store.Create(name, progression)
    if errors.Is(err, store.ErrDuplicate) {
        return nil, &ValidationError{Msg: "exercise with this name already exists"}
    }
    return e, err
}
```

- **DO** normalize inputs (trim whitespace, lowercase) before validating.
- **DO** return `*ValidationError` for user-caused failures so handlers can map them to `400`.
- **DO** translate store sentinel errors at the service boundary so handlers never import the `store` package for error checks. Two acceptable forms:
  - Wrap into a typed error (e.g. `*ValidationError`) when the error maps to a user-facing message.
  - Alias the sentinel (`var ErrNotFound = store.ErrNotFound`) when no additional context is needed — `errors.Is` resolves through aliases correctly.
- **DON'T** let raw `store.ErrNotFound` or `store.ErrDuplicate` propagate to handlers without going through a service-level symbol first.
- **DON'T** write SQL in a service. Delegate to the store.

### Transactions

When a service operation spans multiple tables, it manages the transaction:

```go
type ProgramService struct {
    db    *sql.DB
    store *store.ProgramStore
}

func (s *ProgramService) CreateFull(...) (*store.Program, error) {
    tx, err := s.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() //nolint:errcheck
    // ... operations ...
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit: %w", err)
    }
    return result, nil
}
```

- **DO** call `defer tx.Rollback()` immediately after `tx.Begin()` — it is a no-op after `Commit`.
- **DON'T** manage transactions inside a store method — stores receive a `*sql.DB`, not a `*sql.Tx`.

---

## Stores

Stores own data access: raw SQL, scanning rows, and database error translation.

```go
type ExerciseStore struct {
    db *sql.DB
}

func NewExerciseStore(db *sql.DB) *ExerciseStore {
    return &ExerciseStore{db: db}
}
```

### Sentinel Errors

```go
var ErrNotFound  = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")
```

- **DO** declare `ErrNotFound` and `ErrDuplicate` in each store package that needs them.
- **DO** return `ErrNotFound` when `RowsAffected() == 0` on UPDATE or DELETE.
- **DO** detect PostgreSQL unique constraint violations (`pgErr.Code == "23505"`) and return `ErrDuplicate`.
- **DON'T** return raw `pgconn.PgError` values — translate them at the store boundary.

### Query Pattern

```go
func (s *ExerciseStore) List() ([]Exercise, error) {
    rows, err := s.db.Query(`SELECT id, name FROM exercises ORDER BY name`)
    if err != nil {
        return nil, fmt.Errorf("list exercises: %w", err)
    }
    defer rows.Close()

    exercises := []Exercise{}
    for rows.Next() {
        var e Exercise
        if err := rows.Scan(&e.ID, &e.Name); err != nil {
            return nil, fmt.Errorf("scan exercise: %w", err)
        }
        exercises = append(exercises, e)
    }
    return exercises, rows.Err()
}
```

- **DO** use positional parameters `$1`, `$2`, ... for all query arguments.
- **DO** wrap errors with `fmt.Errorf("operation name: %w", err)`.
- **DO** call `defer rows.Close()` immediately after a successful `Query` call.
- **DO** initialize result slices as `[]T{}` (not `nil`) so empty results serialize as `[]` not `null`.
- **DO** return `rows.Err()` as the final error from any row iteration.
- **DON'T** use string formatting to build queries — always use parameterized queries.

---

## Domain Types

All domain structs live in `backend/internal/store/types.go`.

```go
type Exercise struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type ExerciseDetail struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Progression *string `json:"progression,omitempty"`
}
```

- **DO** define list types (e.g., `Exercise`) and detail types (e.g., `ExerciseDetail`) separately when the detail includes optional or nested fields.
- **DO** use `*T` pointer fields with `omitempty` for nullable columns.
- **DO** use `int64` for primary and foreign keys on data/entity tables (exercises, programs, workouts, etc.).
- **DO** use `uuid.UUID` for primary keys on identity tables (users, sessions, API keys) — these follow the `UUID PRIMARY KEY` SQL convention.
- **DON'T** add computed or presentation fields to store types — those belong in a service or handler response struct.
- **DON'T** define domain types inside handler or service files.

---

## Error Handling

### Wrapping

```go
// DO
return nil, fmt.Errorf("list exercises: %w", err)

// DON'T
return nil, err
return nil, errors.New("list exercises: " + err.Error())
```

- **DO** wrap every error with `fmt.Errorf("context: %w", err)` before returning up the call stack.
- **DON'T** swallow errors or return bare `err` without context.

### Checking

```go
// Sentinel errors
if errors.Is(err, store.ErrNotFound) { ... }

// Typed errors
var ve *service.ValidationError
if errors.As(err, &ve) { ... }
```

- **DO** use `errors.Is` for sentinel errors and `errors.As` for typed errors.
- **DON'T** compare errors with `==` or check error strings.

---

## Middleware

Middleware takes `http.Handler` and returns `http.Handler`.

```go
func OAuth(us *store.UserStore, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
        if !ok || token == "" {
            // respond 401 and return
        }
        next.ServeHTTP(w, r)
    })
}
```

- **DO** write middleware as functions that accept dependencies and `next http.Handler`.
- **DON'T** use a global middleware registry or middleware structs with `ServeHTTP`.

---

## API Contract

### Routing

- **DO** use the Go 1.22 `METHOD /path` syntax in `HandleFunc`.
- **DO** follow REST conventions: `GET` lists, `POST` creates, `PUT` updates, `DELETE` removes.
- **DON'T** use path verbs: ~~`POST /exercises/create`~~, ~~`GET /exercises/list`~~.

### JSON Responses

All keys in `snake_case`. Success and error shapes are fixed:

```json
// Success
{ "id": 1, "name": "Squat" }

// Error
{ "error": "exercise not found" }
```

- **DO** use `jsonOK(w, data)` for 200 responses.
- **DO** use `jsonError(w, message, code)` for all error responses.
- **DO** use `jsonResponse(w, data, http.StatusCreated)` for 201 responses.
- **DON'T** write ad-hoc response shapes — always use the helpers.

---

## MCP Integration

The MCP server is an alternative interface to the same service layer used by HTTP handlers.

- **DO** group MCP tools by resource in `register[Resource]Tools` functions.
- **DO** call the same service methods that HTTP handlers call — no direct store access from MCP tools.
- **DON'T** duplicate business logic between the MCP layer and HTTP handlers.
