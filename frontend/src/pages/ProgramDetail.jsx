// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import {
	closestCenter,
	DndContext,
	DragOverlay,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import {
	SortableContext,
	useSortable,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { useAuth } from "../Auth.jsx";
import {
	Accordion,
	AccordionContent,
	AccordionDragHandle,
	AccordionItem,
	AccordionTrigger,
} from "../components/ui/Accordion.jsx";
import { Button } from "../components/ui/Button.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { Markdown } from "../components/ui/Markdown.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ToggleGroup } from "../components/ui/ToggleGroup.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { useSortableGroups } from "../hooks/useSortableGroups.js";
import { cn } from "../lib/utils.js";
import { apiFetch } from "../lib/api.js";

// ── Constants ─────────────────────────────────────────────────────────────────

const LATERALITY_OPTIONS = [
	{ value: "bilateral", label: "Bilateral" },
	{ value: "unilateral", label: "Unilateral" },
	{ value: "left", label: "Left" },
	{ value: "right", label: "Right" },
	{ value: "alternating", label: "Alternating" },
];

// ── Icons ─────────────────────────────────────────────────────────────────────

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

function PlusIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M8 1.75a.75.75 0 0 1 .75.75v4.75h4.75a.75.75 0 0 1 0 1.5H8.75v4.75a.75.75 0 0 1-1.5 0V8.75H2.5a.75.75 0 0 1 0-1.5h4.75V2.5A.75.75 0 0 1 8 1.75Z"
				fill="currentColor"
			/>
		</svg>
	);
}

function TrashIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				fill-rule="evenodd"
				clip-rule="evenodd"
				d="M6.5 1.75a.25.25 0 0 1 .25-.25h2.5a.25.25 0 0 1 .25.25V3h-3V1.75ZM5 3V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75V3h2.25a.75.75 0 0 1 0 1.5H14v8.75A1.75 1.75 0 0 1 12.25 15h-8.5A1.75 1.75 0 0 1 2 13.25V4.5H.75a.75.75 0 0 1 0-1.5H5ZM3.5 4.5v8.75c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25V4.5h-9Z"
				fill="currentColor"
			/>
		</svg>
	);
}

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
				<Tooltip>
					<TooltipTrigger>
						<Button
							variant="outline"
							size="icon"
							aria-label="Edit exercise"
							onClick={(e) => {
								e.currentTarget.blur();
								onEdit(exercise);
							}}
						>
							<PencilIcon />
						</Button>
					</TooltipTrigger>
					<TooltipContent>Edit</TooltipContent>
				</Tooltip>
				<Tooltip>
					<TooltipTrigger>
						<Button
							variant="outline"
							size="icon"
							aria-label="Remove exercise"
							onClick={(e) => {
								e.currentTarget.blur();
								onRemove(exercise);
							}}
						>
							<TrashIcon />
						</Button>
					</TooltipTrigger>
					<TooltipContent>Remove</TooltipContent>
				</Tooltip>
			</div>
		</div>
	);
}

// ── ProgramDetail (shell) ─────────────────────────────────────────────────────
// Fetches program data, then renders ProgramDetailInner keyed by programId so
// that useSortableGroups is seeded fresh whenever the program changes.

