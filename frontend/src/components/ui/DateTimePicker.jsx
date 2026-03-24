// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils";

const fieldClass = (className, inline) =>
	cn(
		"text-sm placeholder:text-(--color-muted) transition-colors disabled:opacity-50 disabled:pointer-events-none",
		inline
			? "bg-transparent border-0 rounded-md px-1 -mx-1 hover:bg-(--color-bg) focus:bg-(--color-bg) focus:outline-none focus:ring-0 focus:ring-offset-0 text-right cursor-pointer"
			: "w-full h-10 rounded-lg border px-3 bg-(--color-surface) border-(--color-border) text-(--color-text) focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-(--color-accent)",
		className,
	);

export function DateTimePicker({
	label,
	id,
	inline,
	containerClass,
	inputRef,
	class: className,
	...props
}) {
	return (
		<div class={cn("flex flex-col gap-1.5", containerClass)}>
			{label && (
				<label
					for={id}
					class="text-sm font-medium"
					style={{ color: "var(--color-text)" }}
				>
					{label}
				</label>
			)}
			<input
				type="datetime-local"
				ref={inputRef}
				id={id}
				class={fieldClass(
					cn(
						!inline &&
							"read-only:bg-(--color-bg) read-only:cursor-default read-only:focus:ring-0 read-only:focus:ring-offset-0",
						className,
					),
					inline,
				)}
				{...props}
			/>
		</div>
	);
}
