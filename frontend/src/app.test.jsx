/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { App } from "./app.jsx";

describe("App", () => {
	it("renders the Cove heading", () => {
		render(<App />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});
});
