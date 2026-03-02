// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { PageTitle } from "./PageTitle.jsx";

describe("PageTitle", () => {
	it("renders children correctly", () => {
		render(<PageTitle>Test Title</PageTitle>);
		expect(
			screen.getByRole("heading", { name: "Test Title" }),
		).toBeInTheDocument();
	});

	it("applies custom class names", () => {
		const { container } = render(
			<PageTitle class="custom-class">Title</PageTitle>,
		);
		expect(container.firstChild).toHaveClass("custom-class");
	});

	it("renders as h1", () => {
		render(<PageTitle>Title</PageTitle>);
		const heading = screen.getByRole("heading", { name: "Title" });
		expect(heading.tagName).toBe("H1");
	});
});
