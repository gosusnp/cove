// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { defineConfig } from "@playwright/test";

export default defineConfig({
	testDir: "./e2e",
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	use: {
		baseURL: "http://localhost:8080",
		trace: "on-first-retry",
	},
	projects: [
		// Auth setup — runs first, saves session to playwright/.auth/user.json
		{
			name: "setup",
			testMatch: /setup\/.*\.setup\.js/,
		},
		// Authenticated tests — depend on setup
		{
			name: "auth",
			dependencies: ["setup"],
			use: { storageState: "playwright/.auth/user.json" },
			testMatch: /tests\/.*\.spec\.js/,
		},
		// Unauthenticated tests — no session, no dependency on setup
		{
			name: "noauth",
			testMatch: /noauth\/.*\.spec\.js/,
		},
	],
	webServer: {
		command:
			"sh -c 'set -a && . ../backend/.env && set +a && cd ../backend && bin/cove'",
		url: "http://localhost:8080",
		reuseExistingServer: !process.env.CI,
		timeout: 30_000,
	},
});
