-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP TABLE IF EXISTS user_tokens;

CREATE TABLE user_sessions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id),
    token      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
