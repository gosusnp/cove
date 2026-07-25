/**
 * Copyright (c) 2026 Jimmy Ma
 * SPDX-License-Identifier: Elastic-2.0
 */

import { render } from "preact";
import "./app.css";
import { App } from "./App.jsx";
import { clearBundleIfNeeded } from "./lib/bundleManager.js";

(async () => {
	const shouldMount = await clearBundleIfNeeded();
	if (!shouldMount) return;
	render(<App />, document.getElementById("app"));
})();
