# Architecture Guide

Structural rules for how the Cove backend is organized and how layers interact.

---

## Project Layout

```
android/          — Capacitor Android project (mobile builds)
backend/          — Go module (github.com/gosusnp/cove/backend)
  cmd/llm/        — standalone CLI for iterating on LLM prompt workflows without the HTTP server
  internal/
    db/           — database connection and migrations/
    handlers/     — HTTP layer; helpers.go provides jsonOK, jsonError, jsonResponse, pathID
    service/      — business logic and ValidationError
    domain/       — domain entities, logic, and hardened types
    store/        — SQL queries and database error translation
    middleware/   — APIKey and OAuth auth, request logging
    workers/      — background job orchestration; ports.go defines data-tier boundaries
    crypto/       — Encryptor interface, AES-256-GCM impl, EncryptedField[T] hardened type, SensitiveString
    llm/          — LLM Client interface and OpenAI-compatible implementation
    llm/prompts/  — shared prompt builders and embedded templates, grouped by domain (fitness, cooking)
    fdc/          — USDA FoodData Central API client (ingredient search and import)
    markdown/     — program-to-markdown rendering (used by MCP)
    httputil/     — shared HTTP helpers
    testutil/     — shared test infrastructure (testcontainers + pgtestdb)
  main.go         — HTTP server entry point
frontend/         — Preact + Vite + Tailwind v4
  capacitor.config.json — Capacitor configuration for mobile asset bundling
  src/
    __mocks__/    — jsdom-compatible mocks for Radix packages used in tests
    components/
      shared/     — cross-cutting app components (Nav, ActivityPicker, …)
      ui/         — gold standard primitives (Button, Dialog, Switch, NavigationMenu, NavigationMenuBrand, TopBar, Avatar)
    hooks/        — signal-based hooks (useDialog, …)
    lib/          — shared utilities (cn, apiFetch)
    pages/        — route-level page components
docs/             — architecture and style guides
```

---

## Layer Hierarchy

```
Handlers  →  Services  →  Stores  →  Database
Workers   →  Ports     →  (Services or remote HTTP)
```

This is a strict, one-direction dependency chain.

- **DO** inject dependencies downward only: handlers depend on services, services depend on stores.
- **DON'T** let a store call a service, or a handler query the database directly.
- **DON'T** skip layers — a handler must not instantiate a store or call SQL.
- **DON'T** let services share state or call each other. Coordination belongs at the handler or a dedicated service.
- **DO** allow a service to depend on multiple stores when it needs to coordinate across resources (e.g. `ProgramService` depends on both `ProgramStore` and `ExerciseStore`).
- **DON'T** let workers import `internal/service` or `internal/store` directly — all data access goes through a port.

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
- **DO** use `jsonOK`, `jsonError`, `jsonResponse`, and `internalError` from `handlers/helpers.go` for all responses.
- **DO** use `pathID[Type](r, "id")` from `handlers/helpers.go` to parse path parameters.
- **DO** map `service.ErrUnauthorized` → `401 Unauthorized` and `service.ErrNotFound` → `404 Not Found` in the handler's error helper.
- **DO** use `internalError(w, r, err)` for all unexpected server-side failures — it logs the error with method and path, then responds with `500 Internal Server Error`. Never use `jsonError(..., http.StatusInternalServerError)` directly.
- **DON'T** perform business logic or validation in a handler — delegate to the service.
- **DON'T** write inline `json.Marshal` or `w.Write` calls — use the helpers.
- **DON'T** pass `err.Error()` to `jsonError` — internal error details must not be sent to the client.

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
- **Exception — patch types:** `[Entity]Patch` structs that use `domain.Optional[T]` to carry partial-update semantics are defined in the service package and decoded directly in the handler. Duplicating them as handler DTOs adds no value. See `WorkoutSessionPatch` and `UserPreferencesPatch` for the established pattern.

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

