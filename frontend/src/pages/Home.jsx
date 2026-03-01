// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { PageTitle } from "../components/ui/PageTitle.jsx";

export function Home() {
	return (
		<main class="flex flex-1 flex-col items-center justify-center gap-3 px-4">
			<PageTitle class="text-6xl tracking-tight">Cove</PageTitle>
			<p class="text-sm" style={{ color: "var(--color-muted)" }}>
				Your space.
			</p>
		</main>
	);
}
