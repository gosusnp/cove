// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils.js";

// TagSelector renders a set of independent pill-shaped toggle buttons for
// multi-select use cases (e.g. session labels). Each tag can be toggled on
// and off independently; any combination is valid.
//
// Props:
//   label    — optional text label rendered above the pills
//   value    — string[] of currently selected values
//   onChange — called with the new string[] after a toggle
//   options  — [{ value: string, label: string }]
//   disabled — disables all buttons
//   class    — extra classes on the root element
export function TagSelector({
	label,
	value = [],
	onChange,
	options = [],
	disabled = false,
	class: className,
}) {
	function handleClick(optionValue) {
		if (disabled) return;
		if (value.includes(optionValue)) {
			onChange(value.filter((v) => v !== optionValue));
		} else {
			onChange([...value, optionValue]);
		}
	}

	return (
		<div class={cn("flex flex-col gap-1.5", className)}>
			{label && (
				<span class="text-sm font-medium text-(--color-text)">{label}</span>
			)}
			<div class="flex flex-wrap gap-2">
				{options.map((option) => {
					const isActive = value.includes(option.value);
					return (
						<button
							key={option.value}
							type="button"
							aria-pressed={isActive}
							onClick={() => handleClick(option.value)}
							disabled={disabled}
							class={cn(
								"h-7 px-3 rounded-full text-xs font-medium border transition-colors touch-manipulation",
								"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-(--color-accent)",
								"disabled:opacity-50 disabled:pointer-events-none",
								isActive
									? "bg-(--color-accent) text-white border-(--color-accent)"
									: "bg-transparent text-(--color-muted) border-(--color-border) hover:text-(--color-text) hover:border-(--color-text)",
							)}
						>
							{option.label}
						</button>
					);
				})}
			</div>
		</div>
	);
}