func (s *ExerciseService) Create(ctx context.Context, name string, progression *string) (*domain.Exercise, error) {
    name = normalizeName(name)
    if name == "" {
        return nil, &ValidationError{Msg: "name is required"}
    }
    var ex *domain.Exercise
    err := withScopedTx(ctx, s.db, func(q store.Querier) error {
        var err error
        ex, err = s.store.Create(ctx, q, name, progression)
        if errors.Is(err, store.ErrDuplicate) {
            return &ValidationError{Msg: "exercise with this name already exists"}
        }
        return err
    })
    return ex, err
}
```

- **DO** normalize inputs (trim whitespace, lowercase) before validating.
- **DO** return `*ValidationError` for user-caused failures so handlers can map them to `400`.
- **DO** translate store sentinel errors at the service boundary so handlers never import the `store` package for error checks. Two acceptable forms:
  - Wrap into a typed error (e.g. `*ValidationError`) when the error maps to a user-facing message.
  - Alias the sentinel (`var ErrNotFound = store.ErrNotFound`) when no additional context is needed — `errors.Is` resolves through aliases correctly.
- **DON'T** let raw `store.ErrNotFound` or `store.ErrDuplicate` propagate to handlers without going through a service-level symbol first.
- **DON'T** write SQL in a service. Delegate to the store.

### Scoped Transactions (RLS-gated operations)

For operations on tenant-owned data, use the `withScopedTx` helper (internal to the `service` package). It:
1. Extracts the `Identity` from the context, returning `ErrUnauthorized` if absent.
2. Starts a transaction and wraps it in a `ScopedQuerier`.
3. Sets PostgreSQL session variables (`app.current_org_id`, `app.current_user_id`) used by RLS policies and bookkeeping triggers.

```go
func (s *ExerciseService) Create(ctx context.Context, name string, ...) (*domain.Exercise, error) {
    var ex *domain.Exercise
    err := withScopedTx(ctx, s.db, func(q store.Querier) error {
        var err error
        ex, err = s.store.Create(ctx, q, name, ...)
        return err
    })
    return ex, err
}
```

### Plain Transactions (non-RLS operations)

When a service operation spans multiple stores but does not require RLS, manage the transaction lifecycle directly:

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

### Defense-in-Depth

Even with RLS enabled, stores **must** include explicit `org_id` and `is_public` filters in their SQL. This makes the access intent clear and provides a second layer of isolation.

```go
func (s *ExerciseStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ExerciseID) (*domain.Exercise, error) {
    var e domain.Exercise
    err := q.QueryRowContext(ctx, `
        SELECT id, name, ... FROM exercises
        WHERE id = $1 AND (org_id = $2 OR is_public = true)
    `, id, orgID).Scan(...)
    // ...
}
```

- **DO** pass `orgID` explicitly to store methods that access tenant-owned data.
- **DO** filter with `WHERE org_id = $N` (or `org_id = $N OR is_public = true` for readable resources).
- **DON'T** rely solely on RLS policies for data isolation.

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

To prevent logic errors (e.g., passing a `ProgramID` where an `ExerciseID` is required), we use generic wrapper types with phantom types. There are two variants depending on the underlying database column type:

```go
// internal/domain/types.go

// ID[T] wraps uuid.UUID — use for identity tables (users, orgs, sessions, API keys).
type ID[T any] struct {
    uuid.UUID
}

type UserID ID[struct{ userID struct{} }]
type OrgID  ID[struct{ orgID struct{} }]

// IntID[T] wraps int64 — use for resource tables with BIGSERIAL primary keys.
type IntID[T any] int64

