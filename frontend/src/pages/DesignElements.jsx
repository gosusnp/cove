// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Dev-only component showcase. Accessible at /design-elements when VITE_COVE_ENV=dev is set.

import {
	closestCenter,
	DndContext,
	DragOverlay,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import {
	SortableContext,
	useSortable,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { useLocation } from "preact-iso";
import {
	Accordion,
	AccordionContent,
	AccordionDragHandle,
	AccordionItem,
	AccordionTrigger,
} from "../components/ui/Accordion.jsx";
import {
	CheckList,
	CheckListItem,
	CheckListSection,
} from "../components/ui/CheckList.jsx";
import { Avatar } from "../components/ui/Avatar.jsx";
import { Button } from "../components/ui/Button.jsx";
import { Combobox } from "../components/ui/Combobox.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import { EditableMarkdown } from "../components/ui/EditableMarkdown.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
	DialogTrigger,
} from "../components/ui/Dialog.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { ListItem } from "../components/ui/ListItem.jsx";
import {
	NavigationMenu,
	NavigationMenuBrand,
	NavigationMenuItem,
	NavigationMenuLink,
} from "../components/ui/NavigationMenu.jsx";
import { DateTimePicker } from "../components/ui/DateTimePicker.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import {
	Section as CardSection,
	Divider,
	Row,
} from "../components/ui/Section.jsx";
import { Switch } from "../components/ui/Switch.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { ToggleGroup } from "../components/ui/ToggleGroup.jsx";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "../components/ui/Tooltip.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { useSortableGroups } from "../hooks/useSortableGroups.js";
import { cn } from "../lib/utils.js";

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

// ── Accordion demo — exercise sets ────────────────────────────────────────────

const INITIAL_SETS = [
	{
		id: "set-1",
		name: "Lower body",
		exercises: [
			{ id: "ex-1", name: "Back squat", sets: 4, reps: 5 },
			{ id: "ex-2", name: "Romanian deadlift", sets: 3, reps: 8 },
			{ id: "ex-3", name: "Leg press", sets: 3, reps: 12 },
		],
	},
	{
		id: "set-2",
		name: "Upper body push",
		exercises: [
			{ id: "ex-4", name: "Bench press", sets: 4, reps: 6 },
			{ id: "ex-5", name: "Overhead press", sets: 3, reps: 8 },
		],
	},
	{
		id: "set-3",
		name: "Upper body pull",
		exercises: [
			{ id: "ex-6", name: "Pull-ups", sets: 4, reps: 8 },
			{ id: "ex-7", name: "Barbell row", sets: 3, reps: 6 },
		],
	},
];

// A sortable exercise row used inside an accordion set.
function SortableExercise({ id, name, sets, reps, setId }) {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id, data: { type: "exercise", setId } });

	return (
		<div
			ref={setNodeRef}
			style={{ transform: CSS.Transform.toString(transform), transition }}
			class={cn(
				"flex items-center gap-3 py-2 px-1 rounded-lg",
				"border border-transparent",
				isDragging
					? "opacity-50 border-(--color-border) bg-(--color-bg)"
					: "hover:bg-(--color-bg)",
			)}
		>
			<button
				type="button"
				class="flex items-center justify-center w-5 h-5 shrink-0 cursor-grab active:cursor-grabbing text-(--color-muted) hover:text-(--color-text)"
				aria-label="Drag to reorder"
				{...listeners}
				{...attributes}
			>
				<GripIcon />
			</button>
			<span class="flex-1 text-sm text-(--color-text)">{name}</span>
			<span
				class="text-xs tabular-nums"
				style={{ color: "var(--color-muted)" }}
			>
				{sets}×{reps}
			</span>
		</div>
	);
}

function GripIcon() {
	return (
		<svg
			width="12"
			height="12"
			viewBox="0 0 14 14"
			fill="none"
			aria-hidden="true"
		>
			<circle cx="5" cy="3" r="1" fill="currentColor" />
			<circle cx="9" cy="3" r="1" fill="currentColor" />
			<circle cx="5" cy="7" r="1" fill="currentColor" />
			<circle cx="9" cy="7" r="1" fill="currentColor" />
			<circle cx="5" cy="11" r="1" fill="currentColor" />
			<circle cx="9" cy="11" r="1" fill="currentColor" />
		</svg>
	);
}

