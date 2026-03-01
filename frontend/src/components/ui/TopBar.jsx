// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function TopBar({ children }) {
	return (
		<header
			class="hidden md:flex fixed top-0 inset-x-0 z-50 items-stretch justify-between px-6"
			style={{
				height: "var(--nav-h-desktop)",
				background: "var(--color-surface)",
				borderBottom: "1px solid var(--color-border)",
			}}
		>
			{children}
		</header>
	);
}
