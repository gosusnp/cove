// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render } from "@testing-library/preact";
import { LocationProvider } from "preact-iso";
import { vi } from "vitest";
import { AuthContext } from "./Auth.jsx";

export function withProviders(ui, { path = "/", user = null } = {}) {
	window.history.pushState({}, "", path);
	const auth = {
		user,
		token: user ? "tok" : null,
		login: vi.fn(),
		logout: vi.fn(),
		updateUser: vi.fn(),
	};
	return {
		...render(
			<LocationProvider>
				<AuthContext.Provider value={auth}>{ui}</AuthContext.Provider>
			</LocationProvider>,
		),
		auth,
	};
}
