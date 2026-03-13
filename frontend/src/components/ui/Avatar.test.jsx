// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Avatar } from "./Avatar.jsx";

describe("Avatar", () => {
	it("renders initials", () => {
		render(<Avatar initials="JS" />);
		expect(screen.getByText("JS")).toBeInTheDocument();
	});

	it("renders with role=img", () => {
		render(<Avatar initials="JS" label="Jane Smith" />);
		expect(screen.getByRole("img")).toBeInTheDocument();
	});

	it("sets aria-label from label prop", () => {
		render(<Avatar initials="JS" label="Jane Smith" />);
		expect(screen.getByLabelText("Jane Smith")).toBeInTheDocument();
	});
});
