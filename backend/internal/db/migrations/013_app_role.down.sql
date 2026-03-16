-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'cove_app_role') THEN
        DROP ROLE cove_app_role;
    END IF;
EXCEPTION
    WHEN insufficient_privilege THEN NULL;
END $$;
