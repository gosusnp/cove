// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

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
	async function handleConfirm(e) {
		e.currentTarget.blur();
		await onConfirm();
		openSignal.value = false;
	}

	return (
		<Dialog openSignal={openSignal}>
			<DialogContent>
				<DialogTitle>{title}</DialogTitle>
				{description && <DialogDescription>{description}</DialogDescription>}
				<div class="mt-6 flex justify-end gap-2">
					<DialogClose>
						<Button variant="outline" size="sm">
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
