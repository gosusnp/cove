-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Restore BIGSERIAL sequences for program_sets and program_exercises.

CREATE SEQUENCE IF NOT EXISTS program_sets_id_seq;
ALTER TABLE program_sets ALTER COLUMN id SET DEFAULT nextval('program_sets_id_seq');
ALTER SEQUENCE program_sets_id_seq OWNED BY program_sets.id;
SELECT setval('program_sets_id_seq', COALESCE((SELECT MAX(id) FROM program_sets), 0) + 1, false);

CREATE SEQUENCE IF NOT EXISTS program_exercises_id_seq;
ALTER TABLE program_exercises ALTER COLUMN id SET DEFAULT nextval('program_exercises_id_seq');
ALTER SEQUENCE program_exercises_id_seq OWNED BY program_exercises.id;
SELECT setval('program_exercises_id_seq', COALESCE((SELECT MAX(id) FROM program_exercises), 0) + 1, false);
