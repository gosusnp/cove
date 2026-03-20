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
import { ActivityPicker } from "../components/shared/ActivityPicker.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { ProgramDetail } from "./ProgramDetail.jsx";
import { apiFetch } from "../lib/api.js";

// ─── ProgramList ────────────────────────────────────────────────────────────

function ProgramList({ programs, selectedId, onSelect, onNew, error }) {
	return (
		<div class="flex flex-col">
			{/* Header */}
			<div
				class="flex items-center justify-between px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Programs
				</h2>
				<Button variant="primary" size="sm" onClick={onNew}>
					+ New
				</Button>
			</div>

			{/* Program rows */}
			{error && (
				<p class="px-4 py-3 text-sm" style={{ color: "var(--color-error)" }}>
					{error}
				</p>
			)}
			{!error && programs.length === 0 ? (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No programs yet.
				</p>
			) : (
				!error &&
				programs.map((p, i) => (
					<ListItem
						key={p.id}
						label={p.name}
						active={p.id === selectedId}
						isLast={i === programs.length - 1}
						onClick={() => onSelect(p.id)}
					/>
				))
			)}
		</div>
	);
}

// ─── Programs (page) ─────────────────────────────────────────────────────────

export function Programs() {
	const { user } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const programs = useSignal([]);
	const loading = useSignal(true);

	// New program dialog
	const programDialog = useDialog();
	const nameField = useSignal("");
	const activityField = useSignal("");
	const saving = useSignal(false);
	const formError = useSignal("");

	const fetchError = useSignal("");

	const fetchPrograms = async () => {
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await apiFetch("/api/programs");
			if (!r.ok) throw new Error("Failed to fetch programs");
			programs.value = await r.json();
		} catch (err) {
			fetchError.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		fetchPrograms();
	}, [!!user]);

	const handleSelect = (id) => {
		route(`/programs/${id}`);
	};

	const openNew = () => {
		nameField.value = "";
		activityField.value = "";
		formError.value = "";
		programDialog.show();
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (!nameField.value.trim()) {
			formError.value = "Name is required.";
			return;
		}
		saving.value = true;
		formError.value = "";
		try {
			const r = await apiFetch("/api/programs", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name: nameField.value.trim(),
					activity: activityField.value || undefined,
				}),
			});
			if (!r.ok) {
				const data = await r.json();
				throw new Error(data.error || "Failed to create program");
			}
			await fetchPrograms();
			programDialog.hide();
		} catch (err) {
			formError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const handleProgramDeleted = () => {
		if (selectedId) {
			route("/programs");
		}
		fetchPrograms();
	};

	return (
		<>
			<ListDetail
				hasDetail={!!selectedId}
				emptyState="Select a program to view its sets."
				list={
					<ProgramList
						programs={programs.value}
						selectedId={selectedId}
						onSelect={handleSelect}
						onNew={openNew}
						error={fetchError.value}
					/>
				}
				detail={
					selectedId ? (
						<ProgramDetail
							programId={selectedId}
							onProgramUpdated={fetchPrograms}
							onProgramDeleted={handleProgramDeleted}
						/>
					) : null
				}
			/>

			{/* New Program dialog */}
			<Dialog openSignal={programDialog.open}>
				<DialogContent>
					<form onSubmit={handleSave}>
						<DialogTitle>New Program</DialogTitle>
						<div class="mt-4 flex flex-col gap-4">
							<TextField
								id="program-name"
								label="Name"
								value={nameField.value}
								onInput={(e) => {
									nameField.value = e.target.value;
								}}
								autoFocus
							/>
							<ActivityPicker
								value={activityField.value}
								onChange={(v) => {
									activityField.value = v;
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
