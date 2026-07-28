// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { Capacitor } from "@capacitor/core";
import { CapacitorUpdater } from "@capgo/capacitor-updater";

// On native platforms, manages the Capgo bundle state for the current build:
//
// - Non-dev builds: pin the Capgo channel and call notifyAppReady
//   fire-and-forget. These must not delay the initial render since this runs
//   every time the WebView/JS context restarts (background resume, OTA bundle
//   swap), not just on cold start. Failures are only logged since there is no
//   render path left to surface them.
//
// - Dev builds: autoUpdate is disabled in capacitor.config.js, so setChannel
//   is skipped (it requires server contact and would throw). Instead, reset
//   any stale non-builtin bundle to the locally built assets, blocking until
//   done so the builtin always wins before the first render.
export async function clearBundleIfNeeded() {
	if (!Capacitor.isNativePlatform()) return true;

	const channel = import.meta.env.VITE_COVE_ENV;
	if (!channel) throw new Error("VITE_COVE_ENV is not set");

	if (channel !== "dev") {
		CapacitorUpdater.notifyAppReady().catch((err) =>
			console.warn("notifyAppReady failed:", err),
		);
		CapacitorUpdater.setChannel({ channel }).catch((err) =>
			console.warn("setChannel failed:", err),
		);
		return true;
	}

	const { bundle } = await CapacitorUpdater.current();
	const { bundles } = await CapacitorUpdater.list();
	const remoteBundles = bundles.filter((b) => b.id !== "builtin");

	if (bundle.id !== "builtin") {
		console.warn(
			"[Dev] Remote OTA bundle active. Resetting to builtin assets...",
		);
		await CapacitorUpdater.reset({ toLastSuccessful: false });
		for (const b of remoteBundles) {
			await CapacitorUpdater.delete({ id: b.id }).catch((err) =>
				console.warn("[Dev] Could not delete bundle", b.id, err),
			);
		}
		window.location.reload();
		return false;
	}

	// Builtin is already active — clean up any stale stored bundles without reloading.
	for (const b of remoteBundles) {
		await CapacitorUpdater.delete({ id: b.id }).catch((err) =>
			console.warn("[Dev] Could not delete bundle", b.id, err),
		);
	}
	return true;
}
