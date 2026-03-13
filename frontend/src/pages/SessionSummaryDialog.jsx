// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import {
	Dialog,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { Button } from "../components/ui/Button.jsx";
import { TextField } from "../components/ui/TextField.jsx";

// Formats elapsed seconds as Xh YYm ZZs or Mm SSs.
function formatDuration(totalSeconds) {
	const h = Math.floor(totalSeconds / 3600);
	const m = Math.floor((totalSeconds % 3600) / 60);
	const s = totalSeconds % 60;
	if (h > 0) {
		return `${h}h ${String(m).padStart(2, "0")}m ${String(s).padStart(2, "0")}s`;
	}
	return `${m}m ${String(s).padStart(2, "0")}s`;
}

function formatDate(d) {
	return d.toLocaleDateString(undefined, {
		weekday: "long",
		month: "long",
		day: "numeric",
	});
}

function SummaryRow({ label, value }) {
	return (
		<div class="flex justify-between gap-4">
			<span style={{ color: "var(--color-muted)" }}>{label}</span>
			<span
				class="font-medium text-right"
				style={{ color: "var(--color-text)" }}
			>
				{value}
			</span>
		</div>
	);
}

// SessionSummaryDialog opens when the athlete taps "End Session".
// On mobile it covers the full screen; on sm+ it appears as a centered dialog.
// Props:
//   openSignal     — Preact signal controlling open state
//   completedAt    — Date captured when End Session was tapped
//   elapsed        — seconds elapsed at that moment (number)
//   programName    — string or null (structured session)
//   notesSignal    — Preact signal for freeform notes (shared with tracker)
//   effortSignal   — Preact signal for perceived effort (1–10 or null)
//   saving         — boolean
//   saveError      — string or ""
//   onCancel       — called when user cancels (resumes timer)
//   onSave         — called when user confirms (writes to DB)
export function SessionSummaryDialog({
	openSignal,
	completedAt,
	elapsed,
	programName,
	notesSignal,
	effortSignal,
	saving,
	saveError,
	onCancel,
	onSave,
}) {
	return (
		<Dialog
			openSignal={openSignal}
			onOpenChange={(v) => {
				if (!v) onCancel();
			}}
		>
			<DialogContent fullscreen>
				{/* Header */}
				<div
					class="flex items-center px-6 py-4 border-b shrink-0"
					style={{ borderColor: "var(--color-border)" }}
				>
					<DialogTitle>Session Summary</DialogTitle>
				</div>

				{/* Scrollable body */}
				<div class="flex flex-col gap-5 px-6 py-5 overflow-y-auto flex-1">
					{/* Read-only session info */}
					<div class="flex flex-col gap-2 text-sm">
						<SummaryRow
							label="Date"
							value={formatDate(completedAt ?? new Date())}
						/>
						<SummaryRow label="Duration" value={formatDuration(elapsed)} />
						{programName && <SummaryRow label="Program" value={programName} />}
					</div>

					<div
						class="border-t"
						style={{ borderColor: "var(--color-border)" }}
					/>

					{/* Perceived effort slider */}
					<div class="flex flex-col gap-3">
						<div class="flex items-center justify-between">
							<p
								class="text-sm font-medium"
								style={{ color: "var(--color-text)" }}
							>
								Perceived Effort{" "}
								<span style={{ color: "var(--color-muted)" }}>(optional)</span>
							</p>
							{effortSignal.value != null && (
								<span
									class="text-sm font-semibold tabular-nums"
									style={{ color: "var(--color-accent)" }}
								>
									{effortSignal.value} / 10
								</span>
							)}
						</div>
						<input
							type="range"
							min="1"
							max="10"
							step="1"
							value={effortSignal.value ?? ""}
							onInput={(e) => {
								effortSignal.value = Number(e.target.value);
							}}
							class="w-full accent-(--color-accent)"
							aria-label="Perceived effort 1 to 10"
						/>
						<div
							class="flex justify-between text-xs"
							style={{ color: "var(--color-muted)" }}
						>
							<span>Easy</span>
							<span>Max effort</span>
						</div>
					</div>

					<div
						class="border-t"
						style={{ borderColor: "var(--color-border)" }}
					/>

					{/* Notes */}
					{/* Speech-to-text placeholder: this textarea could accept voice input
					    via the Web Speech API (SpeechRecognition) in a future iteration. */}
					<TextField
						id="summary-notes"
						label="Notes"
						multiline
						value={notesSignal.value}
						onInput={(e) => {
							notesSignal.value = e.target.value;
						}}
						placeholder="How did it feel? Any PRs or observations?"
						rows={4}
					/>

					{saveError && (
						<p class="text-sm" style={{ color: "var(--color-error, #dc2626)" }}>
							{saveError}
						</p>
					)}
				</div>

				{/* Footer actions */}
				<div
					class="flex gap-3 justify-end px-6 py-4 border-t shrink-0"
					style={{ borderColor: "var(--color-border)" }}
				>
					<Button variant="outline" size="md" onClick={onCancel}>
						Cancel
					</Button>
					<Button
						variant="primary"
						size="md"
						onClick={onSave}
						disabled={saving}
					>
						Save Session
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
