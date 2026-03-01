// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { LocationProvider } from "preact-iso";
import { AuthContext } from "./auth.jsx";
import { App } from "./App.jsx";
import { Login } from "./pages/Login.jsx";
import { Home } from "./pages/Home.jsx";
import { Settings } from "./pages/Settings.jsx";
import { Nav } from "./components/Nav.jsx";

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

function withProviders(ui, { path = "/", user = null } = {}) {
	window.history.pushState({}, "", path);
	const auth = {
		user,
		token: user ? "tok" : null,
		login: vi.fn(),
		logout: vi.fn(),
	};
	return {
		...render(
			<LocationProvider>
				<AuthContext.Provider value={auth}>{ui}</AuthContext.Provider>
			</LocationProvider>,
		),
		auth,
	};
}

describe("App", () => {
	it("renders the nav on the login page", () => {
		window.history.pushState({}, "", "/login");
		render(<App />);
		expect(screen.getByRole("banner")).toBeInTheDocument();
	});
});

describe("Login", () => {
	it("renders the Cove heading", () => {
		render(<Login />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the Continue with Google link pointing to /auth/login", () => {
		render(<Login />);
		const link = screen.getByRole("link", { name: /continue with google/i });
		expect(link).toHaveAttribute("href", "/auth/login");
	});
});

describe("Home", () => {
	it("renders the heading", () => {
		render(<Home />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the tagline", () => {
		render(<Home />);
		expect(screen.getByText("Your space.")).toBeInTheDocument();
	});
});

describe("Settings", () => {
	it("renders the page heading", () => {
		withProviders(<Settings />, { user: MOCK_USER });
		expect(
			screen.getByRole("heading", { name: "Settings" }),
		).toBeInTheDocument();
	});

	it("shows the signed-in user email", () => {
		withProviders(<Settings />, { user: MOCK_USER });
		expect(screen.getByText(MOCK_USER.email)).toBeInTheDocument();
	});

	it("shows the signed-in user name", () => {
		withProviders(<Settings />, { user: MOCK_USER });
		expect(screen.getAllByText(MOCK_USER.name).length).toBeGreaterThan(0);
	});

	it("calls logout and redirects on sign out", () => {
		const { auth } = withProviders(<Settings />, {
			path: "/settings",
			user: MOCK_USER,
		});
		fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
		expect(auth.logout).toHaveBeenCalled();
	});
});

describe("Nav", () => {
	it("has a Cove link pointing to /", () => {
		withProviders(<Nav />);
		const link = screen.getByRole("link", { name: "Cove" });
		expect(link).toHaveAttribute("href", "/");
	});

	it("avatar links to /settings when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		const link = screen.getByRole("link", { name: "Account settings" });
		expect(link).toHaveAttribute("href", "/settings");
	});

	it("shows user initials avatar when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		// Initials from "Jane Smith" → "JS"
		expect(screen.getByText("JS")).toBeInTheDocument();
	});

	it("shows user email as avatar aria-label when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		expect(screen.getByLabelText(MOCK_USER.email)).toBeInTheDocument();
	});
});
