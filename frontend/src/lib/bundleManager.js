// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { Capacitor } from "@capacitor/core";
import { CapacitorUpdater } from "@capgo/capacitor-updater";

// On native platforms, pins the Capgo channel to the build's own channel. On
// dev builds, this blocks so a stale remote OTA bundle can be reset to
// builtin assets before it ever flashes on screen. On other builds, channel
// pinning and notifyAppReady are fire-and-forget (chained, not awaited):
// they must not delay the initial render, since this runs every time the
// WebView/JS context restarts (background resume, OTA bundle swap), not
// just on cold start. Failures are only logged since there's no render path
// left to surface them to.
export async function clearBundleIfNeeded() {
	if (!Capacitor.isNativePlatform()) return true;

	const channel = import.meta.env.VITE_COVE_ENV;
	if (!channel) throw new Error("VITE_COVE_ENV is not set");

	if (channel !== "dev") {
		CapacitorUpdater.setChannel({ channel })
			.then(() => CapacitorUpdater.notifyAppReady())
			.catch((err) =>
				console.warn("Bundle manager background call failed:", err),
			);
		return true;
	}

	await CapacitorUpdater.setChannel({ channel });
	const { bundle } = await CapacitorUpdater.current();
	if (bundle.id !== "builtin") {
		console.warn(
			"[Dev] Remote OTA bundle detected. Resetting to builtin assets...",
		);
		await CapacitorUpdater.reset({ toLastSuccessful: false });
		window.location.reload();
		return false;
	}

	return true;
}
