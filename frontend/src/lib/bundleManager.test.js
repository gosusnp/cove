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

	it("sets channel and mounts normally on dev native build with builtin bundle", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "dev");
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "builtin" } });

		const result = await clearBundleIfNeeded();

		expect(result).toBe(true);
		expect(CapacitorUpdater.setChannel).toHaveBeenCalledWith({
			channel: "dev",
		});
		expect(CapacitorUpdater.notifyAppReady).not.toHaveBeenCalled();
		expect(CapacitorUpdater.reset).not.toHaveBeenCalled();
	});

	it("resets and reloads on dev native build with non-builtin bundle", async () => {
		Capacitor.isNativePlatform.mockReturnValue(true);
		vi.stubEnv("VITE_COVE_ENV", "dev");
		CapacitorUpdater.current.mockResolvedValue({ bundle: { id: "v1.2.3" } });

		const result = await clearBundleIfNeeded();

		expect(result).toBe(false);
		expect(CapacitorUpdater.setChannel).toHaveBeenCalledWith({
			channel: "dev",
		});
		expect(CapacitorUpdater.reset).toHaveBeenCalledWith({
			toLastSuccessful: false,
		});
		expect(reloadMock).toHaveBeenCalled();
		expect(CapacitorUpdater.notifyAppReady).not.toHaveBeenCalled();
	});
});
