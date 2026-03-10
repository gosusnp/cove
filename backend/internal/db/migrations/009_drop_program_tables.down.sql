-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Recreate program_sets and program_exercises without data (rollback path only).
-- Uses BIGINT (not BIGSERIAL) since sequences were dropped in migration 008.

CREATE TABLE program_sets (
    id BIGINT PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES programs(id),
    name TEXT,
    rounds INTEGER DEFAULT 1,
    intra_set_rest_seconds INTEGER,
    sort_order INTEGER
);

CREATE TABLE program_exercises (
    id BIGINT PRIMARY KEY,
    program_set_id BIGINT NOT NULL REFERENCES program_sets(id),
    exercise_id BIGINT NOT NULL REFERENCES exercises(id),
    laterality TEXT,
    target_reps INTEGER,
    target_duration_seconds INTEGER,
    target_weight_kg NUMERIC(5,2),
    sort_order INTEGER
);

-- Restore the nullable FK reference columns on session tables
ALTER TABLE session_sets ADD COLUMN IF NOT EXISTS program_set_id BIGINT REFERENCES program_sets(id);
ALTER TABLE session_exercises ADD COLUMN IF NOT EXISTS program_exercise_id BIGINT REFERENCES program_exercises(id);