type ExerciseID IntID[struct{ exerciseID struct{} }]
type ProgramID  IntID[struct{ programID struct{} }]
```

Choose the wrapper based on the SQL column type:

| SQL column | Go wrapper | When to use |
|---|---|---|
| `UUID PRIMARY KEY` | `ID[T]` | Identity/auth tables: users, orgs, sessions, API keys |
| `BIGSERIAL PRIMARY KEY` | `IntID[T]` | Resource tables: exercises, programs, workout sessions |

- **DO** use hardened IDs for all new entities — never use raw `uuid.UUID`, `int64`, or `string` for primary keys.
- **DO** define the phantom type using an unexported field in a struct: `struct{ name struct{} }`.
- **DO** use `ID[T]` for `Scan`, `Value`, and `MarshalJSON` support on UUID columns.
- **DO** use `IntID[T]` for `Scan`, `Value`, and `MarshalJSON` support on `BIGSERIAL` columns.

### Entities

```go
type User struct {
    ID        UserID    `json:"id"`
    Email     Email     `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

- **DO** define full types (e.g., `Exercise`, `Program`) and lite types (e.g., `ExerciseLite`, `ProgramLite`) separately when a trimmed projection is needed — e.g. for list endpoints that don't need the full hierarchy.
- **DO** use `*T` pointer fields with `omitempty` for nullable columns.
- **DO** use hardened `ID[T]` or `IntID[T]` for primary keys on all new tables — see the Type-Safe Identifiers section for which to choose.
- **DO** use `ID[T]` (UUID) for identity tables (users, orgs, sessions, API keys) and `IntID[T]` (int64) for resource tables (exercises, programs, sessions).
- **DO** use `time.Time` for all timestamp columns (`created_at`, `updated_at`) — never `string`.
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

## Multi-Tenancy & RLS

Cove uses PostgreSQL Row Level Security to isolate tenant data.

1. **Required Columns**: All tenant-owned tables must have `org_id` and `created_by` columns marked `NOT NULL`.
2. **Bookkeeping Trigger**: Attach the `update_bookkeeping_columns` trigger to automate `updated_at`, `updated_by`, and default `org_id`/`created_by` from the session variables.
3. **Policies**:
    - `SELECT`: `USING (org_id = current_app_org_id() OR is_public = true)`
    - `INSERT`: `WITH CHECK (org_id = current_app_org_id())`
    - `UPDATE`: `USING (...) WITH CHECK (...)` — restricted to same org
    - `DELETE`: `USING (org_id = current_app_org_id())`
4. **Session Variables**: Set via `ScopedQuerier` inside `withScopedTx` before each query:
    - `app.current_org_id` — read by `current_app_org_id()`
    - `app.current_user_id` — read by `current_app_user_id()`

---

## Middleware

Middleware takes `http.Handler` and returns `http.Handler`.

```go
func OAuth(us *store.UserStore, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Cookie-first: browser sessions use an HttpOnly cookie.
        // Bearer fallback: MCP clients and API keys use the Authorization header.
        var token string
        if c, err := r.Cookie("cove_session"); err == nil {
            token = c.Value
        } else if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
            token = t
        }
        if token == "" {
            // respond 401 and return
        }
        next.ServeHTTP(w, r)
    })
}
```

- **DO** write middleware as functions that accept dependencies and `next http.Handler`.
- **DON'T** use a global middleware registry or middleware structs with `ServeHTTP`.
- **DO** check the `cove_session` cookie before the `Authorization` header — browser clients set the cookie; API/MCP clients set the header.

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

## Credential & Sensitive Data Hygiene

These rules apply everywhere in the codebase — backend, frontend, tests, and logs.

### Session tokens

- **DO** deliver session tokens as `HttpOnly; Secure; SameSite=Strict` cookies set by the server. The browser sends them automatically; JS never sees them.
- **DON'T** write session tokens, API keys, or any auth credential to `localStorage`, `sessionStorage`, a JS variable, a URL parameter, or a log line — not even in development.
- **DO** use `secureCookies: true` in production (`COVE_DEV == ""`). The flag is wired in `main.go`; tests pass `false` to allow HTTP.

### Frontend

- **DON'T** read or write auth state from `localStorage` or `sessionStorage`. Auth state is bootstrapped by `GET /api/users/me` on mount.
- **DON'T** expose tokens in the `AuthContext` or any component prop. The context carries only the user object, `logout`, and `updateUser`.
- **DO** use `apiFetch` from `src/lib/api.js` for all API calls — it adds `credentials: "include"` so the session cookie is sent automatically.

### Logging

- **DON'T** log request headers, cookie values, or any field that could contain a credential.
- **DON'T** include user-supplied free-text fields (session notes, exercise descriptions) in error messages or structured log output.
- **DO** log only opaque identifiers (IDs, status codes, durations) — never the values they reference.

---

## Sensitive Data at Rest

Some domain fields contain user-private data (e.g. session notes, perceived effort) that must be encrypted in the database. Cove uses AES-256-GCM via the `internal/crypto` package.

### Design Principles

- **Plaintext never leaves the handler scope in normal read flows.** The only way to read sensitive fields is via `UseSensitiveData` — a callback pattern that decrypts into a `*T`, calls the handler, then zeros the struct in place before returning.
- **String fields use `crypto.SensitiveString` (`[]byte`-backed).** Go string interning means `*string` fields cannot be reliably zeroed; `SensitiveString` avoids that risk. Zeroing the `*T` struct after the callback wipes the backing bytes.
- **The service injects the encryptor; it never decrypts for output.** This keeps the plaintext lifetime as short as possible. **Exception — read-modify-write:** a service `Patch` method may call `UseSensitiveData` internally to merge existing encrypted state with the incoming patch before re-encrypting. This is the only sanctioned case for decryption in the service layer.
- **The store is a pure byte pass-through.** It has no knowledge of encryption — it scans and writes raw `[]byte`.
- **User ID is bound into every ciphertext via GCM additional data.** `UseSensitiveData` passes `ws.UserID.UUID[:]` as AAD; decryption will fail if the stored user_id does not match, preventing row-swapping attacks.
- **Empty sensitive data is stored as NULL.** `[Entity]SensitiveData` must implement `IsEmpty() bool`; the service skips encryption and writes `NULL` when it returns true, avoiding spurious ciphertext for empty payloads.

### Naming Conventions

| Concept | Convention |
|---|---|
| DB column | `sensitive_data BYTEA` |
| Domain field | `sensitiveData` (unexported) |
| Sensitive payload struct | `[Entity]SensitiveData` (e.g. `SessionSensitiveData`) |
| Decrypt callback method | `UseSensitiveData(ctx, fn)` on the domain entity |
| Store params field | `SensitiveData [Entity]SensitiveData` |

### Layer Responsibilities

```
Handler/Worker → calls ws.UseSensitiveData(ctx, fn) — primary site where plaintext exists
Service        → encrypts on write; calls ws.SetEncryptor after store read; may decrypt
                 only in Patch (read-modify-write) to merge before re-encrypting
