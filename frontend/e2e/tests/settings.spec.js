// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { test, expect } from "@playwright/test";

test("settings page renders without errors", async ({ page }) => {
	const errors = [];
	page.on("pageerror", (err) => errors.push(err.message));

	await page.goto("/settings");
	await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
	expect(errors).toHaveLength(0);
});

test("settings shows user email", async ({ page }) => {
	await page.goto("/settings");
	await expect(page.getByText("test@cove.dev").first()).toBeVisible();
});

test("sign out navigates to login", async ({ page }) => {
	await page.goto("/settings");
	await page.getByRole("button", { name: "Sign out" }).click();
	await expect(page).toHaveURL("/login");
});

test("can navigate from home to settings via avatar", async ({ page }) => {
	await page.goto("/");
	await page.getByRole("link", { name: "Account settings" }).click();
	await expect(page).toHaveURL("/settings");
	await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
});
