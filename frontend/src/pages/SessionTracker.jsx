// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { useLocation } from "preact-iso";
import { KeepAwake } from "@capacitor-community/keep-awake";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import {
	CheckList,
	CheckListItem,
	CheckListSection,
} from "../components/ui/CheckList.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { SessionSummaryDialog } from "./SessionSummaryDialog.jsx";
import { apiFetch } from "../lib/api.js";
import { ActivityPicker } from "../components/shared/ActivityPicker.jsx";
import {
	convertFitnessWeight,
	useUnitPreferences,
} from "../hooks/useUnitPreferences.js";
import { useDialog } from "../hooks/useDialog.js";

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

// Flatten a program's sets × rounds into CheckList sections.
// Each section has a label and the list of exercises for that round.
function flattenProgram(program) {
	const sections = [];
	for (const set of program.sets ?? []) {
		const rounds = set.rounds ?? 1;
		for (let r = 1; r <= rounds; r++) {
			const roundPart = rounds > 1 ? `Round ${r} of ${rounds}` : null;
			const label = [set.name, roundPart].filter(Boolean).join(" · ");
			sections.push({ label, exercises: set.exercises ?? [] });
		}
	}
	return sections;
}

// Format the primary label for a program exercise.
// Quantity prefix (reps and/or duration) comes before the name.
function formatLabel(ex) {
	const prefix = [
		ex.reps ? `${ex.reps}x` : null,
		ex.duration_s ? `${ex.duration_s}s` : null,
	]
		.filter(Boolean)
		.join(" ");
	return prefix ? `${prefix} ${ex.name}` : ex.name;
}

// Format the subtitle line for a program exercise (weight and laterality).
// Returns an empty string when neither is set; the subtitle slot is always
// rendered to keep item height consistent and maximise the swipe target.
//
// This is a consumption context: weight is converted to the viewer's preferred
// unit regardless of the stored unit. Edit contexts (e.g. ProgramDetail) render
// the stored unit directly without conversion.
function formatSubtitle(ex, fitnessWeightUnit) {
	const parts = [];
	if (ex.weight) {
		const fromUnit = ex.weight_unit ?? "kg";
		const converted = convertFitnessWeight(
			ex.weight,
			fromUnit,
			fitnessWeightUnit,
		);
		parts.push(`${converted} ${fitnessWeightUnit}`);
	}
	if (ex.laterality) parts.push(ex.laterality);
	return parts.join(" · ");
}

// Build a synthetic program object for free-form sessions (no program selected).
// Produces a single set of 1 round with one exercise named after the activity,
// so the tracking screen has a consistent CheckList UI regardless of program use.
function syntheticProgram(activityName) {
	return {
		name: activityName,
		sets: [
			{
				name: activityName,
				rounds: 1,
				exercises: [{ name: activityName }],
			},
		],
	};
}