Store          → scans sensitive_data into ws.SensitiveDataScanner(); writes raw []byte
```

### Domain Entity Pattern

```go
// String fields use *crypto.SensitiveString instead of *string so their backing
// bytes can be explicitly zeroed after use.
type SessionSensitiveData struct {
    PerceivedEffort  *int                    `json:"perceived_effort,omitempty"`
    SessionNotes     *crypto.SensitiveString `json:"session_notes,omitempty"`
    // ...
}

// IsEmpty is the single source of truth for "no sensitive data set".
// The service uses it to write NULL to the DB instead of encrypting {}.
// Update this whenever a new field is added to the struct.
func (s SessionSensitiveData) IsEmpty() bool {
    return s.PerceivedEffort == nil && s.SessionNotes == nil // ...
}

type WorkoutSession struct {
    // ... public fields ...
    // sensitiveData contains a sync.Mutex; always use *WorkoutSession, never copy.
    sensitiveData crypto.EncryptedField[SessionSensitiveData] `json:"-"`
}

// UseSensitiveData is the only way to access sensitive fields. fn receives a
// *SessionSensitiveData; the struct is zeroed in place after fn returns,
// wiping SensitiveString backing bytes. UserID is passed as GCM AAD.
func (ws *WorkoutSession) UseSensitiveData(ctx context.Context, fn func(SessionSensitiveData) error) error {
    return ws.sensitiveData.Use(ctx, func(p *SessionSensitiveData) error {
        return fn(*p)
    }, ws.UserID.UUID[:])
}

