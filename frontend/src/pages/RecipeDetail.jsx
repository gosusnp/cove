// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { Plus, Trash2, X } from "lucide-preact";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "../components/ui/Accordion.jsx";
import { Button } from "../components/ui/Button.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";

// ─── IngredientRow ────────────────────────────────────────────────────────────

function IngredientRow({ ingredient, prepId, onUpdated, onDeleted }) {
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
			const r = await apiFetch(
				`/api/preparations/${prepId}/ingredients/${ingredient.id}`,
				{
					method: "PUT",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						ingredient_id: ingredient.ingredient_id,
						name: name.value.trim() || ingredient.name,
						amount: parseFloat(amount.value) || ingredient.amount,
						unit: unit.value.trim(),
						prep: prep.value.trim() || undefined,
					}),
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
							value={amount.value}
							onInput={(e) => {
								amount.value = e.target.value;
							}}
						/>
					</div>
					<div style={{ width: "80px" }}>
						<TextField
							label="Unit"
							value={unit.value}
							onInput={(e) => {
								unit.value = e.target.value;
							}}
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
						{ingredient.unit ? ` ${ingredient.unit}` : ""}
					</span>{" "}
					{ingredient.name}
					{ingredient.prep && (
						<span class="text-xs ml-1" style={{ color: "var(--color-muted)" }}>
							({ingredient.prep})
						</span>
					)}
				</span>
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
			</div>
			{error.value && (
				<p class="text-xs px-1" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}
		</>
	);
}

// ─── FDCSearch ────────────────────────────────────────────────────────────────

function FDCSearch({ initialQuery = "", onSelect }) {
	const query = useSignal(initialQuery);
	const results = useSignal(null);
	const loading = useSignal(false);
	const error = useSignal("");

	const search = async () => {
		if (!query.value.trim()) return;
		loading.value = true;
		error.value = "";
		try {
			const r = await apiFetch(
				`/api/fdc/search?q=${encodeURIComponent(query.value.trim())}`,
			);
			if (!r.ok) throw new Error("Search failed");
			const data = await r.json();
			results.value = data.foods ?? [];
		} catch (err) {
			error.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		if (initialQuery.trim()) search();
	}, []);

	return (
		<div class="flex flex-col gap-2">
			<div class="flex gap-2 items-end">
				<div class="flex-1">
					<TextField
						id="fdc-query"
						label="Search FDC"
						value={query.value}
						onInput={(e) => {
							query.value = e.target.value;
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								search();
							}
						}}
					/>
				</div>
				<Button
					size="sm"
					type="button"
					onClick={search}
					disabled={loading.value}
				>
					{loading.value ? "…" : "Search"}
				</Button>
			</div>
			{error.value && (
				<p class="text-xs" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}
			{results.value !== null && results.value.length === 0 && (
				<p class="text-xs" style={{ color: "var(--color-muted)" }}>
					No results found.
				</p>
			)}
			{results.value !== null && results.value.length > 0 && (
				<div
					class="flex flex-col overflow-y-auto rounded border max-h-48"
					style={{ borderColor: "var(--color-border)" }}
				>
					{results.value.map((food) => (
						<button
							key={food.fdc_id}
							type="button"
							class="flex items-center gap-2 px-2 py-1.5 text-left text-xs hover:bg-[var(--color-bg)] transition-colors"
							onClick={() => onSelect(food)}
						>
							<span class="flex-1" style={{ color: "var(--color-text)" }}>
								{food.name}
							</span>
							<span
								class="shrink-0 rounded px-1 py-0.5 text-xs"
								style={
									food.data_type === "Foundation"
										? { background: "var(--color-accent)", color: "#fff" }
										: {
												background: "var(--color-surface)",
												color: "var(--color-muted)",
											}
								}
							>
								{food.data_type === "Foundation" ? "Foundation" : "SR Legacy"}
							</span>
						</button>
					))}
				</div>
			)}
		</div>
	);
}

// ─── AddIngredientForm ────────────────────────────────────────────────────────

