-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Recreate session_sets and session_exercises without data (rollback path only).
-- Uses BIGINT (not BIGSERIAL) since sequences were dropped with the tables.

CREATE TABLE session_sets (
    id BIGINT PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id),
    name TEXT,
    round_number INTEGER,
    sort_order INTEGER
);

CREATE TABLE session_exercises (
    id BIGINT PRIMARY KEY,
    session_set_id BIGINT NOT NULL REFERENCES session_sets(id),
    exercise_id BIGINT REFERENCES exercises(id),
    name TEXT,
    laterality TEXT,
    target_reps INTEGER,
    target_duration_seconds INTEGER,
    target_weight_kg NUMERIC(5,2),
    actual_reps INTEGER,
    actual_duration_seconds INTEGER,
    actual_weight_kg NUMERIC(5,2),
    sort_order INTEGER
);
