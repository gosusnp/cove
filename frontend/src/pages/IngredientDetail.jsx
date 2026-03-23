// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { Pencil, RefreshCw, Trash2 } from "lucide-preact";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { FDCSearch } from "../components/shared/FDCSearch.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";

// ─── IngredientDetailInner ────────────────────────────────────────────────────

function MacroRow({ label, value, unit, last }) {
	return (
		<Row label={label} last={last}>
			<span class="text-sm" style={{ color: "var(--color-text)" }}>
				{value != null ? `${value} ${unit}` : "—"}
			</span>
		</Row>
	);
}

function IngredientDetailInner({
	ingredient: initialIngredient,
	onUpdated,
	onDeleted,
}) {
	const ingredient = useSignal(initialIngredient);

	// ── Name editing ─────────────────────────────────────────────────────────
	const editingName = useSignal(false);
	const editName = useSignal(initialIngredient.name);
	const nameSaving = useSignal(false);
	const nameError = useSignal("");
	const nameInputRef = useRef(null);

	useEffect(() => {
		if (editingName.value) nameInputRef.current?.focus();
	}, [editingName.value]);

	const saveName = async () => {
		const trimmed = editName.value.trim();
		if (!trimmed) {
			nameError.value = "Name is required.";
			return;
		}
		nameSaving.value = true;
		nameError.value = "";
		try {
			const ing = ingredient.value;
			const r = await apiFetch(`/api/ingredients/${ing.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: trimmed,
					fdc_id: ing.fdc_id ?? undefined,
					calories_per_100g: ing.calories_per_100g,
					protein_per_100g: ing.protein_per_100g,
					fat_per_100g: ing.fat_per_100g,
					carbs_per_100g: ing.carbs_per_100g,
					density_g_per_ml: ing.density_g_per_ml ?? undefined,
					is_public: ing.is_public,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to save");
			}
			ingredient.value = await r.json();
			editingName.value = false;
			onUpdated?.(ingredient.value);
		} catch (err) {
			nameError.value = err.message;
		} finally {
			nameSaving.value = false;
		}
	};

	// ── FDC ──────────────────────────────────────────────────────────────────
	const changingFDC = useSignal(false);
	const syncing = useSignal(false);
	const fdcError = useSignal("");

	const handleFDCSelect = async (food) => {
		syncing.value = true;
		fdcError.value = "";
		try {
			const ing = ingredient.value;
			const r = await apiFetch(`/api/ingredients/${ing.id}`, {
				method: "PUT",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: ing.name,
					fdc_id: food.fdc_id,
					calories_per_100g: food.calories_per_100g,
					protein_per_100g: food.protein_per_100g,
					fat_per_100g: food.fat_per_100g,
					carbs_per_100g: food.carbs_per_100g,
					density_g_per_ml: food.density_g_per_ml ?? undefined,
					is_public: ing.is_public,
				}),
			});
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Failed to update");
			}
			ingredient.value = await r.json();
			changingFDC.value = false;
			onUpdated?.(ingredient.value);
		} catch (err) {
			fdcError.value = err.message;
		} finally {
			syncing.value = false;
		}
	};

	const handleSync = async () => {
		syncing.value = true;
		fdcError.value = "";
		try {
			const r = await apiFetch(
				`/api/ingredients/${ingredient.value.id}/fdc-sync`,
				{ method: "POST" },
			);
			if (!r.ok) {
				const d = await r.json();
				throw new Error(d.error || "Sync failed");
			}
			ingredient.value = await r.json();
			onUpdated?.(ingredient.value);
		} catch (err) {
			fdcError.value = err.message;
		} finally {
			syncing.value = false;
		}
	};

	// ── Delete ────────────────────────────────────────────────────────────────
	const deleteDialog = useDialog();
	const deleteError = useSignal("");

	const handleDelete = async () => {
		deleteError.value = "";
		const r = await apiFetch(`/api/ingredients/${ingredient.value.id}`, {
			method: "DELETE",
		});
		if (!r.ok) {
			deleteError.value = "Failed to delete ingredient.";
			throw new Error("Failed to delete ingredient");
		}
		onDeleted?.();
	};

	const ing = ingredient.value;

	return (
		<>
			<div class="p-6 max-w-3xl mx-auto w-full flex flex-col gap-6">
				{/* Name */}
				{editingName.value ? (
					<div class="flex flex-col gap-1">
						<div class="flex items-center gap-3">
							<TextField
								inline
								inputRef={nameInputRef}
								containerClass="flex-1 min-w-0"
								class="text-2xl font-semibold"
								value={editName.value}
								onInput={(e) => {
									editName.value = e.target.value;
								}}
								onKeyDown={(e) => {
									if (e.key === "Escape") {
										editingName.value = false;
										nameError.value = "";
									}
									if (e.key === "Enter") saveName();
								}}
							/>
							<div class="flex gap-2 shrink-0">
								<Button
									size="sm"
									onClick={saveName}
									disabled={nameSaving.value}
								>
									{nameSaving.value ? "Saving…" : "Save"}
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={() => {
										editingName.value = false;
										nameError.value = "";
									}}
									disabled={nameSaving.value}
								>
									Cancel
								</Button>
							</div>
						</div>
						{nameError.value && (
							<p class="text-xs" style={{ color: "var(--color-error)" }}>
								{nameError.value}
							</p>
						)}
					</div>
				) : (
					<div class="group flex items-center justify-between w-full gap-2">
						<h1
							class="text-2xl font-semibold"
							style={{ color: "var(--color-text)" }}
						>
							{ing.name}
						</h1>
						<div class="flex items-center gap-1 shrink-0">
							<Button
								variant="ghost"
								size="icon"
								class="opacity-0 group-hover:opacity-30 transition-opacity"
								onClick={() => {
									editName.value = ing.name;
									editingName.value = true;
								}}
								aria-label="Edit ingredient name"
							>
								<Pencil
									size={14}
									style={{ color: "var(--color-muted)" }}
									aria-hidden="true"
								/>
							</Button>
							<Button
								variant="ghost"
								size="icon"
								aria-label="Delete ingredient"
								onClick={deleteDialog.show}
							>
								<Trash2 size={14} aria-hidden="true" />
							</Button>
						</div>
					</div>
				)}

				{/* FDC */}
				<Section title="FDC">
					{changingFDC.value ? (
						<div class="p-3">
							<FDCSearch
								onSelect={handleFDCSelect}
								onCancel={() => {
									changingFDC.value = false;
									fdcError.value = "";
								}}
							/>
						</div>
					) : (
						<Row label="FDC ID" last>
							<div class="flex items-center gap-2">
								<span class="text-sm" style={{ color: "var(--color-text)" }}>
									{ing.fdc_id ?? "—"}
								</span>
								<Button
									variant="outline"
									size="sm"
									onClick={() => {
										changingFDC.value = true;
										fdcError.value = "";
									}}
									disabled={syncing.value}
								>
									Change
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={handleSync}
									disabled={syncing.value || ing.fdc_id == null}
									aria-label="Sync nutrition from FDC"
								>
									<RefreshCw size={13} aria-hidden="true" />
									{syncing.value ? "Syncing…" : "Sync"}
								</Button>
							</div>
						</Row>
					)}
					{fdcError.value && (
						<p
							class="px-3 pb-2 text-xs"
							style={{ color: "var(--color-error)" }}
						>
							{fdcError.value}
						</p>
					)}
				</Section>

				{/* Nutrition */}
				<Section title="Nutrition per 100 g">
					<MacroRow
						label="Calories"
						value={ing.calories_per_100g}
						unit="kcal"
					/>
					<MacroRow label="Protein" value={ing.protein_per_100g} unit="g" />
					<MacroRow label="Fat" value={ing.fat_per_100g} unit="g" />
					<MacroRow
						label="Carbs"
						value={ing.carbs_per_100g}
						unit="g"
						last={ing.density_g_per_ml == null}
					/>
					{ing.density_g_per_ml != null && (
						<MacroRow
							label="Density"
							value={ing.density_g_per_ml}
							unit="g/ml"
							last
						/>
					)}
				</Section>

				{deleteError.value && (
					<p class="text-sm" style={{ color: "var(--color-error)" }}>
						{deleteError.value}
					</p>
				)}
			</div>

			<ConfirmDialog
				openSignal={deleteDialog.open}
				title="Delete Ingredient"
				description="This will permanently delete the ingredient."
				confirmLabel="Delete"
				onConfirm={handleDelete}
			/>
		</>
	);
}

// ─── IngredientDetail (shell) ─────────────────────────────────────────────────

export function IngredientDetail({ ingredientId, onUpdated, onDeleted }) {
	const ingredient = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");

	useEffect(() => {
		if (!ingredientId) return;
		loading.value = true;
		error.value = "";
		apiFetch(`/api/ingredients/${ingredientId}`)
			.then((r) => {
				if (!r.ok) throw new Error("Failed to load ingredient");
				return r.json();
			})
			.then((data) => {
				ingredient.value = data;
			})
			.catch((err) => {
				error.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [ingredientId]);

	if (loading.value) {
		return (
			<div class="flex flex-1 items-center justify-center p-8">
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Loading…
				</p>
			</div>
		);
	}
	if (error.value) {
		return (
			<div class="flex flex-1 items-center justify-center p-8">
				<p class="text-sm" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			</div>
		);
	}
	if (!ingredient.value) return null;

	return (
		<IngredientDetailInner
			key={ingredientId}
			ingredient={ingredient.value}
			onUpdated={onUpdated}
			onDeleted={onDeleted}
		/>
	);
}
