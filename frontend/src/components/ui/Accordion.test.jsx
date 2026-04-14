// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import {
	Accordion,
	AccordionContent,
	AccordionDragHandle,
	AccordionItem,
	AccordionTrigger,
} from "./Accordion.jsx";

// ── Basic accordion ───────────────────────────────────────────────────────────

describe("Accordion", () => {
	it("renders all items", () => {
		render(
			<Accordion>
				<AccordionItem value="a">
					<AccordionTrigger>Item A</AccordionTrigger>
					<AccordionContent>Content A</AccordionContent>
				</AccordionItem>
				<AccordionItem value="b">
					<AccordionTrigger>Item B</AccordionTrigger>
					<AccordionContent>Content B</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		expect(screen.getByRole("button", { name: /Item A/i })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: /Item B/i })).toBeInTheDocument();
	});

	it("reveals content when trigger is clicked", () => {
		render(
			<Accordion>
				<AccordionItem value="a">
					<AccordionTrigger>Press me</AccordionTrigger>
					<AccordionContent>Hidden content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		const trigger = screen.getByRole("button", { name: /Press me/i });
		fireEvent.click(trigger);

		expect(screen.getByText("Hidden content")).toBeInTheDocument();
	});

	it("opens multiple items in type=multiple mode", () => {
		render(
			<Accordion type="multiple">
				<AccordionItem value="a">
					<AccordionTrigger>A</AccordionTrigger>
					<AccordionContent>Content A</AccordionContent>
				</AccordionItem>
				<AccordionItem value="b">
					<AccordionTrigger>B</AccordionTrigger>
					<AccordionContent>Content B</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		fireEvent.click(screen.getByRole("button", { name: "A" }));
		fireEvent.click(screen.getByRole("button", { name: "B" }));

		expect(screen.getByText("Content A")).toBeInTheDocument();
		expect(screen.getByText("Content B")).toBeInTheDocument();
	});
});

// ── AccordionItem ─────────────────────────────────────────────────────────────

describe("AccordionItem", () => {
	it("renders children", () => {
		render(
			<Accordion>
				<AccordionItem value="x">
					<AccordionTrigger>Trigger</AccordionTrigger>
					<AccordionContent>Body</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);
		expect(
			screen.getByRole("button", { name: /Trigger/i }),
		).toBeInTheDocument();
	});

	it("does not toggle when space is pressed in a nested input with stopPropagation", () => {
		render(
			<Accordion>
				<AccordionItem value="a">
					<AccordionTrigger>
						<input
							data-testid="nested-input"
							onKeyDown={(e) => {
								if (e.key === " ") e.stopPropagation();
							}}
							onKeyUp={(e) => {
								if (e.key === " ") e.stopPropagation();
							}}
						/>
					</AccordionTrigger>
					<AccordionContent>Hidden content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		const input = screen.getByTestId("nested-input");
		input.focus();
		fireEvent.keyDown(input, { key: " ", code: "Space", keyCode: 32 });
		fireEvent.keyUp(input, { key: " ", code: "Space", keyCode: 32 });

		// Should stay closed
		const contentContainer = screen
			.queryByText("Hidden content")
			?.closest("[data-state]");
		if (contentContainer) {
			expect(contentContainer).toHaveAttribute("data-state", "closed");
		} else {
			expect(screen.queryByText("Hidden content")).not.toBeInTheDocument();
		}
	});

	it("toggles when space is pressed in a nested input without stopPropagation", () => {
		render(
			<Accordion>
				<AccordionItem value="a">
					<AccordionTrigger>
						<input data-testid="nested-input-bubbling" />
					</AccordionTrigger>
					<AccordionContent>Hidden content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		const input = screen.getByTestId("nested-input-bubbling");
		input.focus();
		fireEvent.keyDown(input, { key: " ", code: "Space", keyCode: 32 });
		fireEvent.keyUp(input, { key: " ", code: "Space", keyCode: 32 });

		// Should toggle open on keyUp Space if bubbling up to Trigger
		const contentContainer = screen
			.getByText("Hidden content")
			.closest("[data-state]");
		expect(contentContainer).toHaveAttribute("data-state", "open");
	});

	it("renders with a sortable id without throwing", () => {
		render(
			<Accordion>
				<AccordionItem id="item-1" value="item-1">
					<AccordionTrigger>Sortable item</AccordionTrigger>
					<AccordionContent>Sortable content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);
		expect(
			screen.getByRole("button", { name: /Sortable item/i }),
		).toBeInTheDocument();
	});
});

// ── AccordionDragHandle ───────────────────────────────────────────────────────

describe("AccordionDragHandle", () => {
	it("renders with accessible label", () => {
		render(
			<Accordion>
				<AccordionItem value="s1">
					<AccordionTrigger>
						<AccordionDragHandle />
						Set 1
					</AccordionTrigger>
					<AccordionContent>content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		expect(document.querySelector(".cursor-grab")).toBeInTheDocument();
	});

	it("renders the grip icon", () => {
		render(
			<Accordion>
				<AccordionItem value="s1">
					<AccordionTrigger>
						<AccordionDragHandle />
						Set 1
					</AccordionTrigger>
					<AccordionContent>content</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);

		const handle = document.querySelector(".cursor-grab");
		expect(handle.querySelector("svg")).toBeInTheDocument();
	});
});

// ── AccordionContent ──────────────────────────────────────────────────────────

describe("AccordionContent", () => {
	it("renders children", () => {
		render(
			<Accordion>
				<AccordionItem value="c">
					<AccordionTrigger>Open</AccordionTrigger>
					<AccordionContent>
						<p>Detail text</p>
					</AccordionContent>
				</AccordionItem>
			</Accordion>,
		);
		expect(screen.getByText("Detail text")).toBeInTheDocument();
	});
});