export function ProgramDetail({ programId, onProgramUpdated }) {
	const { user } = useAuth();

	const rawProgram = useSignal(null);
	const loading = useSignal(true);
	const fetchError = useSignal("");
	const refreshKey = useSignal(0);

	const fetchProgram = async () => {
		if (!programId || !user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await apiFetch(`/api/programs/${programId}`);
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
	}, [programId, user]);

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

	if (!rawProgram.value) return null;

	return (
		<ProgramDetailInner
			key={`${programId}-${refreshKey.value}`}
			program={rawProgram.value}
			onRefresh={fetchProgram}
			onProgramUpdated={onProgramUpdated}
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

function ProgramDetailInner({
	program: initialProgram,
	onRefresh,
	onProgramUpdated,
}) {
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
	const programDescription = useSignal(initialProgram.description ?? "");

	// ── Inline edit ───────────────────────────────────────────────────────────
	// editingField: null | "name" | "description"
	const editingField = useSignal(null);
	const editValue = useSignal("");
	const editSaving = useSignal(false);
	const editError = useSignal("");

	const startEdit = (field) => {
		editingField.value = field;
		editValue.value =
			field === "name" ? programName.value : programDescription.value;
		editError.value = "";
	};

	const cancelEdit = () => {
		editingField.value = null;
		editError.value = "";
	};

	const nameInputRef = useRef(null);
	const descInputRef = useRef(null);

	useEffect(() => {
		if (editingField.value === "name") nameInputRef.current?.focus();
		else if (editingField.value === "description")
			descInputRef.current?.focus();
	}, [editingField.value]);

	const saveEdit = async () => {
		if (editingField.value === "name" && !editValue.value.trim()) {
			editError.value = "Name is required.";
			return;
		}
		editSaving.value = true;
		editError.value = "";
		const name =
			editingField.value === "name"
				? editValue.value.trim()
				: programName.value;
		const desc =
			editingField.value === "description"
				? editValue.value.trim()
				: programDescription.value;
		try {
			const r = await apiFetch(`/api/programs/${initialProgram.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ name, description: desc || undefined }),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save");
			}
			programName.value = name;
			programDescription.value = desc;
			editingField.value = null;
			onProgramUpdated?.();
		} catch (err) {
			editError.value = err.message;
		} finally {
			editSaving.value = false;
		}
	};

	// ── Exercises list for the combobox ───────────────────────────────────────
	const exerciseOptions = useSignal([]);

	useEffect(() => {
		apiFetch("/api/exercises")
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
	}, []);

	// ── Add / Edit Set dialog ─────────────────────────────────────────────────
	const setDialog = useDialog();
	const reorderError = useSignal("");

	const editingSet = useSignal(null);
	const setNameField = useSignal("");
	const setRoundsField = useSignal("3");
	const setRestField = useSignal("");
	const setFormError = useSignal("");
	const setFormSaving = useSignal(false);

	const openAddSet = () => {
		editingSet.value = null;
		setNameField.value = "";
		setRoundsField.value = "3";
		setRestField.value = "";
		setFormError.value = "";
		setDialog.show();
	};

	const openEditSet = (set) => {
		editingSet.value = set;
		setNameField.value = set.name ?? "";
		setRoundsField.value = String(set.rounds ?? 3);
		setRestField.value = set.rest_s != null ? String(set.rest_s) : "";
		setFormError.value = "";
		setDialog.show();
	};

	const handleSaveSet = async (e) => {
		e.preventDefault();
		const rounds = parseInt(setRoundsField.value, 10);
		const restRaw = setRestField.value.trim();
		const rest = restRaw === "" ? null : parseInt(restRaw, 10);
		if (Number.isNaN(rounds) || rounds < 1) {
			setFormError.value = "Rounds must be at least 1.";
			return;
		}
		if (rest !== null && (Number.isNaN(rest) || rest < 0)) {
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
		try {
			const r = await apiFetch(url, {
				method,
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: setNameField.value.trim() || undefined,
					rounds,
					rest_s: rest ?? undefined,
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
		const r = await apiFetch(
			`/api/programs/${initialProgram.id}/sets/${s._rawId}`,
			{
				method: "DELETE",
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

		const isEdit = !!editingPex.value;
		const url = isEdit
			? `/api/programs/${initialProgram.id}/sets/${pexSetId.value}/exercises/${editingPex.value._rawId}`
			: `/api/programs/${initialProgram.id}/sets/${pexSetId.value}/exercises`;
		const method = isEdit ? "PUT" : "POST";

		const repsVal = pexReps.value !== "" ? parseInt(pexReps.value, 10) : null;
		const durVal =
			pexDuration.value !== "" ? parseInt(pexDuration.value, 10) : null;
		const weightVal =
			pexWeight.value !== "" ? parseFloat(pexWeight.value) : null;

		try {
			const r = await apiFetch(url, {
				method,
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					exercise_id: pexExerciseId.value,
					laterality: pexLaterality.value ?? undefined,
					reps: repsVal,
					duration_s: durVal,
					weight_kg: weightVal,
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
		const r = await apiFetch(
			`/api/programs/${initialProgram.id}/sets/${setRawId}/exercises/${pex._rawId}`,
			{
				method: "DELETE",
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

		if (!over || active.id === over.id) return;

		let finalSets;
		if (!activeType) {
			// ── Set reorder ──────────────────────────────────────────────────
			const oldIdx = setsSnapshot.findIndex((s) => s.id === active.id);
			const newIdx = setsSnapshot.findIndex((s) => s.id === over.id);
			if (oldIdx < 0 || newIdx < 0) return;

			finalSets = [...setsSnapshot];
			const [moved] = finalSets.splice(oldIdx, 1);
			finalSets.splice(newIdx, 0, moved);
		} else if (activeType === "exercise") {
			// ── Exercise reorder / cross-set move ────────────────────────────
			// For cross-set moves, setsSnapshot already has the exercise in the
			// destination set (handleDragOver applied the optimistic update).
			// For same-set reorders, apply the position swap manually.
			const srcSetId = active.data.current?.setId;
			const destSet = setsSnapshot.find((s) =>
				s.exercises.some((e) => e.id === active.id),
			);
			if (!destSet) return;

			finalSets = setsSnapshot.map((s) => {
				if (s.id === destSet.id && srcSetId === destSet.id) {
					const exs = [...s.exercises];
					const oldIdx = exs.findIndex((e) => e.id === active.id);
					const newIdx = exs.findIndex((e) => e.id === over.id);
					if (oldIdx >= 0 && newIdx >= 0 && oldIdx !== newIdx) {
						const [mv] = exs.splice(oldIdx, 1);
						exs.splice(newIdx, 0, mv);
					}
					return { ...s, exercises: exs };
				}
				return s;
			});
		} else {
			return;
		}

		const structure = finalSets.map((s) => ({
			set_id: s._rawId,
			exercise_ids: s.exercises.map((e) => e._rawId),
		}));

		reorderError.value = "";
		apiFetch(`/api/programs/${initialProgram.id}/structure`, {
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(structure),
		})
			.then((r) => {
				if (!r.ok) {
					return r.json().then((j) => {
						reorderError.value = j.error ?? "Failed to save order.";
						onRefresh();
					});
				}
			})
			.catch(() => {
				reorderError.value = "Failed to save order.";
				onRefresh();
			});
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
				<div class="flex items-start justify-between gap-4 flex-wrap">
					<div class="flex flex-col gap-1 flex-1 min-w-0">
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
							<button
								type="button"
								class="text-left group flex items-center justify-between w-full cursor-text"
								onClick={() => startEdit("name")}
								aria-label="Edit program name"
							>
								<h1
									class="text-2xl font-semibold"
									style={{ color: "var(--color-text)" }}
								>
									{programName.value}
								</h1>
								<PencilIcon
									class="opacity-0 group-hover:opacity-30 transition-opacity shrink-0"
									style={{ color: "var(--color-muted)" }}
									aria-hidden="true"
								/>
							</button>
						)}

						{/* Description */}
						{editingField.value === "description" ? (
							<div class="flex flex-col gap-2">
								<TextField
									inline
									multiline
									inputRef={descInputRef}
									class="text-sm"
									rows="3"
									value={editValue.value}
									onInput={(e) => {
										editValue.value = e.target.value;
									}}
									onKeyDown={(e) => {
										if (e.key === "Escape") cancelEdit();
									}}
								/>
								{editError.value && (
									<p class="text-xs" style={{ color: "var(--color-error)" }}>
										{editError.value}
									</p>
								)}
								<div class="flex justify-end gap-2">
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
						) : (
							<div class="group flex items-start justify-between w-full gap-2">
								{programDescription.value ? (
									<Markdown>{programDescription.value}</Markdown>
								) : (
									<p
										class="text-sm italic"
										style={{ color: "var(--color-muted)", opacity: 0.5 }}
									>
										Add a description…
									</p>
								)}
								<button
									type="button"
									class="opacity-0 group-hover:opacity-30 transition-opacity shrink-0 cursor-pointer"
									onClick={() => startEdit("description")}
									aria-label="Edit program description"
								>
									<PencilIcon
										style={{ color: "var(--color-muted)" }}
										aria-hidden="true"
									/>
								</button>
							</div>
						)}
					</div>
					<div class="flex gap-2 shrink-0">
						<Button
							variant="primary"
							size="sm"
							onClick={(e) => {
								e.currentTarget.blur();
								openAddSet();
							}}
						>
							+ Add Set
						</Button>
					</div>
				</div>

				{/* Structure preview */}
				{initialProgram.structure && (
					<details class="rounded-lg border border-(--color-border) bg-(--color-surface)">
						<summary
							class="cursor-pointer select-none px-4 py-3 text-sm font-medium"
							style={{ color: "var(--color-text)" }}
						>
							Structure preview
						</summary>
						<pre
							class="px-4 pb-4 text-xs leading-relaxed whitespace-pre-wrap"
							style={{ color: "var(--color-muted)" }}
						>
							{initialProgram.structure}
						</pre>
					</details>
				)}

				{/* Sets */}
				{reorderError.value && (
					<p class="text-sm" style={{ color: "var(--color-error)" }}>
						{reorderError.value}
					</p>
				)}
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
												{set.rounds}×
												{set.rest_s != null ? ` · ${set.rest_s}s rest` : ""}
											</span>
											<div class="flex gap-1">
												<Tooltip>
													<TooltipTrigger>
														<Button
															variant="outline"
															size="icon"
															aria-label="Add exercise"
															onClick={(e) => {
																e.stopPropagation();
																e.currentTarget.blur();
																openAddPex(set.id);
															}}
														>
															<PlusIcon />
														</Button>
													</TooltipTrigger>
													<TooltipContent>Add Exercise</TooltipContent>
												</Tooltip>
												<Tooltip>
													<TooltipTrigger>
														<Button
															variant="outline"
															size="icon"
															aria-label="Edit set"
															onClick={(e) => {
																e.stopPropagation();
																e.currentTarget.blur();
																openEditSet(set);
															}}
														>
															<PencilIcon />
														</Button>
													</TooltipTrigger>
													<TooltipContent>Edit</TooltipContent>
												</Tooltip>
												<Tooltip>
													<TooltipTrigger>
														<Button
															variant="outline"
															size="icon"
															aria-label="Delete set"
															onClick={(e) => {
																e.stopPropagation();
																e.currentTarget.blur();
																openDeleteSet(set);
															}}
														>
															<TrashIcon />
														</Button>
													</TooltipTrigger>
													<TooltipContent>Delete</TooltipContent>
												</Tooltip>
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
								label="Rest between sets (sec, optional)"
								type="number"
								min="0"
								value={setRestField.value}
								onInput={(e) => {
									setRestField.value = e.target.value;
								}}
							/>
							{setFormError.value && (
								<p class="text-sm" style={{ color: "var(--color-error)" }}>
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
								<p class="text-sm" style={{ color: "var(--color-error)" }}>
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
