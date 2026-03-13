// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Switch } from "./Switch.jsx";

describe("Switch", () => {
	it("renders a switch", () => {
		render(<Switch />);
		expect(screen.getByRole("switch")).toBeInTheDocument();
	});

	it("is unchecked by default", () => {
		render(<Switch />);
		expect(screen.getByRole("switch")).toHaveAttribute("aria-checked", "false");
	});

	it("toggles when clicked", () => {
		render(<Switch />);
		const sw = screen.getByRole("switch");
		fireEvent.click(sw);
		expect(sw).toHaveAttribute("aria-checked", "true");
	});

	it("reflects a provided checkedSignal", () => {
		const checked = signal(true);
		render(<Switch checkedSignal={checked} />);
		expect(screen.getByRole("switch")).toHaveAttribute("aria-checked", "true");
	});

	it("updates the checkedSignal when toggled", () => {
		const checked = signal(false);
		render(<Switch checkedSignal={checked} />);
		fireEvent.click(screen.getByRole("switch"));
		expect(checked.value).toBe(true);
	});
});
