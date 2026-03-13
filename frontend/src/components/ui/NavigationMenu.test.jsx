// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import {
	NavigationMenu,
	NavigationMenuBrand,
	NavigationMenuItem,
	NavigationMenuLink,
} from "./NavigationMenu.jsx";

describe("NavigationMenuLink", () => {
	it("renders a link with the given href", () => {
		render(
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuLink href="/settings">Settings</NavigationMenuLink>
				</NavigationMenuItem>
			</NavigationMenu>,
		);
		expect(screen.getByRole("link", { name: "Settings" })).toHaveAttribute(
			"href",
			"/settings",
		);
	});

	it("sets aria-current=page when active", () => {
		render(
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuLink href="/settings" active>
						Settings
					</NavigationMenuLink>
				</NavigationMenuItem>
			</NavigationMenu>,
		);
		expect(screen.getByRole("link")).toHaveAttribute("aria-current", "page");
	});

	it("omits aria-current when not active", () => {
		render(
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuLink href="/settings">Settings</NavigationMenuLink>
				</NavigationMenuItem>
			</NavigationMenu>,
		);
		expect(screen.getByRole("link")).not.toHaveAttribute("aria-current");
	});
});

describe("NavigationMenuBrand", () => {
	it("renders a link with the given href", () => {
		render(
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuBrand href="/">Cove</NavigationMenuBrand>
				</NavigationMenuItem>
			</NavigationMenu>,
		);
		expect(screen.getByRole("link", { name: "Cove" })).toHaveAttribute(
			"href",
			"/",
		);
	});
});
