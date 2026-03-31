-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP FUNCTION IF EXISTS cove.current_app_is_service();
ALTER TABLE cove.user_tokens ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE cove.users DROP CONSTRAINT IF EXISTS users_identity_check;
ALTER TABLE cove.users DROP COLUMN IF EXISTS is_service_account;
ALTER TABLE cove.users ALTER COLUMN google_sub SET NOT NULL;
ALTER TABLE cove.users ALTER COLUMN email SET NOT NULL;
