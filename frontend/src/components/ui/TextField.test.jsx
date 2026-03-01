// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { TextField } from "./TextField.jsx";

describe("TextField", () => {
	it("renders an input", () => {
		render(<TextField id="t" />);
		expect(screen.getByRole("textbox")).toBeInTheDocument();
	});

	it("renders a label when provided", () => {
		render(<TextField id="t" label="Token name" />);
		expect(screen.getByLabelText("Token name")).toBeInTheDocument();
	});

	it("does not render a label element when omitted", () => {
		render(<TextField id="t" placeholder="Placeholder" />);
		expect(screen.queryByRole("label")).not.toBeInTheDocument();
	});

	it("forwards placeholder", () => {
		render(<TextField id="t" placeholder="Enter a name" />);
		expect(screen.getByPlaceholderText("Enter a name")).toBeInTheDocument();
	});

	it("is disabled when disabled prop is set", () => {
		render(<TextField id="t" label="Name" disabled />);
		expect(screen.getByRole("textbox")).toBeDisabled();
	});

	it("is read-only when readOnly prop is set", () => {
		render(<TextField id="t" label="Token" value="pat_abc" readOnly />);
		expect(screen.getByRole("textbox")).toHaveAttribute("readonly");
	});

	it("forwards value", () => {
		render(<TextField id="t" value="pat_abc123" readOnly />);
		expect(screen.getByRole("textbox")).toHaveValue("pat_abc123");
	});
});
