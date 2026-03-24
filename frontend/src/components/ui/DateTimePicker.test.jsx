// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { DateTimePicker } from "./DateTimePicker.jsx";

describe("DateTimePicker", () => {
	it("renders a datetime-local input", () => {
		const { container } = render(<DateTimePicker id="t" />);
		const input = container.querySelector('input[type="datetime-local"]');
		expect(input).toBeInTheDocument();
	});

	it("renders a label when provided", () => {
		render(<DateTimePicker id="t" label="Start Date" />);
		expect(screen.getByLabelText("Start Date")).toBeInTheDocument();
	});

	it("forwards value", () => {
		const val = "2026-03-24T12:00";
		const { container } = render(<DateTimePicker id="t" value={val} />);
		const input = container.querySelector('input[type="datetime-local"]');
		expect(input).toHaveValue(val);
	});

	it("is disabled when disabled prop is set", () => {
		render(<DateTimePicker id="t" label="Start Date" disabled />);
		expect(screen.getByLabelText("Start Date")).toBeDisabled();
	});

	it("is read-only when readOnly prop is set", () => {
		render(<DateTimePicker id="t" label="Start Date" readOnly />);
		expect(screen.getByLabelText("Start Date")).toHaveAttribute("readonly");
	});

	describe("inline variant", () => {
		it("renders an input without border or background classes", () => {
			render(<DateTimePicker id="t" label="Start Date" inline />);
			const input = screen.getByLabelText("Start Date");
			expect(input.className).toContain("border-b-2");
			expect(input.className).not.toContain("rounded-lg");
		});

		it("forwards inputRef to the underlying input", () => {
			const ref = { current: null };
			render(<DateTimePicker id="t" inputRef={ref} />);
			expect(ref.current).toBeInstanceOf(HTMLInputElement);
			expect(ref.current.type).toBe("datetime-local");
		});
	});
});
