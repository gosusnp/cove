// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function ListDetail({ list, detail, emptyState, hasDetail }) {
	return (
		<div class="flex h-[calc(100vh-var(--nav-h-mobile))] md:h-[calc(100vh-var(--nav-h-desktop))]">
			{/* Left panel — list */}
			<div
				class={[
					"shrink-0 overflow-y-auto",
					"w-full md:w-[260px]",
					"border-r border-(--color-border)",
					hasDetail ? "hidden md:block" : "block",
				].join(" ")}
				style={{ background: "var(--color-surface)" }}
			>
				{list}
			</div>

			{/* Right panel — detail or empty state */}
			<div
				class={[
					"flex-1 overflow-y-auto",
					hasDetail
						? "block"
						: "hidden md:flex md:items-center md:justify-center",
				].join(" ")}
				style={{ background: "var(--color-bg)" }}
			>
				{detail ?? (
					<p
						class="hidden md:block text-sm"
						style={{ color: "var(--color-muted)" }}
					>
						{emptyState}
					</p>
				)}
			</div>
		</div>
	);
}