export function SessionTracker() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { fitnessWeightUnit } = useUnitPreferences();

	// Programs for optional selector.
	const programs = useSignal([]);
	const selectedProgramId = useSignal("");

	// Timer state.
	const elapsed = useSignal(0); // seconds
	const running = useSignal(false);
	const intervalRef = useRef(null);
	const segmentStartRef = useRef(null); // Date.now() when current segment began
	const accumulatedRef = useRef(0); // total seconds from all prior segments

	// Session state.
	const sessionId = useSignal(null); // null until session is created
	const saving = useSignal(false);
	const saveError = useSignal("");

	// Notes field.
	const notes = useSignal("");

	// Activity field (pre-filled from selected program).
	const activity = useSignal("");

	// Summary dialog state.
	const showSummary = useSignal(false);
	const perceivedEffort = useSignal(null); // 1–10 or null
	const summaryElapsed = useSignal(0); // elapsed captured when End Session is tapped
	const completedAtRef = useRef(null); // Date captured when End Session is tapped

	// Full program detail (fetched on selection to get structure field).
	const programDetail = useSignal(null);

	// Programs being tracked in the current session (for CheckList rendering).
	// Populated on Start and appended to via "Add Program".
	const sessionPrograms = useSignal([]);

	// Add Program dialog state.
	const addProgramDialog = useDialog();
	const addProgramId = useSignal("");
	const addingProgram = useSignal(false);
	const addProgramError = useSignal("");

	// Load programs list for the selector.
	useEffect(() => {
		if (!user) return;
		apiFetch("/api/programs")
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then((data) => {
				programs.value = data;
			})
			.catch(() => {});
	}, [!!user]);

	// Fetch full program detail (including structure) when selection changes.
	useEffect(() => {
		programDetail.value = null;
		if (!selectedProgramId.value) {
			activity.value = "";
			return;
		}
		apiFetch(`/api/programs/${selectedProgramId.value}`)
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then((data) => {
				programDetail.value = data;
				activity.value = data.activity ?? "";
			})
			.catch(() => {});
	}, [user, selectedProgramId.value]);

	// Wake lock: keep screen on while timer is running.
	useEffect(() => {
		if (running.value) {
			KeepAwake.keepAwake().catch(() => {});
			return () => {
				KeepAwake.allowSleep().catch(() => {});
			};
		}
	}, [running.value]);

	// Timer tick: compute elapsed from timestamp so backgrounding doesn't lose time.
	useEffect(() => {
		if (running.value) {
			intervalRef.current = setInterval(() => {
				if (!segmentStartRef.current) return;
				elapsed.value =
					Math.floor((Date.now() - segmentStartRef.current) / 1000) +
					accumulatedRef.current;
			}, 1000);
		} else {
			clearInterval(intervalRef.current);
		}
		return () => clearInterval(intervalRef.current);
	}, [running.value]);

	// Snap elapsed immediately when the app is foregrounded.
	useEffect(() => {
		function handleVisibility() {
			if (!document.hidden && running.value && segmentStartRef.current) {
				elapsed.value =
					Math.floor((Date.now() - segmentStartRef.current) / 1000) +
					accumulatedRef.current;
			}
		}
		document.addEventListener("visibilitychange", handleVisibility);
		return () =>
			document.removeEventListener("visibilitychange", handleVisibility);
	}, []);

	// Start session: create server record, then start timer.
	async function handleStart() {
		if (!user) return;
		saving.value = true;
		saveError.value = "";
		try {
			// If a program is selected but its detail hasn't loaded yet (e.g. the
			// user tapped Start before the background fetch resolved), fetch it now.
			let prog = programDetail.value;
			if (selectedProgramId.value && !prog) {
				const dr = await apiFetch(`/api/programs/${selectedProgramId.value}`);
				if (dr.ok) {
					prog = await dr.json();
					programDetail.value = prog;
					activity.value = prog.activity ?? "";
				}
			}
			const body = {
				started_at: new Date().toISOString(),
				...(prog && {
					program_id: prog.id,
					program_name: prog.name,
					program_structure: prog.structure ?? null,
				}),
			};
			const r = await apiFetch("/api/sessions", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(body),
			});
			if (!r.ok) throw new Error("Failed to start session");
			const data = await r.json();
			sessionId.value = data.id;
			segmentStartRef.current = Date.now();
			running.value = true;

			// Initialise the session program list for CheckList rendering.
			if (prog) {
				sessionPrograms.value = [prog];
			} else if (activity.value) {
				sessionPrograms.value = [syntheticProgram(activity.value)];
			}
		} catch (err) {
			saveError.value = err.message;
		} finally {
			saving.value = false;
		}
	}

	function handlePause() {
		accumulatedRef.current += Math.floor(
			(Date.now() - segmentStartRef.current) / 1000,
		);
		elapsed.value = accumulatedRef.current;
		segmentStartRef.current = null;
		running.value = false;
	}

	function handleResume() {
		segmentStartRef.current = Date.now();
		running.value = true;
	}

	// End Session tapped: pause timer and open summary dialog.
	function handleStopClick() {
		if (sessionId.value == null) return;
		const now = Date.now();
		if (segmentStartRef.current !== null) {
			accumulatedRef.current += Math.floor(
				(now - segmentStartRef.current) / 1000,
			);
		}
		elapsed.value = accumulatedRef.current;
		segmentStartRef.current = null;
		running.value = false;
		completedAtRef.current = new Date(now);
		summaryElapsed.value = accumulatedRef.current;
		showSummary.value = true;
	}

	// Cancel from summary dialog: resume timer.
	function handleSummaryCancel() {
		showSummary.value = false;
		segmentStartRef.current = Date.now();
		running.value = true;
		completedAtRef.current = null;
		saveError.value = "";
	}

	// Save Session: patch with completed_at, duration, notes, effort, then navigate.
	// Program name and structure are already up-to-date on the server (set on POST
	// and updated via PATCH on each mid-session add), so they are omitted here.
	async function handleSave() {
		if (sessionId.value == null) return;
		saving.value = true;
		saveError.value = "";
		try {
			const body = {
				completed_at: completedAtRef.current.toISOString(),
				duration_s: summaryElapsed.value,
				session_notes: notes.value || null,
				perceived_effort: perceivedEffort.value,
				activity: activity.value || null,
			};
			const r = await apiFetch(`/api/sessions/${sessionId.value}`, {
				method: "PATCH",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(body),
			});
			if (!r.ok) throw new Error("Failed to save session");
			route(`/sessions/${sessionId.value}`);
		} catch (err) {
			saving.value = false;
			saveError.value = err.message;
		}
	}

	// Add a program to the active session mid-workout.
	// Fetches the selected program, appends it to the local render list, and
	// PATCHes the session with accumulated program name and structure.
	async function handleAddProgram() {
		const pid = addProgramId.value;
		if (!pid || !sessionId.value) return;
		addingProgram.value = true;
		addProgramError.value = "";
		try {
			const r = await apiFetch(`/api/programs/${pid}`);
			if (!r.ok) throw new Error("Failed to load program");
			const prog = await r.json();

			const allPrograms = [...sessionPrograms.value, prog];
			const accName = allPrograms.map((p) => p.name).join(", ");
			const accStructure = allPrograms
				.map((p) => p.structure)
				.filter(Boolean)
				.join("\n\n");

			const patch = { program_name: accName };
			if (accStructure) patch.program_structure = accStructure;

			const pr = await apiFetch(`/api/sessions/${sessionId.value}`, {
				method: "PATCH",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(patch),
			});
			if (!pr.ok) throw new Error("Failed to update session");

			sessionPrograms.value = allPrograms;
			addProgramId.value = "";
			addProgramDialog.hide();
		} catch (err) {
			addProgramError.value = err.message;
		} finally {
			addingProgram.value = false;
		}
	}

	const notStarted = sessionId.value == null;
	const started = sessionId.value != null;

	const programOptions = programs.value.map((p) => ({
		value: String(p.id),
		label: p.name,
	}));

	// Accumulated program name across all session programs (for summary dialog).
	const accumulatedProgramName =
		sessionPrograms.value.length > 0
			? sessionPrograms.value.map((p) => p.name).join(", ")
			: null;

	return (
		<>
			<div class="page-fixed" style={{ background: "var(--color-bg)" }}>
				{/* Timer zone — flex-none so it never scrolls or shrinks */}
				<div
					class="flex-none flex flex-col gap-4 px-4 py-5 border-b"
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

				{/*
				 * Scrollable content zone — flex-1 min-h-0 overflow-y-auto is the
				 * scroll container. CheckList sticky headers use sticky top-0 which
				 * sticks them to the top of this div (visually just below the timer).
				 */}
				<div
					class="flex-1 min-h-0 overflow-y-auto"
					style={{ touchAction: "pan-y" }}
				>
					<div class="page-content">
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
								<ActivityPicker
									value={activity.value || null}
									onChange={(v) => {
										activity.value = v;
									}}
								/>
							</div>
						)}

						{/* Program tracker — one CheckList per program in the session */}
						{started && sessionPrograms.value.length > 0 && (
							<div class="flex flex-col gap-4">
								{sessionPrograms.value.map((prog, pi) => (
									<div key={pi} class="flex flex-col gap-2">
										{sessionPrograms.value.length > 1 && (
											<p
												class="text-xs font-medium uppercase tracking-wide px-1"
												style={{ color: "var(--color-muted)" }}
											>
												{prog.name}
											</p>
										)}
										<CheckList>
											{flattenProgram(prog).map(({ label, exercises }, i) => (
												<CheckListSection key={i} label={label}>
													{exercises.map((ex, j) => (
														<CheckListItem
															key={j}
															subtitle={formatSubtitle(ex, fitnessWeightUnit)}
														>
															{formatLabel(ex)}
														</CheckListItem>
													))}
												</CheckListSection>
											))}
										</CheckList>
									</div>
								))}
							</div>
						)}

						{/* Add Program — visible once session is running */}
						{started && (
							<Button
								variant="outline"
								size="sm"
								onClick={addProgramDialog.show}
							>
								+ Add Program
							</Button>
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
			</div>

			<SessionSummaryDialog
				openSignal={showSummary}
				completedAt={completedAtRef.current}
				elapsed={summaryElapsed.value}
				programName={accumulatedProgramName}
				notesSignal={notes}
				effortSignal={perceivedEffort}
				activitySignal={activity}
				saving={saving.value}
				saveError={saveError.value}
				onCancel={handleSummaryCancel}
				onSave={handleSave}
			/>

			{/* Add Program dialog */}
			<Dialog
				openSignal={addProgramDialog.open}
				onOpenChange={(v) => {
					if (!v) addProgramError.value = "";
				}}
			>
				<DialogContent>
					<DialogTitle>Add Program</DialogTitle>
					<div class="flex flex-col gap-4 mt-4">
						<Combobox
							label="Program"
							value={addProgramId.value}
							onChange={(v) => {
								addProgramId.value = v;
							}}
							options={programOptions}
							placeholder="Select a program…"
						/>
						{addProgramError.value && (
							<p class="text-sm" style={{ color: "var(--color-error)" }}>
								{addProgramError.value}
							</p>
						)}
						<div class="flex justify-end gap-2">
							<DialogClose>
								<Button variant="outline" size="sm">
									Cancel
								</Button>
							</DialogClose>
							<Button
								variant="primary"
								size="sm"
								disabled={!addProgramId.value || addingProgram.value}
								onClick={handleAddProgram}
							>
								Add
							</Button>
						</div>
					</div>
				</DialogContent>
			</Dialog>
		</>
	);
}
