// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { Pencil, Plus, Trash2, X } from "lucide-preact";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "../components/ui/Accordion.jsx";
import { Button } from "../components/ui/Button.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import {
	Dialog,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import { Markdown } from "../components/ui/Markdown.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { FDCSearch } from "../components/shared/FDCSearch.jsx";
import { useDialog } from "../hooks/useDialog.js";
import {
	DISPLAY_STEPS,
	convertUnit,
	useUnitPreferences,
} from "../hooks/useUnitPreferences.js";
import { apiFetch } from "../lib/api.js";

// ── Cooking unit options ───────────────────────────────────────────────────────

const COOKING_UNIT_OPTIONS = [
	{ value: "g", label: "g — gram" },
	{ value: "kg", label: "kg" },
	{ value: "oz", label: "oz — ounce" },
	{ value: "lb", label: "lb — pound" },
	{ value: "ml", label: "ml" },
	{ value: "l", label: "l — liter" },
	{ value: "tsp", label: "tsp — teaspoon" },
	{ value: "tbsp", label: "tbsp — tablespoon" },
	{ value: "fl_oz", label: "fl oz" },
	{ value: "cup", label: "cup" },
	{ value: "unit", label: "unit — each" },
];

// ─── IngredientRow ────────────────────────────────────────────────────────────

