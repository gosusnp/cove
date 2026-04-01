// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useLocation } from "preact-iso";
import { useAuth } from "../Auth.jsx";
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
import { Row, Section } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ToggleGroup } from "../components/ui/ToggleGroup.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { timeAgo } from "../lib/utils";
import { apiFetch } from "../lib/api.js";

function initials(user) {
	if (user.display_name) {
		return user.display_name
			.split(" ")
			.map((p) => p[0])
			.join("")
			.toUpperCase()
			.slice(0, 2);
	}
	if (user.first_name || user.last_name) {
		return [user.first_name?.[0], user.last_name?.[0]]
			.filter(Boolean)
			.join("")
			.toUpperCase();
	}
	if (user.email) {
		return user.email[0].toUpperCase();
	}
	return "?";
}

function displayLabel(user) {
	if (user?.display_name) return user.display_name;
	if (user?.first_name || user?.last_name)
		return [user.first_name, user.last_name].filter(Boolean).join(" ");
	return null;
}

export function Settings() {
	const { user, logout, updateUser } = useAuth();
	const { route } = useLocation();

	// ── User profile ────────────────────────────────────────────────────
	useEffect(() => {
		if (!user) return;
		apiFetch("/api/users/me")
			.then((r) => {
				if (r.ok) return r.json().then(updateUser);
			})
			.catch(() => {});
	}, []);

	// ── Name fields ──────────────────────────────────────────────────────
	const displayName = useSignal(user?.display_name ?? "");
	const firstName = useSignal(user?.first_name ?? "");
	const lastName = useSignal(user?.last_name ?? "");

	// ── Unit fields ───────────────────────────────────────────────────────
	const fitnessUnit = useSignal(user?.fitness_unit_system ?? "metric");
	const cookingUnit = useSignal(user?.cooking_unit_system ?? "metric");

	useEffect(() => {
		if (!user) return;
		displayName.value = user.display_name ?? "";
		firstName.value = user.first_name ?? "";
		lastName.value = user.last_name ?? "";
		fitnessUnit.value = user.fitness_unit_system ?? "metric";
		cookingUnit.value = user.cooking_unit_system ?? "metric";
	}, [user?.id]);

	async function saveName(field, signal) {
		const prev = user?.[field] ?? "";
		const r = await apiFetch("/api/users/me", {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ [field]: signal.value.trim() || null }),
		});
		if (r.ok) updateUser(await r.json());
		else signal.value = prev;
	}

	async function saveUnit(field, signal, value) {
		const prev = signal.value;
		signal.value = value;
		const r = await apiFetch("/api/users/me", {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ [field]: value }),
		});
		if (r.ok) updateUser(await r.json());
		else signal.value = prev;
	}

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
		if (!user) return;
		apiFetch("/api/users/tokens")
			.then((r) => (r.ok ? r.json() : []))
			.then((data) => {
				tokens.value = data;
			})
			.catch(() => {})
			.finally(() => {
				tokensLoading.value = false;
			});
	}, [!!user]);

	// ── Sessions ─────────────────────────────────────────────────────────
	const sessions = useSignal([]);
	const sessionsLoading = useSignal(true);
	const sessionsError = useSignal(false);

	useEffect(() => {
		if (!user) return;
		apiFetch("/api/users/sessions")
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
	}, [!!user]);

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
			const r = await apiFetch("/api/users/tokens", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
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
		await apiFetch(`/api/users/tokens/${id}`, {
			method: "DELETE",
		});
	}

	async function handleDeleteSession(id) {
		const s = sessions.value.find((sess) => sess.id === id);
		if (s?.is_current) {
			handleSignOut();
			return;
		}
		const r = await apiFetch(`/api/users/sessions/${id}`, {
			method: "DELETE",
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
		<main class="flex flex-1 flex-col gap-8 px-4 py-6 max-w-lg mx-auto w-full">
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
						{displayLabel(user) && (
							<span class="font-medium" style={{ color: "var(--color-text)" }}>
								{displayLabel(user)}
							</span>
						)}
						<span class="text-sm" style={{ color: "var(--color-muted)" }}>
							{user?.email ?? "—"}
						</span>
					</div>
				</div>
				<Row label="Display name">
					<TextField
						value={displayName.value}
						placeholder="e.g. Jon Doe"
						onInput={(e) => {
							displayName.value = e.target.value;
						}}
						onBlur={() => saveName("display_name", displayName)}
					/>
				</Row>
				<Row label="First name">
					<TextField
						value={firstName.value}
						placeholder="First"
						onInput={(e) => {
							firstName.value = e.target.value;
						}}
						onBlur={() => saveName("first_name", firstName)}
					/>
				</Row>
				<Row label="Last name" last>
					<TextField
						value={lastName.value}
						placeholder="Last"
						onInput={(e) => {
							lastName.value = e.target.value;
						}}
						onBlur={() => saveName("last_name", lastName)}
					/>
				</Row>
			</Section>

			<Section title="Units">
				<Row label="Fitness weight">
					<ToggleGroup
						value={fitnessUnit.value}
						onChange={(v) => saveUnit("fitness_unit_system", fitnessUnit, v)}
						options={[
							{ value: "metric", label: "kg" },
							{ value: "imperial", label: "lb" },
						]}
					/>
				</Row>
				<Row label="Cooking" last>
					<ToggleGroup
						value={cookingUnit.value}
						onChange={(v) => saveUnit("cooking_unit_system", cookingUnit, v)}
						options={[
							{ value: "metric", label: "Metric" },
							{ value: "imperial", label: "Imperial" },
							{ value: "us_customary", label: "US (cups)" },
						]}
					/>
				</Row>
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

			<Section
				title="API Tokens"
				action={
					<Button
						variant="outline"
						size="sm"
						onClick={openCreateDialog}
						disabled={tokensLoading.value}
					>
						Create
					</Button>
				}
			>
				{tokens.value.length === 0 && !tokensLoading.value ? (
					<Row label="No tokens yet" last />
				) : (
					tokens.value.map((t, i) => (
						<Row
							key={t.id}
							label={t.name}
							sublabel={`Created ${new Date(t.created_at).toLocaleDateString()} · Last used ${t.last_used_at ? timeAgo(t.last_used_at) : "never"}`}
							last={i === tokens.value.length - 1}
						>
							<Button
								variant="destructive"
								size="sm"
								onClick={() => handleDelete(t.id)}
							>
								Delete
							</Button>
						</Row>
					))
				)}
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
							sublabel={`Active ${timeAgo(s.last_used_at ?? s.created_at)} · ${s.last_ip_masked ?? s.initial_ip_masked ?? "Unknown IP"}`}
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
							<div class="mt-4 flex flex-col gap-1.5">
								<label
									for="created-token"
									class="text-sm font-medium"
									style={{ color: "var(--color-text)" }}
								>
									Your new token
								</label>
								<div class="flex gap-2">
									<TextField
										id="created-token"
										value={createdToken.value}
										readOnly
										class="font-mono text-xs"
									/>
									<Button
										variant="outline"
										onClick={handleCopy}
										class="shrink-0"
									>
										{copied.value ? "Copied!" : "Copy"}
									</Button>
								</div>
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
