// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { createContext } from "preact";
import { useContext, useEffect, useState } from "preact/hooks";

const STORAGE_KEY = "cove_session";

function readSession() {
	// Prefer a token delivered via OAuth redirect query param.
	const params = new URLSearchParams(window.location.search);
	const token = params.get("token");
	if (token) {
		const s = { token, user: {} };
		localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
		window.history.replaceState({}, "", window.location.pathname);
		return s;
	}
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? JSON.parse(raw) : null;
	} catch {
		return null;
	}
}

export const AuthContext = createContext(null);

export function AuthProvider({ children }) {
	const [session, setSession] = useState(readSession);

	function login(token, user) {
		const s = { token, user };
		localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
		setSession(s);
	}

	function logout() {
		localStorage.removeItem(STORAGE_KEY);
		setSession(null);
	}

	function updateUser(user) {
		setSession((s) => {
			if (!s) return s;
			const next = { ...s, user };
			localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
			return next;
		});
	}

	useEffect(() => {
		const token = session?.token;
		if (!token) return;
		fetch("/api/users/me", {
			headers: { Authorization: `Bearer ${token}` },
		})
			.then((r) => {
				if (r.status === 401) {
					logout();
					return;
				}
				return r.json().then(updateUser);
			})
			.catch(() => {});
	}, []);

	return (
		<AuthContext.Provider
			value={{
				user: session?.user ?? null,
				token: session?.token ?? null,
				login,
				logout,
				updateUser,
			}}
		>
			{children}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	return useContext(AuthContext);
}
