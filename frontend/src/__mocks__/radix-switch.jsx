// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-switch.

export const Root = ({
	children,
	checked,
	onCheckedChange,
	disabled,
	className,
	...props
}) => (
	<button
		role="switch"
		aria-checked={checked}
		data-state={checked ? "checked" : "unchecked"}
		disabled={disabled}
		onClick={() => onCheckedChange?.(!checked)}
		className={className}
		{...props}
	>
		{children}
	</button>
);

export const Thumb = ({ className }) => <span className={className} />;
