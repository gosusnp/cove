run:
	VITE_COVE_ENV=dev npm run dev

watch:
	npm run watch

build:
	npm run build

preview:
	npm run preview

lint:
	npm run lint

fmt:
	npm run format

license-fix:
	cd ../backend && go tool addlicense -f ../.license_header.txt -ignore "node_modules/**" -ignore "dist/**" ../frontend

test:
	npm run test

test.e2e:
	npm run e2e

e2e.ui:
	npm run e2e:ui

fix: license-fix fmt

check: lint test

install:
	npm install
	npx playwright install --with-deps

.PHONY: run watch build preview lint fmt license-fix test test.e2e e2e.ui fix check install
