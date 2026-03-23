-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE TABLE cove.program_versions (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES cove.programs(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES cove.orgs(id),
    snapshot JSONB NOT NULL,
    created_by UUID NOT NULL REFERENCES cove.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_program_versions_list ON cove.program_versions (program_id, created_at DESC);
CREATE INDEX idx_program_versions_rls ON cove.program_versions (org_id);

-- Trigger to capture history on update
CREATE OR REPLACE FUNCTION cove.program_history_trigger() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO cove.program_versions (program_id, org_id, snapshot, created_by)
    VALUES (
        OLD.id,
        OLD.org_id,
        jsonb_build_object(
            'name', OLD.name,
            'description', OLD.description,
            'activity', OLD.activity,
            'is_public', OLD.is_public,
            'sets', OLD.sets
        ),
        cove.current_app_user_id()
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER program_history
BEFORE UPDATE ON cove.programs
FOR EACH ROW EXECUTE FUNCTION cove.program_history_trigger();

-- RLS
ALTER TABLE cove.program_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE cove.program_versions FORCE ROW LEVEL SECURITY;

CREATE POLICY program_versions_select ON cove.program_versions
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE POLICY program_versions_insert ON cove.program_versions
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY program_versions_delete ON cove.program_versions
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());
