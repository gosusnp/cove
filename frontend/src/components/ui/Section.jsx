// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export function Section({ title, children }) {
	return (
		<section class="flex flex-col gap-3">
			<h2
				class="text-xs font-semibold uppercase tracking-widest"
				style={{ color: "var(--color-muted)" }}
			>
				{title}
			</h2>
			<div
				class="rounded-xl overflow-hidden"
				style={{
					background: "var(--color-surface)",
					border: "1px solid var(--color-border)",
				}}
			>
				{children}
			</div>
		</section>
	);
}

export function Row({ label, sublabel, children, last }) {
	return (
		<div
			class="flex items-center justify-between px-4 py-3 gap-4"
			style={last ? {} : { borderBottom: "1px solid var(--color-border)" }}
		>
			<div class="flex flex-col gap-0.5 min-w-0">
				<span class="text-sm" style={{ color: "var(--color-text)" }}>
					{label}
				</span>
				{sublabel && (
					<span class="text-xs" style={{ color: "var(--color-muted)" }}>
						{sublabel}
					</span>
				)}
			</div>
			<div
				class="flex items-center gap-2 text-sm shrink-0"
				style={{ color: "var(--color-muted)" }}
			>
				{children}
			</div>
		</div>
	);
}

export function Divider() {
	return <hr style={{ borderColor: "var(--color-border)" }} />;
}
