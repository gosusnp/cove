-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Drop the outdated relational session tables replaced by the Markdown snapshot model.
-- session_exercises references session_sets, so it must be dropped first.
DROP TABLE session_exercises;
DROP TABLE session_sets;
