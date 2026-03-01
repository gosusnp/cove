preview: build
	cd backend && set -a && . .env && set +a && ./bin/cove

# Must not run in parallel because backend.build depends on the frontend.build
build: frontend.build backend.build

# Runs backend at :8080 in dev move and watch rebuild UI assets and backend
# Develop at localhost:8080. OAuth redirect URL must point to localhost:8080/auth/callback.
dev:
	$(MAKE) -j2 backend.watch frontend.watch

check: backend.check frontend.check

fix: backend.fix frontend.fix

test.e2e:
	$(MAKE) -C frontend test.e2e

##################################################
# Devenv dispatch
devenv.up:
	$(MAKE) -C backend devenv.up

devenv.down:
	$(MAKE) -C backend devenv.down

devenv.reset:
	$(MAKE) -C backend devenv.reset


##################################################
# Backend dispatch
backend.build:
	$(MAKE) -C backend build

backend.check:
	$(MAKE) -C backend check

backend.fix:
	$(MAKE) -C backend fix

backend.run:
	$(MAKE) -C backend run

backend.watch:
	$(MAKE) -C backend watch


##################################################
# Frontend dispatch
frontend.build:
	$(MAKE) -C frontend build

frontend.check:
	$(MAKE) -C frontend check

frontend.fix:
	$(MAKE) -C frontend fix

frontend.run:
	$(MAKE) -C frontend run

frontend.watch:
	$(MAKE) -C frontend watch

.PHONY: preview dev build check fix test.e2e devenv.up devenv.down devenv.reset backend.build backend.check backend.fix backend.run backend.watch frontend.build frontend.check frontend.fix frontend.run frontend.watch
