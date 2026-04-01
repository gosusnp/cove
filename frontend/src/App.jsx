// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useEffect } from "preact/hooks";
import { LocationProvider, Route, Router, useLocation } from "preact-iso";
import { AuthProvider, useAuth } from "./Auth.jsx";
import { Nav } from "./components/shared/Nav.jsx";
import { AdminServiceAccounts } from "./pages/AdminServiceAccounts.jsx";
import { DesignElements } from "./pages/DesignElements.jsx";
import { Exercises } from "./pages/Exercises.jsx";
import { Home } from "./pages/Home.jsx";
import { Ingredients } from "./pages/Ingredients.jsx";
import { Recipes } from "./pages/Recipes.jsx";
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
	"/cook",
	"/admin",
];

function Layout() {
	const { url, route } = useLocation();
	const { user, loading } = useAuth();

	const isProtected = PROTECTED_ROUTES.some(
		(p) => url === p || url.startsWith(`${p}/`),
	);
	const isAdminRoute = url === "/admin" || url.startsWith("/admin/");

	useEffect(() => {
		if (loading) return;
		if (!user && isProtected) {
			route("/login");
		} else if (user && url === "/login") {
			route("/");
		} else if (user && isAdminRoute && !user.is_admin) {
			route("/");
		}
	}, [user, url, loading]);

	// Suppress render until auth is resolved to avoid a flash of wrong content.
	if (loading) return null;
	if (!user && isProtected) return null;
	if (user && url === "/login") return null;
	if (user && isAdminRoute && !user.is_admin) return null;

	return (
		<div class="app-shell flex flex-col min-h-dvh">
			<Nav />
			<Router>
				<Route path="/login" component={Login} />
				<Route path="/" component={Home} />
				<Route path="/cook/ingredients" component={Ingredients} />
				<Route path="/cook/ingredients/:id" component={Ingredients} />
				<Route path="/cook/recipes" component={Recipes} />
				<Route path="/cook/recipes/:id" component={Recipes} />
				<Route path="/exercises" component={Exercises} />
				<Route path="/exercises/:id" component={Exercises} />
				<Route path="/programs" component={Programs} />
				<Route path="/programs/:id" component={Programs} />
				<Route path="/sessions" component={Sessions} />
				<Route path="/sessions/:id" component={Sessions} />
				<Route path="/workout" component={SessionTracker} />
				<Route path="/settings" component={Settings} />
				<Route
					path="/admin/service-accounts"
					component={AdminServiceAccounts}
				/>
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
