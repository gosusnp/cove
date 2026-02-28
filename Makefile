preview: build
	cd backend && set -a && . .env && set +a && ./bin/cove

run.backend:
	$(MAKE) -C backend run

run.frontend:
	$(MAKE) -C frontend run

build:
	$(MAKE) -C frontend build
	$(MAKE) -C backend build
	$(MAKE) -C backend build-mcp

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

.PHONY: preview run.backend run.frontend build test lint fmt check fix devenv.up devenv.down devenv.reset
