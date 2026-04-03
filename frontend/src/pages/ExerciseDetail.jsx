// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { Pencil, Trash2 } from "lucide-preact";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import { Switch } from "../components/ui/Switch.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";

// ── ExerciseDetail (shell) ────────────────────────────────────────────────────
// Fetches exercise data, then renders ExerciseDetailInner.

export function ExerciseDetail({
	exerciseId,
	onExerciseUpdated,
	onExerciseDeleted,
}) {
	const { user } = useAuth();

	const exercise = useSignal(null);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	const fetchExercise = async () => {
		if (!exerciseId || !user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await apiFetch(`/api/exercises/${exerciseId}`);
			if (!r.ok) throw new Error("Failed to load exercise");
			exercise.value = await r.json();
		} catch (err) {
			fetchError.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchExercise();
	}, [exerciseId, user]);

	if (loading.value) {
		return (
			<div class="flex flex-1 items-center justify-center p-8">
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Loading…
				</p>
			</div>
		);
	}

	if (fetchError.value) {
		return (
			<div class="flex flex-1 items-center justify-center p-8">
				<p class="text-sm" style={{ color: "var(--color-error)" }}>
					{fetchError.value}
				</p>
			</div>
		);
	}

	if (!exercise.value) return null;

	return (
		<ExerciseDetailInner
			key={exerciseId}
			exercise={exercise.value}
			onExerciseUpdated={onExerciseUpdated}
			onExerciseDeleted={onExerciseDeleted}
		/>
	);
}

// ── ExerciseDetailInner ───────────────────────────────────────────────────────

