// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { useLocation } from "preact-iso";
import { useAuth } from "../auth.jsx";
import { Avatar } from "../components/ui/Avatar.jsx";
import { Button } from "../components/ui/Button.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Section, Row } from "../components/ui/Section.jsx";
import { timeAgo } from "../lib/utils";
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";

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

export function Settings() {
	const { user, token, logout, updateUser } = useAuth();
	const { route } = useLocation();

	// ── User profile ────────────────────────────────────────────────────
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

	// ── Tokens ───────────────────────────────────────────────────────────
	const tokens = useSignal([]);
	const tokensLoading = useSignal(true);
	const createDialog = useDialog();
	const tokenName = useSignal("");
	const creating = useSignal(false);
	const createError = useSignal("");
	const createdToken = useSignal(null);
	const copied = useSignal(false);

	useEffect(() => {
		if (!token) return;
		fetch("/api/users/tokens", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => (r.ok ? r.json() : []))
			.then((data) => {
				tokens.value = data;
			})
			.catch(() => {})
			.finally(() => {
				tokensLoading.value = false;
			});
	}, [token]);

	// ── Sessions ─────────────────────────────────────────────────────────
	const sessions = useSignal([]);
	const sessionsLoading = useSignal(true);
	const sessionsError = useSignal(false);

	useEffect(() => {
		if (!token) return;
		fetch("/api/users/sessions", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then((data) => {
				sessions.value = data;
			})
			.catch(() => {
				sessionsError.value = true;
			})
			.finally(() => {
				sessionsLoading.value = false;
			});
	}, [token]);

	function openCreateDialog() {
		tokenName.value = "";
		createError.value = "";
		createdToken.value = null;
		copied.value = false;
		createDialog.show();
	}

	async function handleCreate(e) {
		e.preventDefault();
		const name = tokenName.value.trim();
		if (!name) {
			createError.value = "Token name is required.";
			return;
		}
		creating.value = true;
		createError.value = "";
		try {
			const r = await fetch("/api/users/tokens", {
				method: "POST",
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ name }),
			});
			if (!r.ok) {
				createError.value = "Failed to create token. Try again.";
				return;
			}
			const data = await r.json();
			tokens.value = [
				...tokens.value,
				{ id: data.id, name: data.name, created_at: data.created_at },
			];
			createdToken.value = data.token;
		} catch {
			createError.value = "Failed to create token. Try again.";
		} finally {
			creating.value = false;
		}
	}

	async function handleDelete(id) {
		tokens.value = tokens.value.filter((t) => t.id !== id);
		await fetch(`/api/users/tokens/${id}`, {
			method: "DELETE",
			headers: { Authorization: `Bearer ${token}` },
		});
	}

	async function handleDeleteSession(id) {
		const s = sessions.value.find((sess) => sess.id === id);
		if (s?.is_current) {
			handleSignOut();
			return;
		}
		const r = await fetch(`/api/users/sessions/${id}`, {
			method: "DELETE",
			headers: { Authorization: `Bearer ${token}` },
		});
		if (r.ok) {
			sessions.value = sessions.value.filter((sess) => sess.id !== id);
		}
	}

	function handleCopy() {
		navigator.clipboard.writeText(createdToken.value);
		copied.value = true;
		setTimeout(() => {
			copied.value = false;
		}, 2000);
	}

	function handleSignOut() {
		logout();
		route("/login");
	}

	return (
		<main class="flex flex-1 flex-col gap-6 px-4 py-6 max-w-lg mx-auto w-full">
			<PageTitle>Settings</PageTitle>

			<Section title="Profile">
				<div
					class="flex items-center gap-4 px-4 py-4"
					style={{ borderBottom: "1px solid var(--color-border)" }}
				>
					<Avatar
						initials={user ? initials(user) : "?"}
						label={user?.email}
						class="w-14 h-14 text-xl shrink-0"
					/>
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

			<Section title="API Tokens">
				{tokens.value.map((t) => (
					<Row
						key={t.id}
						label={t.name}
						sublabel={`Created ${new Date(t.created_at).toLocaleDateString()} · Last used ${t.last_used_at ? timeAgo(t.last_used_at) : "never"}`}
					>
						<Button
							variant="destructive"
							size="sm"
							onClick={() => handleDelete(t.id)}
						>
							Delete
						</Button>
					</Row>
				))}
				<Row label="New token" last>
					<Button
						variant="outline"
						size="sm"
						onClick={openCreateDialog}
						disabled={tokensLoading.value}
					>
						Create
					</Button>
				</Row>
			</Section>

			<Section title="Active Sessions">
				{sessionsLoading.value ? (
					<Row label="Loading…" last />
				) : sessionsError.value ? (
					<Row label="Could not load sessions." last />
				) : sessions.value.length === 0 ? (
					<Row label="No active sessions" last />
				) : (
					sessions.value.map((s, i) => (
						<Row
							key={s.id}
							label={`${s.last_browser ?? s.initial_browser ?? "Unknown Browser"} on ${s.last_os ?? s.initial_os ?? "Unknown OS"}${s.is_current ? " · Current" : ""}`}
							sublabel={`${s.last_ip_masked ?? s.initial_ip_masked ?? "Unknown IP"} · Active ${timeAgo(s.last_used_at ?? s.created_at)}`}
							last={i === sessions.value.length - 1}
						>
							<Button
								variant="destructive"
								size="sm"
								onClick={() => handleDeleteSession(s.id)}
							>
								{s.is_current ? "Sign out" : "Revoke"}
							</Button>
						</Row>
					))
				)}
			</Section>

			<Section title="Account">
				<Row label="Sign out" last>
					<Button variant="destructive" size="sm" onClick={handleSignOut}>
						Sign out
					</Button>
				</Row>
			</Section>

			<Dialog openSignal={createDialog.open}>
				<DialogContent>
					{createdToken.value === null ? (
						<form onSubmit={handleCreate}>
							<DialogTitle>New API Token</DialogTitle>
							<DialogDescription>
								Enter a name to identify this token.
							</DialogDescription>
							<div class="mt-4">
								<TextField
									id="token-name"
									label="Token name"
									placeholder="e.g. CI pipeline"
									value={tokenName.value}
									onInput={(e) => {
										tokenName.value = e.target.value;
									}}
									autoFocus
								/>
								{createError.value && (
									<p class="text-sm mt-2" style={{ color: "#dc2626" }}>
										{createError.value}
									</p>
								)}
							</div>
							<div class="mt-6 flex justify-end gap-2">
								<DialogClose>
									<Button variant="ghost" size="sm" type="button">
										Cancel
									</Button>
								</DialogClose>
								<Button size="sm" type="submit" disabled={creating.value}>
									{creating.value ? "Creating…" : "Create token"}
								</Button>
							</div>
						</form>
					) : (
						<>
							<DialogTitle>Token created</DialogTitle>
							<DialogDescription>
								Copy this token now — it won't be shown again.
							</DialogDescription>
							<div class="mt-4 flex gap-2 items-end">
								<TextField
									id="created-token"
									label="Your new token"
									value={createdToken.value}
									readOnly
									class="font-mono text-xs"
								/>
								<Button
									variant="outline"
									size="sm"
									onClick={handleCopy}
									class="shrink-0"
								>
									{copied.value ? "Copied!" : "Copy"}
								</Button>
							</div>
							<div class="mt-6 flex justify-end">
								<DialogClose>
									<Button size="sm">Done</Button>
								</DialogClose>
							</div>
						</>
					)}
				</DialogContent>
			</Dialog>
		</main>
	);
}
