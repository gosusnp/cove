run:
	VITE_COVE_ENV=dev npm run dev

watch:
	npm run watch

build:
	npm run build
	npm run build:ota-bundle

preview:
	npm run preview

lint:
	npm run lint

fmt:
	npm run format

fmt-check:
	npx biome format .

license-check:
	cd ../backend && go tool addlicense -check -f ../.license_header.txt -ignore "**/node_modules/**" -ignore "**/dist/**" -ignore "**/dist-mobile/**" ../frontend

license-fix:
	cd ../backend && go tool addlicense -f ../.license_header.txt -ignore "**/node_modules/**" -ignore "**/dist/**" -ignore "**/dist-mobile/**" ../frontend

test:
	npm run test

test.e2e:
	npm run e2e

e2e.ui:
	npm run e2e:ui

fix: license-fix fmt

check: lint fmt-check license-check test ota-bundle

pre-commit: fix check

ota-bundle:
	npm run build:ota-bundle

android.build:
	npm run cap:sync

android.dev:
	npm run cap:dev

android.run:
	npm run cap:run

install:
	npm install
	npx playwright install --with-deps

.PHONY: run watch build preview lint fmt fmt-check license-fix license-check test test.e2e e2e.ui fix check install ota-bundle android.build android.dev android.run
