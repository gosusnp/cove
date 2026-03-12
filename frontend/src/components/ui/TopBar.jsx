// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function TopBar({ brand, children }) {
	return (
		<header
			class="hidden md:flex fixed top-0 inset-x-0 z-50 items-stretch justify-between"
			style={{
				height: "var(--nav-h-desktop)",
				background: "var(--color-surface)",
				borderBottom: "1px solid var(--color-border)",
			}}
		>
			<div
				class="shrink-0 flex items-center"
				style={{ width: "var(--sidebar-w-desktop)" }}
			>
				{brand}
			</div>
			{children}
		</header>
	);
}
