/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import { resolve } from "node:path";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vitest/config";

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
			"@capgo/capacitor-updater": resolve(
				import.meta.dirname,
				"src/__mocks__/capgo-capacitor-updater.js",
			),
			"@capacitor/core": resolve(
				import.meta.dirname,
				"src/__mocks__/capacitor-core.js",
			),
			"@capacitor-community/keep-awake": resolve(
				import.meta.dirname,
				"src/__mocks__/capacitor-keep-awake.js",
			),
			"@capgo/capacitor-social-login": resolve(
				import.meta.dirname,
				"src/__mocks__/capacitor-social-login.js",
			),
			"react-markdown": resolve(
				import.meta.dirname,
				"src/__mocks__/react-markdown.jsx",
			),
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
			"@radix-ui/react-accordion": resolve(
				import.meta.dirname,
				"src/__mocks__/radix-accordion.jsx",
			),
			"@dnd-kit/core": resolve(
				import.meta.dirname,
				"src/__mocks__/dnd-kit-core.jsx",
			),
			"@dnd-kit/sortable": resolve(
				import.meta.dirname,
				"src/__mocks__/dnd-kit-sortable.js",
			),
			"@dnd-kit/utilities": resolve(
				import.meta.dirname,
				"src/__mocks__/dnd-kit-utilities.js",
			),
		},
		coverage: {
			provider: "v8",
			reporter: ["text", "html"],
		},
	},
});
