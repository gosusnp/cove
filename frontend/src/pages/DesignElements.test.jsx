// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { LocationProvider } from "preact-iso";
import { DesignElements } from "./DesignElements.jsx";

// Mock the UI components to avoid issues with Radix UI in jsdom/vitest.
// Radix UI components often use APIs not fully implemented in jsdom,
// causing "Invalid hook call" when used with Preact in a test environment.
vi.mock("../components/ui/Dialog.jsx", () => ({
	Dialog: ({ children }) => <div data-testid="mock-dialog">{children}</div>,
	DialogTrigger: ({ children }) => (
		<button type="button" data-testid="mock-dialog-trigger">
			{children}
		</button>
	),
	DialogContent: ({ children }) => (
		<div data-testid="mock-dialog-content">{children}</div>
	),
	DialogTitle: ({ children }) => <h2>{children}</h2>,
	DialogDescription: ({ children }) => <p>{children}</p>,
	DialogClose: ({ children }) => (
		<button type="button" data-testid="mock-dialog-close">
			{children}
		</button>
	),
}));

vi.mock("../components/ui/Switch.jsx", () => ({
	Switch: () => (
		<input type="checkbox" readOnly checked={false} aria-checked="false" />
	),
}));

vi.mock("../components/ui/Tooltip.jsx", () => ({
	Tooltip: ({ children }) => <div>{children}</div>,
	TooltipTrigger: ({ children }) => <div>{children}</div>,
	TooltipContent: ({ children }) => <div role="tooltip">{children}</div>,
}));

vi.mock("../components/ui/Section.jsx", () => ({
	Section: ({ title, children }) => (
		<section>
			<h2>{title}</h2>
			{children}
		</section>
	),
	Row: ({ label, children }) => (
		<div>
			<span>{label}</span>
			{children}
		</div>
	),
	Divider: () => <hr />,
}));

function withProviders(ui) {
	return render(<LocationProvider>{ui}</LocationProvider>);
}

describe("DesignElements", () => {
	it("renders when VITE_COVE_ENV is dev", () => {
		// Mock the env var. In Vitest, we can sometimes set this directly if it's evaluated at runtime.
		import.meta.env.VITE_COVE_ENV = "dev";

		withProviders(<DesignElements />);
		expect(
			screen.getByRole("heading", { name: "Design Elements" }),
		).toBeInTheDocument();
	});

	it("renders all sections", () => {
		import.meta.env.VITE_COVE_ENV = "dev";
		withProviders(<DesignElements />);
		expect(
			screen.getByRole("heading", { name: "Top Navigation" }),
		).toBeInTheDocument();
		expect(screen.getByRole("heading", { name: "Button" })).toBeInTheDocument();
		expect(screen.getByRole("heading", { name: "Dialog" })).toBeInTheDocument();
		expect(screen.getByRole("heading", { name: "Switch" })).toBeInTheDocument();
		expect(screen.getByRole("heading", { name: "Avatar" })).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "TextField" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Tooltip" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Section + Row" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "PageTitle" }),
		).toBeInTheDocument();
	});

	it("renders different button variants", () => {
		import.meta.env.VITE_COVE_ENV = "dev";
		withProviders(<DesignElements />);
		// Section headers also have some text, but we check for the specific size labels
		// 3 variants (primary/outline/ghost) + 1 destructive = 4 "Small" buttons
		expect(screen.getAllByText("Small").length).toBe(4);
		// 3 variants (primary/outline/ghost) + 1 destructive = 4 "Medium" buttons
		expect(screen.getAllByText("Medium").length).toBe(4);
		expect(screen.getAllByText("Large").length).toBe(3);
	});

	it("renders switches", () => {
		import.meta.env.VITE_COVE_ENV = "dev";
		withProviders(<DesignElements />);
		const switches = screen.getAllByRole("checkbox");
		expect(switches.length).toBeGreaterThanOrEqual(4);
	});

	it("renders avatars", () => {
		import.meta.env.VITE_COVE_ENV = "dev";
		withProviders(<DesignElements />);
		expect(screen.getAllByText("JM").length).toBeGreaterThan(0);
		expect(
			screen.getAllByLabelText("jimmy@example.com").length,
		).toBeGreaterThan(0);
	});

	it("does not render when VITE_COVE_ENV is not dev", () => {
		// We set it to prod to ensure it hides.
		import.meta.env.VITE_COVE_ENV = "prod";
		const { container } = withProviders(<DesignElements />);
		expect(container.firstChild).toBeNull();
	});
});
