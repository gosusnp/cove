# Cove

Personal fitness tracking app with an MCP interface for AI-assisted workout planning.

## Docs

Read before writing any code:

- `docs/architecture.md` — project layout, layer hierarchy, handler/service/store rules, error handling, MCP integration
- `docs/style.md` — file naming, identifiers, JSON tags, SQL conventions, file headers, frontend conventions

## Commands

Root commands operate on both backend and frontend:

| Task | Command |
|---|---|
| Start dev database | `make devenv.up` |
| Run backend | `make backend.run` |
| Run frontend | `make frontend.run` |
| Lint + test | `make check` |
| Format + license headers | `make fix` |

When working on one layer only, run from the relevant subdirectory to avoid triggering both:

```sh
cd backend && make check   # or: test, fix
cd frontend && make check  # or: test, fix
```

Backend tests require Docker (testcontainers). Migrations run automatically on startup.

## Environment

Configured via `backend/.env`. Required for OAuth: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`. See `README.md` for full variable reference.

## Definition of Done

Before considering a task complete:

1. Run `make fix` — formats code and applies license headers
2. Run `make check` — must pass with no lint errors and no failing tests
