-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE TABLE cove.preparations (
    id           BIGSERIAL PRIMARY KEY,
    name         CITEXT NOT NULL,
    description  TEXT,
    yield_amount NUMERIC(8,3) NOT NULL,
    yield_unit   TEXT NOT NULL,
    steps        JSONB NOT NULL DEFAULT '[]',
    is_public    BOOLEAN NOT NULL DEFAULT FALSE,
    org_id       UUID NOT NULL REFERENCES cove.orgs(id),
    created_by   UUID NOT NULL REFERENCES cove.users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by   UUID REFERENCES cove.users(id),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_preparations_rls ON cove.preparations (org_id, is_public);

CREATE TRIGGER preparations_bookkeeping
BEFORE INSERT OR UPDATE ON cove.preparations
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE cove.preparations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.preparations FORCE ROW LEVEL SECURITY;

CREATE POLICY preparations_select ON cove.preparations
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id() OR is_public = true);

CREATE POLICY preparations_insert ON cove.preparations
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY preparations_update ON cove.preparations
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY preparations_delete ON cove.preparations
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE TABLE cove.preparation_ingredients (
    id             BIGSERIAL PRIMARY KEY,
    preparation_id BIGINT NOT NULL REFERENCES cove.preparations(id) ON DELETE CASCADE,
    ingredient_id  BIGINT NOT NULL REFERENCES cove.ingredients(id),
    name           TEXT NOT NULL,
    amount         NUMERIC(8,3) NOT NULL,
    unit           TEXT NOT NULL,
    prep           TEXT,
    org_id         UUID NOT NULL REFERENCES cove.orgs(id),
    created_by     UUID NOT NULL REFERENCES cove.users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by     UUID REFERENCES cove.users(id),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_preparation_ingredients_preparation_id ON cove.preparation_ingredients (preparation_id);
CREATE INDEX idx_preparation_ingredients_rls ON cove.preparation_ingredients (org_id);

CREATE TRIGGER preparation_ingredients_bookkeeping
BEFORE INSERT OR UPDATE ON cove.preparation_ingredients
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE cove.preparation_ingredients ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.preparation_ingredients FORCE ROW LEVEL SECURITY;

CREATE POLICY preparation_ingredients_select ON cove.preparation_ingredients
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE POLICY preparation_ingredients_insert ON cove.preparation_ingredients
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY preparation_ingredients_update ON cove.preparation_ingredients
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY preparation_ingredients_delete ON cove.preparation_ingredients
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE TABLE cove.recipes (
    id           BIGSERIAL PRIMARY KEY,
    name         CITEXT NOT NULL,
    description  TEXT,
    yield_amount NUMERIC(8,3),
    yield_unit   TEXT,
    servings     INT NOT NULL,
    is_public    BOOLEAN NOT NULL DEFAULT FALSE,
    org_id       UUID NOT NULL REFERENCES cove.orgs(id),
    created_by   UUID NOT NULL REFERENCES cove.users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by   UUID REFERENCES cove.users(id),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recipes_rls ON cove.recipes (org_id, is_public);

CREATE TRIGGER recipes_bookkeeping
BEFORE INSERT OR UPDATE ON cove.recipes
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE cove.recipes ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.recipes FORCE ROW LEVEL SECURITY;

CREATE POLICY recipes_select ON cove.recipes
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id() OR is_public = true);

CREATE POLICY recipes_insert ON cove.recipes
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY recipes_update ON cove.recipes
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY recipes_delete ON cove.recipes
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE TABLE cove.recipe_preparations (
    id             BIGSERIAL PRIMARY KEY,
    recipe_id      BIGINT NOT NULL REFERENCES cove.recipes(id) ON DELETE CASCADE,
    preparation_id BIGINT NOT NULL REFERENCES cove.preparations(id),
    position       INT NOT NULL,
    amount         NUMERIC(8,3) NOT NULL,
    unit           TEXT NOT NULL,
    org_id         UUID NOT NULL REFERENCES cove.orgs(id),
    created_by     UUID NOT NULL REFERENCES cove.users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by     UUID REFERENCES cove.users(id),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recipe_preparations_recipe_id ON cove.recipe_preparations (recipe_id);
CREATE INDEX idx_recipe_preparations_rls ON cove.recipe_preparations (org_id);

CREATE TRIGGER recipe_preparations_bookkeeping
BEFORE INSERT OR UPDATE ON cove.recipe_preparations
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE cove.recipe_preparations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.recipe_preparations FORCE ROW LEVEL SECURITY;

CREATE POLICY recipe_preparations_select ON cove.recipe_preparations
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE POLICY recipe_preparations_insert ON cove.recipe_preparations
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY recipe_preparations_update ON cove.recipe_preparations
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY recipe_preparations_delete ON cove.recipe_preparations
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());
