// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { computed, useSignal } from "@preact/signals";
import { createContext } from "preact";
import { useContext, useEffect, useMemo, useRef } from "preact/hooks";
import { cn } from "../../lib/utils";

const SWIPE_THRESHOLD = 56;

// Internal context — CheckListSection provides, CheckListItem consumes.
// Items register their checked signal so the section can derive allDone
// and support swipe-to-check-all on the header.
const SectionCtx = createContext(null);

// ── CheckList ──────────────────────────────────────────────────────────────────

// Scroll container. Constrain height externally to enable sticky section headers.
// overflow-x-clip clips the button overshoot during swipe gestures without
// creating a scroll context (unlike overflow-x-hidden), which would break sticky.
export function CheckList({ class: className, children }) {
	return (
		<div
			class={cn(
				"rounded-xl overflow-x-clip",
				"border border-(--color-border)",
				"bg-(--color-surface)",
				className,
			)}
		>
			{children}
		</div>
	);
}

// ── CheckListSection ───────────────────────────────────────────────────────────

// Groups items under a sticky, swipeable label.
// Swipe right on the header to check all items in the section.
// The header strikes through and shows a checkmark when all items are done.
export function CheckListSection({ label, children }) {
	// Set of each item's `checked` signal — populated by CheckListItem on mount.
	const items = useRef(new Set());
	// Bumped on register/unregister so the computed re-subscribes to new items.
	const version = useSignal(0);

	// allDone is a computed signal: true when every registered item is checked.
	// Created once in a ref so it isn't recreated on re-renders.
	const allDoneRef = useRef(null);
	if (!allDoneRef.current) {
		allDoneRef.current = computed(() => {
			void version.value; // subscribe to registration changes
			if (items.current.size === 0) return false;
			return [...items.current].every((s) => s.value);
		});
	}
	const allDone = allDoneRef.current.value;

	function register(checkedSignal) {
		items.current.add(checkedSignal);
		version.value++;
		return () => {
			items.current.delete(checkedSignal);
			version.value++;
		};
	}

	function checkAll() {
		for (const s of items.current) {
			s.value = true;
		}
	}

	// ── Header swipe-right to check all ───────────────────────────────────────

	const fillOpacity = useSignal(0);
	const translateX = useSignal(0);
	const swiping = useSignal(false);

	const touchStartX = useRef(0);
	const touchStartY = useRef(0);
	const swipeIntent = useRef(null); // null | 'h' | 'v'
	const swipeHandled = useRef(false);
	const dx = useRef(0);
	// Cleanup fn for the per-gesture document-level touchmove listener.
	const touchMoveCleanup = useRef(null);
	// Remove any in-flight listener when the component unmounts (e.g. in tests
	// that fire touchStart but no touchEnd, or during hot-reload).
	useEffect(() => () => touchMoveCleanup.current?.(), []);

	const touchMoveRef = useRef(null);
	touchMoveRef.current = (e) => {
		const deltaX = e.touches[0].clientX - touchStartX.current;
		const deltaY = e.touches[0].clientY - touchStartY.current;

		if (swipeIntent.current === null) {
			if (Math.abs(deltaX) > 6 || Math.abs(deltaY) > 6) {
				swipeIntent.current = Math.abs(deltaX) >= Math.abs(deltaY) ? "h" : "v";
				// Claim the gesture immediately so the compositor doesn't commit to scroll.
				// Claim for any horizontal intent — the deltaX > 0 directionality
				// check belongs on the processing step below, not here. On real
				// devices the first detected deltaX can be slightly negative even
				// during a rightward swipe due to touch noise.
				if (swipeIntent.current === "h") {
					e.preventDefault();
				}
			}
			return;
		}

		// Only handle rightward horizontal swipes.
		if (swipeIntent.current !== "h" || deltaX <= 0) return;

		e.preventDefault();
		swiping.value = true;
		dx.current = deltaX;
		translateX.value = Math.min(SWIPE_THRESHOLD * 1.3, deltaX);
		fillOpacity.value = Math.min(0.1, (deltaX / SWIPE_THRESHOLD) * 0.1);
	};

	function onTouchStart(e) {
		// Nothing to do when all items are already checked.
		if (allDoneRef.current.value) return;

		touchStartX.current = e.touches[0].clientX;
		touchStartY.current = e.touches[0].clientY;
		swipeIntent.current = null;
		swiping.value = false;
		swipeHandled.current = false;
		dx.current = 0;

		// Attach non-passive touchmove to document for this gesture.
		// A document-level non-passive listener is the only reliable way to call
		// e.preventDefault() before the mobile browser's compositor thread commits
		// to a scroll — element-level listeners inside scroll containers are
		// evaluated too late on iOS Safari and Android WebView.
		const handler = (ev) => touchMoveRef.current(ev);
		document.addEventListener("touchmove", handler, { passive: false });
		touchMoveCleanup.current = () =>
			document.removeEventListener("touchmove", handler);
	}

	function onTouchEnd() {
		touchMoveCleanup.current?.();
		touchMoveCleanup.current = null;
		if (swiping.value && dx.current >= SWIPE_THRESHOLD) {
			checkAll();
			swipeHandled.current = true;
		}
		translateX.value = 0;
		fillOpacity.value = 0;
		swiping.value = false;
		swipeIntent.current = null;
	}

	function onTouchCancel() {
		touchMoveCleanup.current?.();
		touchMoveCleanup.current = null;
		translateX.value = 0;
		fillOpacity.value = 0;
		swiping.value = false;
		swipeIntent.current = null;
	}

	// Stable context value — items capture this in a [] useEffect, so it must
	// not change identity across re-renders of CheckListSection.
	const ctx = useMemo(() => ({ register }), []);

	return (
		<SectionCtx.Provider value={ctx}>
			<div>
				<button
					type="button"
					class="sticky top-0 z-10 relative overflow-hidden w-full flex items-center gap-2 px-4 py-1.5 border-b border-(--color-border) bg-transparent border-x-0 border-t-0 text-xs font-semibold uppercase tracking-widest select-none cursor-default text-left"
					style={{
						background: "var(--color-bg)",
						color: "var(--color-muted)",
						touchAction: "pan-y",
						willChange: "transform",
						transform:
							translateX.value !== 0
								? `translateX(${translateX.value}px)`
								: undefined,
						transition: swiping.value ? "none" : "transform 0.2s ease-out",
					}}
					onTouchStart={onTouchStart}
					onTouchEnd={onTouchEnd}
					onTouchCancel={onTouchCancel}
					onClick={checkAll}
				>
					{/* Swipe fill — grows as user drags right */}
					<span
						class="absolute inset-0 pointer-events-none"
						style={{
							background: "var(--color-accent)",
							opacity: fillOpacity.value,
							transition: swiping.value ? "none" : "opacity 0.2s ease-out",
						}}
					/>

					<span
						class={cn("relative flex-1", allDone && "line-through")}
						style={{
							opacity: allDone ? 0.5 : 1,
							transition: "opacity 0.15s ease-out",
						}}
					>
						{label}
					</span>

					<span
						class={cn(
							"relative shrink-0 transition-opacity duration-150",
							allDone ? "opacity-100" : "opacity-0",
						)}
						style={{ color: "var(--color-accent)" }}
						aria-hidden="true"
					>
						<CheckIcon />
					</span>
				</button>
				<div>{children}</div>
			</div>
		</SectionCtx.Provider>
	);
}

