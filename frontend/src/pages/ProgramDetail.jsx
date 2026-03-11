// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect, useRef } from "preact/hooks";
import { useSignal } from "@preact/signals";
import {
	DndContext,
	closestCenter,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
	DragOverlay,
} from "@dnd-kit/core";
import {
	SortableContext,
	verticalListSortingStrategy,
	useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import {
	Accordion,
	AccordionItem,
	AccordionTrigger,
	AccordionContent,
	AccordionDragHandle,
} from "../components/ui/Accordion.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ToggleGroup } from "../components/ui/ToggleGroup.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { useSortableGroups } from "../hooks/useSortableGroups.js";
import { cn } from "../lib/utils.js";

// ── Constants ─────────────────────────────────────────────────────────────────

const LATERALITY_OPTIONS = [
	{ value: "bilateral", label: "Bilateral" },
	{ value: "unilateral", label: "Unilateral" },
	{ value: "left", label: "Left" },
	{ value: "right", label: "Right" },
	{ value: "alternating", label: "Alternating" },
];

// ── Weight tag ────────────────────────────────────────────────────────────────

function WeightTag({ weight_kg }) {
	if (weight_kg == null || weight_kg === 0) {
		return (
			<span class="text-xs" style={{ color: "var(--color-muted)" }}>
				bodyweight
			</span>
		);
	}
	if (weight_kg > 0) {
		return (
			<span
				class="text-xs font-medium"
				style={{ color: "var(--color-success, #1a6e1a)" }}
			>
				+{weight_kg}kg
			</span>
		);
	}
	return (
		<span
			class="text-xs font-medium"
			style={{ color: "var(--color-info, #0055aa)" }}
		>
			{weight_kg}kg (assisted)
		</span>
	);
}

// ── GripIcon ──────────────────────────────────────────────────────────────────

function GripIcon() {
	return (
		<svg
			width="12"
			height="12"
			viewBox="0 0 14 14"
			fill="none"
			aria-hidden="true"
		>
			<circle cx="5" cy="3" r="1" fill="currentColor" />
			<circle cx="9" cy="3" r="1" fill="currentColor" />
			<circle cx="5" cy="7" r="1" fill="currentColor" />
			<circle cx="9" cy="7" r="1" fill="currentColor" />
			<circle cx="5" cy="11" r="1" fill="currentColor" />
			<circle cx="9" cy="11" r="1" fill="currentColor" />
		</svg>
	);
}

// ── Sentinel placeholder ──────────────────────────────────────────────────────
// Registered as a sortable node so closestCenter has a stable DOM target when a
// set is empty. Prevents the oscillation where an exercise rapidly moves in/out
// of an empty set because the only collision target is the AccordionItem itself.

const SENTINEL_PREFIX = "placeholder-";

function sentinelId(prefixedSetId) {
	return `${SENTINEL_PREFIX}${prefixedSetId}`;
}

function SentinelDropTarget({ id, setId, hidden }) {
	// data.type="exercise" and data.setId make useSortableGroups.handleDragOver
	// treat this node as a drop target inside the set, routing the move correctly.
	const { setNodeRef } = useSortable({
		id,
		data: { type: "exercise", setId },
	});
	// Always render so the DOM node stays registered with dnd-kit even while the
	// last exercise is being dragged out. Hidden visually when set has exercises.
	return (
		<div
			ref={setNodeRef}
			class="flex items-center justify-center text-xs"
			style={{
				color: "var(--color-muted)",
				height: hidden ? 0 : "2rem",
				overflow: "hidden",
			}}
			aria-hidden={hidden ? "true" : undefined}
		>
			Drop exercises here
		</div>
	);
}

// ── Sortable exercise row ─────────────────────────────────────────────────────

function SortableExerciseRow({ exercise, setId, onEdit, onRemove }) {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id: exercise.id, data: { type: "exercise", setId } });

	return (
		<div
			ref={setNodeRef}
			style={{ transform: CSS.Transform.toString(transform), transition }}
			class={cn(
				"flex items-center gap-3 py-2 px-1 rounded-lg",
				"border border-transparent",
				isDragging
					? "opacity-50 border-(--color-border) bg-(--color-bg)"
					: "hover:bg-(--color-bg)",
			)}
		>
			<button
				type="button"
				class="flex items-center justify-center w-5 h-5 shrink-0 cursor-grab active:cursor-grabbing text-(--color-muted) hover:text-(--color-text)"
				aria-label="Drag to reorder"
				{...listeners}
				{...attributes}
			>
				<GripIcon />
			</button>
			<span class="flex-1 text-sm text-(--color-text)">{exercise.name}</span>
			{exercise.laterality && (
				<span
					class="text-xs tabular-nums"
					style={{ color: "var(--color-muted)" }}
				>
					{exercise.laterality}
				</span>
			)}
			{exercise.reps != null && (
				<span
					class="text-xs tabular-nums"
					style={{ color: "var(--color-muted)" }}
				>
					{exercise.reps} reps
				</span>
			)}
			{exercise.duration_s != null && (
				<span
					class="text-xs tabular-nums"
					style={{ color: "var(--color-muted)" }}
				>
					{exercise.duration_s}s
				</span>
			)}
			<WeightTag weight_kg={exercise.weight_kg} />
			<div class="flex gap-1 shrink-0">
				<Button variant="ghost" size="sm" onClick={() => onEdit(exercise)}>
					Edit
				</Button>
				<Button
					variant="destructive"
					size="sm"
					onClick={() => onRemove(exercise)}
				>
					Del
				</Button>
			</div>
		</div>
	);
}

