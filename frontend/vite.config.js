/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import { defineConfig } from "vitest/config";
import { resolve } from "node:path";
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
		alias: {
			"@radix-ui/react-navigation-menu": resolve(
				import.meta.dirname,
				"src/__mocks__/radix-navigation-menu.jsx",
			),
			"@radix-ui/react-tooltip": resolve(
				import.meta.dirname,
				"src/__mocks__/radix-tooltip.jsx",
			),
			"@radix-ui/react-switch": resolve(
				import.meta.dirname,
				"src/__mocks__/radix-switch.jsx",
			),
			"@radix-ui/react-dialog": resolve(
				import.meta.dirname,
				"src/__mocks__/radix-dialog.jsx",
			),
		},
		coverage: {
			provider: "v8",
			reporter: ["text", "html"],
		},
	},
});
