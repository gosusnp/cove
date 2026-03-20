// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Trash2 } from "lucide-preact";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";
import { ActivityPicker } from "../components/shared/ActivityPicker.jsx";

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

export function SessionDetail({ sessionId, onDelete }) {
	const { user } = useAuth();
	const session = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");
	const deleteDialog = useDialog();

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
		s.program_name && { label: "Planned program", value: s.program_name },
		s.started_at && { label: "Started", value: formatDate(s.started_at) },
		s.completed_at && { label: "Completed", value: formatDate(s.completed_at) },
		s.duration_s && { label: "Duration", value: formatDuration(s.duration_s) },
		s.perceived_effort != null && {
			label: "Perceived effort",
			value: `${s.perceived_effort} / 10`,
		},
	].filter(Boolean);

	const pageTitle = s.program_name ?? s.activity ?? "Session";

	async function handleDelete() {
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "DELETE",
		});
		if (!r.ok) throw new Error("Failed to delete session");
		onDelete(sessionId);
	}

	async function saveNotes(notes) {
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ session_notes: notes || null }),
		});
		if (!r.ok) throw new Error("Failed to save notes");
		session.value = await r.json();
	}

	async function saveActivity(activity) {
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ activity: activity || null }),
		});
		if (!r.ok) throw new Error("Failed to save activity");
		session.value = await r.json();
	}

	return (
		<div class="max-w-2xl mx-auto px-4 py-6 flex flex-col gap-6">
			<div class="flex items-center justify-between gap-2">
				<PageTitle>{pageTitle}</PageTitle>
				<Tooltip>
					<TooltipTrigger>
						<Button
							variant="ghost"
							size="icon"
							aria-label="Delete session"
							onClick={deleteDialog.show}
						>
							<Trash2 size={16} aria-hidden="true" />
						</Button>
					</TooltipTrigger>
					<TooltipContent>Delete session</TooltipContent>
				</Tooltip>
			</div>

			<Section title="Overview">
				<Row label="Activity" last={overviewRows.length === 0}>
					<ActivityPicker
						value={s.activity ?? ""}
						onChange={saveActivity}
						label=""
						class="w-48"
					/>
				</Row>
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

			<Section title="Notes">
				<div class="px-4 py-3">
					<EditableMarkdown
						value={session.value.session_notes ?? null}
						placeholder="Add notes…"
						variant="plain"
						editLabel="Edit session notes"
						onSave={saveNotes}
					/>
				</div>
			</Section>

			{s.program_structure && (
				<Section title="Planned program">
					<div
						class="px-4 py-3 text-sm whitespace-pre-wrap"
						style={{ color: "var(--color-text)" }}
					>
						{s.program_structure}
					</div>
				</Section>
			)}
			<ConfirmDialog
				openSignal={deleteDialog.open}
				title="Delete Session"
				description="This will permanently delete the session. This cannot be undone."
				confirmLabel="Delete"
				onConfirm={handleDelete}
			/>
		</div>
	);
}
