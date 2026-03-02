// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { Login } from "./Login.jsx";
import { withProviders } from "../test-utils.jsx";

describe("Login", () => {
	it("renders the Cove heading", () => {
		withProviders(<Login />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the Continue with Google link pointing to /auth/login", () => {
		withProviders(<Login />);
		const link = screen.getByRole("link", { name: /continue with google/i });
		expect(link).toHaveAttribute("href", "/auth/login");
	});
});
