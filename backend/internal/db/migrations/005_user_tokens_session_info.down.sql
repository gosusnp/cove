-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE user_tokens DROP COLUMN initial_ip_masked;
ALTER TABLE user_tokens DROP COLUMN initial_browser;
ALTER TABLE user_tokens DROP COLUMN initial_os;
ALTER TABLE user_tokens DROP COLUMN last_ip_masked;
ALTER TABLE user_tokens DROP COLUMN last_browser;
ALTER TABLE user_tokens DROP COLUMN last_os;
