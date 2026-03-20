// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useLocation } from "preact-iso";
import { useAuth } from "../../Auth.jsx";
import { cn } from "../../lib/utils.js";
import { Avatar } from "../ui/Avatar.jsx";
import {
	NavigationMenu,
	NavigationMenuBrand,
	NavigationMenuItem,
	NavigationMenuLink,
} from "../ui/NavigationMenu.jsx";
import { TopBar } from "../ui/TopBar.jsx";

const NAV_SECTIONS = [
	{
		label: "Train",
		items: [
			{ label: "Workout", href: "/workout" },
			{ label: "Review Sessions", href: "/sessions" },
		],
	},
	{
		label: "Program",
		items: [
			{ label: "Build Programs", href: "/programs" },
			{ label: "Exercises", href: "/exercises" },
		],
	},
];

export const TRAIN_NAV_ITEMS = NAV_SECTIONS[0].items;

const PAGE_TITLES = [
	{ label: "Workout", href: "/workout" },
	{ label: "Exercises", href: "/exercises" },
	{ label: "Build Programs", href: "/programs" },
	{ label: "Review Sessions", href: "/sessions" },
	{ label: "Settings", href: "/settings" },
];

function pageTitle(url) {
	const match = PAGE_TITLES.find(
		({ href }) => url === href || url.startsWith(`${href}/`),
	);
	return match ? match.label : null;
}

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

function isActive(href, url) {
	return href === "/"
		? url === "/"
		: url === href || url.startsWith(`${href}/`);
}

function SidebarLink({ href, active, children }) {
	return (
		<a
			href={href}
			aria-current={active ? "page" : undefined}
			class={cn(
				"flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors",
				"active:scale-95 touch-manipulation select-none",
				active
					? "bg-(--color-bg) text-(--color-accent)"
					: "text-(--color-muted) hover:text-(--color-text) hover:bg-(--color-bg)",
			)}
			style={{ textDecoration: "none" }}
		>
			{children}
		</a>
	);
}

function DesktopSidebar({ user, url }) {
	return (
		<aside
			class="hidden md:flex flex-col fixed top-0 left-0 bottom-0 z-40"
			style={{
				width: "var(--sidebar-w-desktop)",
				paddingTop: "var(--nav-h-desktop)",
				background: "var(--color-surface)",
				borderRight: "1px solid var(--color-border)",
			}}
		>
			<nav class="flex flex-col gap-1 px-2 py-4">
				{user &&
					NAV_SECTIONS.map(({ label, items }, i) => (
						<>
							<p
								class={`px-4 py-1 text-xs font-semibold uppercase tracking-wider${i > 0 ? " mt-2" : ""}`}
								style={{ color: "var(--color-text)" }}
							>
								{label}
							</p>
							{items.map(({ label: itemLabel, href }) => (
								<SidebarLink
									key={href}
									href={href}
									active={isActive(href, url)}
								>
									{itemLabel}
								</SidebarLink>
							))}
						</>
					))}
			</nav>
		</aside>
	);
}

function MobileBottomBar({ user, url }) {
	return (
		<nav
			class="md:hidden fixed bottom-0 inset-x-0 z-50 flex items-stretch justify-around"
			style={{
				height: "var(--nav-h-mobile)",
				background: "var(--color-surface)",
				borderTop: "1px solid var(--color-border)",
			}}
		>
			{user ? (
				<>
					<a
						href="/"
						aria-current={url === "/" ? "page" : undefined}
						aria-label="Home"
						class={cn(
							"flex flex-1 items-center justify-center text-base font-semibold tracking-tight transition-colors touch-manipulation select-none",
							url === "/" ? "text-(--color-accent)" : "text-(--color-muted)",
						)}
						style={{ textDecoration: "none" }}
					>
						Cove
					</a>
					<a
						href="/settings"
						aria-current={url === "/settings" ? "page" : undefined}
						aria-label="Account settings"
						class={cn(
							"flex flex-1 flex-col items-center justify-center gap-0.5 transition-colors touch-manipulation select-none",
							url === "/settings"
								? "text-(--color-accent)"
								: "text-(--color-muted)",
						)}
						style={{ textDecoration: "none" }}
					>
						<Avatar initials={initials(user)} label={user.email} />
					</a>
				</>
			) : (
				<a
					href="/login"
					aria-current={url === "/login" ? "page" : undefined}
					class={cn(
						"flex flex-1 items-center justify-center text-sm font-medium touch-manipulation select-none",
						url === "/login" ? "text-(--color-accent)" : "text-(--color-muted)",
					)}
					style={{ textDecoration: "none" }}
				>
					Sign in
				</a>
			)}
		</nav>
	);
}

export function Nav() {
	const { user } = useAuth();
	const { url } = useLocation();

	const title = pageTitle(url);

	return (
		<>
			<DesktopSidebar user={user} url={url} />
			<TopBar
				brand={
					<NavigationMenu>
						<NavigationMenuItem>
							<NavigationMenuBrand href="/">Cove</NavigationMenuBrand>
						</NavigationMenuItem>
					</NavigationMenu>
				}
			>
				<div class="flex flex-1 items-center px-6">
					{title && (
						<span
							class="text-sm font-semibold"
							style={{ color: "var(--color-text)" }}
						>
							{title}
						</span>
					)}
				</div>
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
			<MobileBottomBar user={user} url={url} />
		</>
	);
}