// ── CheckListItem ──────────────────────────────────────────────────────────────

// Swipe right to check, swipe right again or swipe left to uncheck.
// Click/tap also toggles.
// defaultChecked sets the initial state for demos and testing.
export function CheckListItem({
	children,
	defaultChecked = false,
	class: className,
}) {
	const ctx = useContext(SectionCtx);
	const checked = useSignal(defaultChecked);
	const translateX = useSignal(0);
	const swiping = useSignal(false);

	// Register this item's checked signal with the parent section.
	useEffect(() => {
		if (!ctx) return;
		return ctx.register(checked);
	}, []);

	const touchStartX = useRef(0);
	const touchStartY = useRef(0);
	const swipeIntent = useRef(null); // null | 'h' | 'v'
	const swipeHandled = useRef(false);
	// Cleanup fn for the per-gesture document-level touchmove listener.
	const touchMoveCleanup = useRef(null);
	// Remove any in-flight listener when the component unmounts (e.g. in tests
	// that fire touchStart but no touchEnd, or during hot-reload).
	useEffect(() => () => touchMoveCleanup.current?.(), []);

	const touchMoveRef = useRef(null);
	touchMoveRef.current = (e) => {
		const dx = e.touches[0].clientX - touchStartX.current;
		const dy = e.touches[0].clientY - touchStartY.current;

		// Determine swipe direction on first meaningful movement.
		if (swipeIntent.current === null) {
			if (Math.abs(dx) > 6 || Math.abs(dy) > 6) {
				swipeIntent.current = Math.abs(dx) >= Math.abs(dy) ? "h" : "v";
				// Claim the gesture immediately so the compositor doesn't commit to scroll.
				if (swipeIntent.current === "h") {
					e.preventDefault();
				}
			}
			return;
		}

		if (swipeIntent.current !== "h") return;

		e.preventDefault();
		swiping.value = true;
		translateX.value = Math.max(
			-SWIPE_THRESHOLD * 1.3,
			Math.min(SWIPE_THRESHOLD * 1.3, dx),
		);
	};

	function onTouchStart(e) {
		touchStartX.current = e.touches[0].clientX;
		touchStartY.current = e.touches[0].clientY;
		swipeIntent.current = null;
		swiping.value = false;
		swipeHandled.current = false;

		// Attach non-passive touchmove to document for this gesture.
		// A document-level non-passive listener is the only reliable way to call
		// e.preventDefault() before the mobile browser's compositor thread commits
		// to a scroll — element-level listeners inside scroll containers are
		// evaluated too late on iOS Safari and Android WebView.
		const handler = (ev) => touchMoveRef.current(ev);
		document.addEventListener("touchmove", handler, { passive: false });
		touchMoveCleanup.current = () =>
			document.removeEventListener("touchmove", handler);
	}

	function onTouchEnd() {
		touchMoveCleanup.current?.();
		touchMoveCleanup.current = null;
		if (swiping.value && Math.abs(translateX.value) >= SWIPE_THRESHOLD) {
			checked.value = !checked.value;
			swipeHandled.current = true;
		}
		translateX.value = 0;
		swiping.value = false;
		swipeIntent.current = null;
	}

	function onTouchCancel() {
		touchMoveCleanup.current?.();
		touchMoveCleanup.current = null;
		translateX.value = 0;
		swiping.value = false;
		swipeIntent.current = null;
	}

	function onClick() {
		// Touch devices fire click after touchend; skip if the swipe already handled it.
		if (swipeHandled.current) {
			swipeHandled.current = false;
			return;
		}
		checked.value = !checked.value;
	}

	// Accent fill intensity: grows as the user drags, stays dim when checked.
	const fillOpacity = checked.value
		? 0.07
		: translateX.value > 0
			? Math.min(0.1, (translateX.value / SWIPE_THRESHOLD) * 0.1)
			: 0;

	return (
		<div class="border-b border-(--color-border) last:border-b-0">
			{/*
			 * overflow-hidden and transform must be on the SAME element to avoid
			 * a WebKit compositing bug where a transformed child inside an
			 * overflow:hidden parent paints on the wrong layer.
			 */}
			<button
				type="button"
				aria-pressed={checked.value}
				class={cn(
					"relative w-full overflow-hidden",
					"flex items-center gap-3 px-4 py-3",
					"cursor-pointer select-none text-left",
					"bg-transparent border-none",
					"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-(--color-accent) focus-visible:ring-inset",
					className,
				)}
				style={{
					transform:
						translateX.value !== 0
							? `translateX(${translateX.value}px)`
							: undefined,
					transition: swiping.value ? "none" : "transform 0.2s ease-out",
					touchAction: "pan-y",
					willChange: "transform",
				}}
				onTouchStart={onTouchStart}
				onTouchEnd={onTouchEnd}
				onTouchCancel={onTouchCancel}
				onClick={onClick}
			>
				{/* Accent fill — clipped by the button's own overflow:hidden */}
				<span
					class="absolute inset-0 pointer-events-none"
					style={{
						background: "var(--color-accent)",
						opacity: fillOpacity,
						transition: swiping.value ? "none" : "opacity 0.2s ease-out",
					}}
				/>

				<span
					class={cn(
						"relative flex-1 text-sm leading-snug",
						checked.value && "line-through",
					)}
					style={{
						color: checked.value ? "var(--color-muted)" : "var(--color-text)",
						transition: "color 0.15s ease-out",
					}}
				>
					{children}
				</span>

				<span
					class={cn(
						"relative shrink-0 transition-opacity duration-150",
						checked.value ? "opacity-100" : "opacity-0",
					)}
					style={{ color: "var(--color-accent)" }}
					aria-hidden="true"
				>
					<CheckIcon />
				</span>
			</button>
		</div>
	);
}

// ── Icons ──────────────────────────────────────────────────────────────────────

function CheckIcon() {
	return (
		<svg
			width="16"
			height="16"
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M3 8.5l3.5 3.5 6.5-7"
				stroke="currentColor"
				stroke-width="1.75"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	);
}