function ExerciseSetsDemo() {
	const {
		sets,
		openValues,
		setOpenValues,
		activeId,
		handleDragStart,
		handleDragOver,
		handleDragEnd,
	} = useSortableGroups(INITIAL_SETS);

	const sensors = useSensors(
		useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
		useSensor(KeyboardSensor),
	);

	const activeExercise = activeId
		? sets.flatMap((s) => s.exercises).find((e) => e.id === activeId)
		: null;
	const activeSet = activeId ? sets.find((s) => s.id === activeId) : null;

	return (
		<DndContext
			sensors={sensors}
			collisionDetection={closestCenter}
			onDragStart={handleDragStart}
			onDragOver={handleDragOver}
			onDragEnd={handleDragEnd}
		>
			<SortableContext
				items={sets.map((s) => s.id)}
				strategy={verticalListSortingStrategy}
			>
				<Accordion
					type="multiple"
					value={openValues}
					onValueChange={setOpenValues}
				>
					{sets.map((set) => (
						<AccordionItem key={set.id} id={set.id} value={set.id}>
							<AccordionTrigger>
								<AccordionDragHandle />
								<span>{set.name}</span>
								<span
									class="ml-auto mr-2 text-xs tabular-nums"
									style={{ color: "var(--color-muted)" }}
								>
									{set.exercises.length} exercise
									{set.exercises.length !== 1 ? "s" : ""}
								</span>
							</AccordionTrigger>
							<AccordionContent>
								<SortableContext
									items={set.exercises.map((e) => e.id)}
									strategy={verticalListSortingStrategy}
								>
									<div class="flex flex-col">
										{set.exercises.map((ex) => (
											<SortableExercise
												key={ex.id}
												id={ex.id}
												name={ex.name}
												sets={ex.sets}
												reps={ex.reps}
												setId={set.id}
											/>
										))}
									</div>
								</SortableContext>
								{set.exercises.length === 0 && (
									<p
										class="text-xs py-2 text-center"
										style={{ color: "var(--color-muted)" }}
									>
										Drop exercises here
									</p>
								)}
							</AccordionContent>
						</AccordionItem>
					))}
				</Accordion>
			</SortableContext>

			<DragOverlay>
				{activeExercise && (
					<div class="flex items-center gap-3 py-2 px-3 rounded-lg shadow-lg bg-(--color-surface) border border-(--color-border) text-sm text-(--color-text)">
						<GripIcon />
						<span>{activeExercise.name}</span>
						<span
							class="ml-auto text-xs tabular-nums"
							style={{ color: "var(--color-muted)" }}
						>
							{activeExercise.sets}×{activeExercise.reps}
						</span>
					</div>
				)}
				{activeSet && (
					<div class="rounded-lg px-4 py-3 shadow-lg bg-(--color-surface) border border-(--color-border) text-sm font-medium text-(--color-text)">
						{activeSet.name}
					</div>
				)}
			</DragOverlay>
		</DndContext>
	);
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function DesignElements() {
	const { route } = useLocation();

	useEffect(() => {
		if (import.meta.env.VITE_COVE_ENV !== "dev") {
			route("/");
		}
	}, []);

	if (import.meta.env.VITE_COVE_ENV !== "dev") return null;

	const dialog = useDialog();
	const confirmDialog = useDialog();
	const confirmResult = useSignal(null);
	const switchA = useSignal(false);
	const switchB = useSignal(true);
	const toggleVal = useSignal("bilateral");
	const toggleNullable = useSignal("unilateral");
	const toggleNoLabel = useSignal(null);
	const comboVal = useSignal("");
	const comboVal2 = useSignal("2");
	const listDetailHasDetail = useSignal(false);

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
				<PreviewRow label="multiline">
					<TextField
						id="ex-multiline"
						label="Notes"
						multiline
						placeholder="Enter notes…"
						rows={4}
					/>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── DateTimePicker ──────────────────────────────── */}
			<PageSection title="DateTimePicker">
				<PreviewRow label="default">
					<DateTimePicker id="dt-default" />
				</PreviewRow>
				<PreviewRow label="with label">
					<DateTimePicker id="dt-label" label="Start date" />
				</PreviewRow>
				<PreviewRow label="disabled">
					<DateTimePicker
						id="dt-disabled"
						label="Disabled"
						value="2026-03-24T12:00"
						disabled
					/>
				</PreviewRow>
				<PreviewRow label="read-only">
					<DateTimePicker
						id="dt-readonly"
						label="Started at (read-only)"
						value="2026-03-24T09:30"
						readOnly
					/>
				</PreviewRow>
				<PreviewRow label="inline">
					<div class="p-4 bg-(--color-surface) border border-(--color-border) rounded-lg">
						<DateTimePicker
							id="dt-inline"
							label="Inline style"
							value="2026-03-24T15:45"
							inline
						/>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />
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

			<Divider />

			{/* ── ToggleGroup ───────────────────────────────── */}
			<PageSection title="ToggleGroup">
				<PreviewRow label="with label, pre-selected">
					<ToggleGroup
						label="Laterality"
						value={toggleVal.value}
						onChange={(v) => (toggleVal.value = v)}
						options={[
							{ value: "bilateral", label: "Bilateral" },
							{ value: "unilateral", label: "Unilateral" },
							{ value: "left", label: "Left" },
							{ value: "right", label: "Right" },
							{ value: "alternating", label: "Alternating" },
						]}
					/>
				</PreviewRow>
				<PreviewRow label="without label">
					<ToggleGroup
						value={toggleNoLabel.value}
						onChange={(v) => (toggleNoLabel.value = v)}
						options={[
							{ value: "easy", label: "Easy" },
							{ value: "moderate", label: "Moderate" },
							{ value: "hard", label: "Hard" },
						]}
					/>
				</PreviewRow>
				<PreviewRow label="nullable — click active to deselect">
					<ToggleGroup
						label="Effort"
						value={toggleNullable.value}
						onChange={(v) => (toggleNullable.value = v)}
						options={[
							{ value: "low", label: "Low" },
							{ value: "moderate", label: "Moderate" },
							{ value: "high", label: "High" },
						]}
						nullable
					/>
				</PreviewRow>
				<PreviewRow label="disabled">
					<ToggleGroup
						label="Disabled"
						value="bilateral"
						onChange={() => {}}
						options={[
							{ value: "bilateral", label: "Bilateral" },
							{ value: "unilateral", label: "Unilateral" },
						]}
						disabled
					/>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Accordion ─────────────────────────────────── */}
			<PageSection title="Accordion">
				<PreviewRow label="basic — static items">
					<div class="w-full">
						<Accordion type="single">
							<AccordionItem value="q1">
								<AccordionTrigger>What is Cove?</AccordionTrigger>
								<AccordionContent>
									Cove is a personal fitness tracking app with an AI-assisted
									workout planning interface via MCP.
								</AccordionContent>
							</AccordionItem>
							<AccordionItem value="q2">
								<AccordionTrigger>
									How are workouts structured?
								</AccordionTrigger>
								<AccordionContent>
									Workouts are composed of sets, each containing one or more
									exercises with prescribed reps and load.
								</AccordionContent>
							</AccordionItem>
							<AccordionItem value="q3">
								<AccordionTrigger>
									Can I track progress over time?
								</AccordionTrigger>
								<AccordionContent>
									Yes — every completed workout is stored and surfaced through
									the dashboard and MCP tools.
								</AccordionContent>
							</AccordionItem>
						</Accordion>
					</div>
				</PreviewRow>

				<PreviewRow label="sortable — exercise sets with drag &amp; drop">
					<div class="w-full">
						<p class="text-xs mb-3" style={{ color: "var(--color-muted)" }}>
							Drag the grip handle on a set header to reorder sets. Drag the
							grip on an exercise to reorder within or move between sets.
						</p>
						<ExerciseSetsDemo />
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── Combobox ──────────────────────────────────── */}
			<PageSection title="Combobox">
				<PreviewRow label="default — empty">
					<div class="w-64">
						<Combobox
							options={[
								{ value: "1", label: "Squat" },
								{ value: "2", label: "Bench Press" },
								{ value: "3", label: "Deadlift" },
								{ value: "4", label: "Romanian Deadlift" },
								{ value: "5", label: "Overhead Press" },
							]}
							value={comboVal.value}
							onChange={(v) => (comboVal.value = v)}
							placeholder="Search exercises..."
						/>
					</div>
				</PreviewRow>
				<PreviewRow label="with label + pre-selected value">
					<div class="w-64">
						<Combobox
							label="Exercise"
							options={[
								{ value: "1", label: "Squat" },
								{ value: "2", label: "Bench Press" },
								{ value: "3", label: "Deadlift" },
							]}
							value={comboVal2.value}
							onChange={(v) => (comboVal2.value = v)}
							placeholder="Search exercises..."
						/>
					</div>
				</PreviewRow>
				<PreviewRow label="disabled">
					<div class="w-64">
						<Combobox
							label="Exercise"
							options={[
								{ value: "1", label: "Squat" },
								{ value: "2", label: "Bench Press" },
							]}
							value="1"
							onChange={() => {}}
							placeholder="Search exercises..."
							disabled
						/>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── ListItem ──────────────────────────────────── */}
			<PageSection title="ListItem">
				<PreviewRow label="default — label only">
					<div
						class="w-full rounded-xl overflow-hidden"
						style={{
							background: "var(--color-surface)",
							border: "1px solid var(--color-border)",
						}}
					>
						<ListItem label="Strength A" onClick={() => {}} />
						<ListItem label="Hypertrophy B" onClick={() => {}} />
						<ListItem label="Deload Week" isLast onClick={() => {}} />
					</div>
				</PreviewRow>
				<PreviewRow label="with sublabel">
					<div
						class="w-full rounded-xl overflow-hidden"
						style={{
							background: "var(--color-surface)",
							border: "1px solid var(--color-border)",
						}}
					>
						<ListItem
							label="Morning session"
							sublabel="Mar 10, 2026"
							onClick={() => {}}
						/>
						<ListItem
							label="Evening session"
							sublabel="Mar 12, 2026"
							isLast
							onClick={() => {}}
						/>
					</div>
				</PreviewRow>
				<PreviewRow label="active state">
					<div
						class="w-full rounded-xl overflow-hidden"
						style={{
							background: "var(--color-surface)",
							border: "1px solid var(--color-border)",
						}}
					>
						<ListItem label="Strength A" active onClick={() => {}} />
						<ListItem label="Hypertrophy B" isLast onClick={() => {}} />
					</div>
				</PreviewRow>
				<PreviewRow label="with actions slot">
					<div
						class="w-full rounded-xl overflow-hidden"
						style={{
							background: "var(--color-surface)",
							border: "1px solid var(--color-border)",
						}}
					>
						<ListItem
							label="Strength A"
							active
							onClick={() => {}}
							actions={
								<Button variant="ghost" size="icon" aria-label="Delete">
									×
								</Button>
							}
						/>
						<ListItem
							label="Hypertrophy B"
							isLast
							onClick={() => {}}
							actions={
								<Button variant="ghost" size="icon" aria-label="Delete">
									×
								</Button>
							}
						/>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── ListDetail ────────────────────────────────── */}
			<PageSection title="ListDetail">
				<PreviewRow label="two-panel master-detail layout">
					<div class="w-full flex flex-col gap-3">
						<div class="flex items-center gap-3">
							<Switch checkedSignal={listDetailHasDetail} />
							<span class="text-sm" style={{ color: "var(--color-muted)" }}>
								hasDetail: {String(listDetailHasDetail.value)}
							</span>
						</div>
						<div
							class="w-full rounded-xl overflow-hidden border border-(--color-border)"
							style={{ height: "320px" }}
						>
							<ListDetail
								list={
									<ul class="p-3 flex flex-col gap-1">
										{["Strength A", "Hypertrophy B", "Deload Week"].map(
											(name) => (
												<li
													key={name}
													class="px-3 py-2 rounded-lg text-sm cursor-pointer hover:bg-(--color-bg)"
													style={{ color: "var(--color-text)" }}
												>
													{name}
												</li>
											),
										)}
									</ul>
								}
								detail={
									listDetailHasDetail.value ? (
										<div class="p-6">
											<p
												class="font-medium"
												style={{ color: "var(--color-text)" }}
											>
												Strength A
											</p>
											<p
												class="text-sm mt-1"
												style={{ color: "var(--color-muted)" }}
											>
												4 sets · 3 exercises
											</p>
										</div>
									) : null
								}
								emptyState="Select a program to view its sets."
								hasDetail={listDetailHasDetail.value}
							/>
						</div>
					</div>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── CheckList ─────────────────────────────────── */}
			<PageSection title="CheckList">
				<PreviewRow label="scrollable — swipe right to check, swipe again or left to uncheck">
					<CheckList class="w-full max-h-[320px] overflow-y-auto">
						<CheckListSection label="Warm-up · Round 1 of 2">
							<CheckListItem defaultChecked>Jumping jacks · 30s</CheckListItem>
							<CheckListItem defaultChecked>Hip circles · 20s</CheckListItem>
							<CheckListItem>Arm swings · 20s</CheckListItem>
						</CheckListSection>
						<CheckListSection label="Warm-up · Round 2 of 2">
							<CheckListItem>Jumping jacks · 30s</CheckListItem>
							<CheckListItem>Hip circles · 20s</CheckListItem>
							<CheckListItem>Arm swings · 20s</CheckListItem>
						</CheckListSection>
						<CheckListSection label="Strength · Round 1 of 3">
							<CheckListItem>Push-ups × 10 reps</CheckListItem>
							<CheckListItem>Pull-ups × 5 reps</CheckListItem>
						</CheckListSection>
						<CheckListSection label="Strength · Round 2 of 3">
							<CheckListItem>Push-ups × 10 reps</CheckListItem>
							<CheckListItem>Pull-ups × 5 reps</CheckListItem>
						</CheckListSection>
						<CheckListSection label="Strength · Round 3 of 3">
							<CheckListItem>Push-ups × 10 reps</CheckListItem>
							<CheckListItem>Pull-ups × 5 reps</CheckListItem>
						</CheckListSection>
					</CheckList>
				</PreviewRow>
				<PreviewRow label="full height — no scroll">
					<CheckList class="w-full">
						<CheckListSection label="Round 1 of 2">
							<CheckListItem>Box jumps × 8 reps</CheckListItem>
							<CheckListItem>Burpees × 6 reps</CheckListItem>
						</CheckListSection>
						<CheckListSection label="Round 2 of 2">
							<CheckListItem>Box jumps × 8 reps</CheckListItem>
							<CheckListItem>Burpees × 6 reps</CheckListItem>
						</CheckListSection>
					</CheckList>
				</PreviewRow>
				<PreviewRow label="with subtitle — program exercise format">
					<CheckList class="w-full">
						<CheckListSection label="Strength · Round 1 of 3">
							<CheckListItem subtitle="60 kg · bilateral">
								10x Back Squat
							</CheckListItem>
							<CheckListItem subtitle="alternating">
								12x Reverse Lunge
							</CheckListItem>
							<CheckListItem subtitle="">30s Plank</CheckListItem>
							<CheckListItem subtitle="">3x 10s Crimp block hold</CheckListItem>
						</CheckListSection>
					</CheckList>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── ConfirmDialog ──────────────────────────────── */}
			<PageSection title="ConfirmDialog">
				<PreviewRow label="destructive action">
					<div class="flex items-center gap-3">
						<Button
							variant="destructive"
							size="sm"
							onClick={confirmDialog.show}
						>
							Delete Program
						</Button>
						{confirmResult.value && (
							<span class="text-sm" style={{ color: "var(--color-muted)" }}>
								{confirmResult.value}
							</span>
						)}
					</div>
					<ConfirmDialog
						openSignal={confirmDialog.open}
						title="Delete Program"
						description="This will permanently delete the program and all its sets."
						confirmLabel="Delete"
						onConfirm={async () => {
							confirmResult.value = `Deleted at ${new Date().toLocaleTimeString()}`;
						}}
					/>
				</PreviewRow>
			</PageSection>

			<Divider />

			{/* ── EditableMarkdown ────────────────────────────── */}
			<EditableMarkdownSection />
		</main>
	);
}

