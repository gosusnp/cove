-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.users
    ADD COLUMN display_name TEXT,
    ADD COLUMN first_name   TEXT,
    ADD COLUMN last_name    TEXT;
