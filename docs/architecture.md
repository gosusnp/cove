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
    domain/       — domain entities, logic, and hardened types
    store/        — SQL queries and database error translation
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
exStore   := store.NewExerciseStore() // Stateless
exSvc     := service.NewExerciseService(database, exStore)
exHandler := handlers.NewExerciseHandler(exSvc)
```

- **DO** write a `New[Type](deps)` constructor for every handler, service, and store.
- **DO** store injected dependencies as unexported struct fields: `svc`, `db`, `store`.
- **DON'T** use package-level variables to share dependencies.
- **DON'T** pass `*sql.DB` to a handler — only to services.

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
    db    *sql.DB
    store *store.ExerciseStore
}

func NewExerciseService(db *sql.DB, s *store.ExerciseStore) *ExerciseService {
    return &ExerciseService{db: db, store: s}
}

func (s *ExerciseService) Create(ctx context.Context, name string, progression *string) (*store.ExerciseDetail, error) {
    name = normalizeName(name)
    if name == "" {
        return nil, &ValidationError{Msg: "name is required"}
    }
    e, err := s.store.Create(ctx, s.db, name, progression)
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

When a service operation spans multiple stores, the service owns the transaction lifecycle and passes it down via the `q store.Querier` argument:

```go
type UserService struct {
    db    *sql.DB
    users *store.UserStore
    orgs  *store.OrgStore
}

func (s *UserService) GetOrCreate(ctx context.Context, email, googleSub string) (*store.User, bool, error) {
    tx, err := s.db.Begin()
    if err != nil {
        return nil, false, fmt.Errorf("begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    user, created, err := s.users.UpsertUser(ctx, tx, id, email, googleSub)
    if err != nil {
        return nil, false, err
    }
    if created {
        if err := s.orgs.CreateOrg(ctx, tx, orgID, email); err != nil {
            return nil, false, err
        }
    }

    return user, created, tx.Commit()
}
```

- **DO** hold `*sql.DB` on the service when it manages transactions.
- **DO** call `defer func() { _ = tx.Rollback() }()` immediately after `tx.Begin()` — it is a no-op after `Commit`.
- **DON'T** manage transactions inside a store method — transaction lifecycle belongs to the service.

---

## Stores

Stores own data access: raw SQL, scanning rows, and database error translation.

Stores are **stateless**. They do not hold a database connection or transaction. Instead, every method accepts a `context.Context` and a `Querier`.

```go
// store/base.go — shared by all stores
type Querier interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
```

Concrete stores are empty structs:

```go
type ExerciseStore struct{}

func NewExerciseStore() *ExerciseStore {
    return &ExerciseStore{}
}
```

- **DO** make stores stateless — they must not hold any internal state or connections.
- **DO** accept `ctx context.Context` and `q Querier` as the first two arguments for every store method.

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
func (s *ExerciseStore) List(ctx context.Context, q Querier) ([]Exercise, error) {
    rows, err := q.QueryContext(ctx, `SELECT id, name FROM exercises ORDER BY name`)
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

## Domain Types & Hardened IDs

All domain entities and identifiers live in `backend/internal/domain/`.

### Type-Safe Identifiers

To prevent logic errors (e.g., passing a `ProgramID` where an `ExerciseID` is required), we use a generic `ID[T]` wrapper with phantom types.

```go
// internal/domain/types.go

type ID[T any] struct {
    uuid.UUID
}

type UserID ID[struct{ userID struct{} }]
type OrgID  ID[struct{ orgID struct{} }]
```

- **DO** use hardened IDs for all new entities.
- **DO** define the phantom type using an unexported field in a struct: `struct{ name struct{} }`.
- **DO** use the `ID[T]` helper for `Scan`, `Value`, and `MarshalJSON` support.

### Entities

```go
type User struct {
    ID        UserID    `json:"id"`
    Email     Email     `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

- **DO** define list types (e.g., `Exercise`) and detail types (e.g., `ExerciseDetail`) separately when the detail includes optional or nested fields.
- **DO** use `*T` pointer fields with `omitempty` for nullable columns.
- **DO** use hardened `ID[T]` (wrapping `uuid.UUID`) for primary keys on all new tables.
- **DO** pay specific attention to identity tables (users, sessions, API keys) which **must** use these hardened UUIDs to follow the `UUID PRIMARY KEY` SQL convention.
- **DON'T** add computed or presentation fields to domain types — those belong in a service or handler response struct.
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
