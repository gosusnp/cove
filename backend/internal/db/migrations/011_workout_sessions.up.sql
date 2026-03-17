-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Replaces the old relational sessions table with a flat, snapshot-based model.
-- program_structure presence implicitly indicates a structured session; no
-- session_type discriminator is needed.

-- Drop the old sessions table (no application code references it yet).
DROP TABLE IF EXISTS sessions;

CREATE TABLE workout_sessions (
    id                 BIGSERIAL    PRIMARY KEY,
    org_id             UUID         NOT NULL REFERENCES orgs(id),
    user_id            UUID         NOT NULL REFERENCES users(id),
    program_id         BIGINT       REFERENCES programs(id),
    program_name       TEXT,
    program_structure  TEXT,
    activity           TEXT,
    duration_s         INTEGER,
    perceived_effort   INTEGER,
    session_notes      TEXT,
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    created_by         UUID         NOT NULL REFERENCES users(id),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by         UUID         REFERENCES users(id),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_sessions_org ON workout_sessions (org_id);

CREATE TRIGGER workout_sessions_bookkeeping
BEFORE INSERT OR UPDATE ON workout_sessions
FOR EACH ROW EXECUTE FUNCTION cove.update_bookkeeping_columns();

ALTER TABLE workout_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workout_sessions FORCE ROW LEVEL SECURITY;

CREATE POLICY workout_sessions_select ON workout_sessions
FOR SELECT TO PUBLIC
USING (org_id = cove.current_app_org_id());

CREATE POLICY workout_sessions_insert ON workout_sessions
FOR INSERT TO PUBLIC
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY workout_sessions_update ON workout_sessions
FOR UPDATE TO PUBLIC
USING (org_id = current_app_org_id())
WITH CHECK (org_id = cove.current_app_org_id());

CREATE POLICY workout_sessions_delete ON workout_sessions
FOR DELETE TO PUBLIC
USING (org_id = cove.current_app_org_id());
