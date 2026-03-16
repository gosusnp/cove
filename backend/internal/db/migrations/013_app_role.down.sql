-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'cove_app_role') THEN
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM cove_app_role';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE USAGE, SELECT ON SEQUENCES FROM cove_app_role';
        REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM cove_app_role;
        REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM cove_app_role;
        DROP ROLE cove_app_role;
    END IF;
EXCEPTION
    WHEN insufficient_privilege THEN NULL;
END $$;
