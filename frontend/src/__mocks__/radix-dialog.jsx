// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for @radix-ui/react-dialog.

import { createContext, useContext } from "preact/compat";

const DialogContext = createContext(null);

export const Root = ({ children, open, onOpenChange }) =>
	open ? (
		<DialogContext.Provider value={{ onOpenChange }}>
			{children}
		</DialogContext.Provider>
	) : null;

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
	const ctx = useContext(DialogContext);
	const handleClose = () => ctx?.onOpenChange?.(false);
	if (asChild) {
		const child = Array.isArray(children) ? children[0] : children;
		const originalOnClick = child?.props?.onClick;
		return {
			...child,
			props: {
				...child?.props,
				onClick: (e) => {
					originalOnClick?.(e);
					handleClose();
				},
			},
		};
	}
	return (
		<button type="button" onClick={handleClose}>
			{children}
		</button>
	);
};
