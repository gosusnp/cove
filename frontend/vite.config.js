/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import { defineConfig } from "vitest/config";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [preact(), tailwindcss()],
	server: {
		proxy: {
			"/auth": "http://localhost:8080",
			"/api": "http://localhost:8080",
		},
	},
	build: {
		outDir: "../backend/ui",
		emptyOutDir: true,
	},
	test: {
		environment: "jsdom",
		setupFiles: ["./src/test-setup.js"],
		exclude: ["e2e/**", "node_modules/**"],
		coverage: {
			provider: "v8",
			reporter: ["text", "html"],
		},
	},
});
