// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Dev-only component showcase. Accessible at /design-elements when VITE_COVE_ENV=dev is set.

import { useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import { useSignal } from "@preact/signals";
import { Button } from "../components/ui/Button.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
	DialogTrigger,
} from "../components/ui/Dialog.jsx";
import { Switch } from "../components/ui/Switch.jsx";
import { useDialog } from "../hooks/useDialog.js";
import {
	NavigationMenu,
	NavigationMenuBrand,
	NavigationMenuItem,
	NavigationMenuLink,
} from "../components/ui/NavigationMenu.jsx";
import { NAV_ITEMS } from "../components/Nav.jsx";
import { Avatar } from "../components/ui/Avatar.jsx";

// ── Nav preview helpers ───────────────────────────────────────────────────────

function NavPreview({ user, activeHref }) {
	return (
		<div
			class="flex items-stretch justify-between px-6 rounded-xl overflow-hidden"
			style={{
				height: "var(--nav-h-desktop)",
				background: "var(--color-surface)",
				border: "1px solid var(--color-border)",
			}}
		>
			<NavigationMenu>
				<NavigationMenuItem>
					<NavigationMenuBrand href="/">Cove</NavigationMenuBrand>
				</NavigationMenuItem>
				{NAV_ITEMS.map(({ label, href }) => (
					<NavigationMenuItem key={href}>
						<NavigationMenuLink href={href} active={href === activeHref}>
							{label}
						</NavigationMenuLink>
					</NavigationMenuItem>
				))}
			</NavigationMenu>

			<NavigationMenu>
				<NavigationMenuItem>
					{user ? (
						<NavigationMenuLink href="/settings">
							<Avatar initials={user.initials} label={user.email} />
						</NavigationMenuLink>
					) : (
						<NavigationMenuLink href="/login">Sign in</NavigationMenuLink>
					)}
				</NavigationMenuItem>
			</NavigationMenu>
		</div>
	);
}

function Section({ title, children }) {
	return (
		<section class="flex flex-col gap-4">
			<h2
				class="text-xs font-semibold uppercase tracking-widest"
				style={{ color: "var(--color-muted)" }}
			>
				{title}
			</h2>
			{children}
		</section>
	);
}

function Row({ label, children }) {
	return (
		<div class="flex flex-col gap-2">
			<span class="text-xs" style={{ color: "var(--color-muted)" }}>
				{label}
			</span>
			<div class="flex flex-wrap items-center gap-3">{children}</div>
		</div>
	);
}

function Divider() {
	return <hr style={{ borderColor: "var(--color-border)" }} />;
}

export function DesignElements() {
	const { route } = useLocation();

	useEffect(() => {
		if (import.meta.env.VITE_COVE_ENV !== "dev") {
			route("/");
		}
	}, []);

	if (import.meta.env.VITE_COVE_ENV !== "dev") return null;

	const dialog = useDialog();
	const switchA = useSignal(false);
	const switchB = useSignal(true);

	return (
		<main class="max-w-2xl mx-auto px-4 py-10 flex flex-col gap-10">
			<div>
				<h1 class="text-2xl font-bold">Design Elements</h1>
				<p class="text-sm mt-1" style={{ color: "var(--color-muted)" }}>
					Only visible when{" "}
					<code class="font-mono bg-(--color-bg) px-1 rounded">
						VITE_COVE_ENV=dev
					</code>{" "}
					is set.
				</p>
			</div>

			<Divider />

			{/* ── Top Navigation ─────────────────────────────── */}
			<Section title="Top Navigation">
				<Row label="signed out">
					<div class="w-full">
						<NavPreview user={null} activeHref="/" />
					</div>
				</Row>
				<Row label="signed in — Home active">
					<div class="w-full">
						<NavPreview
							user={{ initials: "JM", email: "jimmy@example.com" }}
							activeHref="/"
						/>
					</div>
				</Row>
				<Row label="signed in — Exercises active">
					<div class="w-full">
						<NavPreview
							user={{ initials: "JM", email: "jimmy@example.com" }}
							activeHref="/exercises"
						/>
					</div>
				</Row>
			</Section>

			<Divider />

			{/* ── Button ─────────────────────────────────────── */}
			<Section title="Button">
				<Row label="variant=primary">
					<Button size="sm">Small</Button>
					<Button size="md">Medium</Button>
					<Button size="lg">Large</Button>
				</Row>
				<Row label="variant=outline">
					<Button variant="outline" size="sm">
						Small
					</Button>
					<Button variant="outline" size="md">
						Medium
					</Button>
					<Button variant="outline" size="lg">
						Large
					</Button>
				</Row>
				<Row label="variant=ghost">
					<Button variant="ghost" size="sm">
						Small
					</Button>
					<Button variant="ghost" size="md">
						Medium
					</Button>
					<Button variant="ghost" size="lg">
						Large
					</Button>
				</Row>
				<Row label="disabled">
					<Button disabled>Primary</Button>
					<Button variant="outline" disabled>
						Outline
					</Button>
					<Button variant="ghost" disabled>
						Ghost
					</Button>
				</Row>
			</Section>

			<Divider />

			{/* ── Dialog ─────────────────────────────────────── */}
			<Section title="Dialog">
				<Row label="signal-controlled via useDialog()">
					<Dialog openSignal={dialog.open}>
						<DialogTrigger>
							<Button variant="outline">Open Dialog</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogTitle>Example Dialog</DialogTitle>
							<DialogDescription>
								This dialog's open state is managed by a Preact signal via{" "}
								<code class="font-mono">useDialog()</code>.
							</DialogDescription>
							<div class="mt-6 flex justify-end gap-2">
								<DialogClose>
									<Button variant="ghost" size="sm">
										Cancel
									</Button>
								</DialogClose>
								<DialogClose>
									<Button size="sm">Confirm</Button>
								</DialogClose>
							</div>
						</DialogContent>
					</Dialog>
					<Button variant="ghost" size="sm" onClick={dialog.show}>
						Open (imperative)
					</Button>
				</Row>
			</Section>

			<Divider />

			{/* ── Switch ─────────────────────────────────────── */}
			<Section title="Switch">
				<Row label="uncontrolled (internal signal)">
					<Switch />
				</Row>
				<Row label="controlled — off">
					<Switch checkedSignal={switchA} />
					<span class="text-sm" style={{ color: "var(--color-muted)" }}>
						value: {String(switchA.value)}
					</span>
				</Row>
				<Row label="controlled — on">
					<Switch checkedSignal={switchB} />
					<span class="text-sm" style={{ color: "var(--color-muted)" }}>
						value: {String(switchB.value)}
					</span>
				</Row>
				<Row label="disabled">
					<Switch disabled />
					<Switch checkedSignal={switchB} disabled />
				</Row>
			</Section>

			<Divider />

			{/* ── Avatar ──────────────────────────────────────── */}
			<Section title="Avatar">
				<Row label="initials">
					<Avatar initials="JM" />
					<Avatar initials="A" />
				</Row>
				<Row label="with aria-label">
					<Avatar initials="JM" label="jimmy@example.com" />
				</Row>
			</Section>
		</main>
	);
}
