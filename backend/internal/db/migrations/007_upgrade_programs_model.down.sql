-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- 1. Drop triggers
DROP TRIGGER IF EXISTS programs_bookkeeping ON programs;

-- 2. Drop RLS policies
ALTER TABLE programs DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS programs_select ON programs;
DROP POLICY IF EXISTS programs_insert ON programs;
DROP POLICY IF EXISTS programs_update ON programs;
DROP POLICY IF EXISTS programs_delete ON programs;

-- 3. Drop indexes
DROP INDEX IF EXISTS idx_programs_rls;

-- 4. Drop columns
ALTER TABLE programs
DROP COLUMN IF EXISTS org_id CASCADE,
DROP COLUMN IF EXISTS is_public CASCADE,
DROP COLUMN IF EXISTS created_by CASCADE,
DROP COLUMN IF EXISTS created_at CASCADE,
DROP COLUMN IF EXISTS updated_by CASCADE,
DROP COLUMN IF EXISTS updated_at CASCADE,
DROP COLUMN IF EXISTS description CASCADE,
DROP COLUMN IF EXISTS sets CASCADE,
DROP COLUMN IF EXISTS exercise_ids CASCADE;

-- 5. Restore original state
ALTER TABLE programs ADD CONSTRAINT programs_name_key UNIQUE (name);