// SetEncryptor is called by the service after a store read.
func (ws *WorkoutSession) SetEncryptor(enc crypto.Encryptor) {
    ws.sensitiveData.SetEncryptor(enc)
}
```

### Service Pattern

```go
func (s *WorkoutSessionService) Create(ctx context.Context, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
    sensitiveData, err := s.encryptSensitiveData(ctx, p)  // encrypt before store write
    if err != nil { return nil, err }

    var ws *domain.WorkoutSession
    err = withScopedTx(ctx, s.db, func(q store.Querier) error {
        ws, err = s.store.Create(ctx, q, p, sensitiveData)
        return err
    })
    ws.SetEncryptor(s.enc)  // attach encryptor; do NOT decrypt
    return ws, err
}
```

### Handler Pattern

The response DTO uses `*string` for JSON output. Convert `*SensitiveString` → `*string` inline inside the callback; the short-lived `string` copy is collected promptly by the GC.

```go
func (h *WorkoutSessionHandler) get(w http.ResponseWriter, r *http.Request) {
    ws, err := h.svc.Get(r.Context(), id)
    // ...

    var resp workoutSessionResponse
    // copy public fields from ws first ...
    if err := ws.UseSensitiveData(r.Context(), func(s domain.SessionSensitiveData) error {
        resp.PerceivedEffort  = s.PerceivedEffort
        resp.SessionNotes     = stringPtr(s.SessionNotes)  // *SensitiveString → *string
        // ... other sensitive fields ...
        return nil
    }); err != nil {
        jsonError(w, "internal error", http.StatusInternalServerError)
        return
    }
    // sensitive local s is zeroed; resp holds only short-lived *string copies
    jsonOK(w, resp)
}
```

### Rules

- **DO** define a `[Entity]SensitiveData` struct for each entity with sensitive fields.
- **DO** use `*crypto.SensitiveString` (not `*string`) for all string fields in `[Entity]SensitiveData` — `[]byte`-backed for reliable zeroing.
- **DO** implement `IsEmpty() bool` on every `[Entity]SensitiveData` struct, enumerating all fields. The service calls this to skip encryption and write `NULL` when no sensitive data is set.
- **DO** keep `sensitiveData` unexported on the domain entity — expose only `UseSensitiveData` and `SetEncryptor`.
- **DO** use `sensitive_data BYTEA` (nullable) as the DB column — `NULL` means no sensitive data was set.
- **DO** inject `crypto.Encryptor` into the service constructor; wire from `SESSION_ENCRYPTION_KEY` env var in `main.go`.
- **DO** call `SetEncryptor` on every `*[Entity]` returned by the store before handing it to the handler.
- **DON'T** call `UseSensitiveData` in the service except in `Patch` methods — the only sanctioned case is reading existing encrypted state to merge it with an incoming patch before re-encrypting. All other decryption belongs in the handler or worker.
- **DON'T** add sensitive fields directly to the domain entity struct — they must go through `[Entity]SensitiveData`.
- **DON'T** store the result of `UseSensitiveData` in a struct field or return value — use it inline only.
- **DON'T** assign `*SensitiveString` fields directly to response structs — convert via `stringPtr(s.Field)` inside the callback.

### Key Management

| Environment | Implementation |
|---|---|
| Local / CI | `AESEncryptor` from `SESSION_ENCRYPTION_KEY` env var (base64-encoded 32-byte key) |
| Production (k8s) | Swap `AESEncryptor` for a KMS/Vault implementation of `crypto.Encryptor` in `main.go` |

Generate a local key with:
```sh
python3 -c "import os, base64; print(base64.b64encode(os.urandom(32)).decode())"
```

#### Ciphertext Format

Every ciphertext produced by `AESEncryptor` is:

```
[1-byte version][12-byte nonce][AES-256-GCM ciphertext + 16-byte tag]
```

The version byte identifies which key was used to encrypt the payload. On decrypt, the version byte is read first and the matching key is selected from the key map. This enables lazy key rotation without a migration.

#### AAD Binding

The owning entity's `user_id` bytes are passed as GCM additional authenticated data (AAD) on every encrypt and decrypt call. The AAD is bound into the GCM authentication tag — decryption fails if the user_id does not match the one used during encryption. This prevents a ciphertext from one row being replayed against another user's row.

#### Key Versioning and Rotation

`NewAESEncryptor(currentVersion byte, keys map[byte]string)` accepts multiple versioned keys. New rows are always encrypted with the `currentVersion` key. Old rows are decrypted using whichever key version was stored in the ciphertext prefix.

**To rotate to a new key:**

1. Generate a new key and assign it the next version byte (e.g. `1`).
2. Update the env var / secret to include both keys and set `currentVersion` to the new version:
   ```go
   // main.go
   enc, err := crypto.NewAESEncryptor(1, map[byte]string{
       0: oldKey,
       1: newKey,
   })
   ```
3. Deploy. New writes use key `1`; existing rows with version `0` continue to decrypt correctly.
4. Re-encrypt existing rows at your own pace (lazy rotation on read, or a background job).
5. Once no rows with version `0` remain, remove the old key from the map and redeploy.

---

## Units & Measurement

### Core principle: units are part of the data

The `(value, unit)` pair stored on a record represents the author's intent and is never mutated by the API layer. The backend always returns the raw stored unit — no conversion happens in handlers or services on read.

### Authoring vs. consumption

| Context | Behaviour | Examples |
|---|---|---|
| **Authoring / edit** | Render the stored unit as-is | ProgramDetail exercise editor, recipe ingredient form |
| **Consumption / display** | Convert to the viewer's preferred unit at render time | SessionTracker, recipe viewer |

This distinction matters because the author and the viewer may be different people or in different environments with different unit preferences. Preserving the stored unit also prevents rounding errors from accumulating across repeated read-convert-save cycles.

### Backend responsibilities

- `domain.ConvertMass` / `domain.ConvertVolume` — canonical conversion primitives.
- `UnitSystem.DefaultFitnessWeightUnit()` / `DefaultCookingMassUnit()` / `DefaultCookingVolumeUnit()` — canonical mapping from a unit system to the meaningful display unit for each domain (e.g. metric fitness → `kg`, never `g`).

### Frontend responsibilities

- `useUnitPreferences()` — resolves the viewer's preferred units from their profile.
- `DISPLAY_STEPS` / `quantizeForDisplay(value, unit)` in `useUnitPreferences.js` — rounds a converted value to the smallest meaningful increment for the target unit (see table below). Also drives the `step` attribute on amount/weight input fields so browser validation matches display precision.
- `convertFitnessWeight(value, fromUnit, toUnit)` in `useUnitPreferences.js` — converts between kg and lb and applies `quantizeForDisplay`; used in consumption views and on unit toggle in edit views.
- **DO** call `convertFitnessWeight` in consumption contexts (e.g. `SessionTracker`) and when the user explicitly toggles the unit in an edit context (e.g. `ProgramDetail`).
- **DON'T** convert on read in edit contexts — render the stored unit directly without conversion.
- **DON'T** duplicate the meaningful-unit mapping; `useUnitPreferences()` mirrors `DefaultFitnessWeightUnit()` and is the single frontend source of truth.

### Display quantization steps

| Unit | Step | Notes |
|---|---|---|
| `kg` | 0.01 | |
| `lb` | 0.5 | fitness context; 0.1 for cooking if ever needed |
| `g` | 0.1 | scale precision |
| `oz` | 0.1 | |
| `ml` | 1 | |
| `l` | 0.01 | |
| `fl_oz` | 0.1 | |
| `tsp` | 0.125 | 1/8 tsp — smallest common measuring spoon |
| `tbsp` | 0.5 | 1/2 tbsp — smallest common measuring spoon |
| `cup` | 0.125 | 1/8 cup |

---

## Workers

Workers are background job processors that run alongside the HTTP server in the same binary. They use the same service layer as handlers but access it through **ports** — interfaces that define the data-tier boundary.

### Why Ports

All data access from a worker must go through a port, never by importing `internal/service` or `internal/store` directly. This establishes a clean seam for two reasons:

1. **Database connection scaling**: Only the API tier (services) should hold database connections. Workers that bypass services multiply DB clients independently, making connection count harder to reason about as the system scales.
2. **Deployment flexibility**: When the worker is eventually split into a separate process, the port implementation swaps from a local adapter (direct service call) to a remote adapter (HTTP client). The workflow orchestration code does not change.

### Structure

```go
// internal/workers/ports.go — data-tier boundary interfaces
type WorkoutSessionPort interface {
    Get(ctx context.Context, id domain.WorkoutSessionID) (*domain.WorkoutSession, error)
    PatchSummary(ctx context.Context, id domain.WorkoutSessionID, patch WorkoutSessionSummaryPatch) error
}

