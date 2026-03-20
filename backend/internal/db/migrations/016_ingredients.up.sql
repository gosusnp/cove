-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE TABLE cove.ingredients (
    id                BIGSERIAL PRIMARY KEY,
    name              CITEXT NOT NULL,
    fdc_id            INT,
    calories_per_100g NUMERIC(8,2) NOT NULL,
    protein_per_100g  NUMERIC(6,2) NOT NULL,
    fat_per_100g      NUMERIC(6,2) NOT NULL,
    carbs_per_100g    NUMERIC(6,2) NOT NULL,
    density_g_per_ml  NUMERIC(5,3),
    org_id            UUID NOT NULL REFERENCES cove.orgs(id),
    is_public         BOOLEAN NOT NULL DEFAULT FALSE,
    created_by        UUID NOT NULL REFERENCES cove.users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by        UUID REFERENCES cove.users(id),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingredients_rls ON cove.ingredients (org_id, is_public);

CREATE TRIGGER ingredients_bookkeeping
BEFORE INSERT OR UPDATE ON cove.ingredients
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE cove.ingredients ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.ingredients FORCE ROW LEVEL SECURITY;

CREATE POLICY ingredients_select ON cove.ingredients
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = cove.current_app_org_id());

CREATE POLICY ingredients_insert ON cove.ingredients
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY ingredients_update ON cove.ingredients
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY ingredients_delete ON cove.ingredients
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());
