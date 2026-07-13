// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import {
	BoulderingTracker,
	serializeBoulderingEntries,
} from "./BoulderingTracker.jsx";

// ─── serializeBoulderingEntries ───────────────────────────────────────────────

describe("serializeBoulderingEntries", () => {
	it("returns empty string for no entries", () => {
		expect(serializeBoulderingEntries([])).toBe("");
	});

	it("formats a single send", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: [], type: "send" },
			]),
		).toBe("- V5: (1 Send)");
	});

	it("uses plural for multiple sends", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: [], type: "send" },
				{ id: 2, grade: "V5", labels: [], type: "send" },
			]),
		).toBe("- V5: (2 Sends)");
	});

	it("uses plural for multiple attempts", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: [], type: "attempt" },
				{ id: 2, grade: "V5", labels: [], type: "attempt" },
			]),
		).toBe("- V5: (2 Attempts)");
	});

	it("orders counts as attempts then sends then work", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: [], type: "send" },
				{ id: 2, grade: "V5", labels: [], type: "attempt" },
				{ id: 3, grade: "V5", labels: [], type: "work" },
			]),
		).toBe("- V5: (1 Attempt, 1 Send, 1 Work)");
	});

	it("includes title-cased labels in the key", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: ["overhang"], type: "send" },
			]),
		).toBe("- V5 Overhang: (1 Send)");
	});

	it("sorts multiple labels alphabetically in the key", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V7", labels: ["power", "overhang"], type: "attempt" },
			]),
		).toBe("- V7 Overhang Power: (1 Attempt)");
	});

	it("groups entries that share the same grade and labels", () => {
		expect(
			serializeBoulderingEntries([
				{ id: 1, grade: "V5", labels: ["overhang"], type: "attempt" },
				{ id: 2, grade: "V5", labels: ["overhang"], type: "attempt" },
				{ id: 3, grade: "V5", labels: ["overhang"], type: "send" },
			]),
		).toBe("- V5 Overhang: (2 Attempts, 1 Send)");
	});

	it("produces a separate line for each distinct grade+label group", () => {
		const result = serializeBoulderingEntries([
			{ id: 1, grade: "V5", labels: ["overhang"], type: "attempt" },
			{ id: 2, grade: "V7", labels: ["power"], type: "send" },
		]);
		expect(result).toContain("- V5 Overhang: (1 Attempt)");
		expect(result).toContain("- V7 Power: (1 Send)");
	});
});

// ─── BoulderingTracker component ─────────────────────────────────────────────

function renderTracker(initialEntries = []) {
	const entriesSignal = signal(initialEntries);
	render(<BoulderingTracker entriesSignal={entriesSignal} />);
	return entriesSignal;
}

describe("BoulderingTracker", () => {
	it("renders the Bouldering heading", () => {
		renderTracker();
		expect(screen.getByText("Bouldering")).toBeInTheDocument();
	});

	it("renders grade buttons V0 through V17", () => {
		renderTracker();
		for (let i = 0; i <= 17; i++) {
			expect(screen.getByRole("button", { name: `V${i}` })).toBeInTheDocument();
		}
	});

	it("renders style tag buttons", () => {
		renderTracker();
		for (const label of [
			"Overhang",
			"Slab",
			"Cave",
			"Power",
			"Dyno",
			"Crimps",
			"Slopers",
			"Pockets",
			"Pinches",
		]) {
			expect(screen.getByRole("button", { name: label })).toBeInTheDocument();
		}
	});

	it("appends a send entry when Log Send is clicked", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
		expect(entries.value).toHaveLength(1);
		expect(entries.value[0].type).toBe("send");
		expect(entries.value[0].grade).toBe("V5"); // V5 is the component default
	});

	it("appends an attempt entry when Log Attempt is clicked", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Log Attempt" }));
		expect(entries.value).toHaveLength(1);
		expect(entries.value[0].type).toBe("attempt");
	});

	it("appends a work entry when Log Work is clicked", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Log Work" }));
		expect(entries.value).toHaveLength(1);
		expect(entries.value[0].type).toBe("work");
	});

	it("uses the selected grade when logging", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "V9" }));
		fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
		expect(entries.value[0].grade).toBe("V9");
	});

	it("includes selected style labels when logging", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Overhang" }));
		fireEvent.click(screen.getByRole("button", { name: "Power" }));
		fireEvent.click(screen.getByRole("button", { name: "Log Attempt" }));
		expect(entries.value[0].labels).toEqual(
			expect.arrayContaining(["overhang", "power"]),
		);
	});

	it("displays the grade and type for each logged entry", () => {
		renderTracker([
			{ id: 1, grade: "V6", labels: [], type: "send" },
			{ id: 2, grade: "V8", labels: [], type: "attempt" },
		]);
		// Grade also appears in the strip buttons, so use getAllByText
		expect(screen.getAllByText("V6").length).toBeGreaterThanOrEqual(1);
		expect(screen.getAllByText("V8").length).toBeGreaterThanOrEqual(1);
		// Type labels are unique to entry rows ("Send" ≠ button label "Log Send")
		expect(screen.getByText("Send")).toBeInTheDocument();
		expect(screen.getByText("Attempt")).toBeInTheDocument();
	});

	it("displays labels for entries that have them", () => {
		renderTracker([
			{ id: 1, grade: "V5", labels: ["overhang", "slab"], type: "work" },
		]);
		expect(screen.getByText("Overhang · Slab")).toBeInTheDocument();
	});

	it("removes an entry when the ✕ button is clicked", () => {
		const entries = renderTracker([
			{ id: 1, grade: "V5", labels: [], type: "send" },
		]);
		fireEvent.click(screen.getByRole("button", { name: "Remove entry" }));
		expect(entries.value).toHaveLength(0);
	});

	it("only removes the targeted entry when multiple are present", () => {
		const entries = renderTracker([
			{ id: 1, grade: "V5", labels: [], type: "send" },
			{ id: 2, grade: "V7", labels: [], type: "attempt" },
		]);
		const removeButtons = screen.getAllByRole("button", {
			name: "Remove entry",
		});
		fireEvent.click(removeButtons[0]);
		expect(entries.value).toHaveLength(1);
		expect(entries.value[0].id).toBe(2);
	});

	it("retains grade selection after logging (last-used behavior)", () => {
		renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "V9" }));
		fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
		// V9 should still be pressed after logging
		expect(screen.getByRole("button", { name: "V9" })).toHaveAttribute(
			"aria-pressed",
			"true",
		);
	});

	it("retains style label selection after logging (last-used behavior)", () => {
		renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Overhang" }));
		fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
		expect(screen.getByRole("button", { name: "Overhang" })).toHaveAttribute(
			"aria-pressed",
			"true",
		);
	});

	it("deselects a style label when clicked again", () => {
		const entries = renderTracker();
		fireEvent.click(screen.getByRole("button", { name: "Overhang" }));
		fireEvent.click(screen.getByRole("button", { name: "Overhang" })); // toggle off
		fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
		expect(entries.value[0].labels).toHaveLength(0);
	});
});
