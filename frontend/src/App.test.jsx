// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App.jsx";

describe("App", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("renders the nav on the login page", async () => {
		vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false }));
		window.history.pushState({}, "", "/login");
		render(<App />);
		await waitFor(() => expect(screen.getByRole("banner")).toBeInTheDocument());
	});
});
