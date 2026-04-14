-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP INDEX IF EXISTS cove.idx_prep_ingredients_prep_ref_id;

ALTER TABLE cove.preparation_ingredients
    DROP CONSTRAINT IF EXISTS prep_ingredient_exactly_one_ref;

ALTER TABLE cove.preparation_ingredients
    DROP COLUMN IF EXISTS preparation_ref_id;

ALTER TABLE cove.preparation_ingredients
    ALTER COLUMN ingredient_id SET NOT NULL;
