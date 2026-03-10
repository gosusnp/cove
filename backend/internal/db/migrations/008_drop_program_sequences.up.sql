-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Change program_sets.id and program_exercises.id from BIGSERIAL to BIGINT.
-- IDs are now generated in Go by reading the current max from the JSONB column.

ALTER TABLE program_sets ALTER COLUMN id SET DEFAULT 0;
ALTER TABLE program_sets ALTER COLUMN id DROP DEFAULT;

ALTER TABLE program_exercises ALTER COLUMN id SET DEFAULT 0;
ALTER TABLE program_exercises ALTER COLUMN id DROP DEFAULT;

DROP SEQUENCE IF EXISTS program_sets_id_seq;
DROP SEQUENCE IF EXISTS program_exercises_id_seq;
