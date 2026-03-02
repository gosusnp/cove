// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { LocationProvider } from "preact-iso";
import { AuthContext } from "./Auth.jsx";
import { App } from "./App.jsx";
import { Login } from "./pages/Login.jsx";
import { Home } from "./pages/Home.jsx";
import { Nav } from "./components/Nav.jsx";

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

function withProviders(ui, { path = "/", user = null } = {}) {
	window.history.pushState({}, "", path);
	const auth = {
		user,
		token: user ? "tok" : null,
		login: vi.fn(),
		logout: vi.fn(),
		updateUser: vi.fn(),
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
