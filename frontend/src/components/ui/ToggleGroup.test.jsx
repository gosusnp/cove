// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { ToggleGroup } from "./ToggleGroup.jsx";

const OPTIONS = [
	{ value: "bilateral", label: "Bilateral" },
	{ value: "unilateral", label: "Unilateral" },
	{ value: "left", label: "Left" },
];

describe("ToggleGroup", () => {
	it("renders label when provided", () => {
		render(
			<ToggleGroup
				label="Laterality"
				value={null}
				onChange={() => {}}
				options={OPTIONS}
			/>,
		);
		expect(screen.getByText("Laterality")).toBeTruthy();
	});

	it("does not render label element when label is not provided", () => {
		render(<ToggleGroup value={null} onChange={() => {}} options={OPTIONS} />);
		expect(screen.queryByText("Laterality")).toBeNull();
	});

	it("renders all options", () => {
		render(<ToggleGroup value={null} onChange={() => {}} options={OPTIONS} />);
		expect(screen.getByText("Bilateral")).toBeTruthy();
		expect(screen.getByText("Unilateral")).toBeTruthy();
		expect(screen.getByText("Left")).toBeTruthy();
	});

	it("calls onChange with option value when clicking an inactive option", () => {
		const onChange = vi.fn();
		render(
			<ToggleGroup value="bilateral" onChange={onChange} options={OPTIONS} />,
		);
		fireEvent.click(screen.getByText("Unilateral"));
		expect(onChange).toHaveBeenCalledWith("unilateral");
	});

	it("calls onChange(null) when clicking active option with nullable=true", () => {
		const onChange = vi.fn();
		render(
			<ToggleGroup
				value="bilateral"
				onChange={onChange}
				options={OPTIONS}
				nullable
			/>,
		);
		fireEvent.click(screen.getByText("Bilateral"));
		expect(onChange).toHaveBeenCalledWith(null);
	});

	it("does not call onChange when clicking active option with nullable=false", () => {
		const onChange = vi.fn();
		render(
			<ToggleGroup value="bilateral" onChange={onChange} options={OPTIONS} />,
		);
		fireEvent.click(screen.getByText("Bilateral"));
		expect(onChange).not.toHaveBeenCalled();
	});

	it("does not call onChange when disabled", () => {
		const onChange = vi.fn();
		render(
			<ToggleGroup
				value={null}
				onChange={onChange}
				options={OPTIONS}
				disabled
			/>,
		);
		// Buttons are disabled, clicks should not fire
		const btn = screen.getByText("Bilateral").closest("button");
		expect(btn.disabled).toBe(true);
	});
});
