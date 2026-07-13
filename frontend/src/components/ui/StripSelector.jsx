// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect, useRef } from "preact/hooks";
import { cn } from "../../lib/utils.js";

// StripSelector renders a horizontally scrollable selector strip inside a
// visible track. Scrolling selects the item nearest to the center; tapping a
// visible item also selects it. The selected item is auto-centered on change.
//
// The visible window is controlled by the parent via the `class` prop (e.g.
// `class="w-72"`). Without a width constraint the track expands to fit all
// items, which defeats the scrolling affordance.
//
// Props:
//   label    — optional text label rendered above the strip
//   options  — string[] of selectable values
//   value    — currently selected value, or null
//   onChange — called with the new value string on scroll settle or tap
//   disabled — disables all interaction
//   class    — extra classes on the root element; use to set the visible width
export function StripSelector({
	label,
	options = [],
	value,
	onChange,
	disabled = false,
	class: className,
}) {
	const scrollRef = useRef(null);
	const innerRef = useRef(null);
	// Prevent the scrollend that fires after our own scrollTo from re-triggering
	// onChange. Without this, centering a selection causes a feedback loop:
	// scrollTo → scrollend → onChange → centering → scrollTo → ...
	const skipNextScrollEnd = useRef(false);

	// Set edge padding and center the selected item whenever value changes.
	//
	// Edge padding lives on the inner content div (not the scroll container) so
	// both leading and trailing space are part of the scrollable area.
	//
	// The inner div has position:relative, making it the offsetParent for its
	// buttons. btn.offsetLeft is then in scroll-content coordinates — the same
	// space as container.scrollLeft — so all arithmetic is in one frame.
	// Reading offsetLeft after setting paddingInline forces a synchronous
	// reflow, so the measurement reflects the updated layout.
	useEffect(() => {
		const container = scrollRef.current;
		const inner = innerRef.current;
		if (!container || !inner) return;

		const pad = Math.max(0, Math.floor(container.clientWidth / 2 - 22));
		inner.style.paddingInline = `${pad}px`;

		if (value == null) return;
		const idx = options.indexOf(value);
		if (idx === -1) return;
		const btn = inner.querySelectorAll("button")[idx];
		if (!btn || typeof container.scrollTo !== "function") return;

		const target = Math.max(
			0,
			btn.offsetLeft + btn.offsetWidth / 2 - container.clientWidth / 2,
		);
		// Only set the skip flag when we're actually going to move the scroll
		// position. If the container is already at the target, scrollend will
		// not fire and the flag would permanently suppress the next user scroll.
		// Threshold of 1px accounts for sub-pixel layout rounding.
		if (Math.abs(target - container.scrollLeft) >= 1) {
			skipNextScrollEnd.current = true;
		}
		container.scrollTo({ left: target, behavior: "smooth" });
	}, [value, options]);

	// Select the item nearest to center once scroll settles. Skips the
	// scrollend fired by our own programmatic centering scroll.
	// Note: scrollend requires Safari ≥ 17.4 / iOS ≥ 17.4. On older WKWebView
	// the handler simply never fires, so user-scroll selection is silently
	// unavailable. Tap-to-select still works on all browsers.
	useEffect(() => {
		const container = scrollRef.current;
		const inner = innerRef.current;
		if (!container || !inner) return;

		function onScrollEnd() {
			if (skipNextScrollEnd.current) {
				skipNextScrollEnd.current = false;
				return;
			}
			const containerCenter = container.scrollLeft + container.clientWidth / 2;
			const buttons = Array.from(inner.querySelectorAll("button"));
			let nearestIdx = 0;
			let nearestDist = Infinity;
			buttons.forEach((btn, i) => {
				const dist = Math.abs(
					btn.offsetLeft + btn.offsetWidth / 2 - containerCenter,
				);
				if (dist < nearestDist) {
					nearestDist = dist;
					nearestIdx = i;
				}
			});
			const nearest = options[nearestIdx];
			if (nearest !== undefined && nearest !== value) {
				onChange(nearest);
			}
		}

		container.addEventListener("scrollend", onScrollEnd, { passive: true });
		return () => container.removeEventListener("scrollend", onScrollEnd);
	}, [options, value, onChange]);

	return (
		<div class={cn("flex flex-col gap-1.5", className)}>
			{label && (
				<span class="text-sm font-medium text-(--color-text)">{label}</span>
			)}
			<div
				ref={scrollRef}
				class="overflow-x-auto rounded-xl border border-(--color-border) bg-(--color-surface)"
			>
				<div ref={innerRef} class="relative flex gap-1 py-1.5">
					{options.map((option) => {
						const isActive = value === option;
						return (
							<button
								key={option}
								type="button"
								aria-pressed={isActive}
								onClick={() => onChange(option)}
								disabled={disabled}
								class={cn(
									"h-9 min-w-[2.75rem] px-2 rounded-lg text-sm font-medium border transition-colors touch-manipulation shrink-0",
									"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-(--color-accent)",
									"disabled:opacity-50 disabled:pointer-events-none",
									isActive
										? "bg-(--color-accent) text-white border-(--color-accent)"
										: "bg-transparent text-(--color-muted) border-(--color-border) hover:text-(--color-text) hover:border-(--color-text)",
								)}
							>
								{option}
							</button>
						);
					})}
				</div>
			</div>
		</div>
	);
}
