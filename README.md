# Cove

Personal fitness tracking app with an MCP interface for AI-assisted workout planning.

## Prerequisites

- Go 1.22+
- Node.js (recent; used for frontend tooling only)
- Docker (for the dev database)

## Setup

1. **Start the database**
   ```sh
   make devenv.up
   ```

2. **Configure environment**
   Edit `backend/.env` and fill in the Google OAuth credentials:

   | Variable | Default | Required | Notes |
   |---|---|---|---|
   | `DATABASE_URL` | `postgres://cove:covedev@localhost:5432/cove` | yes | Runtime app user (RLS enforced in prod) |
   | `MIGRATION_DATABASE_URL` | _(falls back to `DATABASE_URL`)_ | prod only | Privileged migration user; falls back to `DATABASE_URL` when `COVE_DEV` is set |
   | `GOOGLE_CLIENT_ID` | — | yes | From Google Cloud Console |
   | `GOOGLE_CLIENT_SECRET` | — | yes | From Google Cloud Console |
   | `GOOGLE_REDIRECT_URL` | `http://localhost:8080/auth/callback` | yes | |
   | `SESSION_ENCRYPTION_KEY` | — | yes | Base64-encoded 32-byte key; generate with `python3 -c "import os,base64; print(base64.b64encode(os.urandom(32)).decode())"` |
   | `COVE_ALLOWED_EMAILS` | _(empty = allow all)_ | no | Comma-separated allowlist |
   | `COVE_PORT` | `8080` | no | HTTP listen port |
   | `COVE_DB_SCHEMA` | `cove` | no | Database schema; set to a unique name when sharing a Postgres instance |
   | `COVE_DEV` | _(unset)_ | no | Set to any non-empty value to enable dev mode (disk UI assets, dev login route, relaxed cookie security) |

   All variables support a `_FILE` variant that reads the value from a file (e.g. `SESSION_ENCRYPTION_KEY_FILE=/run/secrets/enc-key`). Set `COVE_SECRETS_DIR` to a directory and each key will be read from a file named after the variable — useful for k8s mounted secrets.

   Migrations run automatically on startup using `MIGRATION_DATABASE_URL`.

3. **Install frontend dependencies**
   ```sh
   cd frontend && npm install
   ```

## Development

Run backend and frontend in separate terminals:

```sh
make run.backend    # Go server on :8080
make run.frontend   # Vite dev server on :5173
```

> The Vite dev server is the working dev setup. The Go binary embeds the built frontend for production (`make preview` to test that path).

## Testing

```sh
make test
```

Backend tests spin up a real Postgres container via testcontainers — Docker must be running.

## Other Commands

| Command | Description |
|---|---|
| `make fix` | Format code and apply license headers |
| `make lint` | Lint backend and frontend |
| `make check` | Lint + test |
| `make devenv.down` | Stop the database |
| `make devenv.reset` | Stop and wipe database volume |
| `make build` | Production build (frontend embedded in Go binary) |

## License

[Elastic License 2.0](LICENSE) — free to use and modify; cannot be offered as a hosted/managed service.
