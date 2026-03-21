// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Row, Section } from "../components/ui/Section.jsx";
import { apiFetch } from "../lib/api.js";

function RecipeDetailInner({ recipe }) {
	return (
		<div class="flex flex-col gap-6 p-4">
			<div>
				<h2
					class="text-xl font-semibold"
					style={{ color: "var(--color-text)" }}
				>
					{recipe.name}
				</h2>
				{recipe.description && (
					<p class="mt-1 text-sm" style={{ color: "var(--color-muted)" }}>
						{recipe.description}
					</p>
				)}
			</div>

			<Section title="Details">
				<Row label="Servings">
					<span class="text-sm" style={{ color: "var(--color-text)" }}>
						{recipe.servings}
					</span>
				</Row>
				{recipe.yield_amount != null && (
					<Row label="Yield">
						<span class="text-sm" style={{ color: "var(--color-text)" }}>
							{recipe.yield_amount}
							{recipe.yield_unit ? ` ${recipe.yield_unit}` : ""}
						</span>
					</Row>
				)}
				<Row label="Visibility" last>
					<span class="text-sm" style={{ color: "var(--color-text)" }}>
						{recipe.is_public ? "Public" : "Private"}
					</span>
				</Row>
			</Section>
		</div>
	);
}

export function RecipeDetail({ recipeId }) {
	const recipe = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");

	useEffect(() => {
		if (!recipeId) return;
		loading.value = true;
		error.value = "";
		apiFetch(`/api/recipes/${recipeId}`)
			.then((r) => {
				if (!r.ok) throw new Error("Failed to load recipe");
				return r.json();
			})
			.then((data) => {
				recipe.value = data;
			})
			.catch((err) => {
				error.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [recipeId]);

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

	return <RecipeDetailInner recipe={recipe.value} />;
}
