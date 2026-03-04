// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @dnd-kit/core.
// dnd-kit uses pointer events and browser layout APIs not available in jsdom.

export function DndContext({ children }) {
	return children;
}

export function DragOverlay({ children }) {
	return children ?? null;
}

export function closestCenter() {
	return null;
}

export function closestCorners() {
	return null;
}

export function KeyboardSensor() {}
export function PointerSensor() {}
export function MouseSensor() {}
export function TouchSensor() {}

export function useSensor() {
	return {};
}

export function useSensors(...sensors) {
	return sensors;
}

export function useDroppable() {
	return { isOver: false, setNodeRef: () => {} };
}

export function useDraggable() {
	return {
		attributes: {},
		listeners: {},
		setNodeRef: () => {},
		transform: null,
		isDragging: false,
	};
}

export function useDndMonitor() {}
