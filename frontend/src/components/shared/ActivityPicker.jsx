// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Combobox } from "../ui/Combobox.jsx";
import { apiFetch } from "../../lib/api.js";

let cachedActivities = null;

export function ActivityPicker({
	value,
	onChange,
	label = "Activity",
	disabled = false,
	class: className,
}) {
	const activities = useSignal([]);

	useEffect(() => {
		if (cachedActivities !== null) {
			activities.value = cachedActivities;
			return;
		}
		apiFetch("/api/activities")
			.then((r) => (r.ok ? r.json() : []))
			.then((data) => {
				cachedActivities = data ?? [];
				activities.value = cachedActivities;
			})
			.catch(() => {});
	}, []);

	const options = activities.value.map((a) => ({ value: a, label: a }));

	return (
		<Combobox
			label={label}
			value={value ?? ""}
			onChange={onChange}
			options={options}
			placeholder="Select or type an activity…"
			freeform={true}
			disabled={disabled}
			class={className}
		/>
	);
}
