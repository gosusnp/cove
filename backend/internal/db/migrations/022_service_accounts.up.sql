-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Make email and google_sub nullable so service accounts (which have neither) can exist.
ALTER TABLE cove.users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE cove.users ALTER COLUMN google_sub DROP NOT NULL;

-- Service account flag and invariant: every row is either a service account or has an email.
ALTER TABLE cove.users ADD COLUMN is_service_account BOOLEAN NOT NULL DEFAULT false;
-- Mutual exclusion: service accounts must have no email; regular users must have one.
-- Prevents a row from being both (gaining OAuth access) or neither.
ALTER TABLE cove.users ADD CONSTRAINT users_identity_check
    CHECK (is_service_account = (email IS NULL));

-- Service account tokens have no org scope; drop the NOT NULL constraint on org_id.
ALTER TABLE cove.user_tokens ALTER COLUMN org_id DROP NOT NULL;

-- Helper function read by withScopedTx and used by RLS policies in task 233.
-- Returns true when the current transaction was opened by a service account.
CREATE OR REPLACE FUNCTION cove.current_app_is_service() RETURNS BOOLEAN AS $$
BEGIN
    RETURN COALESCE(NULLIF(current_setting('app.current_is_service', true), '')::BOOLEAN, false);
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$$ LANGUAGE plpgsql STABLE;
