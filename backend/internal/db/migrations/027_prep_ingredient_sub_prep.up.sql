-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.preparation_ingredients ALTER COLUMN ingredient_id DROP NOT NULL;

ALTER TABLE cove.preparation_ingredients
    ADD COLUMN preparation_ref_id BIGINT REFERENCES cove.preparations(id) ON DELETE RESTRICT;

ALTER TABLE cove.preparation_ingredients
    ADD CONSTRAINT prep_ingredient_exactly_one_ref
    CHECK ((ingredient_id IS NOT NULL)::int + (preparation_ref_id IS NOT NULL)::int = 1);

CREATE INDEX idx_prep_ingredients_prep_ref_id
    ON cove.preparation_ingredients (preparation_ref_id)
    WHERE preparation_ref_id IS NOT NULL;