// ── ProgramDetail (shell) ─────────────────────────────────────────────────────
// Fetches program data, then renders ProgramDetailInner keyed by programId so
// that useSortableGroups is seeded fresh whenever the program changes.

export function ProgramDetail({ programId }) {
	const { token } = useAuth();

	const rawProgram = useSignal(null);
	const loading = useSignal(true);
	const fetchError = useSignal("");
	const refreshKey = useSignal(0);

	const fetchProgram = async () => {
		if (!programId || !token) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await fetch(`/api/programs/${programId}`, {
				headers: { Authorization: `Bearer ${token}` },
			});
			if (!r.ok) throw new Error("Failed to load program");
			rawProgram.value = await r.json();
			refreshKey.value += 1;
		} catch (err) {
			fetchError.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchProgram();
	}, [programId, token]);

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
				<p class="text-sm" style={{ color: "var(--color-error, red)" }}>
					{fetchError.value}
				</p>
			</div>
		);
	}

	if (!rawProgram.value) return null;

	return (
		<ProgramDetailInner
			key={`${programId}-${refreshKey.value}`}
			program={rawProgram.value}
			token={token}
			onRefresh={fetchProgram}
		/>
	);
}

// ── ID helpers ────────────────────────────────────────────────────────────────
// dnd-kit ids must be unique across sets AND exercises. Numeric DB ids can
// collide (set id 5 and exercise id 5 are different entities). Prefix them so
// they never collide, keeping useSortableGroups.handleDragOver working correctly.

function sid(id) {
	return `s-${id}`;
}
function eid(id) {
	return `e-${id}`;
}
function rawId(prefixedId) {
	return Number(String(prefixedId).replace(/^[se]-/, ""));
}

// Transform raw API data into prefixed-id form for dnd-kit.
function toSortableSet(set) {
	return {
		...set,
		id: sid(set.id),
		_rawId: set.id,
		exercises: (set.exercises ?? []).map((ex) => ({
			...ex,
			id: eid(ex.id),
			_rawId: ex.id,
		})),
	};
}

// ── ProgramDetailInner ────────────────────────────────────────────────────────
// Receives already-loaded program data and manages all CRUD + DnD.

