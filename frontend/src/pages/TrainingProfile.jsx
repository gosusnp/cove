// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Button } from "../components/ui/Button.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ToggleGroup } from "../components/ui/ToggleGroup.jsx";
import { apiFetch } from "../lib/api.js";

const LEVEL_OPTIONS = [
	{ value: "beginner", label: "Beginner" },
	{ value: "intermediate", label: "Intermediate" },
	{ value: "advanced", label: "Advanced" },
	{ value: "expert", label: "Expert" },
];

function toDisciplinePayload(d) {
	const years = parseFloat(d.years_practice);
	return {
		name: d.name?.trim() || null,
		years_practice: Number.isNaN(years) ? null : years,
		level: d.level || null,
		notes: d.notes || null,
	};
}

function withKey(d) {
	return { ...d, _key: crypto.randomUUID() };
}

export function TrainingProfile() {
	const loading = useSignal(true);
	const motivation = useSignal(null);
	const constraints = useSignal(null);
	const disciplines = useSignal([]);
	const addError = useSignal(null);

	useEffect(() => {
		apiFetch("/api/users/me/training-profile")
			.then((r) => {
				if (r.status === 404) return null;
				return r.ok ? r.json() : null;
			})
			.then((data) => {
				if (data) {
					motivation.value = data.motivation ?? null;
					constraints.value = data.constraints ?? null;
					disciplines.value = (data.disciplines ?? []).map(withKey);
				}
			})
			.catch(() => {})
			.finally(() => {
				loading.value = false;
			});
	}, []);

	async function patchProfile(body) {
		const r = await apiFetch("/api/users/me/training-profile", {
			method: "PATCH",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body),
		});
		if (!r.ok) throw new Error("Failed to save.");
	}

	async function saveDisciplines(updated) {
		const prev = disciplines.value;
		disciplines.value = updated;
		try {
			await patchProfile({ disciplines: updated.map(toDisciplinePayload) });
		} catch {
			disciplines.value = prev;
			throw new Error("Failed to save.");
		}
	}

	function updateDisciplineField(i, field, value) {
		disciplines.value = disciplines.value.map((d, idx) =>
			idx === i ? { ...d, [field]: value } : d,
		);
	}

	async function addDiscipline() {
		addError.value = null;
		const updated = [
			...disciplines.value,
			withKey({ name: "", years_practice: null, level: null, notes: null }),
		];
		try {
			await saveDisciplines(updated);
		} catch {
			addError.value = "Failed to add discipline.";
		}
	}

	function removeDiscipline(i) {
		saveDisciplines(disciplines.value.filter((_, idx) => idx !== i)).catch(
			() => {},
		);
	}

	if (loading.value) return null;

	return (
		<main class="flex flex-1 flex-col gap-8 px-4 py-6 max-w-lg mx-auto w-full">
			<PageTitle>Training Profile</PageTitle>

			<Section title="Motivation">
				<div class="px-4 py-3">
					<EditableMarkdown
						value={motivation.value}
						placeholder="What keeps you training? Your philosophy and goals…"
						variant="plain"
						onSave={async (v) => {
							await patchProfile({ motivation: v || null });
							motivation.value = v || null;
						}}
					/>
				</div>
			</Section>

			<div class="flex flex-col gap-4">
				<div class="flex items-center justify-between">
					<h2
						class="text-xs font-semibold uppercase tracking-widest"
						style={{ color: "var(--color-muted)" }}
					>
						Disciplines
					</h2>
					<div class="flex items-center gap-2">
						{addError.value && (
							<span class="text-xs" style={{ color: "var(--color-error)" }}>
								{addError.value}
							</span>
						)}
						<Button variant="outline" size="sm" onClick={addDiscipline}>
							Add
						</Button>
					</div>
				</div>
				{disciplines.value.length === 0 && (
					<p class="text-sm" style={{ color: "var(--color-muted)" }}>
						No disciplines added yet.
					</p>
				)}
				{disciplines.value.map((d, i) => (
					<Section
						key={d._key}
						title={d.name || `Discipline ${i + 1}`}
						action={
							<Button
								variant="destructive"
								size="sm"
								onClick={() => removeDiscipline(i)}
							>
								Remove
							</Button>
						}
					>
						<Row label="Name">
							<TextField
								value={d.name ?? ""}
								placeholder="e.g. Climbing"
								onInput={(e) =>
									updateDisciplineField(i, "name", e.target.value)
								}
								onBlur={() =>
									saveDisciplines(disciplines.value).catch(() => {})
								}
							/>
						</Row>
						<Row label="Years of practice">
							<TextField
								type="number"
								min="0"
								step="0.5"
								value={d.years_practice ?? ""}
								placeholder="0"
								class="w-20 text-right"
								onInput={(e) =>
									updateDisciplineField(
										i,
										"years_practice",
										e.target.value === "" ? null : e.target.value,
									)
								}
								onBlur={() =>
									saveDisciplines(disciplines.value).catch(() => {})
								}
							/>
						</Row>
						<Row label="Level">
							<ToggleGroup
								value={d.level ?? null}
								onChange={(v) => {
									const updated = disciplines.value.map((disc, idx) =>
										idx === i ? { ...disc, level: v } : disc,
									);
									saveDisciplines(updated).catch(() => {});
								}}
								options={LEVEL_OPTIONS}
								nullable
							/>
						</Row>
						<div
							class="px-4 py-3"
							style={{ borderTop: "1px solid var(--color-border)" }}
						>
							<p
								class="text-xs font-medium mb-2"
								style={{ color: "var(--color-muted)" }}
							>
								Notes
							</p>
							<EditableMarkdown
								value={d.notes ?? null}
								placeholder="Preferences, current grade, goals…"
								variant="plain"
								onSave={async (v) => {
									const updated = disciplines.value.map((disc, idx) =>
										idx === i ? { ...disc, notes: v || null } : disc,
									);
									await saveDisciplines(updated);
								}}
							/>
						</div>
					</Section>
				))}
			</div>

			<Section title="Constraints">
				<div class="px-4 py-3">
					<EditableMarkdown
						value={constraints.value}
						placeholder="Schedule, time per week, equipment…"
						variant="plain"
						onSave={async (v) => {
							await patchProfile({ constraints: v || null });
							constraints.value = v || null;
						}}
					/>
				</div>
			</Section>
		</main>
	);
}
