// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
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
import { ListItem } from "../components/ui/ListItem.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { ProgramDetail } from "./ProgramDetail.jsx";

// ─── Icons ────────────────────────────────────────────────────────────────────

function PencilIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M11.013 1.427a1.75 1.75 0 0 1 2.474 0l1.086 1.086a1.75 1.75 0 0 1 0 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 0 1-.927-.928l.929-3.25c.081-.286.235-.547.445-.758l8.61-8.609Z"
				fill="currentColor"
			/>
		</svg>
	);
}

function TrashIcon() {
	return (
		<svg
			width="14"
			height="14"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				fill-rule="evenodd"
				clip-rule="evenodd"
				d="M6.5 1.75a.25.25 0 0 1 .25-.25h2.5a.25.25 0 0 1 .25.25V3h-3V1.75ZM5 3V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75V3h2.25a.75.75 0 0 1 0 1.5H14v8.75A1.75 1.75 0 0 1 12.25 15h-8.5A1.75 1.75 0 0 1 2 13.25V4.5H.75a.75.75 0 0 1 0-1.5H5ZM3.5 4.5v8.75c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25V4.5h-9Z"
				fill="currentColor"
			/>
		</svg>
	);
}

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
						actions={
							<>
								<Tooltip>
									<TooltipTrigger>
										<Button
											variant="ghost"
											size="icon"
											aria-label="Rename"
											onClick={() => onRename(p)}
										>
											<PencilIcon />
										</Button>
									</TooltipTrigger>
									<TooltipContent>Rename</TooltipContent>
								</Tooltip>
								<Tooltip>
									<TooltipTrigger>
										<Button
											variant="ghost"
											size="icon"
											aria-label="Delete"
											onClick={() => onDelete(p)}
										>
											<TrashIcon />
										</Button>
									</TooltipTrigger>
									<TooltipContent>Delete</TooltipContent>
								</Tooltip>
							</>
						}
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
		if (!user) return;
		loading.value = true;
		fetchError.value = "";
		try {
			const r = await fetch("/api/programs", {
				credentials: "include",
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
	}, [!!user]);

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
				credentials: "include",
				headers: { "Content-Type": "application/json" },
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
			credentials: "include",
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
