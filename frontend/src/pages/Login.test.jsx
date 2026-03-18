// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { Login } from "./Login.jsx";

describe("Login", () => {
	it("renders the Cove heading", () => {
		withProviders(<Login />);
		expect(screen.getByRole("heading", { name: "Cove" })).toBeInTheDocument();
	});

	it("renders the Continue with Google button", () => {
		withProviders(<Login />);
		expect(
			screen.getByRole("button", { name: /continue with google/i }),
		).toBeInTheDocument();
	});

	it("shows an error message when native sign-in fails", async () => {
		const { Capacitor } = await import("@capacitor/core");
		const { SocialLogin } = await import("@capgo/capacitor-social-login");
		vi.spyOn(Capacitor, "isNativePlatform").mockReturnValue(true);
		vi.spyOn(SocialLogin, "login").mockRejectedValue(new Error("auth failed"));

		withProviders(<Login />);
		const button = screen.getByRole("button", {
			name: /continue with google/i,
		});
		fireEvent.click(button);

		await waitFor(() =>
			expect(
				screen.getByText("Sign-in failed. Please try again."),
			).toBeInTheDocument(),
		);

		vi.restoreAllMocks();
	});
});
