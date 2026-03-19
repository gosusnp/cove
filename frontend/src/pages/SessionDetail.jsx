// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { apiFetch } from "../lib/api.js";

function PencilIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M11.013 1.427a1.75 1.75 0 0 1 2.474 0l1.086 1.086a1.75 1.75 0 0 1 0 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 0 1-.927-.928l.929-3.25c.081-.286.235-.547.445-.758l8.61-8.609Z"
				fill="currentColor"
			/>
		</svg>
	);
}

function formatDate(iso) {
	if (!iso) return null;
	return new Date(iso).toLocaleString(undefined, {
		dateStyle: "medium",
		timeStyle: "short",
	});
}

function formatDuration(seconds) {
	if (!seconds) return null;
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	if (h > 0) return `${h}h ${m}m`;
	if (m > 0) return `${m}m ${s > 0 ? `${s}s` : ""}`.trim();
	return `${s}s`;
}

export function SessionDetail({ sessionId }) {
	const { user } = useAuth();
	const session = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");
	const editingNotes = useSignal(false);
	const notesDraft = useSignal("");
	const notesSaving = useSignal(false);
	const notesError = useSignal("");

	useEffect(() => {
		if (!user || !sessionId) return;
		loading.value = true;
		error.value = "";
		apiFetch(`/api/sessions/${sessionId}`)
			.then((r) => {
				if (!r.ok) throw new Error("Failed to fetch session");
				return r.json();
			})
			.then((data) => {
				session.value = data;
			})
			.catch((err) => {
				error.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [sessionId, user]);

	if (loading.value) {
		return (
			<div class="flex items-center justify-center h-full">
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Loading…
				</p>
			</div>
		);
	}

	if (error.value) {
		return (
			<div class="flex items-center justify-center h-full">
				<p class="text-sm" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			</div>
		);
	}

	const s = session.value;
	if (!s) return null;

	const overviewRows = [
		s.activity && { label: "Activity", value: s.activity },
		s.program_name && { label: "Program", value: s.program_name },
		s.started_at && { label: "Started", value: formatDate(s.started_at) },
		s.completed_at && { label: "Completed", value: formatDate(s.completed_at) },
		s.duration_s && { label: "Duration", value: formatDuration(s.duration_s) },
		s.perceived_effort != null && {
			label: "Perceived effort",
			value: `${s.perceived_effort} / 10`,
		},
	].filter(Boolean);

	const title = s.program_name ?? s.activity ?? "Session";

	function startEditNotes() {
		notesDraft.value = session.value.session_notes ?? "";
		editingNotes.value = true;
	}

	function cancelEditNotes() {
		editingNotes.value = false;
		notesError.value = "";
	}

	async function saveNotes() {
		notesSaving.value = true;
		notesError.value = "";
		try {
			const r = await apiFetch(`/api/sessions/${sessionId}`, {
				method: "PATCH",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ session_notes: notesDraft.value || null }),
			});
			if (!r.ok) throw new Error("Failed to save notes");
			const updated = await r.json();
			session.value = updated;
			editingNotes.value = false;
		} catch (err) {
			notesError.value = err.message;
		} finally {
			notesSaving.value = false;
		}
	}

	return (
		<div class="max-w-2xl mx-auto px-4 py-6 flex flex-col gap-6">
			<PageTitle>{title}</PageTitle>

			{overviewRows.length > 0 && (
				<Section title="Overview">
					{overviewRows.map((row, i) => (
						<Row
							key={row.label}
							label={row.label}
							last={i === overviewRows.length - 1}
						>
							{row.value}
						</Row>
					))}
				</Section>
			)}

			<Section title="Notes">
				<div class="px-4 py-3">
					{editingNotes.value ? (
						<div class="flex flex-col gap-2">
							<TextField
								multiline
								rows={4}
								value={notesDraft.value}
								onInput={(e) => (notesDraft.value = e.target.value)}
								onKeyDown={(e) => {
									if (e.key === "Escape") cancelEditNotes();
								}}
							/>
							{notesError.value && (
								<p class="text-xs" style={{ color: "var(--color-error)" }}>
									{notesError.value}
								</p>
							)}
							<div class="flex justify-end gap-2">
								<Button
									size="sm"
									onClick={saveNotes}
									disabled={notesSaving.value}
								>
									{notesSaving.value ? "Saving…" : "Save"}
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={cancelEditNotes}
									disabled={notesSaving.value}
								>
									Cancel
								</Button>
							</div>
						</div>
					) : (
						<div class="group flex items-start justify-between w-full gap-2">
							{session.value.session_notes ? (
								<p
									class="text-sm whitespace-pre-wrap flex-1"
									style={{ color: "var(--color-text)" }}
								>
									{session.value.session_notes}
								</p>
							) : (
								<p
									class="text-sm italic"
									style={{ color: "var(--color-muted)", opacity: 0.5 }}
								>
									Add notes…
								</p>
							)}
							<Button
								variant="ghost"
								size="icon"
								onClick={startEditNotes}
								aria-label="Edit session notes"
								class="opacity-0 group-hover:opacity-30 transition-opacity shrink-0"
							>
								<PencilIcon />
							</Button>
						</div>
					)}
				</div>
			</Section>

			{s.program_structure && (
				<Section title="Program">
					<div
						class="px-4 py-3 text-sm whitespace-pre-wrap"
						style={{ color: "var(--color-text)" }}
					>
						{s.program_structure}
					</div>
				</Section>
			)}
		</div>
	);
}
