-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE exercises (
    id BIGSERIAL PRIMARY KEY,
    name CITEXT NOT NULL UNIQUE,
    progression TEXT
);

CREATE TABLE programs (
    id BIGSERIAL PRIMARY KEY,
    name CITEXT NOT NULL UNIQUE
);

CREATE TABLE program_sets (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES programs(id),
    name TEXT,
    rounds INTEGER DEFAULT 1,
    intra_set_rest_seconds INTEGER,
    sort_order INTEGER
);

CREATE TABLE program_exercises (
    id BIGSERIAL PRIMARY KEY,
    program_set_id BIGINT NOT NULL REFERENCES program_sets(id),
    exercise_id BIGINT NOT NULL REFERENCES exercises(id),
    laterality TEXT,
    target_reps INTEGER,
    target_duration_seconds INTEGER,
    target_weight_kg NUMERIC(5,2),
    sort_order INTEGER
);

CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT REFERENCES programs(id),
    perceived_effort INTEGER,
    notes TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE session_sets (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id),
    program_set_id BIGINT REFERENCES program_sets(id),
    name TEXT,
    round_number INTEGER,
    sort_order INTEGER
);

CREATE TABLE session_exercises (
    id BIGSERIAL PRIMARY KEY,
    session_set_id BIGINT NOT NULL REFERENCES session_sets(id),
    program_exercise_id BIGINT REFERENCES program_exercises(id),
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
