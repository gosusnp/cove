// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { Check, ChevronDown } from "lucide-preact";
import { useEffect, useId, useRef } from "preact/hooks";
import { cn } from "../../lib/utils";

export function Combobox({
	label,
	value,
	onChange,
	options = [],
	placeholder = "Search...",
	disabled = false,
	freeform = false,
	class: className,
}) {
	const inputId = useId();
	const open = useSignal(false);
	const query = useSignal("");
	const highlightedIndex = useSignal(-1);
	const containerRef = useRef(null);
	const inputRef = useRef(null);

	const opts = options ?? [];
	const selectedOption =
		opts.find((o) => String(o.value) === String(value)) ?? null;

	const q = query.value.toLowerCase();
	const filtered =
		q === "" ? opts : opts.filter((o) => o.label.toLowerCase().includes(q));

	// When freeform, append a "Use '...'" entry if the query doesn't exactly match any option.
	const freeformEntry =
		freeform &&
		query.value.trim() !== "" &&
		!opts.some((o) => o.label.toLowerCase() === q)
			? {
					value: query.value.trim(),
					label: query.value.trim(),
					isFreeform: true,
				}
			: null;
	const displayOptions = freeformEntry
		? [...filtered, freeformEntry]
		: filtered;

	function openDropdown() {
		if (disabled || open.value) return;
		query.value = "";
		highlightedIndex.value = -1;
		open.value = true;
	}

	function closeDropdown() {
		open.value = false;
		query.value = "";
		highlightedIndex.value = -1;
	}

	function selectOption(opt) {
		onChange(opt.value);
		closeDropdown();
	}

	function handleInputChange(e) {
		query.value = e.target.value;
		highlightedIndex.value = -1;
		open.value = true;
	}

	function handleInputClick() {
		if (!open.value) openDropdown();
	}

	function handleKeyDown(e) {
		if (!open.value) {
			if (e.key === "ArrowDown" || e.key === "Enter") openDropdown();
			return;
		}
		const list = displayOptions;
		if (e.key === "ArrowDown") {
			e.preventDefault();
			highlightedIndex.value = Math.min(
				highlightedIndex.value + 1,
				list.length - 1,
			);
		} else if (e.key === "ArrowUp") {
			e.preventDefault();
			highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0);
		} else if (e.key === "Enter") {
			e.preventDefault();
			const idx = highlightedIndex.value;
			if (idx >= 0 && idx < list.length) {
				selectOption(list[idx]);
			} else if (freeform && query.value.trim()) {
				// Accept raw typed value when nothing is highlighted.
				onChange(query.value.trim());
				closeDropdown();
			}
		} else if (e.key === "Escape") {
			closeDropdown();
		}
	}

	// Close when focus leaves the container entirely.
	useEffect(() => {
		function handleDocumentClick(e) {
			if (!containerRef.current?.contains(e.target)) {
				closeDropdown();
			}
		}
		document.addEventListener("mousedown", handleDocumentClick);
		return () => document.removeEventListener("mousedown", handleDocumentClick);
	}, []);

	return (
		<div
			ref={containerRef}
			class={cn("relative flex flex-col gap-1.5", className)}
		>
			{label && (
				<label
					for={inputId}
					class="text-sm font-medium"
					style={{ color: "var(--color-text)" }}
				>
					{label}
				</label>
			)}
			<div class="relative">
				<input
					ref={inputRef}
					id={inputId}
					type="text"
					role="combobox"
					aria-expanded={open.value}
					aria-autocomplete="list"
					value={
						open.value
							? query.value
							: (selectedOption?.label ??
								(freeform && value ? String(value) : ""))
					}
					placeholder={selectedOption ? undefined : placeholder}
					disabled={disabled}
					onInput={handleInputChange}
					onClick={handleInputClick}
					onKeyDown={handleKeyDown}
					class={cn(
						"h-10 w-full rounded-lg border px-3 pr-8 text-sm",
						"bg-(--color-surface) text-(--color-text)",
						"border-(--color-border)",
						"placeholder:text-(--color-muted)",
						"transition-colors",
						"focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-(--color-accent)",
						"disabled:opacity-50 disabled:pointer-events-none",
					)}
				/>
				<span
					aria-hidden="true"
					class="pointer-events-none absolute inset-y-0 right-3 flex items-center"
					style={{ color: "var(--color-muted)" }}
				>
					<ChevronDown
						size={14}
						aria-hidden="true"
						style={{
							transform: open.value ? "rotate(180deg)" : "none",
							transition: "transform 0.15s",
						}}
					/>
				</span>
			</div>

			{/* Always rendered; hidden via display:none so clicks on options are never lost */}
			<div
				role="listbox"
				style={{
					display: open.value ? "block" : "none",
					position: "absolute",
					left: 0,
					right: 0,
					top: label ? "calc(100% - 0.375rem)" : "100%",
					zIndex: 50,
					marginTop: "0.25rem",
					maxHeight: "240px",
					overflowY: "auto",
					borderRadius: "8px",
					border: "1px solid var(--color-border)",
					background: "var(--color-surface)",
					boxShadow: "0 4px 16px rgba(0,0,0,0.10)",
				}}
			>
				{displayOptions.length === 0 ? (
					<div
						class="px-3 py-2 text-sm"
						style={{ color: "var(--color-muted)" }}
					>
						No results
					</div>
				) : (
					displayOptions.map((opt, idx) => {
						const isSelected =
							!opt.isFreeform && String(opt.value) === String(value);
						const isHighlighted = idx === highlightedIndex.value;
						return (
							<div
								key={opt.isFreeform ? `__freeform__${opt.value}` : opt.value}
								role="option"
								tabIndex={-1}
								aria-selected={isSelected}
								onClick={() => selectOption(opt)}
								onKeyDown={(e) => {
									if (e.key === "Enter" || e.key === " ") {
										e.preventDefault();
										selectOption(opt);
									}
								}}
								onMouseEnter={() => {
									highlightedIndex.value = idx;
								}}
								class={cn(
									"flex items-center gap-2 px-3 py-2 text-sm cursor-pointer",
									isHighlighted && "bg-(--color-accent)/10",
								)}
								style={
									isSelected
										? { color: "var(--color-accent)", fontWeight: 500 }
										: { color: "var(--color-text)" }
								}
							>
								<span class="w-4 shrink-0">
									{isSelected && <Check size={14} aria-hidden="true" />}
								</span>
								{opt.isFreeform ? (
									<span style={{ fontStyle: "italic" }}>
										Use &ldquo;{opt.label}&rdquo;
									</span>
								) : (
									opt.label
								)}
							</div>
						);
					})
				)}
			</div>
		</div>
	);
}
