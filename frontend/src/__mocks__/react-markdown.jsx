// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// jsdom-compatible mock for react-markdown.
// Passes children through as text to avoid Preact compat issues in tests.
export default function ReactMarkdown({ children }) {
	return <span>{children}</span>;
}
