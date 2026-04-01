// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App.jsx";

describe("App", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("renders the nav on the login page", async () => {
		vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false }));
		window.history.pushState({}, "", "/login");
		render(<App />);
		await waitFor(() => expect(screen.getByRole("banner")).toBeInTheDocument());
	});

	it("redirects non-admin users away from admin routes", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue({
				ok: true,
				json: () =>
					Promise.resolve({ email: "user@test.com", is_admin: false }),
			}),
		);
		window.history.pushState({}, "", "/admin/service-accounts");
		render(<App />);
		await waitFor(() => expect(window.location.pathname).toBe("/"));
	});
});
