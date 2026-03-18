// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import ReactMarkdown from "react-markdown";
import { cn } from "../../lib/utils";

/**
 * Renders a markdown string as styled HTML.
 * Applies consistent typography using design tokens.
 */
export function Markdown({ children, class: className }) {
	return (
		<div class={cn("markdown-body text-sm", className)}>
			<ReactMarkdown>{children}</ReactMarkdown>
		</div>
	);
}
