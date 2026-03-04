// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Section, Row } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { Switch } from "../components/ui/Switch.jsx";
import { useDialog } from "../hooks/useDialog.js";

export function Exercises() {
	const { token } = useAuth();
	const exercises = useSignal([]);
	const loading = useSignal(true);
	const error = useSignal("");

	// Form state
	const dialog = useDialog();
	const editingId = useSignal(null);
	const name = useSignal("");
	const progression = useSignal("");
	const description = useSignal("");
	const isPublic = useSignal(false);
	const saving = useSignal(false);
	const formError = useSignal("");

	const fetchExercises = async () => {
		if (!token) return;
		loading.value = true;
		try {
			const r = await fetch("/api/exercises", {
				headers: { Authorization: `Bearer ${token}` },
			});
			if (!r.ok) throw new Error("Failed to fetch exercises");
			const data = await r.json();
			exercises.value = data;
		} catch (err) {
			error.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchExercises();
	}, [token]);

	const openCreate = () => {
		editingId.value = null;
		name.value = "";
		progression.value = "";
		description.value = "";
		isPublic.value = false;
		formError.value = "";
		dialog.show();
	};

	const openEdit = (ex) => {
		editingId.value = ex.id;
		name.value = ex.name;
		progression.value = ex.progression || "";
		description.value = ex.description || "";
		isPublic.value = ex.is_public;
		formError.value = "";
		dialog.show();
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (!name.value.trim()) {
			formError.value = "Name is required.";
			return;
		}

		saving.value = true;
		formError.value = "";

		const payload = {
			name: name.value.trim(),
			progression: progression.value.trim() || null,
			description: description.value.trim() || null,
			is_public: isPublic.value,
		};

		const method = editingId.value ? "PUT" : "POST";
		const url = editingId.value
			? `/api/exercises/${editingId.value}`
			: "/api/exercises";

		try {
			const r = await fetch(url, {
				method,
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
			});

			if (!r.ok) {
				const data = await r.json();
				throw new Error(data.error || "Failed to save exercise");
			}

			const saved = await r.json();
			if (editingId.value) {
				exercises.value = exercises.value.map((ex) =>
					ex.id === editingId.value ? saved : ex,
				);
			} else {
				exercises.value = [...exercises.value, saved];
			}
			dialog.hide();
		} catch (err) {
			formError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const handleDelete = async (id) => {
		if (!confirm("Are you sure you want to delete this exercise?")) return;
		try {
			const r = await fetch(`/api/exercises/${id}`, {
				method: "DELETE",
				headers: { Authorization: `Bearer ${token}` },
			});
			if (!r.ok) throw new Error("Failed to delete exercise");
			exercises.value = exercises.value.filter((ex) => ex.id !== id);
		} catch (err) {
			alert(err.message);
		}
	};

	return (
		<main class="flex flex-1 flex-col gap-8 px-4 py-6 max-w-lg mx-auto w-full">
			<PageTitle>Exercises</PageTitle>

			<Section
				title="All Exercises"
				action={
					<Button
						variant="outline"
						size="sm"
						onClick={openCreate}
						disabled={loading.value}
					>
						Create
					</Button>
				}
			>
				{loading.value && exercises.value.length === 0 ? (
					<Row label="Loading exercises..." last />
				) : error.value ? (
					<Row label={error.value} last />
				) : exercises.value.length === 0 ? (
					<Row label="No exercises yet" last />
				) : (
					exercises.value.map((ex, i) => (
						<Row
							key={ex.id}
							label={ex.name}
							sublabel={ex.progression || "No progression info"}
							last={i === exercises.value.length - 1}
						>
							<div class="flex gap-2">
								<Button
									variant="outline"
									size="sm"
									onClick={() => openEdit(ex)}
								>
									Update
								</Button>
								<Button
									variant="destructive"
									size="sm"
									onClick={() => handleDelete(ex.id)}
								>
									Delete
								</Button>
							</div>
						</Row>
					))
				)}
			</Section>

			<Dialog openSignal={dialog.open}>
				<DialogContent>
					<form onSubmit={handleSave}>
						<DialogTitle>
							{editingId.value ? "Update Exercise" : "New Exercise"}
						</DialogTitle>
						<DialogDescription>
							Define the exercise details below.
						</DialogDescription>

						<div class="mt-4 flex flex-col gap-4">
							<TextField
								id="ex-name"
								label="Name"
								placeholder="e.g. Diamond Push-up"
								value={name.value}
								onInput={(e) => {
									name.value = e.target.value;
								}}
								autoFocus
							/>

							<TextField
								id="ex-progression"
								label="Progression"
								placeholder="e.g. Push-up"
								value={progression.value}
								onInput={(e) => {
									progression.value = e.target.value;
								}}
							/>

							<TextField
								id="ex-description"
								label="Description"
								placeholder="Optional description"
								value={description.value}
								onInput={(e) => {
									description.value = e.target.value;
								}}
							/>

							<div class="flex items-center justify-between">
								<label
									for="ex-public"
									class="text-sm font-medium"
									style={{ color: "var(--color-text)" }}
								>
									Make public
								</label>
								<Switch id="ex-public" checkedSignal={isPublic} />
							</div>

							{formError.value && (
								<p class="text-sm" style={{ color: "var(--color-error)" }}>
									{formError.value}
								</p>
							)}
						</div>

						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="ghost" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={saving.value}>
								{saving.value ? "Saving..." : "Save Exercise"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>
		</main>
	);
}
