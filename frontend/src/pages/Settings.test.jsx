// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor, within } from "@testing-library/preact";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Capacitor } from "@capacitor/core";
import { withProviders } from "../test-utils.jsx";
import { Settings } from "./Settings.jsx";
import {
	isHealthConnectAvailable,
	openHealthConnectSettings,
	requestHealthConnectPermission,
} from "../lib/healthConnect.js";

vi.mock("../lib/healthConnect.js", () => ({
	isHealthConnectAvailable: vi.fn(),
	requestHealthConnectPermission: vi.fn(),
	openHealthConnectSettings: vi.fn(),
}));

vi.mock("../components/ui/Dialog.jsx", () => ({
	Dialog: ({ children }) => <div data-testid="mock-dialog">{children}</div>,
	DialogContent: ({ children }) => (
		<div data-testid="mock-dialog-content">{children}</div>
	),
	DialogTitle: ({ children }) => <h2>{children}</h2>,
	DialogDescription: ({ children }) => <p>{children}</p>,
	DialogClose: ({ children }) => (
		<button type="button" data-testid="mock-dialog-close">
			{children}
		</button>
	),
}));

const MOCK_USER = { email: "jane@example.com", display_name: "Jane Smith" };

const renderSettings = (opts = {}) =>
	withProviders(<Settings />, { path: "/settings", user: MOCK_USER, ...opts });

function mockFetch({ tokens = [], sessions = [], meUser = MOCK_USER } = {}) {
	return vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url === "/api/users/tokens" && opts?.method === "POST") {
			const name = JSON.parse(opts.body).name;
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						token: "pat_raw_secret",
						id: "new-id",
						name,
						created_at: "2026-03-01T00:00:00Z",
					}),
			});
		}
		if (url.startsWith("/api/users/tokens")) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(tokens),
			});
		}
		if (url.startsWith("/api/users/sessions")) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(sessions),
			});
		}
		return Promise.resolve({ json: () => Promise.resolve(meUser) });
	});
}

