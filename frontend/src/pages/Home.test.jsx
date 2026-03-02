// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { Home } from "./Home.jsx";
import { withProviders } from "../test-utils.jsx";

describe("Home", () => {
	it("renders the heading", () => {
		withProviders(<Home />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the tagline", () => {
		withProviders(<Home />);
		expect(screen.getByText("Your space.")).toBeInTheDocument();
	});
});
