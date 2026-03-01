preview: build
	cd backend && set -a && . .env && set +a && ./bin/cove

# Runs backend at :8080 in dev move and watch rebuild UI assets
# Develop at localhost:8080. OAuth redirect URL must point to localhost:8080/auth/callback.
dev:
	$(MAKE) -j2 run.backend dev.watch

# build+watch: rebuilds frontend into backend/ui/ on every file change.
# Use this when you need the backend to serve the built frontend (e.g. debugging embedded files).
dev.watch:
	$(MAKE) -C frontend watch

run.backend:
	$(MAKE) -C backend run

run.frontend:
	$(MAKE) -C frontend run

build:
	$(MAKE) -C frontend build
	$(MAKE) -C backend build
	$(MAKE) -C backend build-mcp

test.e2e: build
	$(MAKE) -C frontend test.e2e

test:
	$(MAKE) -C backend test
	$(MAKE) -C frontend test

lint:
	$(MAKE) -C backend lint
	$(MAKE) -C frontend lint

fmt:
	$(MAKE) -C backend fmt
	$(MAKE) -C frontend fmt

fix:
	$(MAKE) -C backend fix
	$(MAKE) -C frontend fix

check:
	$(MAKE) -C backend check
	$(MAKE) -C frontend check

devenv.up:
	$(MAKE) -C backend devenv.up

devenv.down:
	$(MAKE) -C backend devenv.down

devenv.reset:
	$(MAKE) -C backend devenv.reset

.PHONY: preview dev dev.watch run.backend run.frontend build test.e2e test lint fmt check fix devenv.up devenv.down devenv.reset
