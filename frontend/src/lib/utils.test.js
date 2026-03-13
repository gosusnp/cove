// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { afterEach, describe, expect, it, vi } from "vitest";
import { timeAgo } from "./utils";

describe("timeAgo", () => {
	afterEach(() => vi.useRealTimers());

	function setNow(iso) {
		vi.useFakeTimers();
		vi.setSystemTime(new Date(iso));
	}

	it("returns just now for less than a minute ago", () => {
		setNow("2026-03-01T12:00:00Z");
		expect(timeAgo("2026-03-01T11:59:30Z")).toBe("just now");
	});

	it("returns minutes ago", () => {
		setNow("2026-03-01T12:05:00Z");
		expect(timeAgo("2026-03-01T12:00:00Z")).toBe("5 minutes ago");
	});

	it("returns hours ago", () => {
		setNow("2026-03-01T15:00:00Z");
		expect(timeAgo("2026-03-01T12:00:00Z")).toBe("3 hours ago");
	});

	it("returns days ago", () => {
		setNow("2026-03-04T12:00:00Z");
		expect(timeAgo("2026-03-01T12:00:00Z")).toBe("3 days ago");
	});

	it("returns months ago", () => {
		setNow("2026-03-01T12:00:00Z");
		// Dec 30 → Mar 1 = 60 days = floor(60/30) = 2 months
		expect(timeAgo("2025-12-30T12:00:00Z")).toBe("2 months ago");
	});

	it("returns years ago", () => {
		setNow("2026-03-01T12:00:00Z");
		expect(timeAgo("2024-03-01T12:00:00Z")).toBe("2 years ago");
	});

	it("uses auto numeric (last month instead of 1 month ago)", () => {
		setNow("2026-03-01T12:00:00Z");
		expect(timeAgo("2026-01-30T12:00:00Z")).toBe("last month");
	});
});
