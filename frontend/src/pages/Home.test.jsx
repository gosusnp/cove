// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { Home } from "./Home.jsx";

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

describe("Home", () => {
	it("renders the heading when signed out", () => {
		withProviders(<Home />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the tagline when signed out", () => {
		withProviders(<Home />);
		expect(screen.getByText("Your space.")).toBeInTheDocument();
	});

	it("shows Train nav items when signed in", () => {
		withProviders(<Home />, { user: MOCK_USER });
		expect(screen.getByRole("link", { name: "Workout" })).toHaveAttribute(
			"href",
			"/workout",
		);
		expect(
			screen.getByRole("link", { name: "Review Sessions" }),
		).toHaveAttribute("href", "/sessions");
	});

	it("does not show Program nav items when signed in", () => {
		withProviders(<Home />, { user: MOCK_USER });
		expect(screen.queryByRole("link", { name: "Exercises" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Build Programs" })).toBeNull();
	});

	it("hides nav items and shows splash when signed out", () => {
		withProviders(<Home />);
		expect(screen.queryByRole("link", { name: "Workout" })).toBeNull();
		expect(screen.queryByRole("link", { name: "Review Sessions" })).toBeNull();
	});
});
