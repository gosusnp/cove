-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- -----------------------------------------------------------------------------
-- RLS Helpers
-- -----------------------------------------------------------------------------

-- Function to get current user ID from session
CREATE OR REPLACE FUNCTION current_app_user_id() RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_user_id', true), '')::UUID;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to get current org ID from session
CREATE OR REPLACE FUNCTION current_app_org_id() RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_org_id', true), '')::UUID;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Trigger function to update bookkeeping columns
CREATE OR REPLACE FUNCTION update_bookkeeping_columns() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.updated_by = COALESCE(NULLIF(current_app_user_id(), NULL), NEW.updated_by);
    
    -- On INSERT, set created columns from session if not provided
    IF (TG_OP = 'INSERT') THEN
        NEW.created_at = COALESCE(NEW.created_at, NOW());
        NEW.created_by = COALESCE(NEW.created_by, current_app_user_id());
        NEW.org_id = COALESCE(NEW.org_id, current_app_org_id());
    END IF;
    
    -- Validation: Ensure required fields are present
    IF NEW.org_id IS NULL THEN
        RAISE EXCEPTION 'org_id is required';
    END IF;
    IF NEW.created_by IS NULL THEN
        RAISE EXCEPTION 'created_by is required';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------------------
-- Exercise Table Upgrades
-- -----------------------------------------------------------------------------

ALTER TABLE exercises
ADD COLUMN org_id     UUID NOT NULL REFERENCES orgs(id),
ADD COLUMN is_public  BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN created_by UUID NOT NULL REFERENCES users(id),
ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN updated_by UUID REFERENCES users(id),
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN description TEXT;

-- Remove unique constraint on name to support multi-tenancy
ALTER TABLE exercises DROP CONSTRAINT IF EXISTS exercises_name_key;

CREATE INDEX idx_exercises_rls ON exercises (org_id, is_public);

-- -----------------------------------------------------------------------------
-- Triggers
-- -----------------------------------------------------------------------------

CREATE TRIGGER exercises_bookkeeping
BEFORE INSERT OR UPDATE ON exercises
FOR EACH ROW EXECUTE FUNCTION update_bookkeeping_columns();

-- -----------------------------------------------------------------------------
-- RLS Policies
-- -----------------------------------------------------------------------------

ALTER TABLE exercises ENABLE ROW LEVEL SECURITY;
ALTER TABLE exercises FORCE ROW LEVEL SECURITY;

-- Select: users can see exercises from their org or public ones
CREATE POLICY exercises_select ON exercises
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = current_app_org_id());

-- Insert: strictly same org
CREATE POLICY exercises_insert ON exercises
FOR INSERT TO PUBLIC
WITH CHECK (org_id = current_app_org_id());

-- Update: strictly same org
CREATE POLICY exercises_update ON exercises
FOR UPDATE TO PUBLIC
USING (org_id = current_app_org_id())
WITH CHECK (org_id = current_app_org_id());

-- Delete: strictly same org
CREATE POLICY exercises_delete ON exercises
FOR DELETE TO PUBLIC
USING (org_id = current_app_org_id());
