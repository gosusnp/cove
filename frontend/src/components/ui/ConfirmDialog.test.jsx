// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog } from "./ConfirmDialog.jsx";

describe("ConfirmDialog", () => {
	it("renders title and description", () => {
		const open = signal(true);
		render(
			<ConfirmDialog
				openSignal={open}
				title="Delete Program"
				description="This will permanently delete the program."
				onConfirm={vi.fn()}
			/>,
		);
		expect(
			screen.getByRole("heading", { name: "Delete Program" }),
		).toBeInTheDocument();
		expect(
			screen.getByText("This will permanently delete the program."),
		).toBeInTheDocument();
	});

	it("Cancel closes the dialog", () => {
		const open = signal(true);
		render(
			<ConfirmDialog
				openSignal={open}
				title="Delete Program"
				onConfirm={vi.fn()}
			/>,
		);
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(open.value).toBe(false);
	});

	it("confirm button calls onConfirm and closes dialog", async () => {
		const open = signal(true);
		const onConfirm = vi.fn().mockResolvedValue(undefined);
		render(
			<ConfirmDialog
				openSignal={open}
				title="Delete Program"
				onConfirm={onConfirm}
			/>,
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete" }));
		// Wait for async onConfirm to resolve
		await vi.waitFor(() => expect(open.value).toBe(false));
		expect(onConfirm).toHaveBeenCalledOnce();
	});

	it("renders custom confirmLabel", () => {
		const open = signal(true);
		render(
			<ConfirmDialog
				openSignal={open}
				title="Remove Exercise"
				confirmLabel="Remove"
				onConfirm={vi.fn()}
			/>,
		);
		expect(screen.getByRole("button", { name: "Remove" })).toBeInTheDocument();
	});
});
