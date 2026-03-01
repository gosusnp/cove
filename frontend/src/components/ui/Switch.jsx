// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import * as RadixSwitch from "@radix-ui/react-switch";
import { useSignal } from "@preact/signals";
import { cn } from "../../lib/utils";

export function Switch({ checkedSignal, class: className, ...props }) {
	const internal = useSignal(false);
	const checked = checkedSignal ?? internal;

	return (
		<RadixSwitch.Root
			checked={checked.value}
			onCheckedChange={(v) => {
				checked.value = v;
			}}
			className={cn(
				"relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full",
				"border-2 border-transparent transition-colors touch-manipulation",
				"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
				"focus-visible:ring-(--color-accent)",
				"disabled:cursor-not-allowed disabled:opacity-50",
				"data-[state=checked]:bg-(--color-accent)",
				"data-[state=unchecked]:bg-(--color-border)",
				className,
			)}
			{...props}
		>
			<RadixSwitch.Thumb
				className={cn(
					"pointer-events-none block h-5 w-5 rounded-full bg-white shadow-sm",
					"transition-transform",
					"data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0",
				)}
			/>
		</RadixSwitch.Root>
	);
}
