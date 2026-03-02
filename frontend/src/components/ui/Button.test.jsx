// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { Button } from "./Button.jsx";

describe("Button", () => {
	it("renders children", () => {
		render(<Button>Save</Button>);
		expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
	});

	it("is disabled when disabled prop is set", () => {
		render(<Button disabled>Save</Button>);
		expect(screen.getByRole("button")).toBeDisabled();
	});

	it("calls onClick when clicked", () => {
		const onClick = vi.fn();
		render(<Button onClick={onClick}>Save</Button>);
		fireEvent.click(screen.getByRole("button"));
		expect(onClick).toHaveBeenCalled();
	});

	it("forwards type prop", () => {
		render(<Button type="submit">Submit</Button>);
		expect(screen.getByRole("button")).toHaveAttribute("type", "submit");
	});

	it("renders all variants without error", () => {
		for (const variant of ["primary", "outline", "ghost", "destructive"]) {
			render(<Button variant={variant}>{variant}</Button>);
			expect(screen.getByRole("button", { name: variant })).toBeInTheDocument();
		}
	});

	it("renders all sizes without error", () => {
		for (const size of ["sm", "md", "lg"]) {
			render(<Button size={size}>{size}</Button>);
			expect(screen.getByRole("button", { name: size })).toBeInTheDocument();
		}
	});
});
