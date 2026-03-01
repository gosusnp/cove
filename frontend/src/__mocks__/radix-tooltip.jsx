// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-tooltip.
// The real package uses browser APIs that jsdom does not support.

export const Provider = ({ children }) => children;
export const Root = ({ children }) => children;
export const Trigger = ({ children, asChild }) => {
	if (asChild) return children;
	return children;
};
export const Portal = ({ children }) => children;
export const Content = ({ children, className }) => (
	<div class={className}>{children}</div>
);