function IngredientRow({
	ingredient,
	prepId,
	editable = true,
	onUpdated,
	onDeleted,
}) {
	const editing = useSignal(false);
	const name = useSignal(ingredient.name);
	const amount = useSignal(String(ingredient.amount));
	const unit = useSignal(ingredient.unit);
	const prep = useSignal(ingredient.prep ?? "");
	const saving = useSignal(false);
	const error = useSignal("");

	const handleSave = async () => {
		saving.value = true;
		error.value = "";
		try {
			const body =
				ingredient.preparation_ref_id != null
					? {
							preparation_ref_id: ingredient.preparation_ref_id,
							name: name.value.trim() || ingredient.name,
							amount: parseFloat(amount.value) || ingredient.amount,
							unit: unit.value.trim(),
							prep: prep.value.trim() || undefined,
						}
					: {
							ingredient_id: ingredient.ingredient_id,
							name: name.value.trim() || ingredient.name,
							amount: parseFloat(amount.value) || ingredient.amount,
							unit: unit.value.trim(),
							prep: prep.value.trim() || undefined,
						};
			const r = await apiFetch(
				`/api/preparations/${prepId}/ingredients/${ingredient.id}`,
				{
					method: "PUT",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(body),
				},
			);
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to update ingredient");
			}
			const updated = await r.json();
			onUpdated(updated);
			editing.value = false;
		} catch (err) {
			error.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const handleDelete = async () => {
		const r = await apiFetch(
			`/api/preparations/${prepId}/ingredients/${ingredient.id}`,
			{ method: "DELETE" },
		);
		if (!r.ok) {
			const d = await r.json().catch(() => ({}));
			error.value = d.error || "Failed to delete ingredient";
			return;
		}
		onDeleted(ingredient.id);
	};

	if (editing.value) {
		return (
			<div
				class="flex flex-col gap-2 py-2 px-3 rounded-lg"
				style={{ background: "var(--color-bg)" }}
			>
				<div class="flex gap-2">
					<div class="flex-1">
						<TextField
							label="Name"
							value={name.value}
							onInput={(e) => {
								name.value = e.target.value;
							}}
						/>
					</div>
					<div style={{ width: "80px" }}>
						<TextField
							label="Amount"
							type="number"
							step={DISPLAY_STEPS[unit.value] ?? "any"}
							value={amount.value}
							onInput={(e) => {
								amount.value = e.target.value;
							}}
						/>
					</div>
					<div style={{ width: "130px" }}>
						<Combobox
							label="Unit"
							value={unit.value}
							onChange={(v) => {
								const a = parseFloat(amount.value);
								if (!Number.isNaN(a) && amount.value !== "") {
									amount.value = String(
										convertUnit(a, unit.value, v, ingredient.density_g_per_ml),
									);
								}
								unit.value = v;
							}}
							options={COOKING_UNIT_OPTIONS}
							placeholder="unit…"
						/>
					</div>
				</div>
				<TextField
					label="Prep"
					placeholder="diced, melted, room temp…"
					value={prep.value}
					onInput={(e) => {
						prep.value = e.target.value;
					}}
				/>
				{error.value && (
					<p class="text-xs" style={{ color: "var(--color-error)" }}>
						{error.value}
					</p>
				)}
				<div class="flex gap-2 justify-end">
					<Button
						variant="ghost"
						size="sm"
						type="button"
						onClick={() => {
							editing.value = false;
						}}
					>
						Cancel
					</Button>
					<Button
						size="sm"
						type="button"
						disabled={saving.value}
						onClick={handleSave}
					>
						{saving.value ? "Saving…" : "Save"}
					</Button>
				</div>
			</div>
		);
	}

	return (
		<>
			<div class="flex items-center gap-2 py-1.5 px-1 group rounded">
				<span class="flex-1 text-sm" style={{ color: "var(--color-text)" }}>
					<span class="font-medium">
						{ingredient.amount}
						{ingredient.unit && ingredient.unit !== "unit"
							? ` ${ingredient.unit}`
							: ""}
					</span>{" "}
					{ingredient.name}
					{ingredient.prep && (
						<span class="text-xs ml-1" style={{ color: "var(--color-muted)" }}>
							({ingredient.prep})
						</span>
					)}
				</span>
				{editable && (
					<>
						<Button
							variant="ghost"
							size="icon"
							type="button"
							class="opacity-0 group-hover:opacity-100"
							onClick={() => {
								editing.value = true;
							}}
						>
							<span class="text-xs" style={{ color: "var(--color-muted)" }}>
								Edit
							</span>
						</Button>
						<Button
							variant="ghost"
							size="icon"
							type="button"
							class="opacity-0 group-hover:opacity-100"
							onClick={handleDelete}
						>
							<X
								size={14}
								aria-hidden="true"
								style={{ color: "var(--color-muted)" }}
							/>
						</Button>
					</>
				)}
			</div>
			{error.value && (
				<p class="text-xs px-1" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}
		</>
	);
}

// ─── AddIngredientForm ────────────────────────────────────────────────────────

// mode: "select"  — combobox: pick existing or trigger FDC creation
//       "fdc"     — FDC search panel; ingredient not yet created
//       "confirm" — FDC entry chosen; fill amount/unit/prep and save
function AddIngredientForm({ prepId, onAdded, onCancel }) {
	const { cookingMassUnit } = useUnitPreferences();

	const ingredients = useSignal([]);
	const preparations = useSignal([]);
	const selectedKey = useSignal(""); // "ing:<id>" or "prep:<id>"
	const fdcQuery = useSignal("");
	const selectedFDC = useSignal(null);
	const name = useSignal("");
	const amount = useSignal("1");
	const unit = useSignal(cookingMassUnit);
	const prepNote = useSignal("");
	const saving = useSignal(false);
	const error = useSignal("");
	const mode = useSignal("select");
	const density = useSignal(null);

	useEffect(() => {
		apiFetch("/api/ingredients")
			.then((r) => r.json())
			.then((data) => {
				ingredients.value = data;
			});
		apiFetch("/api/preparations")
			.then((r) => r.json())
			.then((data) => {
				preparations.value = data;
			});
	}, []);

	const handleIngredientChange = (val) => {
		if (val.startsWith("ing:")) {
			const id = val.slice("ing:".length);
			selectedKey.value = val;
			const found = ingredients.value.find((i) => String(i.id) === id);
			if (found) {
				name.value = found.name;
				density.value = found.density_g_per_ml ?? null;
			}
		} else if (val.startsWith("prep:")) {
			const id = val.slice("prep:".length);
			selectedKey.value = val;
			const found = preparations.value.find((p) => String(p.id) === id);
			if (found) {
				name.value = found.name;
				density.value = null;
			}
		} else {
			// Freeform "Create '...' from FDC" was selected
			fdcQuery.value = val;
			mode.value = "fdc";
		}
	};

	const handleFDCSelect = (food) => {
		selectedFDC.value = food;
		name.value = food.name;
		mode.value = "confirm";
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (mode.value === "select" && !selectedKey.value) {
			error.value = "Select an ingredient or preparation first";
			return;
		}
		saving.value = true;
		error.value = "";
		try {
			let refField;
			if (mode.value === "select") {
				if (selectedKey.value.startsWith("ing:")) {
					refField = {
						ingredient_id: Number(selectedKey.value.slice("ing:".length)),
					};
				} else if (selectedKey.value.startsWith("prep:")) {
					refField = {
						preparation_ref_id: Number(selectedKey.value.slice("prep:".length)),
					};
				}
			} else {
				// confirm mode — create the FDC ingredient first, then use its id
				const f = selectedFDC.value;
				const r = await apiFetch("/api/ingredients", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						name: name.value.trim(),
						fdc_id: f.fdc_id,
						calories_per_100g: f.calories_per_100g,
						protein_per_100g: f.protein_per_100g,
						fat_per_100g: f.fat_per_100g,
						carbs_per_100g: f.carbs_per_100g,
						is_public: true,
					}),
				});
				if (!r.ok) {
					const d = await r.json();
					throw new Error(d.error || "Failed to create ingredient");
				}
				const created = await r.json();
				refField = { ingredient_id: created.id };
				density.value = created.density_g_per_ml ?? null;
			}

			const r = await apiFetch(`/api/preparations/${prepId}/ingredients`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					...refField,
					name: name.value.trim(),
					amount: parseFloat(amount.value) || 1,
					unit: unit.value.trim(),
					prep: prepNote.value.trim() || undefined,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to add ingredient");
			}
			onAdded(await r.json());
		} catch (err) {
			error.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const options = [
		...ingredients.value.map((i) => ({
			value: `ing:${i.id}`,
			label: i.name,
		})),
		...preparations.value
			.filter((p) => String(p.id) !== String(prepId))
			.map((p) => ({
				value: `prep:${p.id}`,
				label: `↳ ${p.name}`,
			})),
	];

	const showPrepFields =
		mode.value === "confirm" ||
		(mode.value === "select" && !!selectedKey.value);

	return (
		<form
			onSubmit={handleSave}
			class="flex flex-col gap-2 py-2 px-3 rounded-lg"
			style={{ background: "var(--color-bg)" }}
		>
			{mode.value === "select" && (
				<Combobox
					autoFocus
					label="Ingredient or sub-preparation"
					value={selectedKey.value}
					onChange={handleIngredientChange}
					options={options}
					placeholder="Search ingredients…"
					freeform
					freeformLabel={(q) => `Create "${q}" from FDC`}
				/>
			)}

			{mode.value === "fdc" && (
				<FDCSearch initialQuery={fdcQuery.value} onSelect={handleFDCSelect} />
			)}

			{mode.value === "confirm" && (
				<div class="flex flex-col gap-0.5">
					<p class="text-xs font-medium" style={{ color: "var(--color-text)" }}>
						{selectedFDC.value?.name}
					</p>
					<p class="text-xs" style={{ color: "var(--color-muted)" }}>
						{selectedFDC.value?.calories_per_100g} kcal ·{" "}
						{selectedFDC.value?.protein_per_100g}g protein ·{" "}
						{selectedFDC.value?.fat_per_100g}g fat ·{" "}
						{selectedFDC.value?.carbs_per_100g}g carbs (per 100g)
					</p>
				</div>
			)}

			{showPrepFields && (
				<>
					{mode.value === "confirm" && (
						<TextField
							id="ing-name"
							label="Name (display)"
							value={name.value}
							onInput={(e) => {
								name.value = e.target.value;
							}}
						/>
					)}
					<div class="flex gap-2">
						<div class="flex-1">
							<TextField
								id="ing-amount"
								label="Amount"
								type="number"
								step={DISPLAY_STEPS[unit.value] ?? "any"}
								value={amount.value}
								onInput={(e) => {
									amount.value = e.target.value;
								}}
							/>
						</div>
						<div class="flex-1">
							<Combobox
								id="ing-unit"
								label="Unit"
								value={unit.value}
								onChange={(v) => {
									const a = parseFloat(amount.value);
									if (!Number.isNaN(a) && amount.value !== "") {
										amount.value = String(
											convertUnit(a, unit.value, v, density.value),
										);
									}
									unit.value = v;
								}}
								options={COOKING_UNIT_OPTIONS}
								placeholder="unit…"
							/>
						</div>
					</div>
					<TextField
						id="ing-prep"
						label="Prep"
						placeholder="diced, melted, room temp…"
						value={prepNote.value}
						onInput={(e) => {
							prepNote.value = e.target.value;
						}}
					/>
				</>
			)}

			{error.value && (
				<p class="text-xs" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}

			<div class="flex gap-2 justify-end">
				<Button variant="ghost" size="sm" type="button" onClick={onCancel}>
					Cancel
				</Button>
				{mode.value === "fdc" && (
					<Button
						variant="ghost"
						size="sm"
						type="button"
						onClick={() => {
							mode.value = "select";
						}}
					>
						← Back
					</Button>
				)}
				{mode.value === "confirm" && (
					<Button
						variant="ghost"
						size="sm"
						type="button"
						onClick={() => {
							mode.value = "fdc";
						}}
					>
						← Change
					</Button>
				)}
				{showPrepFields && (
					<Button
						data-testid="add-ingredient-submit"
						size="sm"
						type="submit"
						disabled={saving.value}
					>
						{saving.value ? "Adding…" : "Add"}
					</Button>
				)}
			</div>
		</form>
	);
}

// ─── StepRow ──────────────────────────────────────────────────────────────────

function StepRow({
	index,
	description,
	onChange,
	onDelete,
	autoFocus = false,
}) {
	const inputRef = useRef(null);
	const local = useSignal(description);

	useEffect(() => {
		if (autoFocus) inputRef.current?.focus();
	}, []);

	const handleBlur = () => {
		if (local.value.trim() !== description) {
			onChange(index, local.value.trim());
		}
	};

	return (
		<div class="flex items-start gap-2 group py-1">
			<span
				class="shrink-0 text-xs font-medium tabular-nums mt-2.5"
				style={{ color: "var(--color-muted)", width: "1.25rem" }}
			>
				{index + 1}.
			</span>
			<TextField
				multiline
				inputRef={inputRef}
				containerClass="flex-1"
				class="rounded bg-transparent border-transparent p-1 focus:ring-1"
				rows={1}
				value={local.value}
				onInput={(e) => {
					local.value = e.target.value;
					e.target.style.height = "auto";
					e.target.style.height = `${e.target.scrollHeight}px`;
				}}
				onFocus={(e) => {
					e.target.style.borderColor = "var(--color-border)";
				}}
				onBlur={(e) => {
					e.target.style.borderColor = "transparent";
					handleBlur();
				}}
			/>
			<Button
				variant="ghost"
				size="icon"
				type="button"
				class="opacity-0 group-hover:opacity-100 mt-1"
				onClick={() => onDelete(index)}
			>
				<X
					size={14}
					aria-hidden="true"
					style={{ color: "var(--color-muted)" }}
				/>
			</Button>
		</div>
	);
}

// ─── PrepEditModal ─────────────────────────────────────────────────────────────

function PrepEditModal({ prep: initialPrep, onSaved, onClose }) {
	const prep = useSignal({
		...initialPrep,
		steps: (initialPrep.steps ?? []).map((s) => ({
			...s,
			_key: s._key,
		})),
		ingredients: initialPrep.ingredients ?? [],
	});
	const name = useSignal(initialPrep.name);
	const yieldAmount = useSignal(String(initialPrep.yield_amount ?? ""));
	const yieldUnit = useSignal(initialPrep.yield_unit ?? "");
	const description = useSignal(initialPrep.description ?? "");

	const stepsDirty = useSignal(false);
	const saving = useSignal(false);
	const modalError = useSignal("");
	const showAddIngredient = useSignal(false);
	const focusNextStepKey = useSignal(null);

	const saveConfirmDialog = useDialog();
	const cancelConfirmDialog = useDialog();

	const isDirty = () => {
		if (stepsDirty.value) return true;
		if (name.value !== initialPrep.name) return true;
		if (yieldAmount.value !== String(initialPrep.yield_amount ?? ""))
			return true;
		if (yieldUnit.value !== (initialPrep.yield_unit ?? "")) return true;
		if (description.value !== (initialPrep.description ?? "")) return true;
		return false;
	};

	const handleStepChange = (index, newDesc) => {
		prep.value = {
			...prep.value,
			steps: prep.value.steps.map((s, i) =>
				i === index ? { ...s, description: newDesc } : s,
			),
		};
		stepsDirty.value = true;
	};

	const handleStepDelete = (index) => {
		prep.value = {
			...prep.value,
			steps: prep.value.steps.filter((_, i) => i !== index),
		};
		stepsDirty.value = true;
	};

	const handleAddStep = () => {
		const newKey = crypto.randomUUID();
		focusNextStepKey.value = newKey;
		prep.value = {
			...prep.value,
			steps: [...prep.value.steps, { description: "", _key: newKey }],
		};
		stepsDirty.value = true;
	};

	const handleIngredientAdded = (ing) => {
		prep.value = {
			...prep.value,
			ingredients: [...prep.value.ingredients, ing],
		};
		showAddIngredient.value = false;
	};

	const handleIngredientUpdated = (updated) => {
		prep.value = {
			...prep.value,
			ingredients: prep.value.ingredients.map((i) =>
				i.id === updated.id ? updated : i,
			),
		};
	};

	const handleIngredientDeleted = (id) => {
		prep.value = {
			...prep.value,
			ingredients: prep.value.ingredients.filter((i) => i.id !== id),
		};
	};

	const doSave = async () => {
		saving.value = true;
		modalError.value = "";
		try {
			const r = await apiFetch(`/api/preparations/${prep.value.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: name.value.trim() || initialPrep.name,
					description: description.value || undefined,
					yield_amount: parseFloat(yieldAmount.value) || 0,
					yield_unit: yieldUnit.value.trim(),
					// eslint-disable-next-line no-unused-vars
					steps: prep.value.steps.map(({ _key, ...s }) => s),
					is_public: prep.value.is_public,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save preparation");
			}
			const updated = await r.json();
			onSaved({ ...updated, ingredients: prep.value.ingredients });
			onClose();
		} catch (err) {
			modalError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	// Always propagate current ingredient state so the read-only view stays in
	// sync even when ingredient changes were persisted but the modal is cancelled.
	const syncAndClose = () => {
		onSaved({ ...initialPrep, ingredients: prep.value.ingredients });
		onClose();
	};

	const handleSave = () => {
		if (isDirty()) {
			saveConfirmDialog.show();
		} else {
			syncAndClose();
		}
	};

	const handleClose = () => {
		if (isDirty()) {
			cancelConfirmDialog.show();
		} else {
			syncAndClose();
		}
	};

	return (
		<DialogContent fullscreen>
			{/* Header */}
			<div
				class="flex items-center justify-between px-4 py-3 shrink-0 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<DialogTitle>{name.value || initialPrep.name}</DialogTitle>
				<Button
					variant="ghost"
					size="icon"
					type="button"
					onClick={handleClose}
					aria-label="Close"
				>
					<X size={16} aria-hidden="true" />
				</Button>
			</div>

			{/* Scrollable body */}
			<div class="flex-1 overflow-y-auto">
				<div class="flex flex-col gap-5 px-4 py-4">
					{/* Name + Yield */}
					<div class="flex gap-3">
						<div class="flex-1">
							<TextField
								label="Name"
								value={name.value}
								onInput={(e) => {
									name.value = e.target.value;
								}}
							/>
						</div>
						<div style={{ width: "80px" }}>
							<TextField
								label="Yield"
								type="number"
								value={yieldAmount.value}
								onInput={(e) => {
									yieldAmount.value = e.target.value;
								}}
							/>
						</div>
						<div style={{ width: "130px" }}>
							<Combobox
								label="Unit"
								value={yieldUnit.value}
								onChange={(v) => {
									yieldUnit.value = v;
								}}
								options={COOKING_UNIT_OPTIONS}
								placeholder="unit…"
							/>
						</div>
					</div>

					{/* Description */}
					<div>
						<span
							class="text-xs font-semibold uppercase tracking-wider mb-2 block"
							style={{ color: "var(--color-muted)" }}
						>
							Description
						</span>
						<EditableMarkdown
							value={description.value}
							placeholder="Add a description…"
							onSave={(v) => {
								description.value = v;
							}}
						/>
					</div>

					{/* Ingredients */}
					<div>
						<span
							class="text-xs font-semibold uppercase tracking-wider mb-2 block"
							style={{ color: "var(--color-muted)" }}
						>
							Ingredients
						</span>
						<div class="flex flex-col">
							{prep.value.ingredients.map((ing) => (
								<IngredientRow
									key={ing.id}
									ingredient={ing}
									prepId={prep.value.id}
									editable
									onUpdated={handleIngredientUpdated}
									onDeleted={handleIngredientDeleted}
								/>
							))}
						</div>
						{showAddIngredient.value ? (
							<div class="mt-2">
								<AddIngredientForm
									prepId={prep.value.id}
									onAdded={handleIngredientAdded}
									onCancel={() => {
										showAddIngredient.value = false;
									}}
								/>
							</div>
						) : (
							<Button
								variant="ghost"
								type="button"
								class="w-full justify-start text-sm font-normal px-1"
								style={{ color: "var(--color-muted)" }}
								onClick={() => {
									showAddIngredient.value = true;
								}}
							>
								Add ingredient…
							</Button>
						)}
					</div>

					{/* Steps */}
					<div>
						<span
							class="text-xs font-semibold uppercase tracking-wider mb-2 block"
							style={{ color: "var(--color-muted)" }}
						>
							Steps
						</span>
						<div class="flex flex-col">
							{prep.value.steps.map((step, i) => (
								<StepRow
									key={step._key}
									index={i}
									description={step.description}
									onChange={handleStepChange}
									onDelete={handleStepDelete}
									autoFocus={step._key === focusNextStepKey.value}
								/>
							))}
						</div>
						<div class="flex items-start gap-2 py-1">
							<span
								class="shrink-0 text-xs font-medium tabular-nums mt-2.5"
								style={{
									color: "var(--color-muted)",
									width: "1.25rem",
									opacity: 0.5,
								}}
							>
								{prep.value.steps.length + 1}.
							</span>
							<Button
								variant="ghost"
								type="button"
								class="flex-1 justify-start text-sm font-normal px-1"
								style={{ color: "var(--color-muted)" }}
								onClick={handleAddStep}
							>
								Add a step…
							</Button>
						</div>
					</div>

					{/* Error */}
					{modalError.value && (
						<p class="text-xs" style={{ color: "var(--color-error)" }}>
							{modalError.value}
						</p>
					)}
				</div>
			</div>

			{/* Footer */}
			<div
				class="flex gap-2 justify-end px-4 py-3 shrink-0 border-t"
				style={{ borderColor: "var(--color-border)" }}
			>
				<Button variant="ghost" type="button" onClick={handleClose}>
					Cancel
				</Button>
				<Button type="button" disabled={saving.value} onClick={handleSave}>
					{saving.value ? "Saving…" : "Save"}
				</Button>
			</div>

			{/* Confirm save dialog */}
			<ConfirmDialog
				openSignal={saveConfirmDialog.open}
				title="Save changes"
				description="Save all changes to this preparation?"
				confirmLabel="Save"
				onConfirm={doSave}
			/>

			{/* Confirm discard dialog */}
			<ConfirmDialog
				openSignal={cancelConfirmDialog.open}
				title="Discard changes"
				description="You have unsaved changes. Discard them and close?"
				confirmLabel="Discard"
				onConfirm={syncAndClose}
			/>
		</DialogContent>
	);
}

// ─── PreparationSection ───────────────────────────────────────────────────────

function PreparationSection({
	link: initialLink,
	prep: initialPrep,
	onUpdated,
	onRemoved,
}) {
	const prep = useSignal({
		...initialPrep,
		steps: (initialPrep.steps ?? []).map((s) => ({
			...s,
			_key: crypto.randomUUID(),
		})),
	});
	const link = useSignal(initialLink);
	const removeDialog = useDialog();
	const editDialog = useDialog();
	const removeError = useSignal("");

	const handleRemove = async () => {
		const r = await apiFetch(
			`/api/recipes/${link.value.recipe_id}/preparations/${link.value.id}`,
			{ method: "DELETE" },
		);
		if (!r.ok) {
			const d = await r.json().catch(() => ({}));
			removeError.value = d.error || "Failed to remove component";
			return;
		}
		onRemoved(link.value.id);
	};

	const handleSaved = (updated) => {
		prep.value = {
			...updated,
			steps: (updated.steps ?? []).map((s) => ({
				...s,
				_key: crypto.randomUUID(),
			})),
		};
		onUpdated(updated);
	};

	return (
		<>
			<AccordionItem value={String(initialLink.id)}>
				<AccordionTrigger>
					<div class="flex flex-1 items-center gap-2 min-w-0">
						<span class="truncate font-medium text-sm">{prep.value.name}</span>
						{prep.value.yield_amount > 0 && (
							<span
								class="text-xs shrink-0 ml-auto"
								style={{ color: "var(--color-muted)" }}
							>
								{prep.value.yield_amount} {prep.value.yield_unit}
							</span>
						)}
					</div>
					<Button
						variant="ghost"
						size="icon"
						type="button"
						onClick={(e) => {
							e.stopPropagation();
							editDialog.show();
						}}
						aria-label="Edit preparation"
					>
						<Pencil
							size={14}
							aria-hidden="true"
							style={{ color: "var(--color-muted)" }}
						/>
					</Button>
					<Button
						variant="ghost"
						size="icon"
						type="button"
						class="ml-2"
						onClick={(e) => {
							e.stopPropagation();
							removeDialog.show();
						}}
					>
						<Trash2
							size={14}
							aria-hidden="true"
							style={{ color: "var(--color-muted)" }}
						/>
					</Button>
				</AccordionTrigger>
				<AccordionContent>
					<div class="flex flex-col gap-4 px-4 py-3">
						{prep.value.description && (
							<Markdown>{prep.value.description}</Markdown>
						)}
						<div>
							<span
								class="text-xs font-semibold uppercase tracking-wider"
								style={{ color: "var(--color-muted)" }}
							>
								Amount in recipe
							</span>
							<p class="mt-1 text-sm" style={{ color: "var(--color-muted)" }}>
								{link.value.amount} {link.value.unit}
							</p>
						</div>
						{prep.value.ingredients && prep.value.ingredients.length > 0 && (
							<div>
								<span
									class="text-xs font-semibold uppercase tracking-wider mb-2 block"
									style={{ color: "var(--color-muted)" }}
								>
									Ingredients
								</span>
								<div class="flex flex-col">
									{prep.value.ingredients.map((ing) => (
										<IngredientRow
											key={ing.id}
											ingredient={ing}
											prepId={prep.value.id}
											editable={false}
										/>
									))}
								</div>
							</div>
						)}
						{prep.value.steps && prep.value.steps.length > 0 && (
							<div>
								<span
									class="text-xs font-semibold uppercase tracking-wider mb-2 block"
									style={{ color: "var(--color-muted)" }}
								>
									Steps
								</span>
								<div class="flex flex-col gap-1">
									{prep.value.steps.map((step, i) => (
										<p
											key={step._key ?? i}
											class="text-sm"
											style={{ color: "var(--color-text)" }}
										>
											<span
												class="font-medium tabular-nums mr-1"
												style={{ color: "var(--color-muted)" }}
											>
												{i + 1}.
											</span>
											{step.description}
										</p>
									))}
								</div>
							</div>
						)}
					</div>
				</AccordionContent>
			</AccordionItem>

			{removeError.value && (
				<p class="px-4 text-xs" style={{ color: "var(--color-error)" }}>
					{removeError.value}
				</p>
			)}
			<ConfirmDialog
				openSignal={removeDialog.open}
				title="Remove component"
				description={`Remove "${prep.value.name}" from this recipe? The preparation itself will not be deleted.`}
				confirmLabel="Remove"
				onConfirm={handleRemove}
			/>
			<Dialog openSignal={editDialog.open}>
				<PrepEditModal
					prep={prep.value}
					onSaved={handleSaved}
					onClose={editDialog.hide}
				/>
			</Dialog>
		</>
	);
}

// ─── AddComponentForm ─────────────────────────────────────────────────────────

function AddComponentForm({ recipeId, position, onAdded, onCancel }) {
	const preparations = useSignal([]);
	const comboValue = useSignal("");
	const selectedPrepId = useSignal(null); // non-null only when an existing prep is selected
	const saving = useSignal(false);
	const error = useSignal("");

	useEffect(() => {
		apiFetch("/api/preparations")
			.then((r) => r.json())
			.then((data) => {
				preparations.value = data;
			});
	}, []);

	const handleSubmit = async (e) => {
		e.preventDefault();
		const val = comboValue.value.trim();
		if (!val) {
			error.value = "Select or name a preparation.";
			return;
		}

		saving.value = true;
		error.value = "";
		try {
			let preparationId;

			if (selectedPrepId.value !== null) {
				// User picked an existing preparation from the dropdown
				preparationId = selectedPrepId.value;
			} else {
				// User typed a new name — create a new preparation
				const cr = await apiFetch("/api/preparations", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						name: val,
						yield_amount: 1,
						yield_unit: "unit",
					}),
				});
				if (!cr.ok) {
					const d = await cr.json();
					throw new Error(d.error || "Failed to create preparation");
				}
				const created = await cr.json();
				preparationId = created.id;
			}

			// Link to recipe with placeholder amount — editable later from the section
			const lr = await apiFetch(`/api/recipes/${recipeId}/preparations`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					preparation_id: preparationId,
					position,
					amount: 1,
					unit: "unit",
				}),
			});
			if (!lr.ok) {
				const d = await lr.json();
				throw new Error(d.error || "Failed to add component");
			}
			const link = await lr.json();

			// Fetch full preparation detail
			const pr = await apiFetch(`/api/preparations/${preparationId}`);
			const prep = await pr.json();

			onAdded({ link, prep });
		} catch (err) {
			error.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const options = preparations.value.map((p) => ({
		value: `existing:${p.id}`,
		label: p.name,
	}));

	return (
		<form
			onSubmit={handleSubmit}
			class="flex flex-col gap-3 p-4 rounded-xl border"
			style={{
				borderColor: "var(--color-border)",
				background: "var(--color-surface)",
			}}
		>
			<p
				class="text-xs font-semibold uppercase tracking-wider"
				style={{ color: "var(--color-muted)" }}
			>
				Add Component
			</p>
			<Combobox
				label="Preparation"
				value={comboValue.value}
				onChange={(v) => {
					const s = String(v);
					if (s.startsWith("existing:")) {
						selectedPrepId.value = Number(s.slice("existing:".length));
					} else {
						selectedPrepId.value = null;
					}
					comboValue.value = s;
				}}
				options={options}
				placeholder="Search or type a name to create…"
				freeform
			/>
			{error.value && (
				<p class="text-xs" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}
			<div class="flex gap-2 justify-end">
				<Button variant="outline" size="sm" type="button" onClick={onCancel}>
					Cancel
				</Button>
				<Button size="sm" type="submit" disabled={saving.value}>
					{saving.value ? "Adding…" : "Add"}
				</Button>
			</div>
		</form>
	);
}

// ─── RecipeDetail ─────────────────────────────────────────────────────────────

export function RecipeDetail({ recipeId, onRecipeUpdated }) {
	const recipe = useSignal(null);
	const components = useSignal([]); // [{link, prep}]
	const openItems = useSignal([]);
	const loading = useSignal(true);
	const error = useSignal("");
	const showAddForm = useSignal(false);

	// Editable recipe name
	const editingName = useSignal(false);
	const nameValue = useSignal("");
	const nameRef = useRef(null);

	useEffect(() => {
		if (editingName.value) nameRef.current?.focus();
	}, [editingName.value]);

	const loadRecipe = async () => {
		if (!recipeId) return;
		loading.value = true;
		error.value = "";
		try {
			const r = await apiFetch(`/api/recipes/${recipeId}`);
			if (!r.ok) throw new Error("Failed to load recipe");
			const data = await r.json();
			recipe.value = data;
			nameValue.value = data.name;

			// Fetch each preparation in parallel
			const preps = await Promise.all(
				(data.preparations ?? []).map(async (link) => {
					const pr = await apiFetch(`/api/preparations/${link.preparation_id}`);
					const prep = await pr.json();
					return { link, prep };
				}),
			);
			components.value = preps;
			openItems.value = preps.map(({ link }) => String(link.id));
		} catch (err) {
			error.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		loadRecipe();
	}, [recipeId]);

	const saveName = async () => {
		const trimmed = nameValue.value.trim();
		if (!trimmed || trimmed === recipe.value.name) {
			editingName.value = false;
			nameValue.value = recipe.value.name;
			return;
		}
		try {
			const r = await apiFetch(`/api/recipes/${recipeId}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: trimmed,
					description: recipe.value.description,
					yield_amount: recipe.value.yield_amount,
					yield_unit: recipe.value.yield_unit,
					servings: recipe.value.servings,
					is_public: recipe.value.is_public,
				}),
			});
			if (!r.ok) throw new Error("Failed to update recipe");
			const updated = await r.json();
			recipe.value = updated;
			onRecipeUpdated?.();
		} catch {
			nameValue.value = recipe.value.name;
		}
		editingName.value = false;
	};

	const handleComponentAdded = ({ link, prep }) => {
		components.value = [...components.value, { link, prep }];
		openItems.value = [...openItems.value, String(link.id)];
		showAddForm.value = false;
	};

	const handleComponentRemoved = (linkId) => {
		components.value = components.value.filter((c) => c.link.id !== linkId);
	};

	const handlePrepUpdated = (updated) => {
		components.value = components.value.map((c) =>
			c.prep.id === updated.id ? { ...c, prep: updated } : c,
		);
	};

	if (loading.value) {
		return (
			<p class="p-4 text-sm" style={{ color: "var(--color-muted)" }}>
				Loading…
			</p>
		);
	}
	if (error.value) {
		return (
			<p class="p-4 text-sm" style={{ color: "var(--color-error)" }}>
				{error.value}
			</p>
		);
	}
	if (!recipe.value) return null;

	return (
		<div class="page-content">
			{/* Recipe name */}
			<div>
				{editingName.value ? (
					<TextField
						inline
						inputRef={nameRef}
						class="text-xl font-semibold"
						value={nameValue.value}
						onInput={(e) => {
							nameValue.value = e.target.value;
						}}
						onBlur={saveName}
						onKeyDown={(e) => {
							if (e.key === "Enter") saveName();
							if (e.key === "Escape") {
								nameValue.value = recipe.value.name;
								editingName.value = false;
							}
						}}
					/>
				) : (
					<Button
						variant="unstyled"
						type="button"
						class="text-xl font-semibold cursor-text"
						style={{ color: "var(--color-text)" }}
						onClick={() => {
							editingName.value = true;
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter" || e.key === " ") editingName.value = true;
						}}
					>
						{recipe.value.name}
					</Button>
				)}
				<p class="mt-0.5 text-sm" style={{ color: "var(--color-muted)" }}>
					{recipe.value.servings} serving
					{recipe.value.servings !== 1 ? "s" : ""}
				</p>
			</div>

			{/* Preparations */}
			<div class="flex flex-col gap-3">
				<div class="flex items-center justify-between">
					<span
						class="text-xs font-semibold uppercase tracking-wider"
						style={{ color: "var(--color-muted)" }}
					>
						Preparations
					</span>
					{!showAddForm.value && (
						<Button
							variant="ghost"
							size="sm"
							type="button"
							onClick={() => {
								showAddForm.value = true;
							}}
						>
							<Plus size={14} aria-hidden="true" />
							<span class="ml-1">Add</span>
						</Button>
					)}
				</div>

				{components.value.length === 0 && !showAddForm.value && (
					<p class="text-sm" style={{ color: "var(--color-muted)" }}>
						No preparations yet. Add one to get started.
					</p>
				)}

				{components.value.length > 0 && (
					<Accordion
						type="multiple"
						value={openItems.value}
						onValueChange={(v) => {
							openItems.value = v;
						}}
					>
						{components.value.map(({ link, prep }) => (
							<PreparationSection
								key={link.id}
								link={link}
								prep={prep}
								onUpdated={handlePrepUpdated}
								onRemoved={handleComponentRemoved}
							/>
						))}
					</Accordion>
				)}

				{showAddForm.value && (
					<AddComponentForm
						recipeId={recipeId}
						position={components.value.length + 1}
						onAdded={handleComponentAdded}
						onCancel={() => {
							showAddForm.value = false;
						}}
					/>
				)}
			</div>
		</div>
	);
}
