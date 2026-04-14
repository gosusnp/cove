// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-accordion.
// Radix Accordion uses ResizeObserver and CSS custom properties not supported by jsdom.

import { createContext } from "preact";
import { useContext, useState } from "preact/hooks";

const RootCtx = createContext({ open: new Set(), toggle: () => {} });
const ItemCtx = createContext({ value: "" });

export function Root({
	children,
	type = "single",
	defaultValue,
	className,
	...props
}) {
	const init = defaultValue
		? new Set(Array.isArray(defaultValue) ? defaultValue : [defaultValue])
		: new Set();
	const [open, setOpen] = useState(init);
	const toggle = (v) =>
		setOpen((prev) => {
			const next = new Set(type === "multiple" ? prev : []);
			prev.has(v) ? next.delete(v) : next.add(v);
			return next;
		});
	return (
		<RootCtx.Provider value={{ open, toggle }}>
			<div className={className} {...props}>
				{children}
			</div>
		</RootCtx.Provider>
	);
}

export function Item({ children, value, className, style, ...props }) {
	const { open } = useContext(RootCtx);
	return (
		<ItemCtx.Provider value={{ value }}>
			<div
				className={className}
				style={style}
				data-state={open.has(value) ? "open" : "closed"}
				{...props}
			>
				{children}
			</div>
		</ItemCtx.Provider>
	);
}

export function Header({ children, asChild }) {
	if (asChild) return children;
	return <div>{children}</div>;
}

export function Trigger({
	children,
	className,
	onClick,
	onKeyDown,
	onKeyUp,
	...props
}) {
	const { open, toggle } = useContext(RootCtx);
	const { value } = useContext(ItemCtx);
	const isOpen = open.has(value);
	return (
		<button
			type="button"
			className={className}
			data-state={isOpen ? "open" : "closed"}
			onClick={(e) => {
				toggle(value);
				onClick?.(e);
			}}
			onKeyDown={(e) => {
				if (e.key === "Enter") {
					toggle(value);
				}
				onKeyDown?.(e);
			}}
			onKeyUp={(e) => {
				if (e.key === " ") {
					toggle(value);
				}
				onKeyUp?.(e);
			}}
			{...props}
		>
			{children}
		</button>
	);
}

export function Content({ children, className, ...props }) {
	const { open } = useContext(RootCtx);
	const { value } = useContext(ItemCtx);
	return (
		<div
			className={className}
			data-state={open.has(value) ? "open" : "closed"}
			{...props}
		>
			{children}
		</div>
	);
}
