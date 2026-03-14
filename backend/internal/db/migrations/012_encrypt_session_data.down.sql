-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE workout_sessions
    DROP COLUMN sensitive_data,
    ADD COLUMN perceived_effort   INTEGER,
    ADD COLUMN session_notes      TEXT,
    ADD COLUMN program_name       TEXT,
    ADD COLUMN program_structure  TEXT;
