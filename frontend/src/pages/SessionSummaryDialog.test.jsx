// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { SessionSummaryDialog } from "./SessionSummaryDialog.jsx";

// Fixed date for deterministic output.
const FIXED_DATE = new Date("2026-03-12T14:00:00");

function makeProps(overrides = {}) {
	return {
		openSignal: signal(true),
		completedAt: FIXED_DATE,
		elapsed: 0,
		programName: null,
		notesSignal: signal(""),
		effortSignal: signal(null),
		saving: false,
		saveError: "",
		onCancel: vi.fn(),
		onSave: vi.fn(),
		...overrides,
	};
}

describe("SessionSummaryDialog", () => {
	it("renders when open", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		expect(screen.getByRole("dialog")).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Session Summary" }),
		).toBeInTheDocument();
	});

	it("does not render when closed", () => {
		render(
			<SessionSummaryDialog {...makeProps({ openSignal: signal(false) })} />,
		);
		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
	});

	it("shows Cancel and Save Session buttons", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Save Session" }),
		).toBeInTheDocument();
	});

	it("formats duration without hours correctly", () => {
		render(<SessionSummaryDialog {...makeProps({ elapsed: 185 })} />);
		expect(screen.getByText("3m 05s")).toBeInTheDocument();
	});

	it("formats duration with hours correctly", () => {
		render(<SessionSummaryDialog {...makeProps({ elapsed: 3725 })} />);
		expect(screen.getByText("1h 02m 05s")).toBeInTheDocument();
	});

	it("shows program name when provided", () => {
		render(
			<SessionSummaryDialog
				{...makeProps({ programName: "Strength Block" })}
			/>,
		);
		expect(screen.getByText("Program")).toBeInTheDocument();
		expect(screen.getByText("Strength Block")).toBeInTheDocument();
	});

	it("hides program row when programName is null", () => {
		render(<SessionSummaryDialog {...makeProps({ programName: null })} />);
		expect(screen.queryByText("Program")).not.toBeInTheDocument();
	});

	it("renders the perceived effort slider", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		expect(
			screen.getByRole("slider", { name: /perceived effort/i }),
		).toBeInTheDocument();
	});

	it("does not show value display when effort is unset", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		expect(screen.queryByText(/\/ 10/)).not.toBeInTheDocument();
	});

	it("shows value display when effort signal has a value", () => {
		render(
			<SessionSummaryDialog {...makeProps({ effortSignal: signal(7) })} />,
		);
		expect(screen.getByText("7 / 10")).toBeInTheDocument();
	});

	it("updates effortSignal when slider changes", () => {
		const effortSignal = signal(null);
		render(<SessionSummaryDialog {...makeProps({ effortSignal })} />);
		fireEvent.input(screen.getByRole("slider"), { target: { value: "8" } });
		expect(effortSignal.value).toBe(8);
	});

	it("renders the notes textarea with correct placeholder", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		expect(
			screen.getByPlaceholderText("How did it feel? Any PRs or observations?"),
		).toBeInTheDocument();
	});

	it("reflects existing notes in the textarea", () => {
		render(
			<SessionSummaryDialog
				{...makeProps({ notesSignal: signal("Great session!") })}
			/>,
		);
		expect(screen.getByDisplayValue("Great session!")).toBeInTheDocument();
	});

	it("updates notesSignal when user types", () => {
		const notesSignal = signal("");
		render(<SessionSummaryDialog {...makeProps({ notesSignal })} />);
		fireEvent.input(screen.getByRole("textbox", { name: /notes/i }), {
			target: { value: "New PR on squat" },
		});
		expect(notesSignal.value).toBe("New PR on squat");
	});

	it("calls onCancel when Cancel is clicked", () => {
		const onCancel = vi.fn();
		render(<SessionSummaryDialog {...makeProps({ onCancel })} />);
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(onCancel).toHaveBeenCalledOnce();
	});

	it("calls onSave when Save Session is clicked", () => {
		const onSave = vi.fn();
		render(<SessionSummaryDialog {...makeProps({ onSave })} />);
		fireEvent.click(screen.getByRole("button", { name: "Save Session" }));
		expect(onSave).toHaveBeenCalledOnce();
	});

	it("disables Save Session when saving", () => {
		render(<SessionSummaryDialog {...makeProps({ saving: true })} />);
		expect(screen.getByRole("button", { name: "Save Session" })).toBeDisabled();
	});

	it("shows saveError when provided", () => {
		render(
			<SessionSummaryDialog
				{...makeProps({ saveError: "Failed to save session" })}
			/>,
		);
		expect(screen.getByText("Failed to save session")).toBeInTheDocument();
	});

	it("shows the date from completedAt", () => {
		render(<SessionSummaryDialog {...makeProps()} />);
		// FIXED_DATE is 2026-03-12 — the formatted label depends on locale but
		// "Date" row label should always be present.
		expect(screen.getByText("Date")).toBeInTheDocument();
	});
});
