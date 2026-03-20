// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Row, Section } from "../components/ui/Section.jsx";
import { apiFetch } from "../lib/api.js";

function MacroRow({ label, value, unit, last }) {
	return (
		<Row label={label} last={last}>
			<span class="text-sm" style={{ color: "var(--color-text)" }}>
				{value != null ? `${value} ${unit}` : "—"}
			</span>
		</Row>
	);
}

function IngredientDetailInner({ ingredient }) {
	return (
		<div class="flex flex-col gap-6 p-4">
			<div>
				<h2
					class="text-xl font-semibold"
					style={{ color: "var(--color-text)" }}
				>
					{ingredient.name}
				</h2>
				{ingredient.fdc_id != null && (
					<p class="mt-1 text-xs" style={{ color: "var(--color-muted)" }}>
						FDC ID: {ingredient.fdc_id}
					</p>
				)}
			</div>

			<Section title="Nutrition per 100 g">
				<MacroRow
					label="Calories"
					value={ingredient.calories_per_100g}
					unit="kcal"
				/>
				<MacroRow
					label="Protein"
					value={ingredient.protein_per_100g}
					unit="g"
				/>
				<MacroRow label="Fat" value={ingredient.fat_per_100g} unit="g" />
				<MacroRow
					label="Carbs"
					value={ingredient.carbs_per_100g}
					unit="g"
					last={ingredient.density_g_per_ml == null}
				/>
				{ingredient.density_g_per_ml != null && (
					<MacroRow
						label="Density"
						value={ingredient.density_g_per_ml}
						unit="g/ml"
						last
					/>
				)}
			</Section>
		</div>
	);
}

export function IngredientDetail({ ingredientId }) {
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
	if (!ingredient.value) return null;

	return <IngredientDetailInner ingredient={ingredient.value} />;
}
