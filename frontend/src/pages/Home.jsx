// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useAuth } from "../Auth.jsx";
import { NAV_SECTIONS } from "../components/shared/Nav.jsx";
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
			<nav aria-label="Main navigation" class="flex flex-col gap-6">
				{NAV_SECTIONS.map(({ label, items }) => (
					<div key={label}>
						<h2
							class="px-1 pb-1 text-xs font-semibold uppercase tracking-widest"
							style={{ color: "var(--color-muted)" }}
						>
							{label}
						</h2>
						{items.map(({ label: itemLabel, href }) => (
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
								{itemLabel}
							</a>
						))}
					</div>
				))}
			</nav>
		</main>
	);
}
