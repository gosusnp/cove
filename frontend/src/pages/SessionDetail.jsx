// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { Trash2 } from "lucide-preact";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { DateTimePicker } from "../components/ui/DateTimePicker.jsx";
import { TextField } from "../components/ui/TextField.jsx";
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

function toDateTimeLocalValue(iso) {
	if (!iso) return "";
	const d = new Date(iso);
	const pad = (n) => String(n).padStart(2, "0");
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function formatDuration(seconds) {
	if (seconds == null) return null;
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	if (h > 0) return `${h}h ${m}m`;
	if (m > 0) return `${m}m ${s > 0 ? `${s}s` : ""}`.trim();
	return `${s}s`;
}

// Parse a duration string to seconds.
// Accepts "1h 30m", "45m", "1h 30m 15s", "1:30" (h:mm), "1:30:00" (h:mm:ss),
// or a plain integer treated as minutes.
function parseDuration(str) {
	if (!str?.trim()) return null;
	const s = str.trim();

	// H:MM or H:MM:SS
	const parts = s.split(":");
	if (parts.length >= 2 && parts.every((p) => /^\d+$/.test(p.trim()))) {
		const [h, m, sec = "0"] = parts;
		return parseInt(h, 10) * 3600 + parseInt(m, 10) * 60 + parseInt(sec, 10);
	}

	// Xh Ym Zs
	let total = 0;
	let matched = false;
	const hm = s.match(/(\d+)\s*h/i);
	const mm = s.match(/(\d+)\s*m(?!s)/i);
	const sm = s.match(/(\d+)\s*s/i);
	if (hm) {
		total += parseInt(hm[1], 10) * 3600;
		matched = true;
	}
	if (mm) {
		total += parseInt(mm[1], 10) * 60;
		matched = true;
	}
	if (sm) {
		total += parseInt(sm[1], 10);
		matched = true;
	}
	if (matched) return total > 0 ? total : null;

	// Plain number: treat as minutes
	if (/^\d+$/.test(s)) return parseInt(s, 10) * 60 || null;

	return null;
}

export function SessionDetail({ sessionId, onDelete }) {
	const { user } = useAuth();
	const session = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");
	const deleteDialog = useDialog();

	// Duration inline edit state.
	const durationEditing = useSignal(false);
	const durationInput = useSignal("");
	const durationRef = useRef(null);
	const durationSaveError = useSignal("");

	// Program name inline edit state.
	const programNameEditing = useSignal(false);
	const programNameInput = useSignal("");
	const programNameRef = useRef(null);
	const programNameSaveError = useSignal("");

	// Started / completed save error state.
	const startedAtSaveError = useSignal("");
	const completedAtSaveError = useSignal("");

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

	function startEditDuration() {
		durationInput.value = formatDuration(s.duration_s) ?? "";
		durationEditing.value = true;
		setTimeout(() => {
			durationRef.current?.focus();
			durationRef.current?.select();
		}, 0);
	}

	async function saveDuration() {
		const newSeconds = parseDuration(durationInput.value);
		const origSeconds = parseDuration(formatDuration(s.duration_s));
		durationEditing.value = false;
		if (newSeconds === origSeconds) return;
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ duration_s: newSeconds }),
		});
		if (r.ok) {
			session.value = await r.json();
		} else {
			durationSaveError.value = "Save failed";
			setTimeout(() => {
				durationSaveError.value = "";
			}, 3000);
		}
	}

	function startEditProgramName() {
		programNameInput.value = s.program_name ?? "";
		programNameEditing.value = true;
		setTimeout(() => {
			programNameRef.current?.focus();
			programNameRef.current?.select();
		}, 0);
	}

	async function saveProgramName() {
		const newName = programNameInput.value.trim() || null;
		programNameEditing.value = false;
		if (newName === (s.program_name ?? null)) return;
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ program_name: newName }),
		});
		if (r.ok) {
			session.value = await r.json();
		} else {
			programNameSaveError.value = "Save failed";
			setTimeout(() => {
				programNameSaveError.value = "";
			}, 3000);
		}
	}

	async function saveProgramStructure(structure) {
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ program_structure: structure || null }),
		});
		if (!r.ok) throw new Error("Failed to save program structure");
		session.value = await r.json();
	}

	async function saveStartedAt(value) {
		const newISO = value ? new Date(value).toISOString() : null;
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ started_at: newISO }),
		});
		if (r.ok) {
			session.value = await r.json();
		} else {
			startedAtSaveError.value = "Save failed";
			setTimeout(() => {
				startedAtSaveError.value = "";
			}, 3000);
		}
	}

	async function saveCompletedAt(value) {
		const newISO = value ? new Date(value).toISOString() : null;
		const r = await apiFetch(`/api/sessions/${sessionId}`, {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ completed_at: newISO }),
		});
		if (r.ok) {
			session.value = await r.json();
		} else {
			completedAtSaveError.value = "Save failed";
			setTimeout(() => {
				completedAtSaveError.value = "";
			}, 3000);
		}
	}

	const hasPerceivedEffort = s.perceived_effort != null;

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
				<Row label="Activity">
					<ActivityPicker
						value={s.activity ?? ""}
						onChange={saveActivity}
						label=""
						class="w-48"
					/>
				</Row>
				<Row label="Planned program">
					<div class="flex flex-col items-end gap-0.5">
						{programNameSaveError.value && (
							<span class="text-xs" style={{ color: "var(--color-error)" }}>
								{programNameSaveError.value}
							</span>
						)}
						{programNameEditing.value ? (
							<TextField
								inputRef={programNameRef}
								value={programNameInput.value}
								onInput={(e) => {
									programNameInput.value = e.target.value;
								}}
								onBlur={saveProgramName}
								onKeyDown={(e) => {
									if (e.key === "Enter") e.currentTarget.blur();
									else if (e.key === "Escape") programNameEditing.value = false;
								}}
								placeholder="Program name"
								class="w-40 text-right"
							/>
						) : (
							<Button
								variant="unstyled"
								aria-label="Edit planned program name"
								onClick={startEditProgramName}
								class="text-sm hover:underline"
								style={{ color: "var(--color-muted)" }}
							>
								{s.program_name ?? "—"}
							</Button>
						)}
					</div>
				</Row>
				<Row label="Duration">
					<div class="flex flex-col items-end gap-0.5">
						{durationSaveError.value && (
							<span class="text-xs" style={{ color: "var(--color-error)" }}>
								{durationSaveError.value}
							</span>
						)}
						{durationEditing.value ? (
							<TextField
								inputRef={durationRef}
								value={durationInput.value}
								onInput={(e) => {
									durationInput.value = e.target.value;
								}}
								onBlur={saveDuration}
								onKeyDown={(e) => {
									if (e.key === "Enter") e.currentTarget.blur();
									else if (e.key === "Escape") durationEditing.value = false;
								}}
								placeholder="e.g. 1h 30m"
								class="w-40 text-right"
							/>
						) : (
							<Button
								variant="unstyled"
								onClick={startEditDuration}
								class="text-sm hover:underline"
								style={{ color: "var(--color-muted)" }}
							>
								{formatDuration(s.duration_s) ?? "—"}
							</Button>
						)}
					</div>
				</Row>
				<Row label="Started">
					<div class="flex flex-col items-end gap-0.5">
						{startedAtSaveError.value && (
							<span class="text-xs" style={{ color: "var(--color-error)" }}>
								{startedAtSaveError.value}
							</span>
						)}
						<DateTimePicker
							inline
							aria-label="Started at"
							value={toDateTimeLocalValue(s.started_at)}
							onBlur={(e) => {
								if (e.target.value !== toDateTimeLocalValue(s.started_at)) {
									saveStartedAt(e.target.value);
								}
							}}
						/>
					</div>
				</Row>
				<Row label="Completed" last={!hasPerceivedEffort}>
					<div class="flex flex-col items-end gap-0.5">
						{completedAtSaveError.value && (
							<span class="text-xs" style={{ color: "var(--color-error)" }}>
								{completedAtSaveError.value}
							</span>
						)}
						<DateTimePicker
							inline
							aria-label="Completed at"
							value={toDateTimeLocalValue(s.completed_at)}
							onBlur={(e) => {
								if (e.target.value !== toDateTimeLocalValue(s.completed_at)) {
									saveCompletedAt(e.target.value);
								}
							}}
						/>
					</div>
				</Row>
				{hasPerceivedEffort && (
					<Row label="Perceived effort" last>
						{s.perceived_effort} / 10
					</Row>
				)}
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

			<Section title="Planned program">
				<div class="px-4 py-3">
					<EditableMarkdown
						value={session.value.program_structure ?? null}
						placeholder="Add program structure…"
						variant="plain"
						editLabel="Edit program structure"
						onSave={saveProgramStructure}
					/>
				</div>
			</Section>

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
