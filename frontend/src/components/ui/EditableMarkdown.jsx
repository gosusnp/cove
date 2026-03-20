// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { Check, Pencil, X } from "lucide-preact";
import { useEffect, useRef } from "preact/hooks";
import { Button } from "./Button.jsx";
import { Markdown } from "./Markdown.jsx";
import { TextField } from "./TextField.jsx";

/**
 * Inline editable markdown field.
 *
 * Props:
 *   value        {string|null}  — current markdown string (controlled)
 *   placeholder  {string}       — text shown when value is empty
 *   onSave       {(v: string) => Promise<void>}  — called with trimmed value on confirm
 *   variant      {"default"|"plain"}  — "default" shows a bordered card; "plain" is borderless
 *                                       for use inside an existing card (default "default")
 *   minRows      {number}       — minimum textarea rows in edit mode (default 3)
 *   maxRows      {number}       — maximum textarea rows before scrolling (default 12)
 *   resizable    {boolean}      — allow vertical resize of textarea (default true)
 *   editLabel    {string}       — aria-label for the edit button (default "Edit")
 *   disabled     {boolean}      — hides the edit button
 */
export function EditableMarkdown({
	value,
	placeholder = "Add a description…",
	onSave,
	variant = "default",
	minRows = 3,
	maxRows = 12,
	resizable = true,
	editLabel = "Edit",
	disabled = false,
}) {
	const editing = useSignal(false);
	const draft = useSignal("");
	const saving = useSignal(false);
	const error = useSignal(null);
	const inputRef = useRef(null);

	useEffect(() => {
		if (!editing.value || !inputRef.current) return;
		const el = inputRef.current;
		el.focus();
		// Size textarea to content, clamped between minRows and maxRows.
		el.style.height = "auto";
		const lh = parseInt(getComputedStyle(el).lineHeight) || 20;
		const pt = parseInt(getComputedStyle(el).paddingTop) || 0;
		const pb = parseInt(getComputedStyle(el).paddingBottom) || 0;
		const minH = minRows * lh + pt + pb;
		const maxH = maxRows * lh + pt + pb;
		el.style.height = `${Math.min(Math.max(el.scrollHeight, minH), maxH)}px`;
	}, [editing.value]);

	function startEdit() {
		draft.value = value ?? "";
		error.value = null;
		editing.value = true;
	}

	function cancel() {
		editing.value = false;
	}

	async function save() {
		const trimmed = draft.value.trim();
		saving.value = true;
		error.value = null;
		try {
			await onSave(trimmed);
			editing.value = false;
		} catch (e) {
			error.value = e?.message ?? "Failed to save.";
		} finally {
			saving.value = false;
		}
	}

	if (editing.value) {
		return (
			<div class="flex flex-col gap-2">
				<TextField
					multiline
					inputRef={inputRef}
					rows={minRows}
					class={resizable ? "resize-y" : undefined}
					value={draft.value}
					onInput={(e) => (draft.value = e.target.value)}
					onKeyDown={(e) => {
						if (e.key === "Escape") cancel();
					}}
				/>
				{error.value && (
					<p class="text-xs" style={{ color: "var(--color-error)" }}>
						{error.value}
					</p>
				)}
				<div class="flex justify-end gap-1">
					<Button
						variant="outline"
						size="icon"
						onClick={save}
						disabled={saving.value}
						aria-label="Save"
					>
						<Check size={14} aria-hidden="true" />
					</Button>
					<Button
						variant="outline"
						size="icon"
						onClick={cancel}
						disabled={saving.value}
						aria-label="Cancel"
					>
						<X size={14} aria-hidden="true" />
					</Button>
				</div>
			</div>
		);
	}

	const isDefault = variant === "default";

	return (
		<div
			class={
				isDefault
					? "group flex items-start justify-between gap-2 rounded-md border border-(--color-border) bg-(--color-surface) px-3 py-2"
					: "group flex items-start justify-between gap-2"
			}
		>
			{value ? (
				<Markdown class="flex-1">{value}</Markdown>
			) : (
				<p
					class="text-sm italic flex-1"
					style={{ color: "var(--color-muted)", opacity: 0.5 }}
				>
					{placeholder}
				</p>
			)}
			{!disabled && (
				<Button
					variant="outline"
					size="icon"
					onClick={startEdit}
					aria-label={editLabel}
					class="opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
				>
					<Pencil size={14} aria-hidden="true" />
				</Button>
			)}
		</div>
	);
}
