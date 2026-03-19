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

	it("keeps dialog open and shows error when onConfirm throws", async () => {
		const open = signal(true);
		const onConfirm = vi
			.fn()
			.mockRejectedValue(new Error("Failed to delete session"));
		render(
			<ConfirmDialog
				openSignal={open}
				title="Delete Session"
				onConfirm={onConfirm}
			/>,
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete" }));
		await vi.waitFor(() =>
			expect(screen.getByText("Failed to delete session")).toBeInTheDocument(),
		);
		expect(open.value).toBe(true);
	});

	it("clears error when cancel is clicked after a failure", async () => {
		const open = signal(true);
		const onConfirm = vi.fn().mockRejectedValue(new Error("Oops"));
		render(
			<ConfirmDialog
				openSignal={open}
				title="Delete Session"
				onConfirm={onConfirm}
			/>,
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete" }));
		await vi.waitFor(() =>
			expect(screen.getByText("Oops")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(screen.queryByText("Oops")).not.toBeInTheDocument();
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
