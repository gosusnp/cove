// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api.js";

function mockAssign() {
	const fn = vi.fn();
	vi.stubGlobal("location", { ...window.location, assign: fn });
	return fn;
}

describe("apiFetch", () => {
	afterEach(() => vi.restoreAllMocks());

	it("includes credentials on every request", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ status: 200, ok: true });
		await apiFetch("/api/foo");
		expect(fetch).toHaveBeenCalledWith(
			"/api/foo",
			expect.objectContaining({ credentials: "include" }),
		);
	});

	it("returns the response on success", async () => {
		const resp = { status: 200, ok: true };
		vi.spyOn(global, "fetch").mockResolvedValue(resp);
		const r = await apiFetch("/api/foo");
		expect(r).toBe(resp);
	});

	it("merges caller options without overriding credentials", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ status: 200, ok: true });
		await apiFetch("/api/foo", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
		});
		expect(fetch).toHaveBeenCalledWith(
			"/api/foo",
			expect.objectContaining({
				method: "POST",
				credentials: "include",
				headers: { "Content-Type": "application/json" },
			}),
		);
	});

	it("redirects to /login on 401", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ status: 401, ok: false });
		const assignSpy = mockAssign();
		await apiFetch("/api/foo");
		expect(assignSpy).toHaveBeenCalledWith("/login");
	});

	it("still returns the 401 response after redirect", async () => {
		const resp = { status: 401, ok: false };
		vi.spyOn(global, "fetch").mockResolvedValue(resp);
		mockAssign();
		const r = await apiFetch("/api/foo");
		expect(r).toBe(resp);
	});

	it("does not redirect on non-401 error responses", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ status: 500, ok: false });
		const assignSpy = mockAssign();
		await apiFetch("/api/foo");
		expect(assignSpy).not.toHaveBeenCalled();
	});

	it("propagates network errors", async () => {
		vi.spyOn(global, "fetch").mockRejectedValue(new Error("network"));
		await expect(apiFetch("/api/foo")).rejects.toThrow("network");
	});
});
