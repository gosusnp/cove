// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

export const Capacitor = {
	isNativePlatform: () => false,
	getPlatform: () => "web",
};

export function registerPlugin(_name, fallbacks) {
	return fallbacks?.web ?? {};
}
