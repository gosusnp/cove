// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { TopBar } from "./TopBar.jsx";

describe("TopBar", () => {
	it("renders a header element", () => {
		render(<TopBar>content</TopBar>);
		expect(screen.getByRole("banner")).toBeInTheDocument();
	});

	it("renders children", () => {
		render(
			<TopBar>
				<span>Nav content</span>
			</TopBar>,
		);
		expect(screen.getByText("Nav content")).toBeInTheDocument();
	});

	it("renders brand content in the left slot", () => {
		render(<TopBar brand={<span>Cove</span>}>content</TopBar>);
		expect(screen.getByText("Cove")).toBeInTheDocument();
	});
});
