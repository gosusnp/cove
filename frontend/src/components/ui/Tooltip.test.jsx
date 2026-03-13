// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { Tooltip, TooltipContent, TooltipTrigger } from "./Tooltip.jsx";

describe("Tooltip", () => {
	it("renders the trigger", () => {
		render(
			<Tooltip>
				<TooltipTrigger>
					<button type="button">Copy</button>
				</TooltipTrigger>
				<TooltipContent>Copied!</TooltipContent>
			</Tooltip>,
		);
		expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
	});

	it("renders the tooltip content", () => {
		render(
			<Tooltip>
				<TooltipTrigger>
					<button type="button">Copy</button>
				</TooltipTrigger>
				<TooltipContent>Copied!</TooltipContent>
			</Tooltip>,
		);
		expect(screen.getByText("Copied!")).toBeInTheDocument();
	});
});
