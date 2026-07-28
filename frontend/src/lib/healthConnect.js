// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { Capacitor, registerPlugin } from "@capacitor/core";

const HealthConnect = registerPlugin("HealthConnect", {
	web: {
		isAvailable: async () => ({ available: false }),
		requestPermission: async () => ({ granted: false }),
		openSettings: async () => {},
	},
});

export async function isHealthConnectAvailable() {
	if (Capacitor.getPlatform() !== "android") return false;
	const { available } = await HealthConnect.isAvailable();
	return available;
}

export async function requestHealthConnectPermission() {
	if (Capacitor.getPlatform() !== "android") return false;
	const { granted } = await HealthConnect.requestPermission();
	return granted;
}

export async function openHealthConnectSettings() {
	if (Capacitor.getPlatform() !== "android") return;
	await HealthConnect.openSettings();
}
