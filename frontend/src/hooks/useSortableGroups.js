// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useState, useRef } from "preact/hooks";
import { arrayMove } from "@dnd-kit/sortable";

// Delay (ms) before a collapsed group auto-opens while an item hovers over it.
export const HOVER_OPEN_DELAY = 600;

// Manages state and DnD event handlers for sortable groups of sortable items
// (e.g. accordion sets each containing exercises). Consumers wire the returned
// handlers to a <DndContext> and the returned state to an <Accordion> — see
// ExerciseSetsDemo for the canonical usage.
export function useSortableGroups(initialSets, { initialOpen } = {}) {
	const [sets, setSets] = useState(initialSets);
	const [openValues, setOpenValues] = useState(
		() => initialOpen ?? initialSets.map((s) => s.id),
	);
	const [activeId, setActiveId] = useState(null);

	// Pending auto-open timer: { setId, timerId } | null
	const hoverTimer = useRef(null);
	// Exercise's current location during drag, used by the safety-net in handleDragEnd.
	const dragLocation = useRef(null);

	function openSet(setId) {
		setOpenValues((prev) => (prev.includes(setId) ? prev : [...prev, setId]));
	}

	function clearHoverTimer() {
		if (hoverTimer.current) {
			clearTimeout(hoverTimer.current.timerId);
			hoverTimer.current = null;
		}
	}

	// ── Handlers ──────────────────────────────────────────────────────────────

	function handleDragStart({ active }) {
		setActiveId(active.id);
		// Record where the exercise starts so the safety-net in handleDragEnd
		// always knows which set to keep open, even with no cross-container move.
		const set = sets.find((s) => s.exercises.some((e) => e.id === active.id));
		dragLocation.current = set
			? { exerciseId: active.id, setId: set.id }
			: null;
	}

	function handleDragOver({ active, over }) {
		if (!over || active.id === over.id) return;

		const activeType = active.data.current?.type;
		if (activeType !== "exercise") return;

		const activeSet = sets.find((s) =>
			s.exercises.some((e) => e.id === active.id),
		);
		if (!activeSet) return;

		// Resolve destination set: hovering over an exercise → use its setId,
		// hovering over a set container directly → use the set's own id.
		const overType = over.data.current?.type;
		const overSetId =
			overType === "exercise" ? over.data.current?.setId : over.id;

		if (!overSetId || activeSet.id === overSetId) {
			// Staying within the same set — cancel any pending open timer.
			clearHoverTimer();
			return;
		}

		if (!openValues.includes(overSetId)) {
			// Destination is collapsed. Schedule auto-open but don't move the exercise
			// yet — moving into a closed set would make it invisible.
			if (hoverTimer.current?.setId !== overSetId) {
				clearHoverTimer();
				hoverTimer.current = {
					setId: overSetId,
					timerId: setTimeout(() => {
						hoverTimer.current = null;
						setOpenValues((prev) =>
							prev.includes(overSetId) ? prev : [...prev, overSetId],
						);
					}, HOVER_OPEN_DELAY),
				};
			}
			return;
		}

		// Destination is open — clear any pending timer and perform the move.
		clearHoverTimer();
		dragLocation.current = { exerciseId: active.id, setId: overSetId };

		setSets((prev) => {
			const src = prev.find((s) => s.id === activeSet.id);
			const dest = prev.find((s) => s.id === overSetId);
			if (!src || !dest) return prev;

			const exercise = src.exercises.find((e) => e.id === active.id);
			const overIndex = dest.exercises.findIndex((e) => e.id === over.id);
			const insertAt = overIndex >= 0 ? overIndex : dest.exercises.length;

			return prev.map((s) => {
				if (s.id === activeSet.id) {
					return {
						...s,
						exercises: s.exercises.filter((e) => e.id !== active.id),
					};
				}
				if (s.id === overSetId) {
					const next = [...s.exercises];
					next.splice(insertAt, 0, exercise);
					return { ...s, exercises: next };
				}
				return s;
			});
		});
	}

	function handleDragEnd({ active, over }) {
		clearHoverTimer();
		const location = dragLocation.current;
		dragLocation.current = null;
		setActiveId(null);

		if (!over || active.id === over.id) return;

		const activeType = active.data.current?.type;

		if (!activeType) {
			// Set reorder.
			setSets((prev) => {
				const oldIdx = prev.findIndex((s) => s.id === active.id);
				const newIdx = prev.findIndex((s) => s.id === over.id);
				if (oldIdx < 0 || newIdx < 0) return prev;
				return arrayMove(prev, oldIdx, newIdx);
			});
			return;
		}

		if (activeType === "exercise") {
			// Same-container exercise reorder.
			// Cross-container moves were already applied optimistically in handleDragOver.
			setSets((prev) => {
				const set = prev.find((s) =>
					s.exercises.some((e) => e.id === active.id),
				);
				if (!set) return prev;
				const oldIdx = set.exercises.findIndex((e) => e.id === active.id);
				const newIdx = set.exercises.findIndex((e) => e.id === over.id);
				if (oldIdx < 0 || newIdx < 0) return prev;
				return prev.map((s) =>
					s.id === set.id
						? { ...s, exercises: arrayMove(s.exercises, oldIdx, newIdx) }
						: s,
				);
			});

			// Safety net: ensure the set where the exercise landed is visible.
			// Covers the case where the destination was closed at drop time.
			if (location) openSet(location.setId);
		}
	}

	return {
		sets,
		openValues,
		setOpenValues,
		activeId,
		handleDragStart,
		handleDragOver,
		handleDragEnd,
	};
}
