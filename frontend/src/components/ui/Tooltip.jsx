// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import * as RadixTooltip from "@radix-ui/react-tooltip";
import { cn } from "../../lib/utils";

export function Tooltip({ children, ...props }) {
	return (
		<RadixTooltip.Provider delayDuration={300}>
			<RadixTooltip.Root {...props}>{children}</RadixTooltip.Root>
		</RadixTooltip.Provider>
	);
}

export function TooltipTrigger({ children, ...props }) {
	return (
		<RadixTooltip.Trigger asChild {...props}>
			<span class="inline-flex">{children}</span>
		</RadixTooltip.Trigger>
	);
}

export function TooltipContent({ class: className, children, ...props }) {
	return (
		<RadixTooltip.Portal>
			<RadixTooltip.Content
				sideOffset={6}
				className={cn(
					"z-50 rounded-md px-2.5 py-1.5 text-xs font-medium shadow-md",
					"bg-(--color-text) text-(--color-surface)",
					"animate-in fade-in-0 zoom-in-95",
					"data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95",
					"data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2",
					className,
				)}
				{...props}
			>
				{children}
			</RadixTooltip.Content>
		</RadixTooltip.Portal>
	);
}
