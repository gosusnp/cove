// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Markdown } from "./Markdown.jsx";

describe("Markdown", () => {
	it("renders text content", () => {
		render(<Markdown>Hello world</Markdown>);
		expect(screen.getByText("Hello world")).toBeInTheDocument();
	});

	it("applies markdown-body class", () => {
		const { container } = render(<Markdown>text</Markdown>);
		expect(container.firstChild).toHaveClass("markdown-body");
	});

	it("applies additional class prop", () => {
		const { container } = render(<Markdown class="custom">text</Markdown>);
		expect(container.firstChild).toHaveClass("custom");
	});

	it("renders nothing when children is empty", () => {
		const { container } = render(<Markdown>{""}</Markdown>);
		expect(container.firstChild).toBeInTheDocument();
	});
});
