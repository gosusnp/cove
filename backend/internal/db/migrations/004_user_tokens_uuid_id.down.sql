-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE user_tokens DROP COLUMN id;
ALTER TABLE user_tokens ADD COLUMN id BIGSERIAL PRIMARY KEY;
