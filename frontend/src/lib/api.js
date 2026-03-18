// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

export function apiFetch(url, options = {}) {
	return fetch(`${API_BASE}${url}`, {
		credentials: "include",
		...options,
	}).then((r) => {
		if (r.status === 401) {
			window.location.assign("/login");
		}
		return r;
	});
}
