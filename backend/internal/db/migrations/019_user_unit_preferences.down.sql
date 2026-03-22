-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.users
    DROP COLUMN IF EXISTS fitness_unit_system,
    DROP COLUMN IF EXISTS cooking_unit_system;
