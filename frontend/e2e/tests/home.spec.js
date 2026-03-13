// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { expect, test } from "@playwright/test";

test("home page shows 'Your space.'", async ({ page }) => {
	await page.goto("/");
	await expect(page.getByText("Your space.")).toBeVisible();
});

test("home page renders without errors", async ({ page }) => {
	const errors = [];
	page.on("pageerror", (err) => errors.push(err.message));

	await page.goto("/");
	await expect(page.getByRole("heading", { name: "Cove" })).toBeVisible();
	expect(errors).toHaveLength(0);
});

test("authenticated / does not redirect to login", async ({ page }) => {
	await page.goto("/");
	await expect(page).not.toHaveURL("/login");
});

test("nav is visible on home page with Cove link and avatar", async ({
	page,
}) => {
	await page.goto("/");
	await expect(page.getByRole("link", { name: "Cove" })).toBeVisible();
	await expect(
		page.getByRole("link", { name: "Account settings" }),
	).toBeVisible();
});
