// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Divider, Row, Section } from "./Section.jsx";

describe("Section", () => {
	it("renders the title", () => {
		render(
			<Section title="Profile">
				<p>content</p>
			</Section>,
		);
		expect(
			screen.getByRole("heading", { name: "Profile" }),
		).toBeInTheDocument();
	});

	it("renders children", () => {
		render(
			<Section title="Profile">
				<p>content</p>
			</Section>,
		);
		expect(screen.getByText("content")).toBeInTheDocument();
	});

	it("renders the action when provided", () => {
		render(
			<Section title="Profile" action={<button type="button">Add</button>}>
				<p>content</p>
			</Section>,
		);
		expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument();
	});

	it("renders without an action when omitted", () => {
		render(
			<Section title="Profile">
				<p>content</p>
			</Section>,
		);
		expect(screen.queryByRole("button")).not.toBeInTheDocument();
	});
});

describe("Row", () => {
	it("renders the label", () => {
		render(<Row label="Email">value</Row>);
		expect(screen.getByText("Email")).toBeInTheDocument();
	});

	it("renders children", () => {
		render(<Row label="Email">user@example.com</Row>);
		expect(screen.getByText("user@example.com")).toBeInTheDocument();
	});

	it("renders sublabel when provided", () => {
		render(
			<Row label="Token" sublabel="Created Jan 1 · Last used never">
				value
			</Row>,
		);
		expect(
			screen.getByText("Created Jan 1 · Last used never"),
		).toBeInTheDocument();
	});

	it("omits sublabel element when not provided", () => {
		render(<Row label="Token">value</Row>);
		expect(screen.queryByText(/Last used/)).not.toBeInTheDocument();
	});

	it("applies a bottom border when last is not set", () => {
		const { container } = render(<Row label="Email">value</Row>);
		const div = container.firstChild;
		expect(div.style.borderBottom).toMatch(/var\(--color-border\)/);
	});

	it("omits the bottom border when last is set", () => {
		const { container } = render(
			<Row label="Email" last>
				value
			</Row>,
		);
		const div = container.firstChild;
		expect(div.style.borderBottom).toBe("");
	});
});

describe("Divider", () => {
	it("renders an hr", () => {
		render(<Divider />);
		expect(document.querySelector("hr")).toBeInTheDocument();
	});
});
