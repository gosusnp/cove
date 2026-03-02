// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-dialog.

export const Root = ({ children, open }) => (open ? children : null);

export const Trigger = ({ children, asChild }) => {
	if (asChild) return children;
	return <button type="button">{children}</button>;
};

export const Portal = ({ children }) => <>{children}</>;

export const Overlay = ({ className }) => (
	<div data-testid="dialog-overlay" className={className} />
);

export const Content = ({ children, className, ...props }) => (
	<div role="dialog" className={className} {...props}>
		{children}
	</div>
);

export const Title = ({ children, className, ...props }) => (
	<h2 className={className} {...props}>
		{children}
	</h2>
);

export const Description = ({ children, className, ...props }) => (
	<p className={className} {...props}>
		{children}
	</p>
);

export const Close = ({ children, asChild }) => {
	if (asChild) return children;
	return <button type="button">{children}</button>;
};