describe("Settings", () => {
	afterEach(() => {
		vi.restoreAllMocks();
		localStorage.clear();
	});

	describe("profile", () => {
		it("renders the page heading", () => {
			mockFetch();
			renderSettings();
			expect(
				screen.getByRole("heading", { name: "Settings" }),
			).toBeInTheDocument();
		});

		it("shows the signed-in user email", () => {
			mockFetch();
			renderSettings();
			expect(screen.getByText(MOCK_USER.email)).toBeInTheDocument();
		});

		it("shows the signed-in user display name", () => {
			mockFetch();
			renderSettings();
			expect(
				screen.getAllByText(MOCK_USER.display_name).length,
			).toBeGreaterThan(0);
		});
	});

	describe("fetch /api/users/me", () => {
		it("sends credentials", async () => {
			const fetchSpy = mockFetch();
			renderSettings();
			await waitFor(() => expect(fetchSpy).toHaveBeenCalled());
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/users/me",
				expect.objectContaining({ credentials: "include" }),
			);
		});

		it("calls updateUser with API response", async () => {
			const apiUser = {
				email: "api@example.com",
				created_at: "2026-01-01T00:00:00Z",
			};
			vi.spyOn(global, "fetch").mockImplementation((url) => {
				if (url.startsWith("/api/users/tokens")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				if (url.startsWith("/api/users/sessions")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(apiUser),
				});
			});
			const { auth } = renderSettings();
			await waitFor(() =>
				expect(auth.updateUser).toHaveBeenCalledWith(apiUser),
			);
		});

		it("shows auth context user when fetch fails", async () => {
			vi.spyOn(global, "fetch").mockRejectedValue(new Error("network error"));
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(MOCK_USER.email)).toBeInTheDocument(),
			);
		});

		it("redirects to /login when /me returns 401", async () => {
			vi.spyOn(global, "fetch").mockResolvedValue({ status: 401, ok: false });
			const assignSpy = vi.fn();
			vi.stubGlobal("location", { ...window.location, assign: assignSpy });
			renderSettings();
			await waitFor(() => expect(assignSpy).toHaveBeenCalledWith("/login"));
		});
	});

	describe("sign out", () => {
		it("calls logout on sign out", () => {
			mockFetch();
			const { auth } = renderSettings();
			fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
			expect(auth.logout).toHaveBeenCalled();
		});
	});

	describe("API tokens", () => {
		it("shows empty state when there are no tokens", async () => {
			mockFetch({ tokens: [] });
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText("No tokens yet")).toBeInTheDocument(),
			);
		});

		it("renders tokens returned by the API", async () => {
			mockFetch({
				tokens: [
					{
						id: "id1",
						name: "CI pipeline",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: "2026-03-01T10:00:00Z",
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText("CI pipeline")).toBeInTheDocument(),
			);
		});

		it("shows never in sublabel when last_used_at is null", async () => {
			mockFetch({
				tokens: [
					{
						id: "id1",
						name: "CI pipeline",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: null,
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Last used never/)).toBeInTheDocument(),
			);
		});

		it("shows relative time in sublabel when last_used_at is set", async () => {
			mockFetch({
				tokens: [
					{
						id: "id1",
						name: "CI pipeline",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: "2026-02-01T00:00:00Z",
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Last used .+/)).toBeInTheDocument(),
			);
		});

		it("calls DELETE when a token is deleted", async () => {
			const fetchSpy = mockFetch({
				tokens: [
					{
						id: "abc-123",
						name: "My token",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: null,
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText("My token")).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Delete" }));
			expect(fetchSpy).toHaveBeenCalledWith("/api/users/tokens/abc-123", {
				method: "DELETE",
				credentials: "include",
			});
		});

		it("removes a deleted token from the list", async () => {
			mockFetch({
				tokens: [
					{
						id: "abc-123",
						name: "My token",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: null,
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText("My token")).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Delete" }));
			await waitFor(() =>
				expect(screen.queryByText("My token")).not.toBeInTheDocument(),
			);
		});

		it("creates a token and reveals the raw value", async () => {
			mockFetch({ tokens: [] });
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Create" }),
				).not.toBeDisabled(),
			);

			fireEvent.input(screen.getByLabelText("Token name"), {
				target: { value: "Deploy key" },
			});
			fireEvent.submit(screen.getByLabelText("Token name").closest("form"));

			await waitFor(() =>
				expect(screen.getByDisplayValue("pat_raw_secret")).toBeInTheDocument(),
			);
		});

		it("copies token to clipboard and shows Copied! feedback", async () => {
			const writeText = vi.fn().mockResolvedValue(undefined);
			Object.defineProperty(navigator, "clipboard", {
				value: { writeText },
				configurable: true,
			});

			mockFetch({ tokens: [] });
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Create" }),
				).not.toBeDisabled(),
			);

			fireEvent.input(screen.getByLabelText("Token name"), {
				target: { value: "Deploy key" },
			});
			fireEvent.submit(screen.getByLabelText("Token name").closest("form"));
			await waitFor(() =>
				expect(screen.getByDisplayValue("pat_raw_secret")).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Copy" }));
			expect(writeText).toHaveBeenCalledWith("pat_raw_secret");
			expect(
				screen.getByRole("button", { name: "Copied!" }),
			).toBeInTheDocument();
		});

		it("shows error when token creation API fails", async () => {
			vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
				if (url === "/api/users/tokens" && opts?.method === "POST") {
					return Promise.resolve({ ok: false });
				}
				if (url.startsWith("/api/users/tokens")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				if (url.startsWith("/api/users/sessions")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				return Promise.resolve({ json: () => Promise.resolve(MOCK_USER) });
			});
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Create" }),
				).not.toBeDisabled(),
			);

			fireEvent.input(screen.getByLabelText("Token name"), {
				target: { value: "Deploy key" },
			});
			fireEvent.submit(screen.getByLabelText("Token name").closest("form"));

			await waitFor(() =>
				expect(
					screen.getByText("Failed to create token. Try again."),
				).toBeInTheDocument(),
			);
		});

		it("shows a validation error when submitting with no name", async () => {
			mockFetch({ tokens: [] });
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Create" }),
				).not.toBeDisabled(),
			);

			fireEvent.submit(screen.getByLabelText("Token name").closest("form"));

			expect(screen.getByText("Token name is required.")).toBeInTheDocument();
		});
	});

	describe("Sessions", () => {
		it("renders sessions returned by the API", async () => {
			mockFetch({
				sessions: [
					{
						id: "s1",
						initial_browser: "Chrome",
						initial_os: "macOS",
						initial_ip_masked: "1.1.1.0",
						is_current: true,
						last_used_at: "2026-03-01T10:00:00Z",
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByText(/Chrome on macOS · Current/),
				).toBeInTheDocument(),
			);
		});

		it("shows last_used_at relative time in sublabel", async () => {
			mockFetch({
				sessions: [
					{
						id: "s1",
						initial_browser: "Chrome",
						initial_os: "macOS",
						initial_ip_masked: "1.1.1.0",
						created_at: "2026-01-01T00:00:00Z",
						last_used_at: "2026-02-01T00:00:00Z",
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Active .+/)).toBeInTheDocument(),
			);
		});

		it("falls back to created_at when last_used_at is null", async () => {
			mockFetch({
				sessions: [
					{
						id: "s1",
						initial_browser: "Chrome",
						initial_os: "macOS",
						initial_ip_masked: "1.1.1.0",
						created_at: "2026-02-01T00:00:00Z",
						last_used_at: null,
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Active .+/)).toBeInTheDocument(),
			);
		});

		it("calls DELETE when a session is revoked", async () => {
			const fetchSpy = mockFetch({
				sessions: [
					{
						id: "abc-456",
						initial_browser: "Firefox",
						initial_os: "Linux",
						is_current: false,
					},
				],
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Firefox on Linux/)).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Revoke" }));
			expect(fetchSpy).toHaveBeenCalledWith("/api/users/sessions/abc-456", {
				method: "DELETE",
				credentials: "include",
			});
		});

		it("signs out when current session is deleted", async () => {
			mockFetch({
				sessions: [
					{
						id: "current-id",
						initial_browser: "Chrome",
						initial_os: "macOS",
						is_current: true,
					},
				],
			});
			const { auth } = renderSettings();
			await waitFor(() =>
				expect(
					screen.getByText(/Chrome on macOS · Current/),
				).toBeInTheDocument(),
			);

			const sessionsSection = screen
				.getByText("Active Sessions")
				.closest("section");
			fireEvent.click(
				within(sessionsSection).getByRole("button", { name: "Sign out" }),
			);

			expect(auth.logout).toHaveBeenCalled();
		});

		it("shows error row when sessions fetch fails", async () => {
			vi.spyOn(global, "fetch").mockImplementation((url) => {
				if (url === "/api/users/sessions") {
					return Promise.reject(new Error("network"));
				}
				if (url.startsWith("/api/users/tokens")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				return Promise.resolve({ json: () => Promise.resolve(MOCK_USER) });
			});
			renderSettings();
			await waitFor(() =>
				expect(
					screen.getByText("Could not load sessions."),
				).toBeInTheDocument(),
			);
		});

		it("keeps session in list when DELETE fails", async () => {
			vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
				if (url === "/api/users/sessions" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve([
								{
									id: "s1",
									initial_browser: "Firefox",
									initial_os: "Linux",
									is_current: false,
								},
							]),
					});
				}
				if (
					url.startsWith("/api/users/sessions/") &&
					opts?.method === "DELETE"
				) {
					return Promise.resolve({ ok: false });
				}
				if (url.startsWith("/api/users/tokens")) {
					return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
				}
				return Promise.resolve({ json: () => Promise.resolve(MOCK_USER) });
			});
			renderSettings();
			await waitFor(() =>
				expect(screen.getByText(/Firefox on Linux/)).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Revoke" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Revoke" }),
				).toBeInTheDocument(),
			);
			expect(screen.getByText(/Firefox on Linux/)).toBeInTheDocument();
		});
	});

	describe("Health Connect", () => {
		beforeEach(() => {
			mockFetch();
		});

		it("does not show the HC section on web", () => {
			renderSettings();
			expect(screen.queryByText("Health Connect")).not.toBeInTheDocument();
		});

		describe("on Android", () => {
			beforeEach(() => {
				vi.spyOn(Capacitor, "getPlatform").mockReturnValue("android");
			});

			it("shows the Health Connect section", () => {
				renderSettings();
				expect(screen.getByText("Health Connect")).toBeInTheDocument();
				expect(
					screen.getByRole("switch", {
						name: "Sync workouts with Health Connect",
					}),
				).toBeInTheDocument();
			});

			it("shows sync enabled when restored from storage", () => {
				localStorage.setItem("hc_sync_enabled", "true");
				renderSettings();
				expect(
					screen.getByRole("switch", {
						name: "Sync workouts with Health Connect",
					}),
				).toHaveAttribute("aria-checked", "true");
			});

			it("enables sync when HC is available and permission is granted", async () => {
				isHealthConnectAvailable.mockResolvedValue(true);
				requestHealthConnectPermission.mockResolvedValue(true);
				renderSettings();
				const toggle = screen.getByRole("switch", {
					name: "Sync workouts with Health Connect",
				});
				fireEvent.click(toggle);
				await waitFor(() =>
					expect(toggle).toHaveAttribute("aria-checked", "true"),
				);
				expect(localStorage.getItem("hc_sync_enabled")).toBe("true");
			});

			it("shows error when Health Connect is not installed", async () => {
				isHealthConnectAvailable.mockResolvedValue(false);
				renderSettings();
				fireEvent.click(
					screen.getByRole("switch", {
						name: "Sync workouts with Health Connect",
					}),
				);
				await waitFor(() =>
					expect(
						screen.getByText("Health Connect is not installed on this device."),
					).toBeInTheDocument(),
				);
				expect(localStorage.getItem("hc_sync_enabled")).toBeNull();
			});

			it("shows error and open settings button when permission is denied", async () => {
				isHealthConnectAvailable.mockResolvedValue(true);
				requestHealthConnectPermission.mockResolvedValue(false);
				renderSettings();
				fireEvent.click(
					screen.getByRole("switch", {
						name: "Sync workouts with Health Connect",
					}),
				);
				await waitFor(() =>
					expect(
						screen.getByText(
							"Permission denied. Open Health Connect and grant access under App permissions.",
						),
					).toBeInTheDocument(),
				);
				expect(
					screen.getByRole("button", { name: "Open settings" }),
				).toBeInTheDocument();
				expect(localStorage.getItem("hc_sync_enabled")).toBeNull();
			});

			it("opens Health Connect settings when button is clicked", async () => {
				isHealthConnectAvailable.mockResolvedValue(true);
				requestHealthConnectPermission.mockResolvedValue(false);
				renderSettings();
				fireEvent.click(
					screen.getByRole("switch", {
						name: "Sync workouts with Health Connect",
					}),
				);
				await waitFor(() =>
					expect(
						screen.getByRole("button", { name: "Open settings" }),
					).toBeInTheDocument(),
				);
				fireEvent.click(screen.getByRole("button", { name: "Open settings" }));
				expect(openHealthConnectSettings).toHaveBeenCalled();
			});

			it("disables sync and clears status when toggled off", async () => {
				localStorage.setItem("hc_sync_enabled", "true");
				renderSettings();
				const toggle = screen.getByRole("switch", {
					name: "Sync workouts with Health Connect",
				});
				expect(toggle).toHaveAttribute("aria-checked", "true");
				fireEvent.click(toggle);
				await waitFor(() =>
					expect(toggle).toHaveAttribute("aria-checked", "false"),
				);
				expect(localStorage.getItem("hc_sync_enabled")).toBeNull();
			});
		});
	});
});
