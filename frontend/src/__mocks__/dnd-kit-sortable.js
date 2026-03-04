// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @dnd-kit/sortable.
// dnd-kit uses pointer events and browser layout APIs not available in jsdom.

export function useSortable() {
	return {
		attributes: { role: "button", tabIndex: 0 },
		listeners: {},
		setNodeRef: () => {},
		transform: null,
		transition: undefined,
		isDragging: false,
	};
}

export function SortableContext({ children }) {
	return children;
}

export function arrayMove(arr, from, to) {
	const next = [...arr];
	const [item] = next.splice(from, 1);
	next.splice(to, 0, item);
	return next;
}

export const verticalListSortingStrategy = () => null;
export const rectSortingStrategy = () => null;
