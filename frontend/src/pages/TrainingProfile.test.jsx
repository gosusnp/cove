// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor, within } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { TrainingProfile } from "./TrainingProfile.jsx";

const MOCK_USER = { email: "jane@example.com", display_name: "Jane Smith" };

const renderPage = (opts = {}) =>
	withProviders(<TrainingProfile />, {
		path: "/train/profile",
		user: MOCK_USER,
		...opts,
	});

function mockFetch(profile = null, { patchOk = true } = {}) {
	return vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url === "/api/users/me/training-profile") {
			if (opts?.method === "PATCH") {
				if (!patchOk) return Promise.resolve({ ok: false });
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({}),
				});
			}
			if (profile === null) {
				return Promise.resolve({ ok: false, status: 404 });
			}
			return Promise.resolve({
				ok: true,
				status: 200,
				json: () => Promise.resolve(profile),
			});
		}
		return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
	});
}

describe("TrainingProfile", () => {
	afterEach(() => vi.restoreAllMocks());

	describe("rendering", () => {
		it("renders the page heading", async () => {
			mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("heading", { name: "Training Profile" }),
				).toBeInTheDocument(),
			);
		});

		it("shows Motivation, Disciplines, and Constraints sections", async () => {
			mockFetch(null);
			renderPage();
			await waitFor(() => {
				expect(screen.getByText("Motivation")).toBeInTheDocument();
				expect(screen.getByText("Disciplines")).toBeInTheDocument();
				expect(screen.getByText("Constraints")).toBeInTheDocument();
			});
		});

		it("shows empty disciplines message when profile has no disciplines", async () => {
			mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByText("No disciplines added yet."),
				).toBeInTheDocument(),
			);
		});

		it("shows Add button for disciplines", async () => {
			mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument(),
			);
		});
	});

	describe("loading profile data", () => {
		it("calls GET /api/users/me/training-profile on mount", async () => {
			const fetchSpy = mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({ credentials: "include" }),
				),
			);
		});

		it("shows motivation text from the API", async () => {
			mockFetch({ motivation: "Stay strong and consistent" });
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByText("Stay strong and consistent"),
				).toBeInTheDocument(),
			);
		});

		it("shows constraints text from the API", async () => {
			mockFetch({ constraints: "3 days per week, home gym only" });
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByText("3 days per week, home gym only"),
				).toBeInTheDocument(),
			);
		});

		it("renders discipline cards for each discipline", async () => {
			mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
					{
						name: "Running",
						years_practice: 2,
						level: "beginner",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() => {
				expect(
					screen.getByRole("heading", { name: "Climbing" }),
				).toBeInTheDocument();
				expect(
					screen.getByRole("heading", { name: "Running" }),
				).toBeInTheDocument();
			});
		});

		it("populates the years_practice field for a discipline", async () => {
			mockFetch({
				disciplines: [
					{ name: "Climbing", years_practice: 7, level: "expert", notes: null },
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByDisplayValue("7")).toBeInTheDocument(),
			);
		});

		it("shows Remove button for each discipline", async () => {
			mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
					{
						name: "Running",
						years_practice: 2,
						level: "beginner",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getAllByRole("button", { name: "Remove" })).toHaveLength(
					2,
				),
			);
		});

		it("shows discipline notes when present", async () => {
			mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: "Currently bouldering V5",
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Currently bouldering V5")).toBeInTheDocument(),
			);
		});

		it("renders without error when profile fetch fails", async () => {
			vi.spyOn(global, "fetch").mockRejectedValue(new Error("network error"));
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Motivation")).toBeInTheDocument(),
			);
		});
	});

	describe("motivation", () => {
		it("calls PATCH with new motivation when saved", async () => {
			const fetchSpy = mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Motivation")).toBeInTheDocument(),
			);

			const section = screen.getByText("Motivation").closest("section");
			fireEvent.click(within(section).getByRole("button", { name: "Edit" }));
			fireEvent.input(within(section).getByRole("textbox"), {
				target: { value: "Train for longevity" },
			});
			fireEvent.click(within(section).getByRole("button", { name: "Save" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: JSON.stringify({ motivation: "Train for longevity" }),
					}),
				),
			);
		});

		it("sends null motivation when saved with empty text", async () => {
			const fetchSpy = mockFetch({ motivation: "Old motivation" });
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Old motivation")).toBeInTheDocument(),
			);

			const section = screen.getByText("Motivation").closest("section");
			fireEvent.click(within(section).getByRole("button", { name: "Edit" }));
			const textarea = within(section).getByRole("textbox");
			fireEvent.input(textarea, { target: { value: "   " } });
			fireEvent.click(within(section).getByRole("button", { name: "Save" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: JSON.stringify({ motivation: null }),
					}),
				),
			);
		});

		it("shows error in EditableMarkdown when PATCH fails", async () => {
			mockFetch(null, { patchOk: false });
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Motivation")).toBeInTheDocument(),
			);

			const section = screen.getByText("Motivation").closest("section");
			fireEvent.click(within(section).getByRole("button", { name: "Edit" }));
			fireEvent.input(within(section).getByRole("textbox"), {
				target: { value: "Will fail" },
			});
			fireEvent.click(within(section).getByRole("button", { name: "Save" }));

			await waitFor(() =>
				expect(screen.getByText("Failed to save.")).toBeInTheDocument(),
			);
		});
	});

	describe("constraints", () => {
		it("calls PATCH with new constraints when saved", async () => {
			const fetchSpy = mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Constraints")).toBeInTheDocument(),
			);

			const section = screen.getByText("Constraints").closest("section");
			fireEvent.click(within(section).getByRole("button", { name: "Edit" }));
			fireEvent.input(within(section).getByRole("textbox"), {
				target: { value: "5 days a week, full gym" },
			});
			fireEvent.click(within(section).getByRole("button", { name: "Save" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: JSON.stringify({ constraints: "5 days a week, full gym" }),
					}),
				),
			);
		});
	});

	describe("add discipline", () => {
		it("calls PATCH with a new empty discipline appended", async () => {
			const fetchSpy = mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: JSON.stringify({
							disciplines: [
								{
									name: null,
									years_practice: null,
									level: null,
									notes: null,
								},
							],
						}),
					}),
				),
			);
		});

		it("shows a new discipline card immediately", async () => {
			mockFetch(null);
			renderPage();
			await waitFor(() =>
				expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(
					screen.queryByText("No disciplines added yet."),
				).not.toBeInTheDocument(),
			);
			expect(
				screen.getByRole("button", { name: "Remove" }),
			).toBeInTheDocument();
		});

		it("shows an error near the Add button when PATCH fails", async () => {
			mockFetch(null, { patchOk: false });
			renderPage();
			await waitFor(() =>
				expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(
					screen.getByText("Failed to add discipline."),
				).toBeInTheDocument(),
			);
		});
	});

	describe("remove discipline", () => {
		it("calls PATCH with the discipline excluded", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Remove" }),
				).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Remove" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: JSON.stringify({ disciplines: [] }),
					}),
				),
			);
		});

		it("removes the discipline card from the UI", async () => {
			mockFetch({
				disciplines: [
					{
						name: "Running",
						years_practice: 2,
						level: "beginner",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("heading", { name: "Running" }),
				).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Remove" }));

			await waitFor(() =>
				expect(
					screen.queryByRole("heading", { name: "Running" }),
				).not.toBeInTheDocument(),
			);
		});

		it("restores the discipline when PATCH fails", async () => {
			vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
				if (
					url === "/api/users/me/training-profile" &&
					opts?.method === "PATCH"
				) {
					return Promise.resolve({ ok: false });
				}
				return Promise.resolve({
					ok: true,
					status: 200,
					json: () =>
						Promise.resolve({
							disciplines: [
								{
									name: "Climbing",
									years_practice: 5,
									level: "advanced",
									notes: null,
								},
							],
						}),
				});
			});
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("heading", { name: "Climbing" }),
				).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Remove" }));

			await waitFor(() =>
				expect(
					screen.getByRole("heading", { name: "Climbing" }),
				).toBeInTheDocument(),
			);
		});
	});

	describe("discipline level", () => {
		it("calls PATCH when level is selected", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "beginner",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Advanced" }),
				).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Advanced" }));

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: expect.stringContaining('"level":"advanced"'),
					}),
				),
			);
		});

		it("all four level options are visible", async () => {
			mockFetch({
				disciplines: [
					{ name: "Climbing", years_practice: 5, level: null, notes: null },
				],
			});
			renderPage();
			await waitFor(() =>
				expect(
					screen.getByRole("heading", { name: "Climbing" }),
				).toBeInTheDocument(),
			);

			expect(
				screen.getByRole("button", { name: "Beginner" }),
			).toBeInTheDocument();
			expect(
				screen.getByRole("button", { name: "Intermediate" }),
			).toBeInTheDocument();
			expect(
				screen.getByRole("button", { name: "Advanced" }),
			).toBeInTheDocument();
			expect(
				screen.getByRole("button", { name: "Expert" }),
			).toBeInTheDocument();
		});
	});

	describe("discipline name", () => {
		it("calls PATCH on name field blur", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByDisplayValue("Climbing")).toBeInTheDocument(),
			);

			const nameInput = screen.getByDisplayValue("Climbing");
			fireEvent.input(nameInput, { target: { value: "Bouldering" } });
			fireEvent.blur(nameInput);

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: expect.stringContaining('"name":"Bouldering"'),
					}),
				),
			);
		});
	});

	describe("discipline years of practice", () => {
		it("calls PATCH on years field blur", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByDisplayValue("5")).toBeInTheDocument(),
			);

			const yearsInput = screen.getByDisplayValue("5");
			fireEvent.input(yearsInput, { target: { value: "8" } });
			fireEvent.blur(yearsInput);

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: expect.stringContaining('"years_practice":8'),
					}),
				),
			);
		});

		it("sends null years_practice when field is cleared", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByDisplayValue("5")).toBeInTheDocument(),
			);

			const yearsInput = screen.getByDisplayValue("5");
			fireEvent.input(yearsInput, { target: { value: "" } });
			fireEvent.blur(yearsInput);

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: expect.stringContaining('"years_practice":null'),
					}),
				),
			);
		});
	});

	describe("discipline notes", () => {
		it("calls PATCH with updated notes when saved", async () => {
			const fetchSpy = mockFetch({
				disciplines: [
					{
						name: "Climbing",
						years_practice: 5,
						level: "advanced",
						notes: null,
					},
				],
			});
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Notes")).toBeInTheDocument(),
			);

			const notesContainer = screen.getByText("Notes").closest("div");
			fireEvent.click(
				within(notesContainer).getByRole("button", { name: "Edit" }),
			);
			fireEvent.input(within(notesContainer).getByRole("textbox"), {
				target: { value: "V5 bouldering" },
			});
			fireEvent.click(
				within(notesContainer).getByRole("button", { name: "Save" }),
			);

			await waitFor(() =>
				expect(fetchSpy).toHaveBeenCalledWith(
					"/api/users/me/training-profile",
					expect.objectContaining({
						method: "PATCH",
						body: expect.stringContaining('"notes":"V5 bouldering"'),
					}),
				),
			);
		});

		it("shows error when notes PATCH fails", async () => {
			mockFetch(
				{
					disciplines: [
						{
							name: "Climbing",
							years_practice: 5,
							level: "advanced",
							notes: null,
						},
					],
				},
				{ patchOk: false },
			);
			renderPage();
			await waitFor(() =>
				expect(screen.getByText("Notes")).toBeInTheDocument(),
			);

			const notesContainer = screen.getByText("Notes").closest("div");
			fireEvent.click(
				within(notesContainer).getByRole("button", { name: "Edit" }),
			);
			fireEvent.input(within(notesContainer).getByRole("textbox"), {
				target: { value: "Some notes" },
			});
			fireEvent.click(
				within(notesContainer).getByRole("button", { name: "Save" }),
			);

			await waitFor(() =>
				expect(screen.getByText("Failed to save.")).toBeInTheDocument(),
			);
		});
	});
});
