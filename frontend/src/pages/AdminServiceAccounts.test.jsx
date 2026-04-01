// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { AdminServiceAccounts } from "./AdminServiceAccounts.jsx";

// ─── Mocks ────────────────────────────────────────────────────────────────────

vi.mock("../components/ui/ConfirmDialog.jsx", () => ({
	ConfirmDialog: ({ onConfirm, openSignal }) =>
		openSignal?.value ? (
			<div>
				<button
					type="button"
					data-testid="confirm-delete-sa"
					onClick={onConfirm}
				>
					Confirm
				</button>
			</div>
		) : null,
}));

vi.mock("../components/ui/Dialog.jsx", () => ({
	Dialog: ({ children }) => <div>{children}</div>,
	DialogContent: ({ children }) => <div>{children}</div>,
	DialogTitle: ({ children }) => <h2>{children}</h2>,
	DialogDescription: ({ children }) => <p>{children}</p>,
	DialogClose: ({ children }) => <div>{children}</div>,
}));

vi.mock("../components/ui/ListDetail.jsx", () => ({
	ListDetail: ({ list, detail, emptyState, hasDetail }) => (
		<div data-testid="mock-list-detail">
			<div data-testid="list-panel">{list}</div>
			<div data-testid="detail-panel">
				{hasDetail ? detail : <p>{emptyState}</p>}
			</div>
		</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const ADMIN_USER = { email: "admin@example.com", is_admin: true };
const MOCK_SA = {
	id: "sa-1",
	name: "CI Bot",
	created_at: "2026-01-01T00:00:00Z",
};
const MOCK_TOKEN = {
	id: "tok-1",
	name: "deploy-key",
	created_at: "2026-01-01T00:00:00Z",
};

function mockFetch({ accounts = [], tokens = [] } = {}) {
	return vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		const method = opts?.method ?? "GET";

		// Token delete: /tokens/{id}
		if (url.includes("/tokens/") && method === "DELETE") {
			return Promise.resolve({ ok: true });
		}
		// Token create: POST /tokens
		if (url.includes("/tokens") && method === "POST") {
			const name = JSON.parse(opts.body).name;
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						id: "new-tok",
						name,
						token: "pat_rawsecret",
						created_at: "2026-01-01T00:00:00Z",
					}),
			});
		}
		// Token list: GET /tokens
		if (url.includes("/tokens")) {
			return Promise.resolve({ ok: true, json: () => Promise.resolve(tokens) });
		}
		// SA delete: DELETE /service-accounts/{id}
		if (url.startsWith("/api/admin/service-accounts/") && method === "DELETE") {
			return Promise.resolve({ ok: true });
		}
		// SA create: POST /service-accounts
		if (url === "/api/admin/service-accounts" && method === "POST") {
			const name = JSON.parse(opts.body).name;
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						id: "new-sa",
						name,
						created_at: "2026-01-01T00:00:00Z",
					}),
			});
		}
		// SA list: GET /service-accounts
		if (url === "/api/admin/service-accounts") {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(accounts),
			});
		}
		return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
	});
}

const renderPage = () =>
	withProviders(<AdminServiceAccounts />, {
		path: "/admin/service-accounts",
		user: ADMIN_USER,
	});

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => vi.restoreAllMocks());

describe("AdminServiceAccounts — list", () => {
	it("renders service account names from the API", async () => {
		mockFetch({ accounts: [MOCK_SA] });
		renderPage();
		await waitFor(() => expect(screen.getByText("CI Bot")).toBeInTheDocument());
	});

	it("shows empty state when no accounts exist", async () => {
		mockFetch({ accounts: [] });
		renderPage();
		await waitFor(() =>
			expect(screen.getByText("No service accounts yet.")).toBeInTheDocument(),
		);
	});

	it("shows detail empty state when no account is selected", async () => {
		mockFetch({ accounts: [MOCK_SA] });
		renderPage();
		await waitFor(() =>
			expect(
				screen.getByText("Select a service account to manage its tokens."),
			).toBeInTheDocument(),
		);
	});
});

