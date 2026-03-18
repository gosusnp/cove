// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { createContext } from "preact";
import { useContext, useEffect, useState } from "preact/hooks";

export const AuthContext = createContext(null);

export function AuthProvider({ children }) {
	const [user, setUser] = useState(null);
	const [loading, setLoading] = useState(true);

	// Bootstrap: check if we have an active session via the HttpOnly cookie.
	useEffect(() => {
		fetch("/api/users/me", { credentials: "include" })
			.then((r) => {
				if (r.ok) return r.json().then(setUser);
			})
			.catch(() => {})
			.finally(() => setLoading(false));
	}, []);

	function logout() {
		fetch("/api/users/logout", {
			method: "POST",
			credentials: "include",
		}).catch(() => {});
		setUser(null);
	}

	function updateUser(u) {
		setUser(u);
	}

	return (
		<AuthContext.Provider value={{ user, loading, logout, updateUser }}>
			{children}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	return useContext(AuthContext);
}
