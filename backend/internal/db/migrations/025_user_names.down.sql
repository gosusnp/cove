-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.users
    DROP COLUMN display_name,
    DROP COLUMN first_name,
    DROP COLUMN last_name;
