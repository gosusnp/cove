// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useAuth } from "../Auth.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import { apiFetch } from "../lib/api.js";

function formatDate(iso) {
	if (!iso) return null;
	return new Date(iso).toLocaleString(undefined, {
		dateStyle: "medium",
		timeStyle: "short",
	});
}

function formatDuration(seconds) {
	if (!seconds) return null;
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	if (h > 0) return `${h}h ${m}m`;
	if (m > 0) return `${m}m ${s > 0 ? `${s}s` : ""}`.trim();
	return `${s}s`;
}

export function SessionDetail({ sessionId }) {
	const { user } = useAuth();
	const session = useSignal(null);
	const loading = useSignal(true);
	const error = useSignal("");

	useEffect(() => {
		if (!user || !sessionId) return;
		loading.value = true;
		error.value = "";
		apiFetch(`/api/sessions/${sessionId}`)
			.then((r) => {
				if (!r.ok) throw new Error("Failed to fetch session");
				return r.json();
			})
			.then((data) => {
				session.value = data;
			})
			.catch((err) => {
				error.value = err.message;
			})
			.finally(() => {
				loading.value = false;
			});
	}, [sessionId, user]);

	if (loading.value) {
		return (
			<div class="flex items-center justify-center h-full">
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Loading…
				</p>
			</div>
		);
	}

	if (error.value) {
		return (
			<div class="flex items-center justify-center h-full">
				<p class="text-sm" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			</div>
		);
	}

	const s = session.value;
	if (!s) return null;

	const overviewRows = [
		s.activity && { label: "Activity", value: s.activity },
		s.program_name && { label: "Program", value: s.program_name },
		s.started_at && { label: "Started", value: formatDate(s.started_at) },
		s.completed_at && { label: "Completed", value: formatDate(s.completed_at) },
		s.duration_s && { label: "Duration", value: formatDuration(s.duration_s) },
		s.perceived_effort != null && {
			label: "Perceived effort",
			value: `${s.perceived_effort} / 10`,
		},
	].filter(Boolean);

	const title = s.program_name ?? s.activity ?? "Session";

	return (
		<div class="max-w-2xl mx-auto px-4 py-6 flex flex-col gap-6">
			<PageTitle>{title}</PageTitle>

			{overviewRows.length > 0 && (
				<Section title="Overview">
					{overviewRows.map((row, i) => (
						<Row
							key={row.label}
							label={row.label}
							last={i === overviewRows.length - 1}
						>
							{row.value}
						</Row>
					))}
				</Section>
			)}

			{s.program_structure && (
				<Section title="Program">
					<div
						class="px-4 py-3 text-sm whitespace-pre-wrap"
						style={{ color: "var(--color-text)" }}
					>
						{s.program_structure}
					</div>
				</Section>
			)}

			{s.session_notes && (
				<Section title="Notes">
					<div
						class="px-4 py-3 text-sm whitespace-pre-wrap"
						style={{ color: "var(--color-text)" }}
					>
						{s.session_notes}
					</div>
				</Section>
			)}
		</div>
	);
}
