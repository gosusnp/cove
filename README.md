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
   | Variable | Default | Notes |
   |---|---|---|
   | `DATABASE_URL` | `postgres://cove:covedev@localhost:5432/cove` | Set by devenv |
   | `COVE_API_KEY` | `dev-key` | Any string for local dev |
   | `GOOGLE_CLIENT_ID` | — | From Google Cloud Console |
   | `GOOGLE_CLIENT_SECRET` | — | From Google Cloud Console |
   | `GOOGLE_REDIRECT_URL` | `http://localhost:8080/auth/callback` | |
   | `COVE_ALLOWED_EMAILS` | _(empty = allow all)_ | Comma-separated allowlist |

   Migrations run automatically on startup.

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
