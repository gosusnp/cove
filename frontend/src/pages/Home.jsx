// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useAuth } from "../Auth.jsx";
import { TRAIN_NAV_ITEMS } from "../components/shared/Nav.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";

export function Home() {
	const { user } = useAuth();

	if (!user) {
		return (
			<main class="flex flex-1 flex-col items-center justify-center gap-3 px-4">
				<PageTitle class="text-6xl tracking-tight">Cove</PageTitle>
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Your space.
				</p>
			</main>
		);
	}

	return (
		<main class="flex flex-1 flex-col px-4 py-6 max-w-sm mx-auto w-full">
			<nav aria-label="Main navigation">
				{TRAIN_NAV_ITEMS.map(({ label, href }) => (
					<a
						key={href}
						href={href}
						class="flex items-center justify-between py-4 text-base font-medium border-b transition-colors hover:text-(--color-accent)"
						style={{
							color: "var(--color-text)",
							borderColor: "var(--color-border)",
							textDecoration: "none",
						}}
					>
						{label}
					</a>
				))}
			</nav>
		</main>
	);
}
