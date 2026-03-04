// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useLocation } from "preact-iso";
import { useAuth } from "../../Auth.jsx";
import { TopBar } from "../ui/TopBar.jsx";
import {
	NavigationMenu,
	NavigationMenuBrand,
	NavigationMenuItem,
	NavigationMenuLink,
} from "../ui/NavigationMenu.jsx";
import { Avatar } from "../ui/Avatar.jsx";

export const NAV_ITEMS = [
	{ label: "Home", href: "/" },
	{ label: "Exercises", href: "/exercises" },
];

function initials(user) {
	if (user.name) {
		return user.name
			.split(" ")
			.map((p) => p[0])
			.join("")
			.toUpperCase()
			.slice(0, 2);
	}
	if (user.email) {
		return user.email[0].toUpperCase();
	}
	return "?";
}

export function Nav() {
	const { user } = useAuth();
	const { url } = useLocation();

	return (
		<TopBar>
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuBrand href="/">Cove</NavigationMenuBrand>
				</NavigationMenuItem>
				{NAV_ITEMS.map(({ label, href }) => (
					<NavigationMenuItem key={href}>
						<NavigationMenuLink href={href} active={url === href}>
							{label}
						</NavigationMenuLink>
					</NavigationMenuItem>
				))}
			</NavigationMenu>

			<NavigationMenu>
				<NavigationMenuItem>
					{user ? (
						<NavigationMenuLink
							href="/settings"
							active={url === "/settings"}
							aria-label="Account settings"
						>
							<Avatar initials={initials(user)} label={user.email} />
						</NavigationMenuLink>
					) : (
						<NavigationMenuLink href="/login" active={url === "/login"}>
							Sign in
						</NavigationMenuLink>
					)}
				</NavigationMenuItem>
			</NavigationMenu>
		</TopBar>
	);
}
