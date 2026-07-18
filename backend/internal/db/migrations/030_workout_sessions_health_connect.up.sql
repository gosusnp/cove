-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.workout_sessions
    ADD COLUMN health_connect_id TEXT UNIQUE,
    ADD COLUMN source             TEXT NOT NULL DEFAULT 'cove',
    ADD COLUMN source_activity    TEXT;
