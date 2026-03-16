-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Create a group role that holds all runtime permissions for the app user.
-- The actual login user (e.g. cove_app) is granted this role outside of
-- migrations, at the infrastructure level (e.g. CNPG managed.roles).
--
-- Role creation requires CREATEROLE privilege (held by cove_migrator in production).
-- In dev/test environments running as an unprivileged user, the role creation and
-- subsequent grants are silently skipped — they are no-ops in those contexts.
DO $$
BEGIN
    CREATE ROLE cove_app_role;
EXCEPTION
    WHEN duplicate_object THEN NULL;
    WHEN insufficient_privilege THEN NULL;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'cove_app_role') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO cove_app_role;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cove_app_role;
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO cove_app_role';
        EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO cove_app_role';
    END IF;
END $$;