function EditableMarkdownSection() {
	const withContent = useSignal(
		"**Strength focus** — 3×5 at 85% 1RM\n\n- Rest 3 min between sets\n- Log RPE after each set",
	);
	const empty = useSignal(null);

	return (
		<PageSection title="EditableMarkdown">
			<PreviewRow label="with content">
				<div class="w-full max-w-md">
					<EditableMarkdown
						value={withContent.value}
						onSave={async (v) => {
							await new Promise((r) => setTimeout(r, 400));
							withContent.value = v || null;
						}}
					/>
				</div>
			</PreviewRow>
			<PreviewRow label="empty (placeholder)">
				<div class="w-full max-w-md">
					<EditableMarkdown
						value={empty.value}
						placeholder="Add a description…"
						onSave={async (v) => {
							await new Promise((r) => setTimeout(r, 400));
							empty.value = v || null;
						}}
					/>
				</div>
			</PreviewRow>
			<PreviewRow label="plain variant (inside a card)">
				<div
					class="w-full max-w-md rounded-xl p-4"
					style={{
						background: "var(--color-surface)",
						border: "1px solid var(--color-border)",
					}}
				>
					<EditableMarkdown
						value={withContent.value}
						variant="plain"
						onSave={async (v) => {
							await new Promise((r) => setTimeout(r, 400));
							withContent.value = v || null;
						}}
					/>
				</div>
			</PreviewRow>
			<PreviewRow label="disabled">
				<div class="w-full max-w-md">
					<EditableMarkdown
						value="Read-only markdown — **no edit button**."
						onSave={async () => {}}
						disabled
					/>
				</div>
			</PreviewRow>
		</PageSection>
	);
}
