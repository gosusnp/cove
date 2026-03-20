-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.programs ADD COLUMN activity CITEXT CHECK (activity <> '');