function ExerciseDetailInner({
	exercise: initialExercise,
	onExerciseUpdated,
	onExerciseDeleted,
}) {
	const name = useSignal(initialExercise.name);
	const description = useSignal(initialExercise.description ?? "");
	const progression = useSignal(initialExercise.progression ?? "");
	const isPublic = useSignal(initialExercise.is_public);

	// editingField: null | "name" | "progression"
	const editingField = useSignal(null);
	const editValue = useSignal("");
	const editSaving = useSignal(false);
	const editError = useSignal("");

	const nameInputRef = useRef(null);
	const progressionInputRef = useRef(null);

	const startEdit = (field) => {
		editingField.value = field;
		editValue.value = field === "name" ? name.value : progression.value;
		editError.value = "";
	};

	const cancelEdit = () => {
		editingField.value = null;
		editError.value = "";
	};

	useEffect(() => {
		if (editingField.value === "name") nameInputRef.current?.focus();
		if (editingField.value === "progression")
			progressionInputRef.current?.focus();
	}, [editingField.value]);

	const saveEdit = async () => {
		if (editingField.value === "name" && !editValue.value.trim()) {
			editError.value = "Name is required.";
			return;
		}
		editSaving.value = true;
		editError.value = "";
		const updatedName =
			editingField.value === "name" ? editValue.value.trim() : name.value;
		const updatedProgression =
			editingField.value === "progression"
				? editValue.value.trim()
				: progression.value;
		try {
			const r = await apiFetch(`/api/exercises/${initialExercise.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: updatedName,
					progression: updatedProgression || null,
					description: description.value || null,
					is_public: isPublic.value,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save");
			}
			if (editingField.value === "name") name.value = updatedName;
			if (editingField.value === "progression")
				progression.value = updatedProgression;
			editingField.value = null;
			onExerciseUpdated?.();
		} catch (err) {
			editError.value = err.message;
		} finally {
			editSaving.value = false;
		}
	};

	const saveDescription = async (desc) => {
		const r = await apiFetch(`/api/exercises/${initialExercise.id}`, {
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				name: name.value,
				progression: progression.value || null,
				description: desc || null,
				is_public: isPublic.value,
			}),
		});
		if (!r.ok) {
			const d = await r.json();
			throw new Error(d.error || "Failed to save");
		}
		description.value = desc;
		onExerciseUpdated?.();
	};

	const savePublic = async (v) => {
		isPublic.value = v;
		const r = await apiFetch(`/api/exercises/${initialExercise.id}`, {
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				name: name.value,
				progression: progression.value || null,
				description: description.value || null,
				is_public: v,
			}),
		});
		if (!r.ok) {
			isPublic.value = !v; // revert on failure
		} else {
			onExerciseUpdated?.();
		}
	};

	const deleteDialog = useDialog();
	const deleteError = useSignal("");

	const handleDelete = async () => {
		deleteError.value = "";
		const r = await apiFetch(`/api/exercises/${initialExercise.id}`, {
			method: "DELETE",
		});
		if (!r.ok) {
			deleteError.value = "Failed to delete exercise.";
			throw new Error("Failed to delete exercise");
		}
		onExerciseDeleted?.();
	};

	return (
		<>
			<div class="page-content">
				{/* Name */}
				{editingField.value === "name" ? (
					<div class="flex flex-col gap-1">
						<div class="flex items-center gap-3">
							<TextField
								inline
								inputRef={nameInputRef}
								containerClass="flex-1 min-w-0"
								class="text-2xl font-semibold"
								value={editValue.value}
								onInput={(e) => {
									editValue.value = e.target.value;
								}}
								onKeyDown={(e) => {
									if (e.key === "Escape") cancelEdit();
									if (e.key === "Enter") saveEdit();
								}}
							/>
							<div class="flex gap-2 shrink-0">
								<Button
									size="sm"
									onClick={saveEdit}
									disabled={editSaving.value}
								>
									{editSaving.value ? "Saving…" : "Save"}
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={cancelEdit}
									disabled={editSaving.value}
								>
									Cancel
								</Button>
							</div>
						</div>
						{editError.value && (
							<p class="text-xs" style={{ color: "var(--color-error)" }}>
								{editError.value}
							</p>
						)}
					</div>
				) : (
					<div class="group flex items-center justify-between w-full gap-2">
						<h1
							class="text-2xl font-semibold"
							style={{ color: "var(--color-text)" }}
						>
							{name.value}
						</h1>
						<div class="flex items-center gap-1 shrink-0">
							<button
								type="button"
								class="opacity-0 group-hover:opacity-30 transition-opacity cursor-pointer"
								onClick={() => startEdit("name")}
								aria-label="Edit exercise name"
							>
								<Pencil
									size={14}
									style={{ color: "var(--color-muted)" }}
									aria-hidden="true"
								/>
							</button>
							<Tooltip>
								<TooltipTrigger>
									<Button
										variant="ghost"
										size="icon"
										aria-label="Delete exercise"
										onClick={deleteDialog.show}
									>
										<Trash2 size={14} aria-hidden="true" />
									</Button>
								</TooltipTrigger>
								<TooltipContent>Delete</TooltipContent>
							</Tooltip>
						</div>
					</div>
				)}

				{/* Description */}
				<div class="flex flex-col gap-1">
					<p
						class="text-xs font-medium uppercase tracking-wide"
						style={{ color: "var(--color-muted)" }}
					>
						Description
					</p>
					<EditableMarkdown
						value={description.value || null}
						placeholder="Add a description…"
						editLabel="Edit exercise description"
						onSave={saveDescription}
					/>
				</div>

				{/* Progression */}
				<div class="flex flex-col gap-1">
					<p
						class="text-xs font-medium uppercase tracking-wide"
						style={{ color: "var(--color-muted)" }}
					>
						Progression
					</p>
					{editingField.value === "progression" ? (
						<div class="flex flex-col gap-1">
							<div class="flex items-center gap-3">
								<TextField
									inline
									inputRef={progressionInputRef}
									containerClass="flex-1 min-w-0"
									value={editValue.value}
									placeholder="e.g. Push-up"
									onInput={(e) => {
										editValue.value = e.target.value;
									}}
									onKeyDown={(e) => {
										if (e.key === "Escape") cancelEdit();
										if (e.key === "Enter") saveEdit();
									}}
								/>
								<div class="flex gap-2 shrink-0">
									<Button
										size="sm"
										onClick={saveEdit}
										disabled={editSaving.value}
									>
										{editSaving.value ? "Saving…" : "Save"}
									</Button>
									<Button
										variant="outline"
										size="sm"
										onClick={cancelEdit}
										disabled={editSaving.value}
									>
										Cancel
									</Button>
								</div>
							</div>
							{editError.value && (
								<p class="text-xs" style={{ color: "var(--color-error)" }}>
									{editError.value}
								</p>
							)}
						</div>
					) : (
						<div class="group flex items-center justify-between w-full gap-2">
							<p
								class="text-sm"
								style={{
									color: progression.value
										? "var(--color-text)"
										: "var(--color-muted)",
								}}
							>
								{progression.value || "No progression set"}
							</p>
							<button
								type="button"
								class="opacity-0 group-hover:opacity-30 transition-opacity shrink-0 cursor-pointer"
								onClick={() => startEdit("progression")}
								aria-label="Edit progression"
							>
								<Pencil
									size={14}
									style={{ color: "var(--color-muted)" }}
									aria-hidden="true"
								/>
							</button>
						</div>
					)}
				</div>

				{/* Delete error */}
				{deleteError.value && (
					<p class="text-sm" style={{ color: "var(--color-error)" }}>
						{deleteError.value}
					</p>
				)}

				{/* Public toggle */}
				<div class="flex items-center justify-between">
					<div>
						<p
							class="text-sm font-medium"
							style={{ color: "var(--color-text)" }}
						>
							Public
						</p>
						<p class="text-xs" style={{ color: "var(--color-muted)" }}>
							Visible to other users
						</p>
					</div>
					<Switch
						id={`ex-public-${initialExercise.id}`}
						checkedSignal={isPublic}
						onCheckedChange={(v) => {
							isPublic.value = v;
							savePublic(v);
						}}
					/>
				</div>
			</div>

			<ConfirmDialog
				openSignal={deleteDialog.open}
				title="Delete Exercise"
				description="This will permanently delete the exercise."
				confirmLabel="Delete"
				onConfirm={handleDelete}
			/>
		</>
	);
}
