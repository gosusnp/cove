-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE TABLE user_training_profiles (
    user_id            UUID         PRIMARY KEY REFERENCES users(id),
    org_id             UUID         NOT NULL REFERENCES orgs(id),
    training_profile   BYTEA,
    created_by         UUID         NOT NULL REFERENCES users(id),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by         UUID         REFERENCES users(id),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TRIGGER user_training_profiles_bookkeeping
BEFORE INSERT OR UPDATE ON user_training_profiles
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE user_training_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_training_profiles FORCE ROW LEVEL SECURITY;

CREATE POLICY user_training_profiles_select ON user_training_profiles
FOR SELECT TO PUBLIC
USING (user_id = cove.current_app_user_id() OR (org_id = cove.current_app_org_id() AND cove.current_app_is_service()));

CREATE POLICY user_training_profiles_insert ON user_training_profiles
FOR INSERT TO PUBLIC
WITH CHECK (user_id = cove.current_app_user_id() AND org_id = cove.current_app_org_id());

CREATE POLICY user_training_profiles_update ON user_training_profiles
FOR UPDATE TO PUBLIC
USING (user_id = cove.current_app_user_id() OR (org_id = cove.current_app_org_id() AND cove.current_app_is_service()))
WITH CHECK (user_id = cove.current_app_user_id() OR (org_id = cove.current_app_org_id() AND cove.current_app_is_service()));

CREATE POLICY user_training_profiles_delete ON user_training_profiles
FOR DELETE TO PUBLIC
USING (user_id = cove.current_app_user_id());
