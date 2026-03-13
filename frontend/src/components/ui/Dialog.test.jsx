// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { signal } from "@preact/signals";
import {
	Dialog,
	DialogContent,
	DialogTitle,
	DialogDescription,
	DialogClose,
} from "./Dialog.jsx";

describe("Dialog", () => {
	it("renders content when open", () => {
		const open = signal(true);
		render(
			<Dialog openSignal={open}>
				<DialogContent>
					<DialogTitle>Settings</DialogTitle>
				</DialogContent>
			</Dialog>,
		);
		expect(
			screen.getByRole("heading", { name: "Settings" }),
		).toBeInTheDocument();
	});

	it("does not render content when closed", () => {
		const open = signal(false);
		render(
			<Dialog openSignal={open}>
				<DialogContent>
					<DialogTitle>Settings</DialogTitle>
				</DialogContent>
			</Dialog>,
		);
		expect(
			screen.queryByRole("heading", { name: "Settings" }),
		).not.toBeInTheDocument();
	});

	it("renders DialogDescription", () => {
		const open = signal(true);
		render(
			<Dialog openSignal={open}>
				<DialogContent>
					<DialogTitle>Title</DialogTitle>
					<DialogDescription>A description</DialogDescription>
				</DialogContent>
			</Dialog>,
		);
		expect(screen.getByText("A description")).toBeInTheDocument();
	});

	it("renders DialogClose", () => {
		const open = signal(true);
		render(
			<Dialog openSignal={open}>
				<DialogContent>
					<DialogTitle>Title</DialogTitle>
					<DialogClose>
						<button type="button">Close</button>
					</DialogClose>
				</DialogContent>
			</Dialog>,
		);
		expect(screen.getByRole("button", { name: "Close" })).toBeInTheDocument();
	});

	it("calls onOpenChange when open state changes", () => {
		const open = signal(true);
		const onOpenChange = vi.fn();
		render(
			<Dialog openSignal={open} onOpenChange={onOpenChange}>
				<DialogContent>
					<DialogTitle>Title</DialogTitle>
					<DialogClose>
						<button type="button">Close</button>
					</DialogClose>
				</DialogContent>
			</Dialog>,
		);
		fireEvent.click(screen.getByRole("button", { name: "Close" }));
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it("renders fullscreen DialogContent", () => {
		const open = signal(true);
		render(
			<Dialog openSignal={open}>
				<DialogContent fullscreen>
					<DialogTitle>Fullscreen</DialogTitle>
				</DialogContent>
			</Dialog>,
		);
		expect(screen.getByRole("dialog")).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Fullscreen" }),
		).toBeInTheDocument();
	});
});
