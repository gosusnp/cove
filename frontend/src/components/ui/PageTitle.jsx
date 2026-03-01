// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cn } from "../../lib/utils";

export function PageTitle({ class: className, children, ...props }) {
	return (
		<h1
			class={cn("text-2xl font-semibold", className)}
			style={{ color: "var(--color-text)" }}
			{...props}
		>
			{children}
		</h1>
	);
}
