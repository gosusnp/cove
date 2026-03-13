// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import * as RadixDialog from "@radix-ui/react-dialog";
import { cn } from "../../lib/utils";

export function Dialog({ openSignal, onOpenChange, children }) {
	return (
		<RadixDialog.Root
			open={openSignal.value}
			onOpenChange={(v) => {
				if (v && document.activeElement instanceof HTMLElement) {
					document.activeElement.blur();
				}
				openSignal.value = v;
				onOpenChange?.(v);
			}}
		>
			{children}
		</RadixDialog.Root>
	);
}

export function DialogTrigger({ children, ...props }) {
	return (
		<RadixDialog.Trigger asChild {...props}>
			{children}
		</RadixDialog.Trigger>
	);
}

export function DialogContent({
	class: className,
	fullscreen,
	children,
	...props
}) {
	return (
		<RadixDialog.Portal>
			<RadixDialog.Overlay className="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm" />
			<RadixDialog.Content
				className={cn(
					"bg-(--color-surface) text-(--color-text) focus:outline-none",
					fullscreen
						? cn(
								// Mobile: full screen
								"fixed inset-0 z-50 flex flex-col overflow-hidden",
								// sm+: centered dialog
								"sm:inset-auto sm:left-1/2 sm:top-1/2",
								"sm:-translate-x-1/2 sm:-translate-y-1/2",
								"sm:w-[calc(100vw-2rem)] sm:max-w-lg sm:max-h-[90vh]",
								"sm:rounded-xl sm:shadow-lg",
							)
						: cn(
								"fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2",
								"w-[calc(100vw-2rem)] max-w-lg",
								"rounded-xl p-6 shadow-lg",
							),
					className,
				)}
				onCloseAutoFocus={(e) => {
					e.preventDefault();
					document.body.focus();
				}}
				{...props}
			>
				{children}
			</RadixDialog.Content>
		</RadixDialog.Portal>
	);
}

export function DialogTitle({ class: className, children, ...props }) {
	return (
		<RadixDialog.Title
			className={cn(
				"text-lg font-semibold leading-none tracking-tight",
				className,
			)}
			{...props}
		>
			{children}
		</RadixDialog.Title>
	);
}

export function DialogDescription({ class: className, children, ...props }) {
	return (
		<RadixDialog.Description
			className={cn("text-sm mt-2", className)}
			style={{ color: "var(--color-muted)" }}
			{...props}
		>
			{children}
		</RadixDialog.Description>
	);
}

export function DialogClose({ children, ...props }) {
	return (
		<RadixDialog.Close asChild {...props}>
			{children}
		</RadixDialog.Close>
	);
}
