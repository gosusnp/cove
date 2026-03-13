// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { ListItem } from "./ListItem.jsx";

describe("ListItem", () => {
	it("renders label", () => {
		render(<ListItem label="Strength A" />);
		expect(screen.getByText("Strength A")).toBeTruthy();
	});

	it("renders sublabel when provided", () => {
		render(<ListItem label="Strength A" sublabel="Mar 12, 2026" />);
		expect(screen.getByText("Mar 12, 2026")).toBeTruthy();
	});

	it("does not render sublabel when omitted", () => {
		render(<ListItem label="Strength A" />);
		expect(screen.queryByText("Mar 12, 2026")).toBeNull();
	});

	it("calls onClick when the row button is clicked", () => {
		const handler = vi.fn();
		render(<ListItem label="Strength A" onClick={handler} />);
		fireEvent.click(screen.getByRole("button"));
		expect(handler).toHaveBeenCalledTimes(1);
	});

	it("applies active styles when active=true", () => {
		const { container } = render(<ListItem label="Active" active />);
		const wrapper = container.firstChild;
		expect(wrapper.style.borderLeft).toBe("3px solid var(--color-accent)");
	});

	it("applies transparent border when active=false", () => {
		const { container } = render(<ListItem label="Inactive" active={false} />);
		const wrapper = container.firstChild;
		expect(wrapper.style.borderLeft).toBe("3px solid transparent");
	});

	it("suppresses bottom border when isLast=true", () => {
		const { container } = render(<ListItem label="Last" isLast />);
		const wrapper = container.firstChild;
		expect(wrapper.style.borderBottom).toBe("");
	});

	it("renders actions slot when provided", () => {
		render(
			<ListItem
				label="With Actions"
				actions={<button type="button">Delete</button>}
			/>,
		);
		expect(screen.getByRole("button", { name: "Delete" })).toBeTruthy();
	});

	it("does not render actions slot when omitted", () => {
		const { container } = render(<ListItem label="No Actions" />);
		// Only the main row button should exist
		const buttons = container.querySelectorAll("button");
		expect(buttons.length).toBe(1);
	});
});
