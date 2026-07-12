// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { Button } from "../components/ui/Button.jsx";
import {
	Dialog,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { DateTimePicker } from "../components/ui/DateTimePicker.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ActivityPicker } from "../components/shared/ActivityPicker.jsx";
import { TagSelector } from "../components/ui/TagSelector.jsx";
import { apiFetch } from "../lib/api.js";
import { useSessionLabels } from "../hooks/useSessionLabels.js";

function toDateTimeLocalValue(d) {
	const pad = (n) => String(n).padStart(2, "0");
	return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// Parses a duration string to seconds.
// Accepts "1h 30m", "45m", "1h 30m 15s", "1:30" (h:mm), "1:30:00" (h:mm:ss),
// or a plain integer treated as minutes.
function parseDuration(str) {
	if (!str?.trim()) return null;
	const s = str.trim();

	const parts = s.split(":");
	if (parts.length >= 2 && parts.every((p) => /^\d+$/.test(p.trim()))) {
		const [h, m, sec = "0"] = parts;
		return parseInt(h, 10) * 3600 + parseInt(m, 10) * 60 + parseInt(sec, 10);
	}

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

	if (/^\d+$/.test(s)) return parseInt(s, 10) * 60 || null;

	return null;
}

// NewSessionDialog lets the user record a past session from the Sessions list.
// Props:
//   openSignal  — Preact signal controlling open state
//   onCreated   — called with the new session object after a successful POST
export function NewSessionDialog({ openSignal, onCreated }) {
	const name = useSignal("");
	const startedAt = useSignal(toDateTimeLocalValue(new Date()));
	const duration = useSignal("");
	const activity = useSignal("");
	const labels = useSignal([]);
	const saving = useSignal(false);
	const saveError = useSignal("");
	const availableLabels = useSessionLabels();

	function resetForm() {
		name.value = "";
		startedAt.value = toDateTimeLocalValue(new Date());
		duration.value = "";
		activity.value = "";
		labels.value = [];
		saveError.value = "";
	}

	function handleCancel() {
		openSignal.value = false;
		resetForm();
	}

	async function handleSave() {
		saving.value = true;
		saveError.value = "";
		try {
			const startDate = new Date(startedAt.value);
			const durationS = parseDuration(duration.value);
			const body = {
				started_at: startDate.toISOString(),
				labels: labels.value,
				...(name.value.trim() && { program_name: name.value.trim() }),
				...(durationS != null && {
					duration_s: durationS,
					completed_at: new Date(
						startDate.getTime() + durationS * 1000,
					).toISOString(),
				}),
				...(activity.value && { activity: activity.value }),
			};
			const r = await apiFetch("/api/sessions", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(body),
			});
			if (!r.ok) throw new Error("Failed to create session");
			const data = await r.json();
			openSignal.value = false;
			resetForm();
			onCreated(data);
		} catch (err) {
			saveError.value = err.message;
		} finally {
			saving.value = false;
		}
	}

	return (
		<Dialog openSignal={openSignal} onOpenChange={(v) => !v && handleCancel()}>
			<DialogContent>
				<DialogTitle>Log Session</DialogTitle>

				<div class="flex flex-col gap-5 mt-5">
					<TextField
						id="new-session-name"
						label="Name"
						placeholder="e.g. Morning run, Push day…"
						value={name.value}
						onInput={(e) => {
							name.value = e.target.value;
						}}
					/>
					<DateTimePicker
						id="new-session-started-at"
						label="Started"
						value={startedAt.value}
						onInput={(e) => {
							startedAt.value = e.target.value;
						}}
					/>
					<TextField
						id="new-session-duration"
						label="Duration"
						placeholder="e.g. 1h 30m, 45m, 1:30"
						value={duration.value}
						onInput={(e) => {
							duration.value = e.target.value;
						}}
					/>
					<ActivityPicker
						value={activity.value}
						onChange={(v) => {
							activity.value = v ?? "";
						}}
					/>
					{availableLabels.length > 0 && (
						<TagSelector
							label="Labels"
							value={labels.value}
							onChange={(v) => {
								labels.value = v;
							}}
							options={availableLabels}
						/>
					)}
					{saveError.value && (
						<p class="text-sm" style={{ color: "var(--color-error, #dc2626)" }}>
							{saveError.value}
						</p>
					)}
				</div>

				<div class="flex gap-3 justify-end mt-6">
					<Button variant="outline" size="md" onClick={handleCancel}>
						Cancel
					</Button>
					<Button
						variant="primary"
						size="md"
						onClick={handleSave}
						disabled={saving.value}
					>
						Log Session
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
