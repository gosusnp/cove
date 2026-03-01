// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useAuth } from "../auth.jsx";

function initials(user) {
	if (user.name) {
		return user.name
			.split(" ")
			.map((p) => p[0])
			.join("")
			.toUpperCase()
			.slice(0, 2);
	}
	if (user.email) {
		return user.email[0].toUpperCase();
	}
	return "?";
}

export function Nav() {
	const { user } = useAuth();

	return (
		<header
			class="hidden md:flex fixed top-0 inset-x-0 z-50 items-center justify-between px-6"
			style={{
				height: "var(--nav-h-desktop)",
				background: "var(--color-surface)",
				borderBottom: "1px solid var(--color-border)",
			}}
		>
			<a
				href="/"
				class="text-base font-semibold tracking-tight"
				style={{ color: "var(--color-text)" }}
			>
				Cove
			</a>
			{user ? (
				<a
					href="/settings"
					class="flex items-center gap-2 rounded-lg px-2 py-1 transition-colors hover:opacity-80"
					aria-label="Account settings"
				>
					<div
						class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold shrink-0"
						style={{
							background: "var(--color-accent)",
							color: "var(--color-surface)",
						}}
					>
						{initials(user)}
					</div>
					<span
						class="text-sm hidden lg:block"
						style={{ color: "var(--color-muted)" }}
					>
						{user.email}
					</span>
				</a>
			) : (
				<a
					href="/login"
					class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors"
					style={{ color: "var(--color-accent)" }}
				>
					Sign in
				</a>
			)}
		</header>
	);
}
