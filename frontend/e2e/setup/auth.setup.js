// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { expect, test as setup } from "@playwright/test";

const AUTH_FILE = "playwright/.auth/user.json";
const STORAGE_KEY = "cove_session";
const TEST_EMAIL = "test@cove.dev";

setup("create authenticated session", async ({ request, page }) => {
	const res = await request.post("/auth/dev-login", {
		data: { email: TEST_EMAIL },
	});
	await expect(res).toBeOK();
	const { token } = await res.json();

	// Navigate to the app so we have an origin to attach localStorage to.
	await page.goto("/login");

	await page.evaluate(
		([key, value]) => localStorage.setItem(key, value),
		[STORAGE_KEY, JSON.stringify({ token, user: { email: TEST_EMAIL } })],
	);

	await page.context().storageState({ path: AUTH_FILE });
});
