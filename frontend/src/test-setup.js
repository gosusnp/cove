/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/preact";
import { afterEach } from "vitest";

afterEach(cleanup);

// Radix UI uses ResizeObserver internally; jsdom does not implement it.
global.ResizeObserver = class ResizeObserver {
	observe() {}
	unobserve() {}
	disconnect() {}
};
