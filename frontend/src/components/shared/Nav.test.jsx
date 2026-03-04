// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { Nav } from "./Nav.jsx";
import { withProviders } from "../../test-utils.jsx";

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
});
