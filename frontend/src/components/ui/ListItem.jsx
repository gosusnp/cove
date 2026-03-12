// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function ListItem({
	label,
	sublabel,
	active,
	isLast,
	onClick,
	actions,
}) {
	return (
		<div
			class="flex items-center justify-between gap-2"
			style={{
				borderBottom: isLast ? undefined : "1px solid var(--color-border)",
				background: active
					? "color-mix(in srgb, var(--color-accent) 10%, transparent)"
					: undefined,
				borderLeft: active
					? "3px solid var(--color-accent)"
					: "3px solid transparent",
			}}
		>
			<button
				type="button"
				class="flex-1 min-w-0 text-left px-4 py-3 cursor-pointer bg-transparent border-none flex flex-col gap-0.5"
				onClick={onClick}
			>
				<span
					class="text-sm font-medium truncate"
					style={{ color: "var(--color-text)" }}
				>
					{label}
				</span>
				{sublabel && (
					<span class="text-xs" style={{ color: "var(--color-muted)" }}>
						{sublabel}
					</span>
				)}
			</button>
			{actions && <div class="flex gap-1 shrink-0 pr-2">{actions}</div>}
		</div>
	);
}
