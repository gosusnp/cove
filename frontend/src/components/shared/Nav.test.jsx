// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { withProviders } from "../../test-utils.jsx";
import { Nav } from "./Nav.jsx";

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

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

	it("hides Exercises and Programs links when not signed in", () => {
		withProviders(<Nav />);
		expect(screen.queryByRole("link", { name: "Exercises" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Programs" })).toBeNull();
	});

	it("shows Sign in links when not signed in", () => {
		withProviders(<Nav />);
		const links = screen.getAllByRole("link", { name: "Sign in" });
		expect(links.length).toBeGreaterThan(0);
		for (const link of links) {
			expect(link).toHaveAttribute("href", "/login");
		}
	});

	it("shows Exercises and Programs links when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		const exerciseLinks = screen.getAllByRole("link", { name: "Exercises" });
		expect(exerciseLinks.length).toBeGreaterThan(0);
		expect(exerciseLinks[0]).toHaveAttribute("href", "/exercises");
		const programLinks = screen.getAllByRole("link", { name: "Programs" });
		expect(programLinks.length).toBeGreaterThan(0);
		expect(programLinks[0]).toHaveAttribute("href", "/programs");
	});
});
