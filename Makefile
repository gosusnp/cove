preview: build
	cd backend && set -a && . .env && set +a && ./bin/cove

# Must not run in parallel because backend.build depends on the frontend.build
build: frontend.build backend.build

# Runs backend at :8080 in dev move and watch rebuild UI assets and backend
# Develop at localhost:8080. OAuth redirect URL must point to localhost:8080/auth/callback.
dev: frontend.ota-bundle
	$(MAKE) -j2 backend.watch frontend.watch

check: frontend.check backend.check android.check

fix: backend.fix frontend.fix android.fix

test.e2e:
	$(MAKE) -C frontend test.e2e

pre-commit-all: frontend.pre-commit android.pre-commit backend.pre-commit

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

backend.pre-commit:
	$(MAKE) -C backend pre-commit


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

frontend.ota-bundle:
	$(MAKE) -C frontend ota-bundle

frontend.pre-commit:
	$(MAKE) -C Frontend pre-commit


##################################################
# Android dispatch
android.build:
	$(MAKE) -C frontend android.build

android.dev:
	$(MAKE) -C frontend android.dev

android.run:
	$(MAKE) -C frontend android.run

android.check:
	$(MAKE) -C android check

android.fix:
	$(MAKE) -C android fix

android.pre-commit:
	$(MAKE) -C android pre-commit

.PHONY: preview dev build check fix test.e2e devenv.up devenv.down devenv.reset backend.build backend.check backend.fix backend.run backend.watch frontend.build frontend.check frontend.fix frontend.run frontend.watch frontend.ota-bundle android.build android.dev android.run android.check android.fix
