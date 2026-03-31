-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Revert service account RLS extensions for all tenant-owned tables.

-- workout_sessions
DROP POLICY workout_sessions_select ON cove.workout_sessions;
CREATE POLICY workout_sessions_select ON cove.workout_sessions
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

DROP POLICY workout_sessions_update ON cove.workout_sessions;
CREATE POLICY workout_sessions_update ON cove.workout_sessions
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- exercises
DROP POLICY exercises_select ON cove.exercises;
CREATE POLICY exercises_select ON cove.exercises
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = cove.current_app_org_id());

DROP POLICY exercises_update ON cove.exercises;
CREATE POLICY exercises_update ON cove.exercises
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- programs
DROP POLICY programs_select ON cove.programs;
CREATE POLICY programs_select ON cove.programs
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = cove.current_app_org_id());

DROP POLICY programs_update ON cove.programs;
CREATE POLICY programs_update ON cove.programs
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- ingredients
DROP POLICY ingredients_select ON cove.ingredients;
CREATE POLICY ingredients_select ON cove.ingredients
FOR SELECT TO PUBLIC
USING (is_public = true OR org_id = cove.current_app_org_id());

DROP POLICY ingredients_update ON cove.ingredients;
CREATE POLICY ingredients_update ON cove.ingredients
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- preparations
DROP POLICY preparations_select ON cove.preparations;
CREATE POLICY preparations_select ON cove.preparations
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id() OR is_public = true);

DROP POLICY preparations_update ON cove.preparations;
CREATE POLICY preparations_update ON cove.preparations
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- preparation_ingredients
DROP POLICY preparation_ingredients_select ON cove.preparation_ingredients;
CREATE POLICY preparation_ingredients_select ON cove.preparation_ingredients
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

DROP POLICY preparation_ingredients_update ON cove.preparation_ingredients;
CREATE POLICY preparation_ingredients_update ON cove.preparation_ingredients
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- recipes
DROP POLICY recipes_select ON cove.recipes;
CREATE POLICY recipes_select ON cove.recipes
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id() OR is_public = true);

DROP POLICY recipes_update ON cove.recipes;
CREATE POLICY recipes_update ON cove.recipes
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- recipe_preparations
DROP POLICY recipe_preparations_select ON cove.recipe_preparations;
CREATE POLICY recipe_preparations_select ON cove.recipe_preparations
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

DROP POLICY recipe_preparations_update ON cove.recipe_preparations;
CREATE POLICY recipe_preparations_update ON cove.recipe_preparations
FOR UPDATE TO PUBLIC
USING (org_id = cove.current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

-- program_versions
DROP POLICY program_versions_select ON cove.program_versions;
CREATE POLICY program_versions_select ON cove.program_versions
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());
