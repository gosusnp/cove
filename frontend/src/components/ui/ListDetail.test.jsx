// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { ListDetail } from "./ListDetail.jsx";

describe("ListDetail", () => {
	it("renders list panel", () => {
		render(
			<ListDetail
				list={<div>Program list</div>}
				detail={null}
				emptyState="Select a program"
				hasDetail={false}
			/>,
		);
		expect(screen.getByText("Program list")).toBeInTheDocument();
	});

	it("renders detail panel when provided", () => {
		render(
			<ListDetail
				list={<div>Program list</div>}
				detail={<div>Program detail</div>}
				emptyState="Select a program"
				hasDetail={true}
			/>,
		);
		expect(screen.getByText("Program detail")).toBeInTheDocument();
	});

	it("shows emptyState when detail is null", () => {
		render(
			<ListDetail
				list={<div>Program list</div>}
				detail={null}
				emptyState="Select a program to view its sets."
				hasDetail={false}
			/>,
		);
		expect(
			screen.getByText("Select a program to view its sets."),
		).toBeInTheDocument();
	});

	it("hides list on mobile when hasDetail is true", () => {
		render(
			<ListDetail
				list={<div>Program list</div>}
				detail={<div>Program detail</div>}
				emptyState="Select a program"
				hasDetail={true}
			/>,
		);
		// The list panel should have the "hidden" class (visible only on md+)
		const listPanel = screen.getByText("Program list").parentElement;
		expect(listPanel.className).toContain("hidden");
		expect(listPanel.className).toContain("md:block");
	});

	it("hides detail on mobile when hasDetail is false", () => {
		render(
			<ListDetail
				list={<div>Program list</div>}
				detail={<div>Program detail</div>}
				emptyState="Select a program"
				hasDetail={false}
			/>,
		);
		// The detail panel should have "hidden" class when hasDetail is false
		const detailPanel = screen.getByText("Program detail").parentElement;
		expect(detailPanel.className).toContain("hidden");
	});
});
