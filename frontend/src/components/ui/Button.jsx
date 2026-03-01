// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { cva } from "class-variance-authority";
import { cn } from "../../lib/utils";

const button = cva(
	[
		"inline-flex items-center justify-center gap-2 font-medium rounded-lg",
		"transition-transform touch-manipulation active:scale-95",
		"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
		"focus-visible:ring-(--color-accent)",
		"disabled:opacity-50 disabled:pointer-events-none",
	],
	{
		variants: {
			variant: {
				primary: ["bg-(--color-accent) text-white", "hover:opacity-90"],
				outline: [
					"border border-(--color-border) bg-(--color-surface)",
					"text-(--color-text) hover:bg-(--color-bg)",
				],
				ghost: ["bg-transparent text-(--color-text)", "hover:bg-(--color-bg)"],
			},
			size: {
				sm: "h-8 px-3 text-sm gap-1.5",
				md: "h-10 px-4 text-sm",
				lg: "h-12 px-6 text-base",
			},
		},
		defaultVariants: {
			variant: "primary",
			size: "md",
		},
	},
);

export function Button({ variant, size, class: className, ...props }) {
	return <button class={cn(button({ variant, size }), className)} {...props} />;
}
