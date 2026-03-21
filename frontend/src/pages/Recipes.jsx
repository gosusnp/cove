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
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";
import { RecipeDetail } from "./RecipeDetail.jsx";

// ─── RecipeList ──────────────────────────────────────────────────────────────

function RecipeList({ recipes, selectedId, onSelect, onNew, loading, error }) {
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
					Recipes
				</h2>
				<Button variant="primary" size="sm" onClick={onNew}>
					+ New
				</Button>
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
			{!loading && !error && recipes.length === 0 && (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No recipes yet.
				</p>
			)}
			{!loading &&
				!error &&
				recipes.map((r, i) => (
					<ListItem
						key={r.id}
						label={r.name}
						sublabel={`${r.servings} serving${r.servings !== 1 ? "s" : ""}`}
						active={r.id === selectedId}
						isLast={i === recipes.length - 1}
						onClick={() => onSelect(r.id)}
					/>
				))}
		</div>
	);
}

// ─── Recipes (page) ───────────────────────────────────────────────────────────

export function Recipes() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const recipes = useSignal([]);
	const loading = useSignal(true);
	const fetchError = useSignal("");

	const recipeDialog = useDialog();
	const nameField = useSignal("");
	const servingsField = useSignal("1");
	const saving = useSignal(false);
	const formError = useSignal("");

	const fetchRecipes = async () => {
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await apiFetch("/api/recipes");
			if (!r.ok) throw new Error("Failed to fetch recipes");
			recipes.value = await r.json();
		} catch (err) {
			fetchError.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchRecipes();
	}, [!!user]);

	const handleSelect = (id) => {
		route(`/cook/recipes/${id}`);
	};

	const openNew = () => {
		nameField.value = "";
		servingsField.value = "1";
		formError.value = "";
		recipeDialog.show();
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (!nameField.value.trim()) {
			formError.value = "Name is required.";
			return;
		}
		const servings = parseInt(servingsField.value, 10);
		if (!servings || servings <= 0) {
			formError.value = "Servings must be a positive number.";
			return;
		}
		saving.value = true;
		formError.value = "";
		try {
			const r = await apiFetch("/api/recipes", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: nameField.value.trim(),
					servings,
				}),
			});
			if (!r.ok) {
				const data = await r.json();
				throw new Error(data.error || "Failed to create recipe");
			}
			const created = await r.json();
			await fetchRecipes();
			recipeDialog.hide();
			route(`/cook/recipes/${created.id}`);
		} catch (err) {
			formError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const handleRecipeDeleted = () => {
		if (selectedId) {
			route("/cook/recipes");
		}
		fetchRecipes();
	};

	return (
		<>
			<ListDetail
				hasDetail={!!selectedId}
				emptyState="Select a recipe to view its details."
				list={
					<RecipeList
						recipes={recipes.value}
						selectedId={selectedId}
						onSelect={handleSelect}
						onNew={openNew}
						loading={loading.value}
						error={fetchError.value}
					/>
				}
				detail={
					selectedId ? (
						<RecipeDetail
							recipeId={selectedId}
							onRecipeUpdated={fetchRecipes}
							onRecipeDeleted={handleRecipeDeleted}
						/>
					) : null
				}
			/>

			{/* New Recipe dialog */}
			<Dialog openSignal={recipeDialog.open}>
				<DialogContent>
					<form onSubmit={handleSave}>
						<DialogTitle>New Recipe</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<TextField
								id="recipe-name"
								label="Name"
								value={nameField.value}
								onInput={(e) => {
									nameField.value = e.target.value;
								}}
								autoFocus
							/>
							<TextField
								id="recipe-servings"
								label="Servings"
								type="number"
								value={servingsField.value}
								onInput={(e) => {
									servingsField.value = e.target.value;
								}}
							/>
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
