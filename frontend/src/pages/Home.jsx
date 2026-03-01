// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function Home() {
	return (
		<main class="flex flex-1 flex-col items-center justify-center gap-3 px-4">
			<h1
				class="text-6xl font-semibold tracking-tight"
				style={{ color: "var(--color-text)" }}
			>
				Cove
			</h1>
			<p class="text-sm" style={{ color: "var(--color-muted)" }}>
				Your space.
			</p>
		</main>
	);
}
