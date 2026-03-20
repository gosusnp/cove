// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useLocation, useRoute } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { ListItem } from "../components/ui/ListItem.jsx";
import { Switch } from "../components/ui/Switch.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";
import { ExerciseDetail } from "./ExerciseDetail.jsx";

// ─── ExerciseList ─────────────────────────────────────────────────────────────

function ExerciseList({ exercises, selectedId, onSelect, onNew, error }) {
	return (
		<div class="flex flex-col">
			{/* Header */}
			<div
				class="flex items-center justify-between px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Exercises
				</h2>
				<Button variant="primary" size="sm" onClick={onNew}>
					+ New
				</Button>
			</div>

			{/* Exercise rows */}
			{error && (
				<p class="px-4 py-3 text-sm" style={{ color: "var(--color-error)" }}>
					{error}
				</p>
			)}
			{!error && exercises.length === 0 ? (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No exercises yet.
				</p>
			) : (
				!error &&
				exercises.map((ex, i) => (
					<ListItem
						key={ex.id}
						label={ex.name}
						sublabel={ex.progression || undefined}
						active={ex.id === selectedId}
						isLast={i === exercises.length - 1}
						onClick={() => onSelect(ex.id)}
					/>
				))
			)}
		</div>
	);
}

// ─── Exercises (page) ─────────────────────────────────────────────────────────

export function Exercises() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const exercises = useSignal([]);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	// New exercise dialog
	const newDialog = useDialog();
	const nameField = useSignal("");
	const progressionField = useSignal("");
	const descriptionField = useSignal("");
	const isPublicField = useSignal(false);
	const saving = useSignal(false);
	const formError = useSignal("");

	const fetchExercises = async () => {
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await apiFetch("/api/exercises");
			if (!r.ok) throw new Error("Failed to fetch exercises");
			exercises.value = await r.json();
		} catch (err) {
			fetchError.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchExercises();
	}, [!!user]);

	const handleSelect = (id) => {
		route(`/exercises/${id}`);
	};

	const openNew = () => {
		nameField.value = "";
		progressionField.value = "";
		descriptionField.value = "";
		isPublicField.value = false;
		formError.value = "";
		newDialog.show();
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (!nameField.value.trim()) {
			formError.value = "Name is required.";
			return;
		}
		saving.value = true;
		formError.value = "";
		try {
			const r = await apiFetch("/api/exercises", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: nameField.value.trim(),
					progression: progressionField.value.trim() || null,
					description: descriptionField.value.trim() || null,
					is_public: isPublicField.value,
				}),
			});
			if (!r.ok) {
				const data = await r.json();
				throw new Error(data.error || "Failed to create exercise");
			}
			const created = await r.json();
			await fetchExercises();
			newDialog.hide();
			route(`/exercises/${created.id}`);
		} catch (err) {
			formError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	return (
		<>
			<ListDetail
				hasDetail={!!selectedId}
				emptyState="Select an exercise to view its details."
				list={
					<ExerciseList
						exercises={exercises.value}
						selectedId={selectedId}
						onSelect={handleSelect}
						onNew={openNew}
						error={fetchError.value}
					/>
				}
				detail={
					selectedId ? (
						<ExerciseDetail
							exerciseId={selectedId}
							onExerciseUpdated={fetchExercises}
							onExerciseDeleted={() => {
								route("/exercises");
								fetchExercises();
							}}
						/>
					) : null
				}
			/>

			{/* New Exercise dialog */}
			<Dialog openSignal={newDialog.open}>
				<DialogContent>
					<form onSubmit={handleSave}>
						<DialogTitle>New Exercise</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<TextField
								id="ex-name"
								label="Name"
								placeholder="e.g. Diamond Push-up"
								value={nameField.value}
								onInput={(e) => {
									nameField.value = e.target.value;
								}}
								autoFocus
							/>
							<TextField
								id="ex-progression"
								label="Progression"
								placeholder="e.g. Push-up"
								value={progressionField.value}
								onInput={(e) => {
									progressionField.value = e.target.value;
								}}
							/>
							<TextField
								id="ex-description"
								label="Description"
								placeholder="Optional description"
								value={descriptionField.value}
								onInput={(e) => {
									descriptionField.value = e.target.value;
								}}
							/>
							<div class="flex items-center justify-between">
								<label
									for="ex-public"
									class="text-sm font-medium"
									style={{ color: "var(--color-text)" }}
								>
									Make public
								</label>
								<Switch id="ex-public" checkedSignal={isPublicField} />
							</div>
							{formError.value && (
								<p class="text-sm" style={{ color: "var(--color-error)" }}>
									{formError.value}
								</p>
							)}
						</div>
						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="outline" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={saving.value}>
								{saving.value ? "Saving…" : "Save"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>
		</>
	);
}
