// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { App } from "./App.jsx";

describe("App", () => {
	it("renders the nav on the login page", () => {
		window.history.pushState({}, "", "/login");
		render(<App />);
		expect(screen.getByRole("banner")).toBeInTheDocument();
	});
});
