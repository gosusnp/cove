// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, waitFor } from "@testing-library/preact";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { AuthProvider, useAuth } from "./Auth.jsx";

function TestComponent({ onLogout }) {
	const { logout } = useAuth();
	return (
		<button
			type="button"
			onClick={() => {
				logout();
				if (onLogout) onLogout();
			}}
		>
			Logout
		</button>
	);
}

describe("AuthContext", () => {
	beforeEach(() => {
		localStorage.clear();
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }),
		);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("calls logout API when session exists", async () => {
		const session = { token: "test-token", user: { id: "1" } };
		localStorage.setItem("cove_session", JSON.stringify(session));

		const { getByText } = render(
			<AuthProvider>
				<TestComponent />
			</AuthProvider>,
		);

		getByText("Logout").click();

		await waitFor(() =>
			expect(global.fetch).toHaveBeenCalledWith(
				"/api/users/logout",
				expect.objectContaining({
					method: "POST",
					headers: { Authorization: "Bearer test-token" },
				}),
			),
		);

		expect(localStorage.getItem("cove_session")).toBeNull();
	});

	it("does not call logout API when no session exists", async () => {
		const { getByText } = render(
			<AuthProvider>
				<TestComponent />
			</AuthProvider>,
		);

		getByText("Logout").click();

		await waitFor(() => {
			expect(global.fetch).not.toHaveBeenCalled();
		});
		expect(localStorage.getItem("cove_session")).toBeNull();
	});
});
