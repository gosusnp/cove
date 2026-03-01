// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils";

export function TextField({ label, id, class: className, ...props }) {
	return (
		<div class="flex flex-col gap-1.5">
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
				id={id}
				class={cn(
					"h-10 w-full rounded-lg border px-3 text-sm",
					"bg-(--color-surface) text-(--color-text)",
					"border-(--color-border)",
					"placeholder:text-(--color-muted)",
					"transition-colors",
					"focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-(--color-accent)",
					"disabled:opacity-50 disabled:pointer-events-none",
					"read-only:bg-(--color-bg) read-only:cursor-default read-only:focus:ring-0 read-only:focus:ring-offset-0",
					className,
				)}
				{...props}
			/>
		</div>
	);
}
