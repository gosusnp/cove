// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Button } from "../ui/Button.jsx";
import { TextField } from "../ui/TextField.jsx";
import { apiFetch } from "../../lib/api.js";

export function FDCSearch({ initialQuery = "", onSelect, onCancel }) {
	const query = useSignal(initialQuery);
	const results = useSignal(null);
	const loading = useSignal(false);
	const error = useSignal("");

	const search = async () => {
		if (!query.value.trim()) return;
		loading.value = true;
		error.value = "";
		try {
			const r = await apiFetch(
				`/api/fdc/search?q=${encodeURIComponent(query.value.trim())}`,
			);
			if (!r.ok) throw new Error("Search failed");
			const data = await r.json();
			results.value = data.foods ?? [];
		} catch (err) {
			error.value = err.message;
		} finally {
			loading.value = false;
		}
	};

	useEffect(() => {
		if (initialQuery.trim()) search();
	}, []);

	return (
		<div class="flex flex-col gap-2">
			<div class="flex gap-2 items-end">
				<div class="flex-1">
					<TextField
						id="fdc-query"
						label="Search FDC"
						value={query.value}
						onInput={(e) => {
							query.value = e.target.value;
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								search();
							}
						}}
					/>
				</div>
				<Button
					size="sm"
					type="button"
					onClick={search}
					disabled={loading.value}
				>
					{loading.value ? "…" : "Search"}
				</Button>
				{onCancel && (
					<Button variant="outline" size="sm" type="button" onClick={onCancel}>
						Cancel
					</Button>
				)}
			</div>
			{error.value && (
				<p class="text-xs" style={{ color: "var(--color-error)" }}>
					{error.value}
				</p>
			)}
			{results.value !== null && results.value.length === 0 && (
				<p class="text-xs" style={{ color: "var(--color-muted)" }}>
					No results found.
				</p>
			)}
			{results.value !== null && results.value.length > 0 && (
				<div
					class="flex flex-col overflow-y-auto rounded border max-h-48"
					style={{ borderColor: "var(--color-border)" }}
				>
					{results.value.map((food) => (
						<button
							key={food.fdc_id}
							type="button"
							class="flex items-center gap-2 px-2 py-1.5 text-left text-xs hover:bg-[var(--color-bg)] transition-colors"
							onClick={() => onSelect(food)}
						>
							<span class="flex-1" style={{ color: "var(--color-text)" }}>
								{food.name}
							</span>
							<span
								class="shrink-0 rounded px-1 py-0.5 text-xs"
								style={
									food.data_type === "Foundation"
										? { background: "var(--color-accent)", color: "#fff" }
										: {
												background: "var(--color-surface)",
												color: "var(--color-muted)",
											}
								}
							>
								{food.data_type === "Foundation" ? "Foundation" : "SR Legacy"}
							</span>
						</button>
					))}
				</div>
			)}
		</div>
	);
}
