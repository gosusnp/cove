-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP TABLE IF EXISTS workout_sessions;

-- Restore the old sessions table (schema only; no data migration on rollback).
CREATE TABLE sessions (
    id               BIGSERIAL    PRIMARY KEY,
    program_id       BIGINT       REFERENCES programs(id),
    perceived_effort INTEGER,
    notes            TEXT,
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  DEFAULT NOW()
);
