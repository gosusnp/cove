// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils";

export function Avatar({ initials, label, class: className }) {
	return (
		<div
			role="img"
			class={cn(
				"w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold shrink-0",
				className,
			)}
			style={{
				background: "var(--color-accent)",
				color: "var(--color-surface)",
			}}
			aria-label={label}
		>
			{initials}
		</div>
	);
}