describe("AdminServiceAccounts — create service account", () => {
	it("shows validation error when name is empty", async () => {
		mockFetch({ accounts: [] });
		renderPage();
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("posts to the API with the entered name", async () => {
		const fetchSpy = mockFetch({ accounts: [] });
		renderPage();
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Deploy Bot" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/admin/service-accounts",
				expect.objectContaining({
					method: "POST",
					body: JSON.stringify({ name: "Deploy Bot" }),
				}),
			),
		);
	});

	it("adds the created account to the list", async () => {
		mockFetch({ accounts: [] });
		renderPage();
		await waitFor(() =>
			expect(screen.getByText("No service accounts yet.")).toBeInTheDocument(),
		);
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Deploy Bot" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		await waitFor(() =>
			expect(screen.getAllByText("Deploy Bot").length).toBeGreaterThan(0),
		);
	});

	it("shows an error when the create API fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === "/api/admin/service-accounts" && opts?.method === "POST") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
		});
		renderPage();
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Deploy Bot" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		await waitFor(() =>
			expect(
				screen.getByText("Failed to create service account."),
			).toBeInTheDocument(),
		);
	});
});

describe("AdminServiceAccounts — token management", () => {
	it("loads tokens when a service account is selected", async () => {
		mockFetch({ accounts: [MOCK_SA], tokens: [MOCK_TOKEN] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));
		fireEvent.click(screen.getByText("CI Bot"));
		await waitFor(() =>
			expect(screen.getByText("deploy-key")).toBeInTheDocument(),
		);
	});

	it("shows empty token state when the account has no tokens", async () => {
		mockFetch({ accounts: [MOCK_SA], tokens: [] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));
		fireEvent.click(screen.getByText("CI Bot"));
		await waitFor(() =>
			expect(screen.getByText("No tokens yet")).toBeInTheDocument(),
		);
	});

	it("creates a token and reveals the raw value once", async () => {
		mockFetch({ accounts: [MOCK_SA], tokens: [] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));
		fireEvent.click(screen.getByText("CI Bot"));
		await waitFor(() => screen.getByText("No tokens yet"));

		fireEvent.input(screen.getByLabelText("Token name"), {
			target: { value: "deploy-key" },
		});
		fireEvent.submit(screen.getByLabelText("Token name").closest("form"));

		await waitFor(() =>
			expect(screen.getByDisplayValue("pat_rawsecret")).toBeInTheDocument(),
		);
	});

	it("shows validation error when token name is empty", async () => {
		mockFetch({ accounts: [MOCK_SA], tokens: [] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));
		fireEvent.click(screen.getByText("CI Bot"));
		await waitFor(() => screen.getByText("No tokens yet"));

		fireEvent.submit(screen.getByLabelText("Token name").closest("form"));
		expect(screen.getByText("Token name is required.")).toBeInTheDocument();
	});

	it("revokes a token and removes it from the list", async () => {
		const fetchSpy = mockFetch({ accounts: [MOCK_SA], tokens: [MOCK_TOKEN] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));
		fireEvent.click(screen.getByText("CI Bot"));
		await waitFor(() => screen.getByText("deploy-key"));

		fireEvent.click(screen.getByRole("button", { name: "Revoke" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				expect.stringContaining("/tokens/tok-1"),
				expect.objectContaining({ method: "DELETE" }),
			),
		);
		await waitFor(() =>
			expect(screen.queryByText("deploy-key")).not.toBeInTheDocument(),
		);
	});
});

describe("AdminServiceAccounts — delete service account", () => {
	it("sends DELETE to the API after confirmation", async () => {
		const fetchSpy = mockFetch({ accounts: [MOCK_SA] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));

		fireEvent.click(screen.getByRole("button", { name: "Delete" }));
		await waitFor(() => screen.getByTestId("confirm-delete-sa"));
		fireEvent.click(screen.getByTestId("confirm-delete-sa"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/admin/service-accounts/sa-1",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});

	it("removes the account from the list after confirmation", async () => {
		mockFetch({ accounts: [MOCK_SA] });
		renderPage();
		await waitFor(() => screen.getByText("CI Bot"));

		fireEvent.click(screen.getByRole("button", { name: "Delete" }));
		await waitFor(() => screen.getByTestId("confirm-delete-sa"));
		fireEvent.click(screen.getByTestId("confirm-delete-sa"));

		await waitFor(() =>
			expect(screen.queryByText("CI Bot")).not.toBeInTheDocument(),
		);
	});
});
