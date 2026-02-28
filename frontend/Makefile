run:
	npm run dev

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

# TODO: add e2e target using Playwright
# - project "dev": runs against vite dev server
# - project "e2e": runs against built Go binary (requires make devenv.up)

fix: license-fix fmt

check: lint test

install:
	npm install

.PHONY: run build preview lint fmt test install
