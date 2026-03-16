-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Create a group role that holds all runtime permissions for the app user.
-- The actual login user (e.g. cove_app) is granted this role outside of
-- migrations, at the infrastructure level (e.g. CNPG managed.roles).
--
-- Role creation requires CREATEROLE privilege (held by cove_migrator in production).
-- In dev/test environments running as an unprivileged user, the role creation is
-- silently skipped — it is a no-op in those contexts.
--
-- Schema-specific grants are applied by db.Migrate() after migrations run,
-- so they are always issued against the correct schema name.
DO $$
BEGIN
    CREATE ROLE cove_app_role;
EXCEPTION
    WHEN duplicate_object THEN NULL;
    WHEN insufficient_privilege THEN NULL;
END $$;
