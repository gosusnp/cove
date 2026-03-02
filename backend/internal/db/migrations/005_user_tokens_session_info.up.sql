-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE user_tokens ADD COLUMN initial_ip_masked TEXT;
ALTER TABLE user_tokens ADD COLUMN initial_browser   TEXT;
ALTER TABLE user_tokens ADD COLUMN initial_os        TEXT;
ALTER TABLE user_tokens ADD COLUMN last_ip_masked    TEXT;
ALTER TABLE user_tokens ADD COLUMN last_browser      TEXT;
ALTER TABLE user_tokens ADD COLUMN last_os           TEXT;
