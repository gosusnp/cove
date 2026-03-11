// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils";

export function ToggleGroup({
	label,
	value,
	onChange,
	options,
	nullable = false,
	disabled = false,
	class: className,
}) {
	function handleClick(optionValue) {
		if (optionValue === value) {
			if (nullable) onChange(null);
		} else {
			onChange(optionValue);
		}
	}

	return (
		<div class={cn("flex flex-col gap-1.5", className)}>
			{label && (
				<span class="text-sm font-medium text-(--color-text)">
					{label}
				</span>
			)}
			<div class="flex flex-wrap">
				{options.map((option, index) => {
					const isActive = option.value === value;
					const isFirst = index === 0;
					const isLast = index === options.length - 1;

					return (
						<button
							key={option.value}
							type="button"
							onClick={() => handleClick(option.value)}
							disabled={disabled}
							class={cn(
								"h-10 px-4 text-sm font-medium border transition-colors",
								"focus-visible:outline-none focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-(--color-accent)",
								"disabled:opacity-50 disabled:cursor-not-allowed",
								isFirst && "rounded-l-lg",
								isLast && "rounded-r-lg",
								!isFirst && "-ml-px",
								isActive
									? "bg-(--color-accent) text-white border-(--color-accent) z-10"
									: "bg-(--color-surface) text-(--color-muted) border-(--color-border) hover:not-disabled:text-(--color-text)",
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
