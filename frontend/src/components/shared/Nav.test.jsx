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
		const links = screen.getAllByRole("link", { name: "Account settings" });
		expect(links.length).toBeGreaterThan(0);
		for (const link of links) {
			expect(link).toHaveAttribute("href", "/settings");
		}
	});

	it("shows user initials avatar when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		// Initials from "Jane Smith" → "JS" (appears in desktop and mobile nav)
		const avatars = screen.getAllByText("JS");
		expect(avatars.length).toBeGreaterThan(0);
	});

	it("shows user email as avatar aria-label when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		const avatars = screen.getAllByLabelText(MOCK_USER.email);
		expect(avatars.length).toBeGreaterThan(0);
	});

	it("hides nav links when not signed in", () => {
		withProviders(<Nav />);
		expect(screen.queryByRole("link", { name: "Exercises" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Build Programs" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Workout" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Review Sessions" })).toBeNull();
	});

	it("shows Sign in links when not signed in", () => {
		withProviders(<Nav />);
		const links = screen.getAllByRole("link", { name: "Sign in" });
		expect(links.length).toBeGreaterThan(0);
		for (const link of links) {
			expect(link).toHaveAttribute("href", "/login");
		}
	});

	it("shows Train and Program nav links when signed in", () => {
		withProviders(<Nav />, { user: MOCK_USER });
		const workoutLinks = screen.getAllByRole("link", { name: "Workout" });
		expect(workoutLinks.length).toBeGreaterThan(0);
		expect(workoutLinks[0]).toHaveAttribute("href", "/workout");
		const sessionLinks = screen.getAllByRole("link", {
			name: "Review Sessions",
		});
		expect(sessionLinks.length).toBeGreaterThan(0);
		expect(sessionLinks[0]).toHaveAttribute("href", "/sessions");
		const exerciseLinks = screen.getAllByRole("link", { name: "Exercises" });
		expect(exerciseLinks.length).toBeGreaterThan(0);
		expect(exerciseLinks[0]).toHaveAttribute("href", "/exercises");
		const programLinks = screen.getAllByRole("link", {
			name: "Build Programs",
		});
		expect(programLinks.length).toBeGreaterThan(0);
		expect(programLinks[0]).toHaveAttribute("href", "/programs");
	});
});
