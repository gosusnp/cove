// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { createContext } from "preact";
import { useContext } from "preact/hooks";
import * as RadixAccordion from "@radix-ui/react-accordion";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { cn } from "../../lib/utils";

// Passes drag listeners/attributes from AccordionItem down to AccordionDragHandle.
const DragCtx = createContext(null);

// ── Accordion ─────────────────────────────────────────────────────────────────

export function Accordion({
	type = "multiple",
	class: className,
	children,
	...props
}) {
	return (
		<RadixAccordion.Root
			type={type}
			className={cn(
				"rounded-xl overflow-hidden",
				"border border-(--color-border)",
				"bg-(--color-surface)",
				className,
			)}
			{...props}
		>
			{children}
		</RadixAccordion.Root>
	);
}

// ── AccordionItem ─────────────────────────────────────────────────────────────

// When `id` is provided the item integrates with @dnd-kit/sortable.
// The consumer must place this inside a <SortableContext> (and a <DndContext>).
export function AccordionItem({
	id,
	value,
	class: className,
	children,
	...props
}) {
	if (id != null) {
		return (
			<SortableAccordionItem id={id} value={value} class={className} {...props}>
				{children}
			</SortableAccordionItem>
		);
	}

	return (
		<DragCtx.Provider value={null}>
			<RadixAccordion.Item
				value={value}
				className={cn(
					"border-b border-(--color-border) last:border-b-0",
					className,
				)}
				{...props}
			>
				{children}
			</RadixAccordion.Item>
		</DragCtx.Provider>
	);
}

function SortableAccordionItem({
	id,
	value,
	class: className,
	children,
	...props
}) {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id });

	return (
		<DragCtx.Provider value={{ listeners, attributes }}>
			<RadixAccordion.Item
				ref={setNodeRef}
				value={value ?? id}
				className={cn(
					"border-b border-(--color-border) last:border-b-0",
					isDragging && "z-10 shadow-md opacity-80",
					className,
				)}
				style={{
					transform: CSS.Transform.toString(transform),
					transition,
				}}
				{...props}
			>
				{children}
			</RadixAccordion.Item>
		</DragCtx.Provider>
	);
}

// ── AccordionTrigger ──────────────────────────────────────────────────────────

export function AccordionTrigger({ class: className, children, ...props }) {
	return (
		<RadixAccordion.Header>
			<RadixAccordion.Trigger
				className={cn(
					"group flex w-full items-center gap-2 px-4 py-3",
					"text-sm font-medium text-(--color-text)",
					"hover:bg-(--color-bg) transition-colors",
					"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-(--color-accent) focus-visible:ring-inset",
					className,
				)}
				{...props}
			>
				{children}
				<ChevronIcon class="ml-auto shrink-0 transition-transform duration-200 group-data-[state=open]:rotate-180" />
			</RadixAccordion.Trigger>
		</RadixAccordion.Header>
	);
}

// ── AccordionContent ──────────────────────────────────────────────────────────

export function AccordionContent({ class: className, children, ...props }) {
	return (
		<RadixAccordion.Content
			className={cn(
				"overflow-hidden text-sm",
				"data-[state=open]:animate-accordion-down data-[state=closed]:animate-accordion-up",
				className,
			)}
			{...props}
		>
			<div class="px-4 pb-3 pt-1">{children}</div>
		</RadixAccordion.Content>
	);
}

// ── AccordionDragHandle ───────────────────────────────────────────────────────

// A drag handle button for reordering AccordionItems. Place it inside
// AccordionTrigger; it picks up drag listeners from the nearest AccordionItem.
export function AccordionDragHandle({ class: className, ...props }) {
	const ctx = useContext(DragCtx);

	return (
		<button
			type="button"
			class={cn(
				"flex items-center justify-center w-6 h-6 shrink-0 rounded",
				"text-(--color-muted) hover:text-(--color-text) transition-colors",
				"cursor-grab active:cursor-grabbing",
				"focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-(--color-accent)",
				className,
			)}
			aria-label="Drag to reorder"
			{...(ctx?.listeners ?? {})}
			{...(ctx?.attributes ?? {})}
			{...props}
		>
			<GripIcon />
		</button>
	);
}

// ── Icons ─────────────────────────────────────────────────────────────────────

function ChevronIcon({ class: className }) {
	return (
		<svg
			class={className}
			width="16"
			height="16"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M4 6l4 4 4-4"
				stroke="currentColor"
				stroke-width="1.5"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	);
}

function GripIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 14 14"
			fill="none"
			aria-hidden="true"
		>
			<circle cx="5" cy="3" r="1" fill="currentColor" />
			<circle cx="9" cy="3" r="1" fill="currentColor" />
			<circle cx="5" cy="7" r="1" fill="currentColor" />
			<circle cx="9" cy="7" r="1" fill="currentColor" />
			<circle cx="5" cy="11" r="1" fill="currentColor" />
			<circle cx="9" cy="11" r="1" fill="currentColor" />
		</svg>
	);
}
