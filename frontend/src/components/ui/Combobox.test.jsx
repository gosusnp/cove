// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { Combobox } from "./Combobox.jsx";

const OPTIONS = [
	{ value: "1", label: "Squat" },
	{ value: "2", label: "Bench Press" },
	{ value: "3", label: "Deadlift" },
];

function setup(props = {}) {
	const onChange = vi.fn();
	const { onChange: _drop, ...rest } = props;
	render(
		<Combobox
			options={OPTIONS}
			value=""
			placeholder="Search exercises..."
			{...rest}
			onChange={props.onChange ?? onChange}
		/>,
	);
	return { onChange: props.onChange ?? onChange };
}

function getListbox() {
	// listbox is always in the DOM; hidden via display:none when closed.
	return document.querySelector("[role=listbox]");
}

function isOpen() {
	return getListbox().style.display !== "none";
}

describe("Combobox", () => {
	it("renders placeholder when no value selected", () => {
		setup();
		expect(screen.getByPlaceholderText("Search exercises...")).toBeTruthy();
	});

	it("renders label when provided", () => {
		setup({ label: "Exercise" });
		expect(screen.getByText("Exercise")).toBeTruthy();
	});

	it("shows selected option label in input", () => {
		setup({ value: "2" });
		expect(screen.getByRole("combobox").value).toBe("Bench Press");
	});

	it("opens dropdown on focus", () => {
		setup();
		expect(isOpen()).toBe(false);
		fireEvent.click(screen.getByRole("combobox"));
		expect(isOpen()).toBe(true);
	});

	it("shows all options when open", () => {
		setup();
		fireEvent.click(screen.getByRole("combobox"));
		expect(isOpen()).toBe(true);
		const opts = document.querySelectorAll("[role=option]");
		expect(opts).toHaveLength(3);
	});

	it("filters options on typing", () => {
		setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		fireEvent.input(input, { target: { value: "bench" } });
		expect(screen.getByText("Bench Press")).toBeTruthy();
		expect(screen.queryByText("Squat")).toBeNull();
		expect(screen.queryByText("Deadlift")).toBeNull();
	});

	it("shows No results when filter matches nothing", () => {
		setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		fireEvent.input(input, { target: { value: "zzz" } });
		expect(screen.getByText("No results")).toBeTruthy();
	});

	it("selects an option on click and calls onChange", () => {
		const { onChange } = setup();
		fireEvent.click(screen.getByRole("combobox"));
		fireEvent.click(screen.getByText("Deadlift"));
		expect(onChange).toHaveBeenCalledWith("3");
	});

	it("closes dropdown after selecting", () => {
		setup();
		fireEvent.click(screen.getByRole("combobox"));
		expect(isOpen()).toBe(true);
		fireEvent.click(screen.getByText("Squat"));
		expect(isOpen()).toBe(false);
	});

	it("ArrowDown highlights next option", () => {
		setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		fireEvent.keyDown(input, { key: "ArrowDown" });
		const opts = document.querySelectorAll("[role=option]");
		expect(opts[0].className).toContain("bg-(--color-accent)/10");
	});

	it("ArrowUp highlights previous option", () => {
		setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		fireEvent.keyDown(input, { key: "ArrowDown" });
		fireEvent.keyDown(input, { key: "ArrowDown" });
		fireEvent.keyDown(input, { key: "ArrowUp" });
		const opts = document.querySelectorAll("[role=option]");
		expect(opts[0].className).toContain("bg-(--color-accent)/10");
	});

	it("Enter selects highlighted option", () => {
		const { onChange } = setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		fireEvent.keyDown(input, { key: "ArrowDown" });
		fireEvent.keyDown(input, { key: "Enter" });
		expect(onChange).toHaveBeenCalledWith("1");
	});

	it("Escape closes dropdown", () => {
		setup();
		const input = screen.getByRole("combobox");
		fireEvent.click(input);
		expect(isOpen()).toBe(true);
		fireEvent.keyDown(input, { key: "Escape" });
		expect(isOpen()).toBe(false);
	});

	it("shows checkmark on selected option", () => {
		setup({ value: "1" });
		fireEvent.click(screen.getByRole("combobox"));
		const opts = document.querySelectorAll("[role=option]");
		expect(opts[0].getAttribute("aria-selected")).toBe("true");
		expect(opts[1].getAttribute("aria-selected")).toBe("false");
	});

	it("closes on outside click", () => {
		setup();
		fireEvent.click(screen.getByRole("combobox"));
		expect(isOpen()).toBe(true);
		fireEvent.mouseDown(document.body);
		expect(isOpen()).toBe(false);
	});

	it("is disabled when disabled prop is set", () => {
		setup({ disabled: true });
		expect(screen.getByRole("combobox").disabled).toBe(true);
	});
});
