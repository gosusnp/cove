// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";

export function useDialog(initialOpen = false) {
	const open = useSignal(initialOpen);

	return {
		open,
		show: () => {
			open.value = true;
		},
		hide: () => {
			open.value = false;
		},
		toggle: () => {
			open.value = !open.value;
		},
	};
}
