-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP TABLE IF EXISTS user_sessions;

CREATE TABLE user_tokens (
    id           BIGSERIAL   PRIMARY KEY,
    user_id      UUID        NOT NULL REFERENCES users(id),
    org_id       UUID        NOT NULL REFERENCES orgs(id),
    kind         TEXT        NOT NULL, -- 'session' | 'pat'
    name         TEXT,                 -- NULL for sessions, required for PATs
    token        TEXT        NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,          -- NULL = never expires (PATs only)
    last_used_at TIMESTAMPTZ
);

CREATE INDEX user_tokens_expires_at_idx ON user_tokens (expires_at)
    WHERE expires_at IS NOT NULL;
