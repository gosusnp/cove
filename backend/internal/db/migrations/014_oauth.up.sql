-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Dynamically registered OAuth 2.0 clients (public clients only, PKCE flow).
-- id is TEXT (not UUID) because client IDs are generated as UUID strings but
-- are opaque identifiers for external clients, not internal identity rows.
CREATE TABLE cove.oauth_clients (
    id            TEXT        PRIMARY KEY,
    name          TEXT        NOT NULL,
    redirect_uris TEXT        NOT NULL, -- JSON-encoded array
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Short-lived authorization codes exchanged for an access token.
CREATE TABLE cove.oauth_codes (
    code           TEXT        PRIMARY KEY,
    client_id      TEXT        NOT NULL REFERENCES cove.oauth_clients(id),
    user_id        UUID        NOT NULL REFERENCES cove.users(id),
    org_id         UUID        NOT NULL REFERENCES cove.orgs(id),
    redirect_uri   TEXT        NOT NULL,
    code_challenge TEXT        NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL,
    used_at        TIMESTAMPTZ
);

CREATE INDEX oauth_codes_expires_at_idx ON cove.oauth_codes (expires_at);