// mode: "select"  — combobox: pick existing or trigger FDC creation
//       "fdc"     — FDC search panel; ingredient not yet created
//       "confirm" — FDC entry chosen; fill amount/unit/prep and save
function AddIngredientForm({ prepId, onAdded, onCancel }) {
	const ingredients = useSignal([]);
	const selectedId = useSignal("");
	const fdcQuery = useSignal("");
	const selectedFDC = useSignal(null);
	const name = useSignal("");
	const amount = useSignal("1");
	const unit = useSignal("");
	const prepNote = useSignal("");
	const saving = useSignal(false);
	const error = useSignal("");
	const mode = useSignal("select");

	useEffect(() => {
		apiFetch("/api/ingredients")
			.then((r) => r.json())
			.then((data) => {
				ingredients.value = data;
			});
	}, []);

	const handleIngredientChange = (val) => {
		if (val.startsWith("existing:")) {
			const id = val.slice("existing:".length);
			selectedId.value = id;
			const found = ingredients.value.find((i) => String(i.id) === id);
			if (found) name.value = found.name;
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
		if (mode.value === "select" && !selectedId.value) {
			error.value = "Select an ingredient first";
			return;
		}
		saving.value = true;
		error.value = "";
		try {
			let ingredientId = Number(selectedId.value);

			if (mode.value === "confirm") {
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
				ingredientId = created.id;
			}

			const r = await apiFetch(`/api/preparations/${prepId}/ingredients`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					ingredient_id: ingredientId,
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

	const options = ingredients.value.map((i) => ({
		value: `existing:${i.id}`,
		label: i.name,
	}));

	const showPrepFields =
		mode.value === "confirm" || (mode.value === "select" && !!selectedId.value);

	return (
		<form
			onSubmit={handleSave}
			class="flex flex-col gap-2 py-2 px-3 rounded-lg"
			style={{ background: "var(--color-bg)" }}
		>
			{mode.value === "select" && (
				<Combobox
					label="Ingredient"
					value={selectedId.value ? `existing:${selectedId.value}` : ""}
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
								value={amount.value}
								onInput={(e) => {
									amount.value = e.target.value;
								}}
							/>
						</div>
						<div class="flex-1">
							<TextField
								id="ing-unit"
								label="Unit (optional)"
								value={unit.value}
								onInput={(e) => {
									unit.value = e.target.value;
								}}
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

function StepRow({ index, description, onChange, onDelete }) {
	const inputRef = useRef(null);
	const local = useSignal(description);

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
	const showAddIngredient = useSignal(false);
	const removeDialog = useDialog();
	const removeError = useSignal("");
	const savingSteps = useSignal(false);

	// Editable link amount/unit
	const editingAmount = useSignal(false);
	const amountValue = useSignal(String(initialLink.amount));
	const unitValue = useSignal(initialLink.unit);

	const saveAmount = async () => {
		editingAmount.value = false;
		const amt = parseFloat(amountValue.value);
		const unit = unitValue.value.trim();
		if (!amt || amt <= 0 || !unit) {
			amountValue.value = String(link.value.amount);
			unitValue.value = link.value.unit;
			return;
		}
		try {
			const r = await apiFetch(
				`/api/recipes/${link.value.recipe_id}/preparations/${link.value.id}`,
				{
					method: "PUT",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						position: link.value.position,
						amount: amt,
						unit,
					}),
				},
			);
			if (!r.ok) throw new Error("Failed to update");
			const updated = await r.json();
			link.value = updated;
		} catch {
			amountValue.value = String(link.value.amount);
			unitValue.value = link.value.unit;
		}
	};

	// Editable name
	const editingName = useSignal(false);
	const nameValue = useSignal(initialPrep.name);
	const nameRef = useRef(null);

	useEffect(() => {
		if (editingName.value) nameRef.current?.focus();
	}, [editingName.value]);

	const saveName = async () => {
		const trimmed = nameValue.value.trim();
		if (!trimmed || trimmed === prep.value.name) {
			editingName.value = false;
			nameValue.value = prep.value.name;
			return;
		}
		try {
			const r = await apiFetch(`/api/preparations/${prep.value.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: trimmed,
					description: prep.value.description,
					yield_amount: prep.value.yield_amount,
					yield_unit: prep.value.yield_unit,
					steps: prep.value.steps,
					is_public: prep.value.is_public,
				}),
			});
			if (!r.ok) throw new Error("Failed to update");
			const updated = await r.json();
			prep.value = updated;
			onUpdated(updated);
		} catch {
			nameValue.value = prep.value.name;
		}
		editingName.value = false;
	};

	const saveDescription = async (newDesc) => {
		const r = await apiFetch(`/api/preparations/${prep.value.id}`, {
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				name: prep.value.name,
				description: newDesc || undefined,
				yield_amount: prep.value.yield_amount,
				yield_unit: prep.value.yield_unit,
				steps: prep.value.steps,
				is_public: prep.value.is_public,
			}),
		});
		if (!r.ok) throw new Error("Failed to save description");
		const updated = await r.json();
		prep.value = updated;
		onUpdated(updated);
	};

	const saveSteps = async (newSteps) => {
		savingSteps.value = true;
		try {
			const r = await apiFetch(`/api/preparations/${prep.value.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: prep.value.name,
					description: prep.value.description,
					yield_amount: prep.value.yield_amount,
					yield_unit: prep.value.yield_unit,
					// eslint-disable-next-line no-unused-vars
					steps: newSteps.map(({ _key, ...s }) => s),
					is_public: prep.value.is_public,
				}),
			});
			if (!r.ok) throw new Error("Failed to save steps");
			const updated = await r.json();
			prep.value = {
				...updated,
				steps: updated.steps.map((s, i) => ({
					...s,
					_key: newSteps[i]?._key ?? crypto.randomUUID(),
				})),
			};
			onUpdated(updated);
		} finally {
			savingSteps.value = false;
		}
	};

	const handleStepChange = (index, newDesc) => {
		const newSteps = prep.value.steps.map((s, i) =>
			i === index ? { ...s, description: newDesc } : s,
		);
		saveSteps(newSteps);
	};

	const handleStepDelete = (index) => {
		const newSteps = prep.value.steps.filter((_, i) => i !== index);
		saveSteps(newSteps);
	};

	const handleAddStep = () => {
		const newSteps = [
			...prep.value.steps,
			{ description: "", _key: crypto.randomUUID() },
		];
		saveSteps(newSteps);
	};

	const handleIngredientAdded = (ing) => {
		prep.value = {
			...prep.value,
			ingredients: [...prep.value.ingredients, ing],
		};
		showAddIngredient.value = false;
		onUpdated(prep.value);
	};

	const handleIngredientUpdated = (updated) => {
		prep.value = {
			...prep.value,
			ingredients: prep.value.ingredients.map((i) =>
				i.id === updated.id ? updated : i,
			),
		};
		onUpdated(prep.value);
	};

	const handleIngredientDeleted = (id) => {
		prep.value = {
			...prep.value,
			ingredients: prep.value.ingredients.filter((i) => i.id !== id),
		};
		onUpdated(prep.value);
	};

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

	return (
		<>
			<AccordionItem value={String(initialLink.id)}>
				<AccordionTrigger>
					<div class="flex flex-1 items-center gap-2 min-w-0">
						{editingName.value ? (
							<TextField
								inline
								inputRef={nameRef}
								containerClass="flex-1 min-w-0"
								class="text-sm font-semibold"
								value={nameValue.value}
								onInput={(e) => {
									nameValue.value = e.target.value;
								}}
								onBlur={saveName}
								onKeyDown={(e) => {
									if (e.key === "Enter") saveName();
									if (e.key === "Escape") {
										nameValue.value = prep.value.name;
										editingName.value = false;
									}
								}}
								onClick={(e) => e.stopPropagation()}
							/>
						) : (
							<Button
								variant="unstyled"
								type="button"
								class="text-sm font-semibold truncate cursor-text"
								style={{ color: "var(--color-text)" }}
								onClick={(e) => {
									e.stopPropagation();
									editingName.value = true;
								}}
								onKeyDown={(e) => {
									if (e.key === "Enter" || e.key === " ") {
										e.stopPropagation();
										editingName.value = true;
									}
								}}
							>
								{prep.value.name}
							</Button>
						)}
						{prep.value.yield_amount > 0 && (
							<span
								class="text-xs shrink-0"
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
					<div class="flex flex-col gap-5 px-4 py-3">
						{/* Description */}
						<EditableMarkdown
							value={prep.value.description ?? ""}
							placeholder="Add a description…"
							onSave={saveDescription}
						/>

						{/* Amount in recipe */}
						<div>
							<span
								class="text-xs font-semibold uppercase tracking-wider"
								style={{ color: "var(--color-muted)" }}
							>
								Amount in recipe
							</span>
							{editingAmount.value ? (
								<div class="flex items-center gap-2 mt-1">
									<TextField
										type="number"
										value={amountValue.value}
										onInput={(e) => {
											amountValue.value = e.target.value;
										}}
										onKeyDown={(e) => {
											if (e.key === "Enter") saveAmount();
											if (e.key === "Escape") {
												amountValue.value = String(link.value.amount);
												unitValue.value = link.value.unit;
												editingAmount.value = false;
											}
										}}
										autoFocus
									/>
									<TextField
										value={unitValue.value}
										placeholder="g, ml, serving…"
										onInput={(e) => {
											unitValue.value = e.target.value;
										}}
										onKeyDown={(e) => {
											if (e.key === "Enter") saveAmount();
											if (e.key === "Escape") {
												amountValue.value = String(link.value.amount);
												unitValue.value = link.value.unit;
												editingAmount.value = false;
											}
										}}
									/>
									<Button size="sm" type="button" onClick={saveAmount}>
										Save
									</Button>
									<Button
										variant="ghost"
										size="sm"
										type="button"
										onClick={() => {
											amountValue.value = String(link.value.amount);
											unitValue.value = link.value.unit;
											editingAmount.value = false;
										}}
									>
										Cancel
									</Button>
								</div>
							) : (
								<Button
									variant="unstyled"
									type="button"
									class="mt-1 text-sm cursor-text"
									style={{ color: "var(--color-muted)" }}
									onClick={() => {
										editingAmount.value = true;
									}}
									onKeyDown={(e) => {
										if (e.key === "Enter" || e.key === " ")
											editingAmount.value = true;
									}}
								>
									{link.value.amount} {link.value.unit}
								</Button>
							)}
						</div>

						{/* Ingredients */}
						<div>
							<div class="flex items-center justify-between mb-2">
								<span
									class="text-xs font-semibold uppercase tracking-wider"
									style={{ color: "var(--color-muted)" }}
								>
									Ingredients
								</span>
								<Button
									variant="ghost"
									size="sm"
									type="button"
									aria-label="Add ingredient"
									onClick={() => {
										showAddIngredient.value = true;
									}}
								>
									<Plus size={14} aria-hidden="true" />
								</Button>
							</div>

							{prep.value.ingredients.length === 0 &&
								!showAddIngredient.value && (
									<p class="text-sm" style={{ color: "var(--color-muted)" }}>
										No ingredients yet.
									</p>
								)}

							<div class="flex flex-col">
								{prep.value.ingredients.map((ing) => (
									<IngredientRow
										key={ing.id}
										ingredient={ing}
										prepId={prep.value.id}
										onUpdated={handleIngredientUpdated}
										onDeleted={handleIngredientDeleted}
									/>
								))}
							</div>

							{showAddIngredient.value && (
								<div class="mt-2">
									<AddIngredientForm
										prepId={prep.value.id}
										onAdded={handleIngredientAdded}
										onCancel={() => {
											showAddIngredient.value = false;
										}}
									/>
								</div>
							)}
						</div>

						{/* Steps */}
						<div>
							<div class="flex items-center justify-between mb-2">
								<span
									class="text-xs font-semibold uppercase tracking-wider"
									style={{ color: "var(--color-muted)" }}
								>
									Steps
								</span>
								<Button
									variant="ghost"
									size="sm"
									type="button"
									onClick={handleAddStep}
									disabled={savingSteps.value}
								>
									<Plus size={14} aria-hidden="true" />
								</Button>
							</div>

							{prep.value.steps.length === 0 && (
								<p class="text-sm" style={{ color: "var(--color-muted)" }}>
									No steps yet.
								</p>
							)}

							<div class="flex flex-col">
								{prep.value.steps.map((step, i) => (
									<StepRow
										key={step._key}
										index={i}
										description={step.description}
										onChange={handleStepChange}
										onDelete={handleStepDelete}
									/>
								))}
							</div>
						</div>
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
						yield_unit: "serving",
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
					unit: "serving",
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
		<div class="flex flex-col gap-6 p-4">
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

			{/* Components */}
			<div class="flex flex-col gap-3">
				<div class="flex items-center justify-between">
					<span
						class="text-xs font-semibold uppercase tracking-wider"
						style={{ color: "var(--color-muted)" }}
					>
						Components
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
						No components yet. Add a preparation to get started.
					</p>
				)}

				{components.value.length > 0 && (
					<Accordion type="multiple">
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
