// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { LocationProvider, Route, Router, useLocation } from "preact-iso";
import { AuthProvider, useAuth } from "./Auth.jsx";
import { Nav } from "./components/shared/Nav.jsx";
import { DesignElements } from "./pages/DesignElements.jsx";
import { Exercises } from "./pages/Exercises.jsx";
import { Home } from "./pages/Home.jsx";
import { Login } from "./pages/Login.jsx";
import { Programs } from "./pages/Programs.jsx";
import { Sessions } from "./pages/Sessions.jsx";
import { SessionTracker } from "./pages/SessionTracker.jsx";
import { Settings } from "./pages/Settings.jsx";

const PROTECTED_ROUTES = [
	"/settings",
	"/exercises",
	"/programs",
	"/sessions",
	"/workout",
];

function Layout() {
	const { url, route } = useLocation();
	const { user } = useAuth();

	const isProtected = PROTECTED_ROUTES.some(
		(p) => url === p || url.startsWith(`${p}/`),
	);

	useEffect(() => {
		if (!user && isProtected) {
			route("/login");
		} else if (user && url === "/login") {
			route("/");
		}
	}, [user, url]);

	// Suppress render until the redirect fires to avoid a flash of wrong content.
	if (!user && isProtected) return null;
	if (user && url === "/login") return null;

	return (
		<div class="app-shell flex flex-col min-h-dvh">
			<Nav />
			<Router>
				<Route path="/login" component={Login} />
				<Route path="/" component={Home} />
				<Route path="/exercises" component={Exercises} />
				<Route path="/programs" component={Programs} />
				<Route path="/programs/:id" component={Programs} />
				<Route path="/sessions" component={Sessions} />
				<Route path="/sessions/:id" component={Sessions} />
				<Route path="/workout" component={SessionTracker} />
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
