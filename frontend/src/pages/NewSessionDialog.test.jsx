// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { fireEvent, render, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NewSessionDialog } from "./NewSessionDialog.jsx";

vi.mock("../hooks/useSessionLabels.js", () => ({
	useSessionLabels: () => [
		{ value: "deload", label: "Deload" },
		{ value: "recovery", label: "Recovery" },
	],
}));

vi.mock("../components/shared/ActivityPicker.jsx", () => ({
	ActivityPicker: ({ onChange }) => (
		<button
			type="button"
			data-testid="mock-activity-picker"
			onClick={() => onChange("strength")}
		>
			Pick Activity
		</button>
	),
}));

afterEach(() => {
	vi.restoreAllMocks();
});

function mockFetch({ ok = true, session = { id: 42 } } = {}) {
	vi.spyOn(global, "fetch").mockResolvedValue({
		ok,
		json: () => Promise.resolve(session),
	});
}

function renderDialog(overrides = {}) {
	const openSignal = signal(true);
	const onCreated = vi.fn();
	render(
		<NewSessionDialog
			openSignal={openSignal}
			onCreated={onCreated}
			{...overrides}
		/>,
	);
	return { openSignal, onCreated };
}

function getPostBody() {
	return JSON.parse(global.fetch.mock.calls[0][1].body);
}

describe("NewSessionDialog", () => {
	it("renders when open", () => {
		renderDialog();
		expect(screen.getByRole("dialog")).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Log Session" }),
		).toBeInTheDocument();
	});

	it("does not render when closed", () => {
		renderDialog({ openSignal: signal(false) });
		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
	});

	it("shows Name, Started, Duration, and Activity fields", () => {
		renderDialog();
		expect(screen.getByLabelText("Name")).toBeInTheDocument();
		expect(screen.getByLabelText("Started")).toBeInTheDocument();
		expect(screen.getByLabelText("Duration")).toBeInTheDocument();
		expect(screen.getByTestId("mock-activity-picker")).toBeInTheDocument();
	});

	it("shows Cancel and Log Session buttons", () => {
		renderDialog();
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Log Session" }),
		).toBeInTheDocument();
	});

	it("pre-fills the Started field with the current time", () => {
		renderDialog();
		expect(screen.getByLabelText("Started").value).toMatch(
			/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/,
		);
	});

	it("closes and resets fields when Cancel is clicked", () => {
		const { openSignal } = renderDialog();

		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Push day" },
		});
		fireEvent.input(screen.getByLabelText("Duration"), {
			target: { value: "45m" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

		expect(openSignal.value).toBe(false);
		expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
	});

	it("POSTs and calls onCreated with the session on success", async () => {
		const session = { id: 42 };
		mockFetch({ session });
		const { onCreated, openSignal } = renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(onCreated).toHaveBeenCalledWith(session));
		expect(openSignal.value).toBe(false);
	});

	it("includes program_name in POST body when name is entered", async () => {
		mockFetch();
		renderDialog();

		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Push day" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody().program_name).toBe("Push day");
	});

	it("omits program_name from POST body when name is empty", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody()).not.toHaveProperty("program_name");
	});

	it("includes duration_s and completed_at when duration is entered", async () => {
		mockFetch();
		renderDialog();

		fireEvent.input(screen.getByLabelText("Started"), {
			target: { value: "2026-03-24T10:00" },
		});
		fireEvent.input(screen.getByLabelText("Duration"), {
			target: { value: "1h 30m" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		const body = getPostBody();
		expect(body.duration_s).toBe(5400);
		expect(new Date(body.completed_at).toISOString()).toBe(
			new Date(
				new Date("2026-03-24T10:00").getTime() + 5400 * 1000,
			).toISOString(),
		);
	});

	it("omits duration_s and completed_at when duration is empty", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		const body = getPostBody();
		expect(body).not.toHaveProperty("duration_s");
		expect(body).not.toHaveProperty("completed_at");
	});

	it("includes activity in POST body when selected", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByTestId("mock-activity-picker"));
		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody().activity).toBe("strength");
	});

	it("omits activity from POST body when not selected", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody()).not.toHaveProperty("activity");
	});

	it("shows an error message when the POST fails", async () => {
		mockFetch({ ok: false });
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() =>
			expect(screen.getByText("Failed to create session")).toBeInTheDocument(),
		);
	});

	it("renders label toggles from useSessionLabels", () => {
		renderDialog();
		expect(screen.getByText("Deload")).toBeInTheDocument();
		expect(screen.getByText("Recovery")).toBeInTheDocument();
	});

	it("includes selected labels in POST body", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByText("Deload"));
		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody().labels).toEqual(["deload"]);
	});

	it("includes empty labels array when no label is selected", async () => {
		mockFetch();
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() => expect(global.fetch).toHaveBeenCalled());
		expect(getPostBody().labels).toEqual([]);
	});

	it("disables Log Session while saving", async () => {
		vi.spyOn(global, "fetch").mockImplementation(
			() => new Promise(() => {}), // never resolves
		);
		renderDialog();

		fireEvent.click(screen.getByRole("button", { name: "Log Session" }));

		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Log Session" }),
			).toBeDisabled(),
		);
	});
});
