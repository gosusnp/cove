// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { TagSelector } from "./TagSelector.jsx";

const OPTIONS = [
	{ value: "deload", label: "Deload" },
	{ value: "recovery", label: "Recovery" },
];

describe("TagSelector", () => {
	it("renders all options", () => {
		render(<TagSelector value={[]} onChange={() => {}} options={OPTIONS} />);
		expect(screen.getByText("Deload")).toBeInTheDocument();
		expect(screen.getByText("Recovery")).toBeInTheDocument();
	});

	it("renders label when provided", () => {
		render(
			<TagSelector
				label="Session type"
				value={[]}
				onChange={() => {}}
				options={OPTIONS}
			/>,
		);
		expect(screen.getByText("Session type")).toBeInTheDocument();
	});

	it("does not render label element when label is not provided", () => {
		render(<TagSelector value={[]} onChange={() => {}} options={OPTIONS} />);
		expect(screen.queryByText("Session type")).toBeNull();
	});

	it("calls onChange with value added when clicking an inactive tag", () => {
		const onChange = vi.fn();
		render(<TagSelector value={[]} onChange={onChange} options={OPTIONS} />);
		fireEvent.click(screen.getByText("Deload"));
		expect(onChange).toHaveBeenCalledWith(["deload"]);
	});

	it("calls onChange with value removed when clicking an active tag", () => {
		const onChange = vi.fn();
		render(
			<TagSelector
				value={["deload", "recovery"]}
				onChange={onChange}
				options={OPTIONS}
			/>,
		);
		fireEvent.click(screen.getByText("Deload"));
		expect(onChange).toHaveBeenCalledWith(["recovery"]);
	});

	it("supports selecting multiple tags independently", () => {
		const onChange = vi.fn();
		render(
			<TagSelector value={["deload"]} onChange={onChange} options={OPTIONS} />,
		);
		fireEvent.click(screen.getByText("Recovery"));
		expect(onChange).toHaveBeenCalledWith(["deload", "recovery"]);
	});

	it("does not call onChange when disabled", () => {
		const onChange = vi.fn();
		render(
			<TagSelector value={[]} onChange={onChange} options={OPTIONS} disabled />,
		);
		expect(screen.getByText("Deload").closest("button")).toBeDisabled();
		expect(screen.getByText("Recovery").closest("button")).toBeDisabled();
		fireEvent.click(screen.getByText("Deload"));
		expect(onChange).not.toHaveBeenCalled();
	});
});
