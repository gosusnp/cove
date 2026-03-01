// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-navigation-menu.
// The real package uses ResizeObserver and CSS custom properties that jsdom
// does not support, causing InvalidCharacterError in tests.

export const Root = ({ children, className }) => (
	<nav class={className}>{children}</nav>
);

export const List = ({ children, className }) => (
	<ul class={className}>{children}</ul>
);

export const Item = ({ children, className }) => (
	<li class={className}>{children}</li>
);

// When asChild is true, Radix's Slot would merge props onto the child element.
// In tests we just render the child directly — it already carries all its props.
export const Link = ({ children, asChild }) => {
	if (asChild) return children;
	return children;
};
