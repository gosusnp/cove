-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Drop FK constraints on sessions (columns are nullable soft-references)
ALTER TABLE session_sets DROP CONSTRAINT IF EXISTS session_sets_program_set_id_fkey;
ALTER TABLE session_exercises DROP CONSTRAINT IF EXISTS session_exercises_program_exercise_id_fkey;

-- Drop the now-orphaned reference columns from sessions
ALTER TABLE session_sets DROP COLUMN IF EXISTS program_set_id;
ALTER TABLE session_exercises DROP COLUMN IF EXISTS program_exercise_id;

-- Drop the tables (program_exercises first due to FK dependency)
DROP TABLE program_exercises;
DROP TABLE program_sets;
