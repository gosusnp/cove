// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
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

		const buttons = screen
			.getAllByRole("button")
			.filter((b) => b.hasAttribute("aria-pressed"));
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

	it("checks on swipe right past threshold", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		// First move: intent detection (> 6px horizontal)
		fireEvent.touchMove(btn, { touches: [{ clientX: 7, clientY: 0 }] });
		// Second move: past SWIPE_THRESHOLD (56px)
		fireEvent.touchMove(btn, { touches: [{ clientX: 60, clientY: 0 }] });
		fireEvent.touchEnd(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("true");
	});

	it("unchecks on swipe right when already checked", () => {
		render(<CheckListItem defaultChecked>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(btn, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(btn, { touches: [{ clientX: 60, clientY: 0 }] });
		fireEvent.touchEnd(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("false");
	});

	it("does not toggle on swipe below threshold", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(btn, { touches: [{ clientX: 7, clientY: 0 }] });
		// 30px — horizontal intent set but below 56px threshold
		fireEvent.touchMove(btn, { touches: [{ clientX: 30, clientY: 0 }] });
		fireEvent.touchEnd(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("false");
	});

	it("does not toggle on vertical swipe", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		// Clearly vertical — dy > dx
		fireEvent.touchMove(btn, { touches: [{ clientX: 2, clientY: 60 }] });
		fireEvent.touchEnd(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("false");
	});
});

describe("CheckListSection swipe-to-check-all", () => {
	it("checks all items on swipe right past threshold", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem>Push-ups × 10</CheckListItem>
				<CheckListItem>Pull-ups × 5</CheckListItem>
			</CheckListSection>,
		);
		const buttons = screen
			.getAllByRole("button")
			.filter((b) => b.hasAttribute("aria-pressed"));
		expect(
			buttons.every((b) => b.getAttribute("aria-pressed") === "false"),
		).toBe(true);

		// Swipe the section header (now a button itself)
		const header = screen.getByText("Strength").closest("[class*=sticky]");
		fireEvent.touchStart(header, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 60, clientY: 0 }] });
		fireEvent.touchEnd(header);

		expect(
			buttons.every((b) => b.getAttribute("aria-pressed") === "true"),
		).toBe(true);
	});

	it("translates the header during swipe and resets on release", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem>Push-ups × 10</CheckListItem>
			</CheckListSection>,
		);
		const header = screen.getByText("Strength").closest("[class*=sticky]");
		fireEvent.touchStart(header, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 40, clientY: 0 }] });
		expect(header.style.transform).toContain("translateX");
		fireEvent.touchEnd(header);
		expect(header.style.transform).toBe("");
	});

	it("does not translate header when all items are already done", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem defaultChecked>Push-ups × 10</CheckListItem>
			</CheckListSection>,
		);
		const header = screen.getByText("Strength").closest("[class*=sticky]");
		fireEvent.touchStart(header, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 60, clientY: 0 }] });
		fireEvent.touchEnd(header);
		expect(header.style.transform).toBe("");
	});

	it("does not check all on swipe below threshold", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem>Push-ups × 10</CheckListItem>
			</CheckListSection>,
		);
		const [btn] = screen
			.getAllByRole("button")
			.filter((b) => b.hasAttribute("aria-pressed"));
		const header = screen.getByText("Strength").closest("[class*=sticky]");
		fireEvent.touchStart(header, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(header, { touches: [{ clientX: 30, clientY: 0 }] });
		fireEvent.touchEnd(header);
		expect(btn.getAttribute("aria-pressed")).toBe("false");
	});

	it("calls preventDefault on the first horizontal touchmove to claim gesture before browser scrolls", () => {
		render(
			<CheckListSection label="Strength">
				<CheckListItem>Push-ups × 10</CheckListItem>
			</CheckListSection>,
		);
		const header = screen.getByText("Strength").closest("[class*=sticky]");
		fireEvent.touchStart(header, { touches: [{ clientX: 0, clientY: 0 }] });
		const spy = vi.spyOn(Event.prototype, "preventDefault");
		// First move — intent detection; preventDefault must fire here, not on the next event
		fireEvent.touchMove(header, { touches: [{ clientX: 7, clientY: 0 }] });
		expect(spy).toHaveBeenCalled();
		spy.mockRestore();
	});
});

describe("CheckListItem gesture claim", () => {
	it("calls preventDefault on first horizontal touchmove to claim gesture before browser scrolls", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		const spy = vi.spyOn(Event.prototype, "preventDefault");
		// First move — intent detection; preventDefault must fire here, not on the next event
		fireEvent.touchMove(btn, { touches: [{ clientX: 7, clientY: 0 }] });
		expect(spy).toHaveBeenCalled();
		spy.mockRestore();
	});

	it("does not call preventDefault on first vertical touchmove so the scroll container can scroll", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		const spy = vi.spyOn(Event.prototype, "preventDefault");
		fireEvent.touchMove(btn, { touches: [{ clientX: 2, clientY: 60 }] });
		expect(spy).not.toHaveBeenCalled();
		spy.mockRestore();
	});

	it("renders subtitle when provided", () => {
		render(
			<CheckListItem subtitle="50 kg · bilateral">Squat × 8</CheckListItem>,
		);
		expect(screen.getByText("50 kg · bilateral")).toBeTruthy();
	});

	it("renders an empty subtitle placeholder when subtitle is an empty string", () => {
		render(<CheckListItem subtitle="">Plank</CheckListItem>);
		// The subtitle span must exist even with empty text to maintain item height.
		const btn = screen.getByRole("button");
		const subtitleSpans = btn.querySelectorAll("span[style*='min-height']");
		expect(subtitleSpans.length).toBe(1);
	});

	it("does not render a subtitle span when subtitle prop is omitted", () => {
		render(<CheckListItem>Push-ups × 10</CheckListItem>);
		const btn = screen.getByRole("button");
		expect(btn.querySelectorAll("span[style*='min-height']").length).toBe(0);
	});

	it("applies line-through to subtitle when checked", () => {
		render(
			<CheckListItem defaultChecked subtitle="50 kg · bilateral">
				Squat × 8
			</CheckListItem>,
		);
		expect(screen.getByText("50 kg · bilateral").className).toContain(
			"line-through",
		);
	});

	it("does not toggle again when a synthesized click fires after a completed swipe", () => {
		render(<CheckListItem>Squat × 8</CheckListItem>);
		const btn = screen.getByRole("button");
		fireEvent.touchStart(btn, { touches: [{ clientX: 0, clientY: 0 }] });
		fireEvent.touchMove(btn, { touches: [{ clientX: 7, clientY: 0 }] });
		fireEvent.touchMove(btn, { touches: [{ clientX: 60, clientY: 0 }] });
		fireEvent.touchEnd(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("true");
		// Touch devices fire a synthetic click after touchend — it must not double-toggle
		fireEvent.click(btn);
		expect(btn.getAttribute("aria-pressed")).toBe("true");
	});
});
