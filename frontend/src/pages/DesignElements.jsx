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
import { Avatar } from "../components/ui/Avatar.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import {
	Section as CardSection,
	Row,
	Divider,
} from "../components/ui/Section.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";

const PREVIEW_NAV_ITEMS = [
	{ label: "Home", href: "/" },
	{ label: "Exercises", href: "/exercises" },
];

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
				{PREVIEW_NAV_ITEMS.map(({ label, href }) => (
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

function PageSection({ title, children }) {
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

function PreviewRow({ label, children }) {
	return (
		<div class="flex flex-col gap-2">
			<span class="text-xs" style={{ color: "var(--color-muted)" }}>
				{label}
			</span>
			<div class="flex flex-wrap items-center gap-3">{children}</div>
		</div>
	);
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
			<PageSection title="Top Navigation">
				<PreviewRow label="signed out">
					<div class="w-full">
						<NavPreview user={null} activeHref="/" />
					</div>
				</PreviewRow>
				<PreviewRow label="signed in — Home active">
					<div class="w-full">
						<NavPreview
							user={{ initials: "JM", email: "jimmy@example.com" }}
							activeHref="/"
						/>
					</div>
				</PreviewRow>
				<PreviewRow label="signed in — Exercises active">
					<div class="w-full">
						<NavPreview
							user={{ initials: "JM", email: "jimmy@example.com" }}
							activeHref="/exercises"
						/>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Button ─────────────────────────────────────── */}
			<PageSection title="Button">
				<PreviewRow label="variant=primary">
					<Button size="sm">Small</Button>
					<Button size="md">Medium</Button>
					<Button size="lg">Large</Button>
				</PreviewRow>
				<PreviewRow label="variant=outline">
					<Button variant="outline" size="sm">
						Small
					</Button>
					<Button variant="outline" size="md">
						Medium
					</Button>
					<Button variant="outline" size="lg">
						Large
					</Button>
				</PreviewRow>
				<PreviewRow label="variant=ghost">
					<Button variant="ghost" size="sm">
						Small
					</Button>
					<Button variant="ghost" size="md">
						Medium
					</Button>
					<Button variant="ghost" size="lg">
						Large
					</Button>
				</PreviewRow>
				<PreviewRow label="variant=destructive">
					<Button variant="destructive" size="sm">
						Small
					</Button>
					<Button variant="destructive" size="md">
						Medium
					</Button>
				</PreviewRow>
				<PreviewRow label="disabled">
					<Button disabled>Primary</Button>
					<Button variant="outline" disabled>
						Outline
					</Button>
					<Button variant="ghost" disabled>
						Ghost
					</Button>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Dialog ─────────────────────────────────────── */}
			<PageSection title="Dialog">
				<PreviewRow label="signal-controlled via useDialog()">
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
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Switch ─────────────────────────────────────── */}
			<PageSection title="Switch">
				<PreviewRow label="uncontrolled (internal signal)">
					<Switch />
				</PreviewRow>
				<PreviewRow label="controlled — off">
					<Switch checkedSignal={switchA} />
					<span class="text-sm" style={{ color: "var(--color-muted)" }}>
						value: {String(switchA.value)}
					</span>
				</PreviewRow>
				<PreviewRow label="controlled — on">
					<Switch checkedSignal={switchB} />
					<span class="text-sm" style={{ color: "var(--color-muted)" }}>
						value: {String(switchB.value)}
					</span>
				</PreviewRow>
				<PreviewRow label="disabled">
					<Switch disabled />
					<Switch checkedSignal={switchB} disabled />
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Avatar ──────────────────────────────────────── */}
			<PageSection title="Avatar">
				<PreviewRow label="initials">
					<Avatar initials="JM" />
					<Avatar initials="A" />
				</PreviewRow>
				<PreviewRow label="with aria-label">
					<Avatar initials="JM" label="jimmy@example.com" />
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── TextField ───────────────────────────────────── */}
			<PageSection title="TextField">
				<PreviewRow label="default">
					<TextField id="ex-default" placeholder="Placeholder text" />
				</PreviewRow>
				<PreviewRow label="with label">
					<TextField
						id="ex-label"
						label="Token name"
						placeholder="e.g. CI pipeline"
					/>
				</PreviewRow>
				<PreviewRow label="disabled">
					<TextField
						id="ex-disabled"
						label="Disabled"
						placeholder="Placeholder"
						disabled
					/>
				</PreviewRow>
				<PreviewRow label="read-only">
					<TextField
						id="ex-readonly"
						label="Token (read-only)"
						value="pat_a1b2c3d4e5f6..."
						readOnly
					/>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Tooltip ─────────────────────────────────────── */}
			<PageSection title="Tooltip">
				<PreviewRow label="default">
					<Tooltip>
						<TooltipTrigger>
							<Button variant="outline" size="sm">
								Hover me
							</Button>
						</TooltipTrigger>
						<TooltipContent>This is a tooltip</TooltipContent>
					</Tooltip>
				</PreviewRow>
				<PreviewRow label="copy pattern">
					<Tooltip>
						<TooltipTrigger>
							<Button variant="ghost" size="sm">
								Copy
							</Button>
						</TooltipTrigger>
						<TooltipContent>Copied!</TooltipContent>
					</Tooltip>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Section / Row ───────────────────────────────── */}
			<PageSection title="Section + Row">
				<PreviewRow label="card section with rows">
					<div class="w-full">
						<CardSection title="Example section">
							<Row label="First item">Value</Row>
							<Row label="Second item">Another value</Row>
							<Row label="Last item" last>
								<Button size="sm" variant="outline">
									Action
								</Button>
							</Row>
						</CardSection>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── PageTitle ─────────────────────────────────── */}
			<PageSection title="PageTitle">
				<PreviewRow label="default (text-2xl)">
					<PageTitle>Settings</PageTitle>
				</PreviewRow>
				<PreviewRow label="hero (text-6xl)">
					<PageTitle class="text-6xl tracking-tight">Cove</PageTitle>
				</PreviewRow>
			</PageSection>
		</main>
	);
}
