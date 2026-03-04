// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { renderHook, act } from "@testing-library/preact";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { useSortableGroups, HOVER_OPEN_DELAY } from "./useSortableGroups.js";

// ── Test data ─────────────────────────────────────────────────────────────────

const SETS = [
	{
		id: "s1",
		name: "Lower",
		exercises: [
			{ id: "e1", name: "Squat" },
			{ id: "e2", name: "RDL" },
		],
	},
	{
		id: "s2",
		name: "Push",
		exercises: [
			{ id: "e3", name: "Bench" },
			{ id: "e4", name: "OHP" },
		],
	},
	{
		id: "s3",
		name: "Pull",
		exercises: [{ id: "e5", name: "Pull-ups" }],
	},
];

// ── Event object helpers ──────────────────────────────────────────────────────

// Builds a dnd-kit active/over object for an exercise being dragged.
function dragExercise(id, setId) {
	return { id, data: { current: { type: "exercise", setId } } };
}

// Builds a dnd-kit active/over object for a set being dragged.
function dragSet(id) {
	return { id, data: { current: {} } };
}

// Builds an over object for hovering over a set container (not an exercise).
function overSet(id) {
	return { id, data: { current: {} } };
}

// Builds an over object for hovering over a specific exercise.
function overExercise(id, setId) {
	return { id, data: { current: { type: "exercise", setId } } };
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function setup(overrides = {}) {
	return renderHook(() => useSortableGroups(SETS, overrides));
}

function exerciseIds(result, setId) {
	return result.current.sets
		.find((s) => s.id === setId)
		.exercises.map((e) => e.id);
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("useSortableGroups — initial state", () => {
	it("sets matches initialSets", () => {
		const { result } = setup();
		expect(result.current.sets).toEqual(SETS);
	});

	it("opens all sets by default", () => {
		const { result } = setup();
		expect(result.current.openValues).toEqual(["s1", "s2", "s3"]);
	});

	it("respects initialOpen option", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		expect(result.current.openValues).toEqual(["s1"]);
	});

	it("activeId starts as null", () => {
		const { result } = setup();
		expect(result.current.activeId).toBeNull();
	});
});

// ── handleDragStart ───────────────────────────────────────────────────────────

describe("useSortableGroups — handleDragStart", () => {
	it("records the active exercise id", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragStart({ active: dragExercise("e1", "s1") }),
		);
		expect(result.current.activeId).toBe("e1");
	});

	it("records the active set id", () => {
		const { result } = setup();
		act(() => result.current.handleDragStart({ active: dragSet("s2") }));
		expect(result.current.activeId).toBe("s2");
	});
});

// ── handleDragOver — no-op cases ─────────────────────────────────────────────

describe("useSortableGroups — handleDragOver no-ops", () => {
	it("ignores when over is null", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: null,
			}),
		);
		expect(result.current.sets).toBe(before);
	});

	it("ignores when active and over are the same item", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overExercise("e1", "s1"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});

	it("ignores when the active item is a set, not an exercise", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragOver({
				active: dragSet("s1"),
				over: overSet("s2"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});

	it("ignores when hovering within the same set", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});
});

// ── handleDragOver — drag into a closed set ───────────────────────────────────

describe("useSortableGroups — handleDragOver into closed set", () => {
	beforeEach(() => vi.useFakeTimers());
	afterEach(() => vi.useRealTimers());

	it("does NOT move the exercise while the destination is closed", () => {
		const { result } = setup({ initialOpen: ["s1"] }); // s2, s3 closed
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(exerciseIds(result, "s1")).toContain("e1");
		expect(exerciseIds(result, "s2")).not.toContain("e1");
	});

	it("opens the closed set after HOVER_OPEN_DELAY", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(result.current.openValues).not.toContain("s2");

		act(() => vi.advanceTimersByTime(HOVER_OPEN_DELAY));
		expect(result.current.openValues).toContain("s2");
	});

	it("does not open early — only after the full delay", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		act(() => vi.advanceTimersByTime(HOVER_OPEN_DELAY - 1));
		expect(result.current.openValues).not.toContain("s2");
	});

	it("re-uses the existing timer when hovering over the same closed set repeatedly", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		const clearSpy = vi.spyOn(globalThis, "clearTimeout");

		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		const callsBefore = clearSpy.mock.calls.length;

		// Second call with same destination — should NOT restart the timer.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(clearSpy.mock.calls.length).toBe(callsBefore);
		clearSpy.mockRestore();
	});

	it("cancels the timer when the exercise returns to its source set", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		// Start hovering over closed s2.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		// Move back to source s1.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		// Timer should have been cancelled — advancing past delay must NOT open s2.
		act(() => vi.advanceTimersByTime(HOVER_OPEN_DELAY));
		expect(result.current.openValues).not.toContain("s2");
	});

	it("cancels the previous timer when switching to a different closed set", () => {
		const { result } = setup({ initialOpen: ["s1"] }); // s2, s3 closed
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		// Switch to hovering over s3 before s2's timer fires.
		act(() => {
			vi.advanceTimersByTime(HOVER_OPEN_DELAY / 2);
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s3"),
			});
		});
		act(() => vi.advanceTimersByTime(HOVER_OPEN_DELAY));

		// s2 must NOT have opened (its timer was cancelled).
		expect(result.current.openValues).not.toContain("s2");
		// s3 MUST have opened (new timer ran to completion).
		expect(result.current.openValues).toContain("s3");
	});
});

