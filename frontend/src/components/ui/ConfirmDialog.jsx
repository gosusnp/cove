// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { Button } from "./Button.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
} from "./Dialog.jsx";

export function ConfirmDialog({
	openSignal,
	title,
	description,
	confirmLabel = "Delete",
	onConfirm,
}) {
	const confirmError = useSignal("");

	async function handleConfirm(e) {
		e.currentTarget.blur();
		confirmError.value = "";
		try {
			await onConfirm();
			openSignal.value = false;
		} catch (err) {
			confirmError.value = err?.message ?? "Something went wrong.";
		}
	}

	function handleCancel() {
		confirmError.value = "";
	}

	return (
		<Dialog openSignal={openSignal}>
			<DialogContent>
				<DialogTitle>{title}</DialogTitle>
				{description && <DialogDescription>{description}</DialogDescription>}
				{confirmError.value && (
					<p class="text-sm mt-2" style={{ color: "var(--color-error)" }}>
						{confirmError.value}
					</p>
				)}
				<div class="mt-6 flex justify-end gap-2">
					<DialogClose>
						<Button variant="outline" size="sm" onClick={handleCancel}>
							Cancel
						</Button>
					</DialogClose>
					<Button variant="destructive" size="sm" onClick={handleConfirm}>
						{confirmLabel}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
