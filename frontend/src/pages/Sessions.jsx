// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { useLocation, useRoute } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { SessionDetail } from "./SessionDetail.jsx";

function formatDate(iso) {
	if (!iso) return null;
	return new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium" });
}

function SessionList({ sessions, selectedId, onSelect, error }) {
	return (
		<div class="flex flex-col">
			<div
				class="flex items-center px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Sessions
				</h2>
			</div>

			{error && (
				<p class="px-4 py-3 text-sm" style={{ color: "var(--color-error)" }}>
					{error}
				</p>
			)}

			{!error && sessions.length === 0 && (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No sessions yet.
				</p>
			)}

			{!error &&
				sessions.map((s, i) => {
					const isActive = s.id === selectedId;
					const label = s.program_name ?? s.activity ?? "Session";
					const sub = formatDate(s.started_at ?? s.created_at);
					return (
						<button
							key={s.id}
							type="button"
							class="w-full text-left px-4 py-3 cursor-pointer bg-transparent border-none flex flex-col gap-0.5"
							style={{
								borderBottom:
									i < sessions.length - 1
										? "1px solid var(--color-border)"
										: undefined,
								background: isActive
									? "color-mix(in srgb, var(--color-accent) 10%, transparent)"
									: undefined,
								borderLeft: isActive
									? "3px solid var(--color-accent)"
									: "3px solid transparent",
							}}
							onClick={() => onSelect(s.id)}
						>
							<span
								class="text-sm font-medium truncate"
								style={{ color: "var(--color-text)" }}
							>
								{label}
							</span>
							{sub && (
								<span class="text-xs" style={{ color: "var(--color-muted)" }}>
									{sub}
								</span>
							)}
						</button>
					);
				})}
		</div>
	);
}

export function Sessions() {
	const { token } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const sessions = useSignal([]);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	useEffect(() => {
		if (!token) return;
		loading.value = true;
		fetchError.value = "";
		fetch("/api/sessions", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => {
				if (!r.ok) throw new Error("Failed to fetch sessions");
				return r.json();
			})
			.then((data) => {
				sessions.value = data;
			})
			.catch((err) => {
				fetchError.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [token]);

	const handleSelect = (id) => {
		route(`/sessions/${id}`);
	};

	return (
		<ListDetail
			hasDetail={!!selectedId}
			emptyState="Select a session to view details."
			list={
				<SessionList
					sessions={sessions.value}
					selectedId={selectedId}
					onSelect={handleSelect}
					error={fetchError.value}
				/>
			}
			detail={selectedId ? <SessionDetail sessionId={selectedId} /> : null}
		/>
	);
}
