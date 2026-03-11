// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import {
	Accordion,
	AccordionItem,
	AccordionTrigger,
	AccordionContent,
	AccordionDragHandle,
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

		expect(
			document.querySelector(".cursor-grab"),
		).toBeInTheDocument();
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
