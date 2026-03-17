-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Drop policies
DROP POLICY IF EXISTS exercises_select ON exercises;
DROP POLICY IF EXISTS exercises_insert ON exercises;
DROP POLICY IF EXISTS exercises_update ON exercises;
DROP POLICY IF EXISTS exercises_delete ON exercises;

-- Disable RLS
ALTER TABLE exercises DISABLE ROW LEVEL SECURITY;

-- Drop trigger
DROP TRIGGER IF EXISTS exercises_bookkeeping ON exercises;

-- Drop index
DROP INDEX IF EXISTS idx_exercises_rls;

-- Drop columns
ALTER TABLE exercises
DROP COLUMN IF EXISTS org_id,
DROP COLUMN IF EXISTS is_public,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS description;

-- Drop RLS helper functions
DROP FUNCTION IF EXISTS cove.update_bookkeeping_columns() CASCADE;
DROP FUNCTION IF EXISTS cove.current_app_org_id() CASCADE;
DROP FUNCTION IF EXISTS cove.current_app_user_id() CASCADE;
