-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- -----------------------------------------------------------------------------
-- Program Table Upgrades
-- -----------------------------------------------------------------------------

ALTER TABLE programs
ADD COLUMN org_id     UUID NOT NULL REFERENCES orgs(id),
ADD COLUMN is_public  BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN created_by UUID NOT NULL REFERENCES users(id),
ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN updated_by UUID REFERENCES users(id),
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN description TEXT,
ADD COLUMN sets JSONB NOT NULL DEFAULT '[]',
ADD COLUMN exercise_ids BIGINT[] NOT NULL DEFAULT '{}';

-- Remove unique constraint on name to support multi-tenancy
ALTER TABLE programs DROP CONSTRAINT IF EXISTS programs_name_key;

CREATE INDEX idx_programs_rls ON programs (org_id, is_public);

-- -----------------------------------------------------------------------------
-- Triggers
-- -----------------------------------------------------------------------------

CREATE TRIGGER programs_bookkeeping
BEFORE INSERT OR UPDATE ON programs
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

-- -----------------------------------------------------------------------------
-- RLS Policies
-- -----------------------------------------------------------------------------

ALTER TABLE programs ENABLE ROW LEVEL SECURITY;
ALTER TABLE programs FORCE ROW LEVEL SECURITY;

-- Select: users can see programs from their org or public ones
CREATE POLICY programs_select ON programs
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = cove.current_app_org_id());

-- Insert: strictly same org
CREATE POLICY programs_insert ON programs
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

-- Update: strictly same org
CREATE POLICY programs_update ON programs
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- Delete: strictly same org
CREATE POLICY programs_delete ON programs
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());
