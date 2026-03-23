// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useLocation, useRoute } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { ListItem } from "../components/ui/ListItem.jsx";
import { apiFetch } from "../lib/api.js";
import { IngredientDetail } from "./IngredientDetail.jsx";

function IngredientList({ ingredients, selectedId, onSelect, loading, error }) {
	return (
		<div class="flex flex-col">
			<div
				class="flex items-center justify-between px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Ingredients
				</h2>
			</div>

			{loading && (
				<p class="px-4 py-3 text-sm" style={{ color: "var(--color-muted)" }}>
					Loading…
				</p>
			)}
			{!loading && error && (
				<p class="px-4 py-3 text-sm" style={{ color: "var(--color-error)" }}>
					{error}
				</p>
			)}
			{!loading && !error && ingredients.length === 0 && (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No ingredients yet.
				</p>
			)}
			{!loading &&
				!error &&
				ingredients.map((ing, i) => (
					<ListItem
						key={ing.id}
						label={ing.name}
						sublabel={`${ing.calories_per_100g} kcal / 100 g`}
						active={ing.id === selectedId}
						isLast={i === ingredients.length - 1}
						onClick={() => onSelect(ing.id)}
					/>
				))}
		</div>
	);
}

export function Ingredients() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const ingredients = useSignal([]);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	useEffect(() => {
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		apiFetch("/api/ingredients")
			.then((r) => {
				if (!r.ok) throw new Error("Failed to fetch ingredients");
				return r.json();
			})
			.then((data) => {
				ingredients.value = data;
			})
			.catch((err) => {
				fetchError.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [!!user]);

	const handleSelect = (id) => {
		route(`/cook/ingredients/${id}`);
	};

	const handleDeleted = () => {
		ingredients.value = ingredients.value.filter((i) => i.id !== selectedId);
		route("/cook/ingredients");
	};

	const handleUpdated = (updated) => {
		ingredients.value = ingredients.value.map((i) =>
			i.id === updated.id ? { ...i, name: updated.name } : i,
		);
	};

	return (
		<ListDetail
			hasDetail={!!selectedId}
			emptyState="Select an ingredient to view its details."
			list={
				<IngredientList
					ingredients={ingredients.value}
					selectedId={selectedId}
					onSelect={handleSelect}
					loading={loading.value}
					error={fetchError.value}
				/>
			}
			detail={
				selectedId ? (
					<IngredientDetail
						ingredientId={selectedId}
						onUpdated={handleUpdated}
						onDeleted={handleDeleted}
					/>
				) : null
			}
		/>
	);
}
