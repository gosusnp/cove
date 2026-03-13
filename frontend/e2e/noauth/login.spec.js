// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { expect, test } from "@playwright/test";

test("shows login page with Cove heading", async ({ page }) => {
	await page.goto("/login");
	await expect(page.getByRole("heading", { name: "Cove" })).toBeVisible();
	await expect(page.getByText("Your space.")).toBeVisible();
});

test("shows Continue with Google button linking to /auth/login", async ({
	page,
}) => {
	await page.goto("/login");
	const link = page.getByRole("link", { name: /continue with google/i });
	await expect(link).toBeVisible();
	await expect(link).toHaveAttribute("href", "/auth/login");
});

test("/ does not redirect when unauthenticated", async ({ page }) => {
	await page.goto("/");
	await expect(page).toHaveURL("/");
});

test("/settings redirects to /login when unauthenticated", async ({ page }) => {
	await page.goto("/settings");
	await expect(page).toHaveURL("/login");
});

test("shows sign in link when unauthenticated", async ({ page }) => {
	await page.goto("/");
	const link = page.getByRole("link", { name: /sign in/i });
	await expect(link).toBeVisible();
	await expect(link).toHaveAttribute("href", "/login");
});