// ── handleDragOver — drag into an open set ────────────────────────────────────

describe("useSortableGroups — handleDragOver into open set", () => {
	it("moves the exercise to the end when hovering over the set container", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		const ids = exerciseIds(result, "s2");
		expect(ids.at(-1)).toBe("e1");
	});

	it("inserts before the target exercise when hovering over a specific exercise", () => {
		const { result } = setup();
		// Hover e1 over e3 (index 0 in s2) → e1 should land at index 0 in s2.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overExercise("e3", "s2"),
			}),
		);
		const ids = exerciseIds(result, "s2");
		expect(ids.indexOf("e1")).toBe(0);
		expect(ids.indexOf("e3")).toBe(1);
	});

	it("removes the exercise from the source set", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(exerciseIds(result, "s1")).not.toContain("e1");
	});

	it("clears any pending open timer for the destination set", () => {
		vi.useFakeTimers();
		const { result } = setup({ initialOpen: ["s1"] }); // s2 starts closed
		// Start a hover timer for s2.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		// Now s2 opens (simulating the timer firing or the user manually opening it).
		act(() => result.current.setOpenValues(["s1", "s2"]));
		// Hover again — this time s2 is open, so the move should happen and timer clear.
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(exerciseIds(result, "s2")).toContain("e1");
		vi.useRealTimers();
	});
});

// ── handleDragEnd — general ───────────────────────────────────────────────────

describe("useSortableGroups — handleDragEnd general", () => {
	beforeEach(() => vi.useFakeTimers());
	afterEach(() => vi.useRealTimers());

	it("clears activeId", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragStart({ active: dragExercise("e1", "s1") }),
		);
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		expect(result.current.activeId).toBeNull();
	});

	it("clears a pending hover timer", () => {
		const { result } = setup({ initialOpen: ["s1"] });
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: null,
			}),
		);
		act(() => vi.advanceTimersByTime(HOVER_OPEN_DELAY));
		expect(result.current.openValues).not.toContain("s2");
	});

	it("is a no-op when over is null", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: null,
			}),
		);
		expect(result.current.sets).toBe(before);
	});

	it("is a no-op when active and over are the same item", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: overExercise("e1", "s1"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});
});

// ── handleDragEnd — set reorder ───────────────────────────────────────────────

describe("useSortableGroups — handleDragEnd set reorder", () => {
	it("moves a set forward in the list", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragEnd({
				active: dragSet("s1"),
				over: overSet("s3"),
			}),
		);
		expect(result.current.sets.map((s) => s.id)).toEqual(["s2", "s3", "s1"]);
	});

	it("moves a set backward in the list", () => {
		const { result } = setup();
		act(() =>
			result.current.handleDragEnd({
				active: dragSet("s3"),
				over: overSet("s1"),
			}),
		);
		expect(result.current.sets.map((s) => s.id)).toEqual(["s3", "s1", "s2"]);
	});
});

// ── handleDragEnd — exercise reorder ─────────────────────────────────────────

describe("useSortableGroups — handleDragEnd exercise reorder", () => {
	it("reorders two exercises within the same set", () => {
		const { result } = setup();
		// e1 is at index 0, e2 at index 1 in s1. Swap them.
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		expect(exerciseIds(result, "s1")).toEqual(["e2", "e1"]);
	});

	it("is a no-op when over.id is not an exercise in the active set (newIdx < 0)", () => {
		const { result } = setup();
		const before = result.current.sets;
		// over.id "s2" is not an exercise id — findIndex returns -1.
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});

	it("is a no-op when the exercise is not found in any set", () => {
		const { result } = setup();
		const before = result.current.sets;
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("nonexistent", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		expect(result.current.sets).toBe(before);
	});
});

// ── handleDragEnd — safety-net open ──────────────────────────────────────────

describe("useSortableGroups — handleDragEnd safety net", () => {
	it("opens the destination set after a same-container reorder (already open — idempotent)", () => {
		const { result } = setup();
		// s1 is already open; reorder e1→e2 within it.
		act(() =>
			result.current.handleDragStart({ active: dragExercise("e1", "s1") }),
		);
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s1"),
				over: overExercise("e2", "s1"),
			}),
		);
		// s1 must still be in openValues.
		expect(result.current.openValues).toContain("s1");
	});

	it("opens the landing set when it was closed at drop time", () => {
		// Start with s1 open, s2 closed.
		const { result } = setup({ initialOpen: ["s1"] });
		// Simulate a cross-container move: dragOver already moved e1 to s2 (open)
		// then s2 somehow got closed between dragOver and dragEnd.
		// We replicate this by: dragging to s2 while open, closing s2, then dropping.
		act(() => {
			// Open s2 so dragOver can proceed.
			result.current.setOpenValues(["s1", "s2"]);
		});
		act(() =>
			result.current.handleDragStart({ active: dragExercise("e1", "s1") }),
		);
		act(() =>
			result.current.handleDragOver({
				active: dragExercise("e1", "s1"),
				over: overSet("s2"),
			}),
		);
		// e1 is now in s2. Close s2 to simulate the edge case.
		act(() => result.current.setOpenValues(["s1"]));
		expect(result.current.openValues).not.toContain("s2");

		// Drop — safety net must reopen s2.
		act(() =>
			result.current.handleDragEnd({
				active: dragExercise("e1", "s2"),
				over: overSet("s2"),
			}),
		);
		expect(result.current.openValues).toContain("s2");
	});
});
