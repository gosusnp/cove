// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { signal } from "@preact/signals";
import { apiFetch } from "../lib/api.js";

// _labels holds { value, label } option objects ready for TagSelector.
// null = not yet fetched; [] = fetch done (empty or failed).
const _labels = signal(null);
let _fetching = false;

// useSessionLabels fetches GET /api/sessions/labels once and caches the result
// in a module-level signal. Returns [] until the fetch resolves. The _fetching
// flag prevents concurrent duplicate requests during the async window where
// _labels.value is still null.
export function useSessionLabels() {
	if (_labels.value === null && !_fetching) {
		_fetching = true;
		apiFetch("/api/sessions/labels")
			.then((r) => r.json())
			.then((d) => {
				_labels.value = d.labels.map((v) => ({
					value: v,
					label: v[0].toUpperCase() + v.slice(1),
				}));
			})
			.catch(() => {
				_labels.value = [];
			});
	}
	return _labels.value ?? [];
}
