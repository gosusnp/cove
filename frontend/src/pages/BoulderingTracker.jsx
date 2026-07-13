// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useRef } from "preact/hooks";
import { Button } from "../components/ui/Button.jsx";
import { StripSelector } from "../components/ui/StripSelector.jsx";
import { TagSelector } from "../components/ui/TagSelector.jsx";

const GRADES = Array.from({ length: 18 }, (_, i) => `V${i}`);

const STYLE_OPTIONS = [
	{ value: "overhang", label: "Overhang" },
	{ value: "slab", label: "Slab" },
	{ value: "cave", label: "Cave" },
	{ value: "power", label: "Power" },
	{ value: "dyno", label: "Dyno" },
	{ value: "crimps", label: "Crimps" },
	{ value: "slopers", label: "Slopers" },
	{ value: "pockets", label: "Pockets" },
	{ value: "pinches", label: "Pinches" },
];

const TYPE_LABELS = { send: "Send", attempt: "Attempt", work: "Work" };

function capitalize(s) {
	return s[0].toUpperCase() + s.slice(1);
}

// Serialize bouldering entries into a structured text summary for session notes.
// Groups entries by (grade + sorted labels) key and counts sends/attempts/work.
export function serializeBoulderingEntries(entries) {
	if (!entries.length) return "";
	const groups = new Map();
	for (const entry of entries) {
		const labelStr = entry.labels.slice().sort().map(capitalize).join(" ");
		const key = labelStr ? `${entry.grade} ${labelStr}` : entry.grade;
		if (!groups.has(key)) groups.set(key, { send: 0, attempt: 0, work: 0 });
		groups.get(key)[entry.type]++;
	}
	return Array.from(groups.entries())
		.map(([key, counts]) => {
			const parts = [];
			if (counts.attempt)
				parts.push(
					`${counts.attempt} Attempt${counts.attempt !== 1 ? "s" : ""}`,
				);
			if (counts.send)
				parts.push(`${counts.send} Send${counts.send !== 1 ? "s" : ""}`);
			if (counts.work) parts.push(`${counts.work} Work`);
			return `- ${key}: (${parts.join(", ")})`;
		})
		.join("\n");
}

// BoulderingTracker renders the sticky grade/style/action controls and the
// chronological list of logged attempts. It manages grade and label selection
// state internally; logged entries are written to entriesSignal so the parent
// (SessionTracker) can serialize them into session notes on save.
export function BoulderingTracker({ entriesSignal }) {
	const grade = useSignal("V5");
	const labels = useSignal([]);
	const nextIdRef = useRef(0);

	function logEntry(type) {
		entriesSignal.value = [
			...entriesSignal.value,
			{
				id: ++nextIdRef.current,
				grade: grade.value,
				labels: labels.value.slice(),
				type,
			},
		];
	}

	function removeEntry(id) {
		entriesSignal.value = entriesSignal.value.filter((e) => e.id !== id);
	}

	return (
		<div class="flex flex-col">
			{/* Sticky grade/style/action controls */}
			<div
				class="sticky top-0 z-10 flex flex-col gap-3 py-3 border-b"
				style={{
					background: "var(--color-bg)",
					borderColor: "var(--color-border)",
				}}
			>
				<p class="text-sm font-semibold" style={{ color: "var(--color-text)" }}>
					Bouldering
				</p>
				<StripSelector
					options={GRADES}
					value={grade.value}
					onChange={(v) => {
						grade.value = v;
					}}
					class="w-full"
				/>
				<TagSelector
					label="Style"
					options={STYLE_OPTIONS}
					value={labels.value}
					onChange={(v) => {
						labels.value = v;
					}}
				/>
				<div class="flex gap-2">
					<Button
						variant="outline"
						size="sm"
						class="flex-1"
						onClick={() => logEntry("send")}
					>
						Log Send
					</Button>
					<Button
						variant="outline"
						size="sm"
						class="flex-1"
						onClick={() => logEntry("attempt")}
					>
						Log Attempt
					</Button>
					<Button
						variant="outline"
						size="sm"
						class="flex-1"
						onClick={() => logEntry("work")}
					>
						Log Work
					</Button>
				</div>
			</div>

			{/* Logged entries — chronological, no grouping */}
			{entriesSignal.value.length > 0 && (
				<ul class="flex flex-col">
					{entriesSignal.value.map((entry) => (
						<li
							key={entry.id}
							class="flex items-center gap-2 py-2.5 border-b text-sm"
							style={{ borderColor: "var(--color-border)" }}
						>
							<span class="font-medium" style={{ color: "var(--color-text)" }}>
								{entry.grade}
							</span>
							{entry.labels.length > 0 && (
								<span style={{ color: "var(--color-muted)" }}>
									{entry.labels.map(capitalize).join(" · ")}
								</span>
							)}
							<span
								class="ml-auto font-medium"
								style={{ color: "var(--color-text)" }}
							>
								{TYPE_LABELS[entry.type]}
							</span>
							<Button
								variant="ghost"
								size="icon"
								aria-label="Remove entry"
								onClick={() => removeEntry(entry.id)}
							>
								✕
							</Button>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
