// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect, useRef } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { useLocation } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { SessionSummaryDialog } from "./SessionSummaryDialog.jsx";

// Formats elapsed seconds as HH:MM:SS.
function formatElapsed(totalSeconds) {
	const h = Math.floor(totalSeconds / 3600);
	const m = Math.floor((totalSeconds % 3600) / 60);
	const s = totalSeconds % 60;
	return [
		String(h).padStart(2, "0"),
		String(m).padStart(2, "0"),
		String(s).padStart(2, "0"),
	].join(":");
}

export function SessionTracker() {
	const { token } = useAuth();
	const { route } = useLocation();

	// Programs for optional selector.
	const programs = useSignal([]);
	const selectedProgramId = useSignal("");

	// Timer state.
	const elapsed = useSignal(0); // seconds
	const running = useSignal(false);
	const intervalRef = useRef(null);

	// Session state.
	const sessionId = useSignal(null); // null until session is created
	const saving = useSignal(false);
	const saveError = useSignal("");

	// Notes field.
	const notes = useSignal("");

	// Summary dialog state.
	const showSummary = useSignal(false);
	const perceivedEffort = useSignal(null); // 1–10 or null
	const summaryElapsed = useSignal(0); // elapsed captured when End Session is tapped
	const completedAtRef = useRef(null); // Date captured when End Session is tapped

	// Program hint strip state (v1 placeholder — expand logic deferred).
	const hintExpanded = useSignal(false);

	// Full program detail (fetched on selection to get structure field).
	const programDetail = useSignal(null);

	// Load programs list for the selector.
	useEffect(() => {
		if (!token) return;
		fetch("/api/programs", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then((data) => {
				programs.value = data;
			})
			.catch(() => {});
	}, [token]);

	// Fetch full program detail (including structure) when selection changes.
	useEffect(() => {
		programDetail.value = null;
		if (!token || !selectedProgramId.value) return;
		fetch(`/api/programs/${selectedProgramId.value}`, {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then((data) => {
				programDetail.value = data;
			})
			.catch(() => {});
	}, [token, selectedProgramId.value]);

	// Timer tick.
	useEffect(() => {
		if (running.value) {
			intervalRef.current = setInterval(() => {
				elapsed.value += 1;
			}, 1000);
		} else {
			clearInterval(intervalRef.current);
		}
		return () => clearInterval(intervalRef.current);
	}, [running.value]);

	// Start session: create server record, then start timer.
	async function handleStart() {
		if (!token) return;
		saving.value = true;
		saveError.value = "";
		try {
			const prog = programDetail.value;
			const body = {
				started_at: new Date().toISOString(),
				...(prog && {
					program_id: prog.id,
					program_name: prog.name,
					program_structure: prog.structure ?? null,
				}),
			};
			const r = await fetch("/api/sessions", {
				method: "POST",
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify(body),
			});
			if (!r.ok) throw new Error("Failed to start session");
			const data = await r.json();
			sessionId.value = data.id;
			running.value = true;
		} catch (err) {
			saveError.value = err.message;
		} finally {
			saving.value = false;
		}
	}

	function handlePause() {
		running.value = false;
	}

	function handleResume() {
		running.value = true;
	}

	// End Session tapped: pause timer and open summary dialog.
	function handleStopClick() {
		if (sessionId.value == null) return;
		running.value = false;
		completedAtRef.current = new Date();
		summaryElapsed.value = elapsed.value;
		showSummary.value = true;
	}

	// Cancel from summary dialog: resume timer.
	function handleSummaryCancel() {
		showSummary.value = false;
		running.value = true;
		completedAtRef.current = null;
		saveError.value = "";
	}

	// Save Session: patch with completed_at, duration, notes, effort, then navigate.
	async function handleSave() {
		if (!token || sessionId.value == null) return;
		saving.value = true;
		saveError.value = "";
		try {
			const prog = programDetail.value;
			const body = {
				completed_at: completedAtRef.current.toISOString(),
				duration_s: summaryElapsed.value,
				session_notes: notes.value || null,
				perceived_effort: perceivedEffort.value,
				...(prog && {
					program_id: prog.id,
					program_name: prog.name,
					program_structure: prog.structure ?? null,
				}),
			};
			const r = await fetch(`/api/sessions/${sessionId.value}`, {
				method: "PATCH",
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify(body),
			});
			if (!r.ok) throw new Error("Failed to save session");
			route(`/sessions/${sessionId.value}`);
		} catch (err) {
			saving.value = false;
			saveError.value = err.message;
		}
	}

	const notStarted = sessionId.value == null;
	const started = sessionId.value != null;

	const programOptions = programs.value.map((p) => ({
		value: String(p.id),
		label: p.name,
	}));

	return (
		<>
			<div
				class="flex flex-col min-h-dvh"
				style={{ background: "var(--color-bg)" }}
			>
				{/* Sticky timer zone */}
				<div
					class="sticky top-0 z-10 flex flex-col gap-4 px-4 py-5 border-b"
					style={{
						background: "var(--color-surface)",
						borderColor: "var(--color-border)",
					}}
				>
					{/* Elapsed time */}
					<div class="flex flex-col items-center gap-1">
						<output
							class="font-mono text-5xl font-semibold tabular-nums tracking-tight"
							style={{ color: "var(--color-text)", display: "block" }}
							aria-live="polite"
							aria-label={`Elapsed time ${formatElapsed(elapsed.value)}`}
						>
							{formatElapsed(elapsed.value)}
						</output>
						{started && (
							<span class="text-xs" style={{ color: "var(--color-muted)" }}>
								{running.value ? "Running" : "Paused"}
							</span>
						)}
					</div>

					{/* Controls */}
					<div class="flex gap-3 justify-center">
						{notStarted && (
							<Button
								variant="primary"
								size="lg"
								onClick={handleStart}
								disabled={saving.value}
							>
								Start
							</Button>
						)}
						{started && running.value && (
							<Button variant="outline" size="md" onClick={handlePause}>
								Pause
							</Button>
						)}
						{started && !running.value && (
							<Button variant="outline" size="md" onClick={handleResume}>
								Resume
							</Button>
						)}
						{started && (
							<Button
								variant="destructive"
								size="md"
								onClick={handleStopClick}
								disabled={saving.value}
							>
								End Session
							</Button>
						)}
					</div>
				</div>

				{/* Scrollable content zone */}
				<div class="flex flex-col gap-6 px-4 py-6 max-w-2xl mx-auto w-full">
					{/* Program selector — only shown before session starts */}
					{notStarted && (
						<div class="flex flex-col gap-2">
							<Combobox
								label="Program (optional)"
								value={selectedProgramId.value}
								onChange={(v) => {
									selectedProgramId.value = v;
								}}
								options={[
									{ value: "", label: "No program" },
									...programOptions,
								]}
								placeholder="Select a program…"
							/>
						</div>
					)}

					{/* Program hint strip (v1 placeholder) */}
					{started && programDetail.value && (
						<div
							class="rounded-xl border overflow-hidden"
							style={{
								background: "var(--color-surface)",
								borderColor: "var(--color-border)",
							}}
						>
							<button
								type="button"
								class="w-full flex items-center justify-between px-4 py-3 text-sm text-left"
								style={{ color: "var(--color-text)" }}
								onClick={() => {
									hintExpanded.value = !hintExpanded.value;
								}}
							>
								<span class="font-medium">
									Program: {programDetail.value.name}
								</span>
								<span style={{ color: "var(--color-muted)" }}>
									{hintExpanded.value ? "▲" : "▼"}
								</span>
							</button>
							{hintExpanded.value && (
								<div
									class="px-4 pb-4 text-sm whitespace-pre-wrap border-t"
									style={{
										borderColor: "var(--color-border)",
										color: "var(--color-text)",
										paddingTop: "0.75rem",
									}}
								>
									{programDetail.value.structure ?? (
										<span style={{ color: "var(--color-muted)" }}>
											No program details available.
										</span>
									)}
								</div>
							)}
						</div>
					)}

					{/* Notes */}
					<TextField
						id="session-notes"
						label="Notes"
						multiline
						value={notes.value}
						onInput={(e) => {
							notes.value = e.target.value;
						}}
						placeholder="What are you working on today?"
						rows={6}
					/>
				</div>
			</div>

			<SessionSummaryDialog
				openSignal={showSummary}
				completedAt={completedAtRef.current}
				elapsed={summaryElapsed.value}
				programName={programDetail.value?.name ?? null}
				notesSignal={notes}
				effortSignal={perceivedEffort}
				saving={saving.value}
				saveError={saveError.value}
				onCancel={handleSummaryCancel}
				onSave={handleSave}
			/>
		</>
	);
}
