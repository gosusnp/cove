// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useLocation, useRoute } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { ListItem } from "../components/ui/ListItem.jsx";
import { SessionDetail } from "./SessionDetail.jsx";
import { apiFetch } from "../lib/api.js";

function formatDate(iso) {
	if (!iso) return null;
	return new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium" });
}

function SessionList({ sessions, selectedId, onSelect, onNew, error }) {
	return (
		<div class="flex flex-col">
			<div
				class="flex items-center justify-between px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Sessions
				</h2>
				<Button variant="primary" size="sm" onClick={onNew}>
					Start
				</Button>
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
					const label = s.program_name ?? s.activity ?? "Session";
					const sub = formatDate(s.started_at ?? s.created_at);
					return (
						<ListItem
							key={s.id}
							label={label}
							sublabel={sub}
							active={s.id === selectedId}
							isLast={i === sessions.length - 1}
							onClick={() => onSelect(s.id)}
						/>
					);
				})}
		</div>
	);
}

export function Sessions() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const sessions = useSignal([]);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	useEffect(() => {
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		apiFetch("/api/sessions")
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
	}, [!!user]);

	const handleSelect = (id) => {
		route(`/sessions/${id}`);
	};

	const handleNew = () => {
		route("/workout");
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
					onNew={handleNew}
					error={fetchError.value}
				/>
			}
			detail={selectedId ? <SessionDetail sessionId={selectedId} /> : null}
		/>
	);
}
