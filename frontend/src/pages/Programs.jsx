// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { useLocation, useRoute } from "preact-iso";
import { useAuth } from "../Auth.jsx";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { ProgramDetail } from "./ProgramDetail.jsx";

// ─── ProgramList ────────────────────────────────────────────────────────────

function ProgramList({
	programs,
	selectedId,
	onSelect,
	onNew,
	onRename,
	onDelete,
	error,
}) {
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
				<p
					class="px-4 py-3 text-sm"
					style={{ color: "var(--color-error, red)" }}
				>
					{error}
				</p>
			)}
			{!error && programs.length === 0 ? (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No programs yet.
				</p>
			) : (
				!error &&
				programs.map((p, i) => {
					const isActive = p.id === selectedId;
					return (
						<div
							key={p.id}
							class="flex items-center justify-between gap-4"
							style={{
								borderBottom:
									i < programs.length - 1
										? "1px solid var(--color-border)"
										: undefined,
								background: isActive
									? "color-mix(in srgb, var(--color-accent) 10%, transparent)"
									: undefined,
								borderLeft: isActive
									? "3px solid var(--color-accent)"
									: "3px solid transparent",
							}}
						>
							<button
								type="button"
								class="flex-1 text-left px-4 py-3 cursor-pointer bg-transparent border-none text-sm truncate min-w-0"
								style={{ color: "var(--color-text)" }}
								onClick={() => onSelect(p.id)}
							>
								{p.name}
							</button>
							<div class="flex gap-2 shrink-0 pr-4">
								<Button variant="ghost" size="sm" onClick={() => onRename(p)}>
									Rename
								</Button>
								<Button
									variant="destructive"
									size="sm"
									onClick={() => onDelete(p)}
								>
									Delete
								</Button>
							</div>
						</div>
					);
				})
			)}
		</div>
	);
}

// ─── Programs (page) ─────────────────────────────────────────────────────────

export function Programs() {
	const { token } = useAuth();
	const { route } = useLocation();
	const { params } = useRoute();
	const selectedId = params?.id ? Number(params.id) : null;

	const programs = useSignal([]);
	const loading = useSignal(true);

	// New / rename dialog
	const programDialog = useDialog();
	const editingProgram = useSignal(null); // null = new, object = rename
	const nameField = useSignal("");
	const saving = useSignal(false);
	const formError = useSignal("");

	// Delete confirm dialog
	const deleteDialog = useDialog();
	const deletingProgram = useSignal(null);

	const fetchError = useSignal("");

	const fetchPrograms = async () => {
		if (!token) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await fetch("/api/programs", {
				headers: { Authorization: `Bearer ${token}` },
			});
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
	}, [token]);

	const handleSelect = (id) => {
		route(`/programs/${id}`);
	};

	const openNew = () => {
		editingProgram.value = null;
		nameField.value = "";
		formError.value = "";
		programDialog.show();
	};

	const openRename = (p) => {
		editingProgram.value = p;
		nameField.value = p.name;
		formError.value = "";
		programDialog.show();
	};

	const openDelete = (p) => {
		deletingProgram.value = p;
		deleteDialog.show();
	};

	const handleSave = async (e) => {
		e.preventDefault();
		if (!nameField.value.trim()) {
			formError.value = "Name is required.";
			return;
		}
		saving.value = true;
		formError.value = "";

		const isEdit = !!editingProgram.value;
		const url = isEdit
			? `/api/programs/${editingProgram.value.id}`
			: "/api/programs";
		const method = isEdit ? "PUT" : "POST";

		try {
			const r = await fetch(url, {
				method,
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ name: nameField.value.trim() }),
			});
			if (!r.ok) {
				const data = await r.json();
				throw new Error(data.error || "Failed to save program");
			}
			await fetchPrograms();
			programDialog.hide();
		} catch (err) {
			formError.value = err.message;
		} finally {
			saving.value = false;
		}
	};

	const handleDelete = async () => {
		const p = deletingProgram.value;
		if (!p) return;
		const r = await fetch(`/api/programs/${p.id}`, {
			method: "DELETE",
			headers: { Authorization: `Bearer ${token}` },
		});
		if (!r.ok) throw new Error("Failed to delete program");
		if (selectedId === p.id) {
			route("/programs");
		}
		await fetchPrograms();
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
						onRename={openRename}
						onDelete={openDelete}
						error={fetchError.value}
					/>
				}
				detail={selectedId ? <ProgramDetail programId={selectedId} /> : null}
			/>

			{/* New / Rename dialog */}
			<Dialog openSignal={programDialog.open}>
				<DialogContent>
					<form onSubmit={handleSave}>
						<DialogTitle>
							{editingProgram.value ? "Rename Program" : "New Program"}
						</DialogTitle>
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
							{formError.value && (
								<p class="text-sm" style={{ color: "var(--color-error, red)" }}>
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

			{/* Delete confirm dialog */}
			<ConfirmDialog
				openSignal={deleteDialog.open}
				title="Delete Program"
				description="This will permanently delete the program and all its sets."
				confirmLabel="Delete"
				onConfirm={handleDelete}
			/>
		</>
	);
}
