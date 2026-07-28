// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@capacitor/core", () => ({
	Capacitor: { isNativePlatform: vi.fn() },
}));

vi.mock("@capgo/capacitor-updater", () => ({
	CapacitorUpdater: {
		setChannel: vi.fn(),
		current: vi.fn(),
		list: vi.fn(),
		delete: vi.fn(),
		reset: vi.fn(),
		notifyAppReady: vi.fn(),
	},
}));

const { Capacitor } = await import("@capacitor/core");
const { CapacitorUpdater } = await import("@capgo/capacitor-updater");
const { clearBundleIfNeeded } = await import("./bundleManager.js");

describe("clearBundleIfNeeded", () => {
	let reloadMock;

	beforeEach(() => {
		reloadMock = vi.fn();
		vi.stubGlobal("location", { reload: reloadMock });
		CapacitorUpdater.setChannel.mockResolvedValue(undefined);
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "builtin" } });
		CapacitorUpdater.list.mockResolvedValue({ bundles: [] });
		CapacitorUpdater.delete.mockResolvedValue(undefined);
		CapacitorUpdater.reset.mockResolvedValue(undefined);
		CapacitorUpdater.notifyAppReady.mockResolvedValue(undefined);
	});

	afterEach(() => {
		vi.unstubAllEnvs();
		vi.unstubAllGlobals();
		vi.clearAllMocks();
	});

	it("throws if VITE_COVE_ENV is not set on a native platform", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "");

		await expect(clearBundleIfNeeded()).rejects.toThrow(
			"VITE_COVE_ENV is not set",
		);
		expect(CapacitorUpdater.setChannel).not.toHaveBeenCalled();
	});

	it("skips all Capgo calls on non-native platforms", async () => {
		Capacitor.isNativePlatform.mockReturnValue(false);
		vi.stubEnv("VITE_COVE_ENV", "prod");

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
		expect(CapacitorUpdater.setChannel).not.toHaveBeenCalled();
		expect(CapacitorUpdater.notifyAppReady).not.toHaveBeenCalled();
	});

	it("sets channel and notifies app ready on prod native build", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "prod");

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
		expect(CapacitorUpdater.setChannel).toHaveBeenCalledWith({
			channel: "prod",
		});
		expect(CapacitorUpdater.notifyAppReady).toHaveBeenCalled();
		expect(CapacitorUpdater.reset).not.toHaveBeenCalled();
	});

	it("does not block mounting on setChannel/notifyAppReady for a non-dev native build", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "prod");
		CapacitorUpdater.setChannel.mockReturnValue(new Promise(() => {}));
		CapacitorUpdater.notifyAppReady.mockReturnValue(new Promise(() => {}));

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
	});

	it("logs setChannel failure but still calls notifyAppReady on a non-dev build", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "prod");
		const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
		const error = new Error("bridge unavailable");
		CapacitorUpdater.setChannel.mockRejectedValue(error);

		const result = await clearBundleIfNeeded();
		await new Promise((resolve) => setTimeout(resolve, 0));

		expect(result).toBe(true);
		expect(CapacitorUpdater.notifyAppReady).toHaveBeenCalled();
		expect(warnSpy).toHaveBeenCalledWith("setChannel failed:", error);
		warnSpy.mockRestore();
	});

	it("mounts normally on dev native build with builtin bundle and no stored bundles", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "dev");
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "builtin" } });
		CapacitorUpdater.list.mockResolvedValue({ bundles: [] });

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
		expect(CapacitorUpdater.setChannel).not.toHaveBeenCalled();
		expect(CapacitorUpdater.notifyAppReady).not.toHaveBeenCalled();
		expect(CapacitorUpdater.delete).not.toHaveBeenCalled();
		expect(CapacitorUpdater.reset).not.toHaveBeenCalled();
	});

	it("cleans up stale stored bundles without reloading when current is builtin", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "dev");
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "builtin" } });
		CapacitorUpdater.list.mockResolvedValue({
			bundles: [{ id: "v1.0.0" }, { id: "v1.1.0" }],
		});

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
		expect(CapacitorUpdater.delete).toHaveBeenCalledWith({ id: "v1.0.0" });
		expect(CapacitorUpdater.delete).toHaveBeenCalledWith({ id: "v1.1.0" });
		expect(CapacitorUpdater.reset).not.toHaveBeenCalled();
		expect(reloadMock).not.toHaveBeenCalled();
	});

	it("deletes stored bundles and resets on dev native build with non-builtin active bundle", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "dev");
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "v1.2.3" } });
		CapacitorUpdater.list.mockResolvedValue({
			bundles: [{ id: "v1.2.3" }],
		});

		const result = await clearBundleIfNeeded();

		expect(result).toBe(false);
		expect(CapacitorUpdater.setChannel).not.toHaveBeenCalled();
		expect(CapacitorUpdater.delete).toHaveBeenCalledWith({ id: "v1.2.3" });
		expect(CapacitorUpdater.reset).toHaveBeenCalledWith({
			toLastSuccessful: false,
		});
		expect(reloadMock).toHaveBeenCalled();
		expect(CapacitorUpdater.notifyAppReady).not.toHaveBeenCalled();
	});
});
