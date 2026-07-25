// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { Capacitor } from "@capacitor/core";
import { CapacitorUpdater } from "@capgo/capacitor-updater";

// On native platforms, pins the Capgo channel to the build's own channel. On
// dev builds, clears any downloaded OTA bundle and reloads to enforce builtin
// assets. On other builds, confirms a successful launch via notifyAppReady.
// Must run before the app is mounted so that a reset and reload can happen
// cleanly without any UI flashing.
export async function clearBundleIfNeeded() {
	if (!Capacitor.isNativePlatform()) return true;

	const channel = import.meta.env.VITE_COVE_ENV;
	if (!channel) throw new Error("VITE_COVE_ENV is not set");
	await CapacitorUpdater.setChannel({ channel });

	if (channel === "dev") {
		const { bundle } = await CapacitorUpdater.current();
		if (bundle.id !== "builtin") {
			console.warn(
				"[Dev] Remote OTA bundle detected. Resetting to builtin assets...",
			);
			await CapacitorUpdater.reset({ toLastSuccessful: false });
			window.location.reload();
			return false;
		}
	} else {
		await CapacitorUpdater.notifyAppReady();
	}

	return true;
}
