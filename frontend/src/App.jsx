// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { LocationProvider, Router, Route, useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import { AuthProvider, useAuth } from "./Auth.jsx";
import { Nav } from "./components/Nav.jsx";
import { Login } from "./pages/Login.jsx";
import { Home } from "./pages/Home.jsx";
import { Settings } from "./pages/Settings.jsx";
import { DesignElements } from "./pages/DesignElements.jsx";

function Layout() {
	const { url, route } = useLocation();
	const { user } = useAuth();

	useEffect(() => {
		if (!user && url === "/settings") {
			route("/login");
		} else if (user && url === "/login") {
			route("/");
		}
	}, [user, url]);

	// Suppress render until the redirect fires to avoid a flash of wrong content.
	if (!user && url === "/settings") return null;
	if (user && url === "/login") return null;

	return (
		<div class="app-shell flex flex-col min-h-dvh">
			<Nav />
			<Router>
				<Route path="/login" component={Login} />
				<Route path="/" component={Home} />
				<Route path="/settings" component={Settings} />
				{import.meta.env.VITE_COVE_ENV === "dev" && (
					<Route path="/design-elements" component={DesignElements} />
				)}
			</Router>
		</div>
	);
}

export function App() {
	return (
		<LocationProvider>
			<AuthProvider>
				<Layout />
			</AuthProvider>
		</LocationProvider>
	);
}
