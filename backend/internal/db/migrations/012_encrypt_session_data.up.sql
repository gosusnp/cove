-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Replace the four plain-text sensitive columns with a single encrypted payload.
ALTER TABLE workout_sessions
    DROP COLUMN perceived_effort,
    DROP COLUMN session_notes,
    DROP COLUMN program_name,
    DROP COLUMN program_structure,
    ADD COLUMN sensitive_data BYTEA;
