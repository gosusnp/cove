// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { StripSelector } from "./StripSelector.jsx";

const GRADES = Array.from({ length: 18 }, (_, i) => `V${i}`);

describe("StripSelector", () => {
	it("renders all options", () => {
		render(<StripSelector options={GRADES} value={null} onChange={() => {}} />);
		for (const grade of GRADES) {
			expect(screen.getByRole("button", { name: grade })).toBeInTheDocument();
		}
	});

	it("marks the selected option as pressed", () => {
		render(<StripSelector options={GRADES} value="V5" onChange={() => {}} />);
		expect(screen.getByRole("button", { name: "V5" })).toHaveAttribute(
			"aria-pressed",
			"true",
		);
		expect(screen.getByRole("button", { name: "V4" })).toHaveAttribute(
			"aria-pressed",
			"false",
		);
	});

	it("calls onChange with the tapped option", () => {
		const onChange = vi.fn();
		render(<StripSelector options={GRADES} value="V5" onChange={onChange} />);
		fireEvent.click(screen.getByRole("button", { name: "V7" }));
		expect(onChange).toHaveBeenCalledWith("V7");
	});

	it("renders buttons as disabled when disabled prop is set", () => {
		const onChange = vi.fn();
		render(
			<StripSelector
				options={GRADES}
				value="V5"
				onChange={onChange}
				disabled
			/>,
		);
		expect(screen.getByRole("button", { name: "V7" })).toBeDisabled();
	});

	it("renders an optional label", () => {
		render(
			<StripSelector
				options={GRADES}
				value={null}
				onChange={() => {}}
				label="Grade"
			/>,
		);
		expect(screen.getByText("Grade")).toBeInTheDocument();
	});
});
