-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE TABLE exercises (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    progression TEXT
);

CREATE TABLE programs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE program_sets (
    id INTEGER PRIMARY KEY,
    program_id INTEGER NOT NULL REFERENCES programs(id),
    name TEXT,
    rounds INTEGER DEFAULT 1,
    intra_set_rest_seconds INTEGER,
    sort_order INTEGER
);

CREATE TABLE program_exercises (
    id INTEGER PRIMARY KEY,
    program_set_id INTEGER NOT NULL REFERENCES program_sets(id),
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    laterality TEXT,
    target_reps INTEGER,
    target_duration_seconds INTEGER,
    target_weight_kg REAL,
    sort_order INTEGER
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    program_id INTEGER REFERENCES programs(id), -- this is for historical purpose, not enforcing FK on by design
    perceived_effort INTEGER,
    notes TEXT,
    started_at TEXT,
    completed_at TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE session_sets (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES sessions(id),
    program_set_id INTEGER REFERENCES program_sets(id), -- this is for historical purpose, not enforcing FK on by design
    name TEXT,
    round_number INTEGER,
    sort_order INTEGER
);

CREATE TABLE session_exercises (
    id INTEGER PRIMARY KEY,
    session_set_id INTEGER NOT NULL REFERENCES session_sets(id),
    program_exercise_id INTEGER REFERENCES program_exercises(id), -- this is for historical purpose, not enforcing FK on by design
    exercise_id INTEGER REFERENCES exercises(id), -- this is for historical purpose, not enforcing FK on by design
    name TEXT,
    laterality TEXT,
    target_reps INTEGER,
    target_duration_seconds INTEGER,
    target_weight_kg REAL,
    actual_reps INTEGER,
    actual_duration_seconds INTEGER,
    actual_weight_kg REAL,
    sort_order INTEGER
);
