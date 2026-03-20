// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { EditableMarkdown } from "./EditableMarkdown.jsx";

describe("EditableMarkdown", () => {
	it("renders markdown content in display mode", () => {
		const { container } = render(
			<EditableMarkdown value="**Hello**" onSave={vi.fn()} />,
		);
		expect(container.querySelector(".markdown-body")).toBeInTheDocument();
	});

	it("renders placeholder when value is empty", () => {
		render(
			<EditableMarkdown
				value={null}
				placeholder="Add a description…"
				onSave={vi.fn()}
			/>,
		);
		expect(screen.getByText("Add a description…")).toBeInTheDocument();
	});

	it("enters edit mode on pencil click", () => {
		render(<EditableMarkdown value="Some notes" onSave={vi.fn()} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		expect(screen.getByRole("textbox")).toBeInTheDocument();
	});

	it("pre-fills textarea with current value", () => {
		render(<EditableMarkdown value="Some notes" onSave={vi.fn()} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		expect(screen.getByRole("textbox")).toHaveValue("Some notes");
	});

	it("cancel returns to display mode without calling onSave", () => {
		const onSave = vi.fn();
		render(<EditableMarkdown value="Some notes" onSave={onSave} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
		expect(onSave).not.toHaveBeenCalled();
	});

	it("Escape cancels edit without calling onSave", () => {
		const onSave = vi.fn();
		render(<EditableMarkdown value="Some notes" onSave={onSave} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		fireEvent.keyDown(screen.getByRole("textbox"), { key: "Escape" });
		expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
		expect(onSave).not.toHaveBeenCalled();
	});

	it("save calls onSave with trimmed value and returns to display mode", async () => {
		const onSave = vi.fn().mockResolvedValue(undefined);
		render(<EditableMarkdown value="Old" onSave={onSave} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		fireEvent.input(screen.getByRole("textbox"), {
			target: { value: "  New content  " },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));
		await vi.waitFor(() =>
			expect(screen.queryByRole("textbox")).not.toBeInTheDocument(),
		);
		expect(onSave).toHaveBeenCalledWith("New content");
	});

	it("shows error and stays in edit mode when onSave rejects", async () => {
		const onSave = vi.fn().mockRejectedValue(new Error("Network error"));
		render(<EditableMarkdown value="Old" onSave={onSave} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		fireEvent.click(screen.getByRole("button", { name: "Save" }));
		await vi.waitFor(() =>
			expect(screen.getByText("Network error")).toBeInTheDocument(),
		);
		expect(screen.getByRole("textbox")).toBeInTheDocument();
	});

	it("hides edit button when disabled", () => {
		render(<EditableMarkdown value="Read only" onSave={vi.fn()} disabled />);
		expect(
			screen.queryByRole("button", { name: "Edit" }),
		).not.toBeInTheDocument();
	});

	it("textarea has resize-y class when resizable (default)", () => {
		render(<EditableMarkdown value="text" onSave={vi.fn()} />);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		expect(screen.getByRole("textbox")).toHaveClass("resize-y");
	});

	it("textarea does not have resize-y when resizable=false", () => {
		render(
			<EditableMarkdown value="text" onSave={vi.fn()} resizable={false} />,
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit" }));
		expect(screen.getByRole("textbox")).not.toHaveClass("resize-y");
	});

	it("default variant renders a bordered container", () => {
		const { container } = render(
			<EditableMarkdown value="text" onSave={vi.fn()} />,
		);
		expect(container.firstChild).toHaveClass("border");
	});

	it("plain variant renders without a border", () => {
		const { container } = render(
			<EditableMarkdown value="text" onSave={vi.fn()} variant="plain" />,
		);
		expect(container.firstChild).not.toHaveClass("border");
	});
});
