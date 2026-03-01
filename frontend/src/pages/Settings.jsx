// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { useLocation } from "preact-iso";
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

function Section({ title, children }) {
	return (
		<section class="flex flex-col gap-3">
			<h2
				class="text-xs font-semibold uppercase tracking-widest"
				style={{ color: "var(--color-muted)" }}
			>
				{title}
			</h2>
			<div
				class="rounded-xl overflow-hidden"
				style={{
					background: "var(--color-surface)",
					border: "1px solid var(--color-border)",
				}}
			>
				{children}
			</div>
		</section>
	);
}

function Row({ label, children, last }) {
	return (
		<div
			class="flex items-center justify-between px-4 py-3 gap-4"
			style={last ? {} : { borderBottom: "1px solid var(--color-border)" }}
		>
			<span class="text-sm" style={{ color: "var(--color-text)" }}>
				{label}
			</span>
			<div
				class="flex items-center gap-2 text-sm"
				style={{ color: "var(--color-muted)" }}
			>
				{children}
			</div>
		</div>
	);
}

export function Settings() {
	const { user, token, logout, updateUser } = useAuth();
	const { route } = useLocation();

	useEffect(() => {
		if (!token) return;
		fetch("/api/users/me", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => {
				if (r.status === 401) {
					logout();
					route("/login");
					return;
				}
				return r.json().then(updateUser);
			})
			.catch(() => {});
	}, [token]);

	function handleSignOut() {
		logout();
		route("/login");
	}

	return (
		<main class="flex flex-1 flex-col gap-6 px-4 py-6 max-w-lg mx-auto w-full">
			<h1 class="text-2xl font-semibold" style={{ color: "var(--color-text)" }}>
				Settings
			</h1>

			<Section title="Profile">
				<div
					class="flex items-center gap-4 px-4 py-4"
					style={{ borderBottom: "1px solid var(--color-border)" }}
				>
					<div
						class="w-14 h-14 rounded-full flex items-center justify-center text-xl font-semibold shrink-0"
						style={{
							background: "var(--color-accent)",
							color: "var(--color-surface)",
						}}
					>
						{user ? initials(user) : "?"}
					</div>
					<div class="flex flex-col gap-0.5">
						{user?.name && (
							<span class="font-medium" style={{ color: "var(--color-text)" }}>
								{user.name}
							</span>
						)}
						<span class="text-sm" style={{ color: "var(--color-muted)" }}>
							{user?.email ?? "—"}
						</span>
					</div>
				</div>
				{user?.name && (
					<Row label="Display name" last>
						<span>{user.name}</span>
					</Row>
				)}
			</Section>

			<Section title="Connected Accounts">
				<Row label="Google" last>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						role="img"
						aria-label="Connected"
					>
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
						<polyline points="22 4 12 14.01 9 11.01" />
					</svg>
					Connected
				</Row>
			</Section>

			<Section title="Account">
				<Row label="Sign out" last>
					<button
						type="button"
						onClick={handleSignOut}
						class="text-sm font-medium px-3 py-1 rounded-lg transition-colors cursor-pointer"
						style={{
							color: "#dc2626",
							background: "#fef2f2",
							border: "1px solid #fecaca",
						}}
					>
						Sign out
					</button>
				</Row>
			</Section>
		</main>
	);
}
