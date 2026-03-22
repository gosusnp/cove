-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.users
    ADD COLUMN fitness_unit_system TEXT,
    ADD COLUMN cooking_unit_system TEXT;
