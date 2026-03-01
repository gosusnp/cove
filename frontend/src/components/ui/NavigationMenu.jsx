// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import * as RadixNav from "@radix-ui/react-navigation-menu";
import { cn } from "../../lib/utils";

export function NavigationMenu({ class: className, children, ...props }) {
	return (
		<RadixNav.Root
			className={cn("relative flex h-full items-stretch", className)}
			{...props}
		>
			<RadixNav.List className="flex items-stretch list-none h-full">
				{children}
			</RadixNav.List>
		</RadixNav.Root>
	);
}

export function NavigationMenuItem({ children }) {
	return (
		<RadixNav.Item className="flex items-stretch">{children}</RadixNav.Item>
	);
}

export function NavigationMenuBrand({ href, children }) {
	return (
		<RadixNav.Link asChild>
			<a
				href={href}
				class="flex h-full items-center px-4 text-base font-semibold tracking-tight"
				style={{ color: "var(--color-text)", textDecoration: "none" }}
			>
				{children}
			</a>
		</RadixNav.Link>
	);
}

export function NavigationMenuLink({
	href,
	active,
	class: className,
	children,
	...props
}) {
	return (
		<RadixNav.Link active={active} asChild>
			<a
				href={href}
				aria-current={active ? "page" : undefined}
				class={cn(
					"flex h-full items-center px-4 text-sm font-medium transition-colors",
					"active:scale-95 touch-manipulation select-none",
					active
						? "bg-(--color-bg) text-(--color-accent)"
						: "text-(--color-muted) hover:text-(--color-text) hover:bg-(--color-bg)",
					className,
				)}
				style={{ textDecoration: "none" }}
				{...props}
			>
				{children}
			</a>
		</RadixNav.Link>
	);
}