function ProgramDetailInner({ program: initialProgram, token, onRefresh }) {
	const sensors = useSensors(
		useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
		useSensor(KeyboardSensor),
	);

	const sortableSets = (initialProgram.sets ?? []).map(toSortableSet);

	// ── Sortable sets state ───────────────────────────────────────────────────
	const {
		sets,
		openValues,
		setOpenValues,
		activeId,
		handleDragStart,
		handleDragOver,
		handleDragEnd: dndHandleDragEnd,
	} = useSortableGroups(sortableSets, {
		initialOpen: sortableSets.map((s) => s.id),
	});

	// Always-current ref so handleDragEnd sees the post-handleDragOver state,
	// not a stale closure snapshot.
	const setsRef = useRef(sets);
	setsRef.current = sets;

	const programName = useSignal(initialProgram.name);

	// ── Exercises list for the combobox ───────────────────────────────────────
	const exerciseOptions = useSignal([]);

	useEffect(() => {
		if (!token) return;
		fetch("/api/exercises", { headers: { Authorization: `Bearer ${token}` } })
			.then((r) => r.json())
			.then((data) => {
				const sorted = [...(data ?? [])].sort((a, b) =>
					a.name.localeCompare(b.name),
				);
				exerciseOptions.value = sorted.map((e) => ({
					value: e.id,
					label: e.name,
				}));
			})
			.catch(() => {});
	}, [token]);

	// ── Rename dialog ─────────────────────────────────────────────────────────
	const renameDialog = useDialog();
	const renameField = useSignal(initialProgram.name);
	const renameSaving = useSignal(false);
	const renameError = useSignal("");

	const openRename = () => {
		renameField.value = programName.value;
		renameError.value = "";
		renameDialog.show();
	};

	const handleRename = async (e) => {
		e.preventDefault();
		if (!renameField.value.trim()) {
			renameError.value = "Name is required.";
			return;
		}
		renameSaving.value = true;
		renameError.value = "";
		try {
			const r = await fetch(`/api/programs/${initialProgram.id}`, {
				method: "PUT",
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ name: renameField.value.trim() }),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to rename program");
			}
			programName.value = renameField.value.trim();
			renameDialog.hide();
		} catch (err) {
			renameError.value = err.message;
		} finally {
			renameSaving.value = false;
		}
	};

	// ── Add / Edit Set dialog ─────────────────────────────────────────────────
	const setDialog = useDialog();
	const editingSet = useSignal(null);
	const setNameField = useSignal("");
	const setRoundsField = useSignal("3");
	const setRestField = useSignal("90");
	const setFormError = useSignal("");
	const setFormSaving = useSignal(false);

	const openAddSet = () => {
		editingSet.value = null;
		setNameField.value = "";
		setRoundsField.value = "3";
		setRestField.value = "90";
		setFormError.value = "";
		setDialog.show();
	};

	const openEditSet = (set) => {
		editingSet.value = set;
		setNameField.value = set.name ?? "";
		setRoundsField.value = String(set.rounds ?? 3);
		setRestField.value = String(set.rest_s ?? 90);
		setFormError.value = "";
		setDialog.show();
	};

	const handleSaveSet = async (e) => {
		e.preventDefault();
		const rounds = parseInt(setRoundsField.value, 10);
		const rest = parseInt(setRestField.value, 10);
		if (Number.isNaN(rounds) || rounds < 1) {
			setFormError.value = "Rounds must be at least 1.";
			return;
		}
		if (Number.isNaN(rest) || rest < 0) {
			setFormError.value = "Rest must be 0 or more seconds.";
			return;
		}
		setFormSaving.value = true;
		setFormError.value = "";
		const isEdit = !!editingSet.value;
		const url = isEdit
			? `/api/programs/${initialProgram.id}/sets/${editingSet.value._rawId}`
			: `/api/programs/${initialProgram.id}/sets`;
		const method = isEdit ? "PUT" : "POST";
		const sortOrder = isEdit ? editingSet.value.sort_order : sets.length;
		try {
			const r = await fetch(url, {
				method,
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					name: setNameField.value.trim() || undefined,
					rounds,
					rest_s: rest,
					sort_order: sortOrder,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save set");
			}
			await onRefresh();
			setDialog.hide();
		} catch (err) {
			setFormError.value = err.message;
		} finally {
			setFormSaving.value = false;
		}
	};

	// ── Delete Set ────────────────────────────────────────────────────────────
	const deleteSetDialog = useDialog();
	const deletingSet = useSignal(null);

	const openDeleteSet = (set) => {
		deletingSet.value = set;
		deleteSetDialog.show();
	};

	const handleDeleteSet = async () => {
		const s = deletingSet.value;
		if (!s) return;
		const r = await fetch(
			`/api/programs/${initialProgram.id}/sets/${s._rawId}`,
			{
				method: "DELETE",
				headers: { Authorization: `Bearer ${token}` },
			},
		);
		if (!r.ok) throw new Error("Failed to delete set");
		await onRefresh();
	};

	// ── Add / Edit Exercise dialog ────────────────────────────────────────────
	const pexDialog = useDialog();
	const editingPex = useSignal(null);
	const pexSetId = useSignal(null);
	const pexExerciseId = useSignal(null);
	const pexLaterality = useSignal(null);
	const pexReps = useSignal("");
	const pexDuration = useSignal("");
	const pexWeight = useSignal("");
	const pexFormError = useSignal("");
	const pexFormSaving = useSignal(false);

	const openAddPex = (prefixedSetId) => {
		editingPex.value = null;
		pexSetId.value = rawId(prefixedSetId);
		pexExerciseId.value = null;
		pexLaterality.value = null;
		pexReps.value = "";
		pexDuration.value = "";
		pexWeight.value = "";
		pexFormError.value = "";
		pexDialog.show();
	};

	const openEditPex = (prefixedSetId, pex) => {
		editingPex.value = pex;
		pexSetId.value = rawId(prefixedSetId);
		pexExerciseId.value = pex.exercise_id;
		pexLaterality.value = pex.laterality ?? null;
		pexReps.value = pex.reps != null ? String(pex.reps) : "";
		pexDuration.value = pex.duration_s != null ? String(pex.duration_s) : "";
		pexWeight.value = pex.weight_kg != null ? String(pex.weight_kg) : "";
		pexFormError.value = "";
		pexDialog.show();
	};

	const handleSavePex = async (e) => {
		e.preventDefault();
		if (!pexExerciseId.value) {
			pexFormError.value = "Exercise is required.";
			return;
		}
		pexFormSaving.value = true;
		pexFormError.value = "";

		const set = sets.find((s) => s._rawId === pexSetId.value);
		const isEdit = !!editingPex.value;
		const url = isEdit
			? `/api/programs/${initialProgram.id}/sets/${pexSetId.value}/exercises/${editingPex.value._rawId}`
			: `/api/programs/${initialProgram.id}/sets/${pexSetId.value}/exercises`;
		const method = isEdit ? "PUT" : "POST";
		const sortOrder = isEdit
			? editingPex.value.sort_order
			: (set?.exercises?.length ?? 0);

		const repsVal = pexReps.value !== "" ? parseInt(pexReps.value, 10) : null;
		const durVal =
			pexDuration.value !== "" ? parseInt(pexDuration.value, 10) : null;
		const weightVal =
			pexWeight.value !== "" ? parseFloat(pexWeight.value) : null;

		try {
			const r = await fetch(url, {
				method,
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					exercise_id: pexExerciseId.value,
					laterality: pexLaterality.value ?? undefined,
					reps: repsVal,
					duration_s: durVal,
					weight_kg: weightVal,
					sort_order: sortOrder,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save exercise");
			}
			await onRefresh();
			pexDialog.hide();
		} catch (err) {
			pexFormError.value = err.message;
		} finally {
			pexFormSaving.value = false;
		}
	};

	// ── Remove Exercise ───────────────────────────────────────────────────────
	const removePexDialog = useDialog();
	const removingPex = useSignal(null);
	const removingPexSetId = useSignal(null);

	const openRemovePex = (prefixedSetId, pex) => {
		removingPex.value = pex;
		removingPexSetId.value = rawId(prefixedSetId);
		removePexDialog.show();
	};

	const handleRemovePex = async () => {
		const pex = removingPex.value;
		const setRawId = removingPexSetId.value;
		if (!pex || !setRawId) return;
		const r = await fetch(
			`/api/programs/${initialProgram.id}/sets/${setRawId}/exercises/${pex._rawId}`,
			{
				method: "DELETE",
				headers: { Authorization: `Bearer ${token}` },
			},
		);
		if (!r.ok) throw new Error("Failed to remove exercise");
		await onRefresh();
	};

	// ── Drag end with API sync ────────────────────────────────────────────────
	const handleDragEnd = (event) => {
		const { active, over } = event;
		const activeType = active.data.current?.type;

		// Read the latest sets via ref — the closure value of `sets` would be
		// stale after handleDragOver's async state update. For cross-set moves
		// the exercise is already in the destination set at this point.
		const setsSnapshot = setsRef.current;

		dndHandleDragEnd(event);

		if (!over) return;

		if (!activeType) {
			// ── Set reorder ──────────────────────────────────────────────────
			if (active.id === over.id) return;
			const oldIdx = setsSnapshot.findIndex((s) => s.id === active.id);
			const newIdx = setsSnapshot.findIndex((s) => s.id === over.id);
			if (oldIdx < 0 || newIdx < 0) return;

			const reordered = [...setsSnapshot];
			const [moved] = reordered.splice(oldIdx, 1);
			reordered.splice(newIdx, 0, moved);

			Promise.all(
				reordered.map((s, i) =>
					fetch(`/api/programs/${initialProgram.id}/sets/${s._rawId}`, {
						method: "PUT",
						headers: {
							Authorization: `Bearer ${token}`,
							"Content-Type": "application/json",
						},
						body: JSON.stringify({
							name: s.name ?? undefined,
							rounds: s.rounds,
							rest_s: s.rest_s,
							sort_order: i,
						}),
					}),
				),
			).catch(() => {});
		} else if (activeType === "exercise") {
			// ── Exercise reorder / cross-set move ────────────────────────────
			// After handleDragOver's optimistic updates, the exercise is already
			// in the destination set inside `setsSnapshot`. Find which set it's in.
			const destSet = setsSnapshot.find((s) =>
				s.exercises.some((e) => e.id === active.id),
			);
			if (!destSet) return;

			// Also persist the source set if it differs (cross-set move).
			const srcSetId = active.data.current?.setId;
			const setsToSync = new Set([destSet.id]);
			if (srcSetId && srcSetId !== destSet.id) setsToSync.add(srcSetId);

			const requests = [];
			for (const setPrefixedId of setsToSync) {
				const s = setsSnapshot.find((x) => x.id === setPrefixedId);
				if (!s) continue;
				// For same-set reorder, reconstruct the final array since we
				// snapshot before dndHandleDragEnd applies arrayMove.
				let exercises = s.exercises;
				if (setPrefixedId === destSet.id && srcSetId === destSet.id) {
					const oldIdx = exercises.findIndex((e) => e.id === active.id);
					const newIdx = exercises.findIndex((e) => e.id === over.id);
					if (oldIdx >= 0 && newIdx >= 0 && oldIdx !== newIdx) {
						exercises = [...exercises];
						const [mv] = exercises.splice(oldIdx, 1);
						exercises.splice(newIdx, 0, mv);
					}
				}
				for (const [i, ex] of exercises.entries()) {
					requests.push(
						fetch(
							`/api/programs/${initialProgram.id}/sets/${s._rawId}/exercises/${ex._rawId}`,
							{
								method: "PUT",
								headers: {
									Authorization: `Bearer ${token}`,
									"Content-Type": "application/json",
								},
								body: JSON.stringify({
									exercise_id: ex.exercise_id,
									laterality: ex.laterality ?? undefined,
									reps: ex.reps ?? null,
									duration_s: ex.duration_s ?? null,
									weight_kg: ex.weight_kg ?? null,
									sort_order: i,
								}),
							},
						),
					);
				}
			}
			Promise.all(requests).catch(() => {});
		}
	};

	// ── Active drag overlays ──────────────────────────────────────────────────
	const activeExercise = activeId
		? sets.flatMap((s) => s.exercises).find((e) => e.id === activeId)
		: null;
	// Sets are identified by prefixed ids ("s-N"); exercises by "e-N".
	// If activeId starts with "s-" it's a set being dragged.
	const activeSet =
		activeId && String(activeId).startsWith("s-")
			? sets.find((s) => s.id === activeId)
			: null;

	// ── Render ────────────────────────────────────────────────────────────────
	return (
		<>
			<div class="p-6 max-w-3xl mx-auto w-full flex flex-col gap-6">
				{/* Header */}
				<div class="flex items-center justify-between gap-4 flex-wrap">
					<PageTitle>{programName.value}</PageTitle>
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onClick={openRename}>
							Rename
						</Button>
						<Button variant="primary" size="sm" onClick={openAddSet}>
							+ Add Set
						</Button>
					</div>
				</div>

				{/* Sets */}
				{sets.length === 0 ? (
					<p class="text-sm" style={{ color: "var(--color-muted)" }}>
						No sets yet. Add a set to start building this program.
					</p>
				) : (
					<DndContext
						sensors={sensors}
						collisionDetection={closestCenter}
						onDragStart={handleDragStart}
						onDragOver={handleDragOver}
						onDragEnd={handleDragEnd}
					>
						<SortableContext
							items={sets.map((s) => s.id)}
							strategy={verticalListSortingStrategy}
						>
							<Accordion
								type="multiple"
								value={openValues}
								onValueChange={setOpenValues}
							>
								{sets.map((set) => (
									<AccordionItem key={set.id} id={set.id} value={set.id}>
										<AccordionTrigger>
											<AccordionDragHandle />
											<span class="flex-1 text-left">
												{set.name || "Unnamed Set"}
											</span>
											<span
												class="ml-auto mr-2 text-xs tabular-nums"
												style={{ color: "var(--color-muted)" }}
											>
												{set.rounds}× · {set.rest_s}s rest
											</span>
											<div class="flex gap-1">
												<Button
													variant="ghost"
													size="sm"
													onClick={(e) => {
														e.stopPropagation();
														openEditSet(set);
													}}
												>
													Edit
												</Button>
												<Button
													variant="destructive"
													size="sm"
													onClick={(e) => {
														e.stopPropagation();
														openDeleteSet(set);
													}}
												>
													Delete
												</Button>
											</div>
										</AccordionTrigger>
										<AccordionContent>
											<SortableContext
												items={[
													sentinelId(set.id),
													...set.exercises.map((e) => e.id),
												]}
												strategy={verticalListSortingStrategy}
											>
												<div class="flex flex-col">
													{/* Sentinel always registered so closestCenter has a
													    stable target while the last exercise is mid-drag */}
													<SentinelDropTarget
														id={sentinelId(set.id)}
														setId={set.id}
														hidden={set.exercises.length > 0}
													/>
													{set.exercises.map((ex) => (
														<SortableExerciseRow
															key={ex.id}
															exercise={ex}
															setId={set.id}
															onEdit={(e) => openEditPex(set.id, e)}
															onRemove={(e) => openRemovePex(set.id, e)}
														/>
													))}
												</div>
											</SortableContext>
											<div class="mt-2">
												<Button
													variant="ghost"
													size="sm"
													onClick={() => openAddPex(set.id)}
												>
													+ Add Exercise
												</Button>
											</div>
										</AccordionContent>
									</AccordionItem>
								))}
							</Accordion>
						</SortableContext>

						<DragOverlay>
							{activeExercise && (
								<div class="flex items-center gap-3 py-2 px-3 rounded-lg shadow-lg bg-(--color-surface) border border-(--color-border) text-sm text-(--color-text)">
									<GripIcon />
									<span>{activeExercise.name}</span>
								</div>
							)}
							{activeSet && (
								<div class="rounded-lg px-4 py-3 shadow-lg bg-(--color-surface) border border-(--color-border) text-sm font-medium text-(--color-text)">
									{activeSet.name || "Unnamed Set"}
								</div>
							)}
						</DragOverlay>
					</DndContext>
				)}
			</div>

			{/* ── Rename dialog ──────────────────────────────────────────────── */}
			<Dialog openSignal={renameDialog.open}>
				<DialogContent>
					<form onSubmit={handleRename}>
						<DialogTitle>Rename Program</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<TextField
								label="Name"
								value={renameField.value}
								onInput={(e) => {
									renameField.value = e.target.value;
								}}
								autoFocus
							/>
							{renameError.value && (
								<p class="text-sm" style={{ color: "var(--color-error, red)" }}>
									{renameError.value}
								</p>
							)}
						</div>
						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="outline" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={renameSaving.value}>
								{renameSaving.value ? "Saving…" : "Save"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>

			{/* ── Add / Edit Set dialog ──────────────────────────────────────── */}
			<Dialog openSignal={setDialog.open}>
				<DialogContent>
					<form onSubmit={handleSaveSet}>
						<DialogTitle>
							{editingSet.value ? "Edit Set" : "Add Set"}
						</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<TextField
								label="Name (optional)"
								value={setNameField.value}
								onInput={(e) => {
									setNameField.value = e.target.value;
								}}
							/>
							<TextField
								label="Rounds"
								type="number"
								min="1"
								value={setRoundsField.value}
								onInput={(e) => {
									setRoundsField.value = e.target.value;
								}}
							/>
							<TextField
								label="Rest between sets (sec)"
								type="number"
								min="0"
								value={setRestField.value}
								onInput={(e) => {
									setRestField.value = e.target.value;
								}}
							/>
							{setFormError.value && (
								<p class="text-sm" style={{ color: "var(--color-error, red)" }}>
									{setFormError.value}
								</p>
							)}
						</div>
						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="outline" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={setFormSaving.value}>
								{setFormSaving.value ? "Saving…" : "Save"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>

			{/* ── Delete Set confirm ─────────────────────────────────────────── */}
			<ConfirmDialog
				openSignal={deleteSetDialog.open}
				title="Delete Set"
				description="This will delete the set and all its exercises."
				confirmLabel="Delete"
				onConfirm={handleDeleteSet}
			/>

			{/* ── Add / Edit Exercise dialog ─────────────────────────────────── */}
			<Dialog openSignal={pexDialog.open}>
				<DialogContent>
					<form onSubmit={handleSavePex}>
						<DialogTitle>
							{editingPex.value ? "Edit Exercise" : "Add Exercise"}
						</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<Combobox
								label="Exercise"
								value={pexExerciseId.value}
								onChange={(v) => {
									pexExerciseId.value = v;
								}}
								options={exerciseOptions.value}
								placeholder="Search exercises..."
								disabled={!!editingPex.value}
							/>
							<ToggleGroup
								label="Laterality"
								value={pexLaterality.value}
								onChange={(v) => {
									pexLaterality.value = v;
								}}
								options={LATERALITY_OPTIONS}
								nullable
							/>
							<TextField
								label="Target Reps"
								type="number"
								min="1"
								placeholder="—"
								value={pexReps.value}
								onInput={(e) => {
									pexReps.value = e.target.value;
								}}
							/>
							<TextField
								label="Target Duration (sec)"
								type="number"
								min="1"
								placeholder="—"
								value={pexDuration.value}
								onInput={(e) => {
									pexDuration.value = e.target.value;
								}}
							/>
							<TextField
								label="Target Weight (kg)"
								type="number"
								step="0.5"
								placeholder="0 = bodyweight"
								value={pexWeight.value}
								onInput={(e) => {
									pexWeight.value = e.target.value;
								}}
							/>
							{pexFormError.value && (
								<p class="text-sm" style={{ color: "var(--color-error, red)" }}>
									{pexFormError.value}
								</p>
							)}
						</div>
						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="outline" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={pexFormSaving.value}>
								{pexFormSaving.value ? "Saving…" : "Save"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>

			{/* ── Remove Exercise confirm ────────────────────────────────────── */}
			<ConfirmDialog
				openSignal={removePexDialog.open}
				title="Remove Exercise"
				description={
					removingPex.value
						? `Remove "${removingPex.value.name}" from this set?`
						: "Remove this exercise from the set?"
				}
				confirmLabel="Remove"
				onConfirm={handleRemovePex}
			/>
		</>
	);
}