// internal/workers/local.go — local adapter (fat-binary deployment)
type LocalWorkoutSessionAdapter struct { svc *service.WorkoutSessionService }

// internal/workers/session_summary.go — workflow orchestration
type SessionSummaryWorker struct {
    sessions  WorkoutSessionPort
    summarize *service.SummarizeService
}
```

The workflow code depends only on `WorkoutSessionPort`. The local adapter is a thin wrapper on `WorkoutSessionService`. When the worker splits to a separate pod, the local adapter is replaced by an HTTP client — zero changes to the workflow.

### Port Design Rules

- **DO** define ports in `internal/workers/ports.go` as narrow interfaces covering only the operations the worker actually needs. A worker that only writes summaries does not need `List` or `Delete`.
- **DO** keep each port scoped to a single resource (`WorkoutSessionPort`, not a catch-all `DataPort`).
- **DON'T** expose `service.WorkoutSessionPatch` or other service types through a port boundary — define worker-owned input types (e.g. `WorkoutSessionSummaryPatch`) and translate in the adapter.
- **DO** require identity in ctx (via `domain.NewContext`) before calling any port method — the same convention as the service layer.
- **DON'T** define ports for pure-computation dependencies (e.g. `SummarizeService`, LLM client) — these are direct dependencies of the worker, not data-tier boundaries.

### Adapters

- **Local adapter** (`LocalWorkoutSessionAdapter`): delegates directly to `WorkoutSessionService`. Used in the fat-binary deployment where worker and API server share a process and DB connection pool.
- **Remote adapter** (future): HTTP client that calls the API. Used when the worker runs as a separate pod. Satisfies the same port interface; workflow code is unchanged.

### Sensitive Data

Workers call `ws.UseSensitiveData(ctx, fn)` the same way handlers do — it is the only sanctioned way to access sensitive fields regardless of call site. Sensitive data never appears in job payloads; job inputs carry only opaque identifiers `(session_id, org_id, user_id)`.

---

## MCP Integration

The MCP server is an alternative interface to the same service layer used by HTTP handlers.

- **DO** group MCP tools by resource in `register[Resource]Tools` functions.
- **DO** call the same service methods that HTTP handlers call — no direct store access from MCP tools.
- **DON'T** duplicate business logic between the MCP layer and HTTP handlers.

---

## Testing Patterns

Cove uses a vertical integration testing strategy. We prioritize testing the full stack (Handlers → Services → Stores) against a real database.

### 1. Test Infrastructure (`testutil`)
All shared testing infrastructure lives in `internal/testutil`.
- **`RunMain(m, dsnPtr)`**: Call this in `TestMain` to handle the PostgreSQL container lifecycle automatically.
- **`NewDB(t)`**: Use this to get an isolated, migrated database for a single test.

### 2. The `TestApp` Pattern
For handler and integration tests, use the `TestApp` struct (found in `internal/handlers/app_test.go`). It provides a pre-wired application stack including all services and the real `OAuth` middleware.

```go
func TestExample(t *testing.T) {
    app := NewTestApp(t) // Starts DB, wires services, applies /api prefix + middleware

    // 1. Seed data directly via services
    p := app.SeedProgram("Strength")

    // 2. Make authenticated request
    r := app.AuthRequest("GET", "/api/programs", nil, userID)
    w := app.Do(r)

    // 3. Verify
    if w.Code != http.StatusOK { ... }
}
```

### 3. Key Helpers
- **`app.Do(r)`**: Executes a request against the fully wired API (with `/api` prefix and OAuth).
- **`app.DoRaw(r)`**: Executes a request against the internal mux (no prefix, no auth).
- **`app.Seed[Entity]`**: Use "Seed" methods to bypass the HTTP layer when setting up prerequisites for a test.
- **`app.AuthRequest(...)`**: Automatically creates a valid session and attaches the `Bearer` token to the request.

### 4. Identity in Context

Since RLS and bookkeeping triggers require session variables, tests **must** provide an `Identity` in the context when calling services or stores directly.

```go
id := &domain.Identity{UserID: uID, OrgID: oID}
ctx := domain.NewContext(context.Background(), id)
```

Use `app.SeedUserWithOrg` to get a real `(UserID, OrgID)` pair, and `app.SeedExerciseForUser` to seed tenant-owned data.

### 5. Coverage Standards
- **Unhappy Paths**: Every handler must have tests for "not found", "invalid body", and "unauthorized" scenarios.
- **Normalization**: Test expectations must account for service-level normalization (e.g., exercise names being trimmed or lowercased).
- **RLS / Defense-in-Depth**: For tenant-owned resources, include a test that verifies a user from one org cannot read or modify data belonging to another org.
