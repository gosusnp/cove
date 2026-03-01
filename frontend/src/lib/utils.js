// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs) {
	return twMerge(clsx(inputs));
}

const TIME_UNITS = [
	{ unit: "year", ms: 365 * 24 * 60 * 60 * 1000 },
	{ unit: "month", ms: 30 * 24 * 60 * 60 * 1000 },
	{ unit: "day", ms: 24 * 60 * 60 * 1000 },
	{ unit: "hour", ms: 60 * 60 * 1000 },
	{ unit: "minute", ms: 60 * 1000 },
];

export function timeAgo(date) {
	const ms = Date.now() - new Date(date).getTime();
	for (const { unit, ms: unitMs } of TIME_UNITS) {
		const n = Math.floor(ms / unitMs);
		if (n >= 1)
			return new Intl.RelativeTimeFormat("en", { numeric: "auto" }).format(
				-n,
				unit,
			);
	}
	return "just now";
}
