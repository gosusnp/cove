-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Add is_admin flag to users. Defaults to false; the application sets it to
-- true on the first real user created (see UpsertUser) and via the admin UI.
ALTER TABLE cove.users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;

-- Bootstrap: promote the sole existing non-service-account user to admin.
-- Safe to run on a fresh database (UPDATE affects 0 rows) or a single-user
-- install (UPDATE affects exactly 1 row).
UPDATE cove.users
SET is_admin = true
WHERE NOT is_service_account
  AND (SELECT COUNT(*) FROM cove.users WHERE NOT is_service_account) = 1;
