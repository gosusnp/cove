// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { CheckList, CheckListItem, CheckListSection } from "./CheckList.jsx";

describe("CheckList", () => {
	it("renders children", () => {
		render(
			<CheckList>
				<span>content</span>
			</CheckList>,
		);
		expect(screen.getByText("content")).toBeTruthy();
	});
});

describe("CheckListSection", () => {
	it("renders the label and children", () => {
		render(
			<CheckListSection label="Round 1 of 3">
				<span>item</span>
			</CheckListSection>,
		);
		expect(screen.getByText("Round 1 of 3")).toBeTruthy();
		expect(screen.getByText("item")).toBeTruthy();
	});

	it("does not strike through the label when the section has no items", () => {
		render(<CheckListSection label="Empty" />);
		expect(screen.getByText("Empty").className).not.toContain("line-through");
	});

	it("does not strike through the label when items are not all done", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem>Push-ups × 10</CheckListItem>
				<CheckListItem>Pull-ups × 5</CheckListItem>
			</CheckListSection>,
		);
		const label = screen.getByText("Strength");
		expect(label.className).not.toContain("line-through");
	});

	it("strikes through the label when all items are done", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem defaultChecked>Push-ups × 10</CheckListItem>
				<CheckListItem defaultChecked>Pull-ups × 5</CheckListItem>
			</CheckListSection>,
		);
		const label = screen.getByText("Strength");
		expect(label.className).toContain("line-through");
	});

	it("strikes through after the last item is checked", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem defaultChecked>Push-ups × 10</CheckListItem>
				<CheckListItem>Pull-ups × 5</CheckListItem>
			</CheckListSection>,
		);
		const label = screen.getByText("Strength");
		expect(label.className).not.toContain("line-through");

		const buttons = screen.getAllByRole("button");
		const unchecked = buttons.find(
			(b) => b.getAttribute("aria-pressed") === "false",
		);
		fireEvent.click(unchecked);

		expect(label.className).toContain("line-through");
	});

	it("removes strike through when an item is unchecked", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem defaultChecked>Push-ups × 10</CheckListItem>
				<CheckListItem defaultChecked>Pull-ups × 5</CheckListItem>
			</CheckListSection>,
		);
		const label = screen.getByText("Strength");
		expect(label.className).toContain("line-through");

		const buttons = screen.getAllByRole("button");
		fireEvent.click(buttons[0]);

		expect(label.className).not.toContain("line-through");
	});
});

describe("CheckListItem", () => {
	it("starts unpressed by default", () => {
		render(<CheckListItem>Push-ups × 10</CheckListItem>);
		expect(screen.getByRole("button").getAttribute("aria-pressed")).toBe(
			"false",
		);
	});

	it("respects defaultChecked=true", () => {
		render(<CheckListItem defaultChecked>Pull-ups × 5</CheckListItem>);
		expect(screen.getByRole("button").getAttribute("aria-pressed")).toBe(
			"true",
		);
	});

	it("toggles to pressed on click", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.click(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("true");
	});

	it("toggles back to unpressed on second click", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.click(btn);
		fireEvent.click(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("false");
	});

	it("applies line-through when checked", () => {
		render(<CheckListItem defaultChecked>Hip thrust × 12</CheckListItem>);
		expect(screen.getByText("Hip thrust × 12").className).toContain(
			"line-through",
		);
	});

	it("removes line-through when unchecked", () => {
		render(<CheckListItem>Hip thrust × 12</CheckListItem>);
		expect(screen.getByText("Hip thrust × 12").className).not.toContain(
			"line-through",
		);
	});
});
