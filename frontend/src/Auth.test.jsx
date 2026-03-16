// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, waitFor } from "@testing-library/preact";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { AuthProvider, useAuth } from "./Auth.jsx";

function TestComponent() {
	const { user, logout } = useAuth();
	return (
		<div>
			<span data-testid="user">{user ? user.email : "none"}</span>
			<button type="button" onClick={logout}>
				Logout
			</button>
		</div>
	);
}

describe("AuthContext", () => {
	beforeEach(() => {
		vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false }));
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("bootstraps user from /api/users/me on mount", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue({
				ok: true,
				json: () => Promise.resolve({ id: "1", email: "test@example.com" }),
			}),
		);

		const { getByTestId } = render(
			<AuthProvider>
				<TestComponent />
			</AuthProvider>,
		);

		await waitFor(() =>
			expect(getByTestId("user").textContent).toBe("test@example.com"),
		);
		expect(global.fetch).toHaveBeenCalledWith(
			"/api/users/me",
			expect.objectContaining({ credentials: "include" }),
		);
	});

	it("remains unauthenticated when /api/users/me returns 401", async () => {
		const { getByTestId } = render(
			<AuthProvider>
				<TestComponent />
			</AuthProvider>,
		);

		await waitFor(() => expect(getByTestId("user").textContent).toBe("none"));
	});

	it("calls logout API with credentials and clears user", async () => {
		vi.stubGlobal(
			"fetch",
			vi
				.fn()
				.mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve({ id: "1", email: "test@example.com" }),
				})
				.mockResolvedValue({ ok: true }),
		);

		const { getByTestId, getByText } = render(
			<AuthProvider>
				<TestComponent />
			</AuthProvider>,
		);

		await waitFor(() =>
			expect(getByTestId("user").textContent).toBe("test@example.com"),
		);

		getByText("Logout").click();

		await waitFor(() => expect(getByTestId("user").textContent).toBe("none"));
		expect(global.fetch).toHaveBeenCalledWith(
			"/api/users/logout",
			expect.objectContaining({ method: "POST", credentials: "include" }),
		);
	});
});
