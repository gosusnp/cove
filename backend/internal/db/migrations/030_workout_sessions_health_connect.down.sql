-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.workout_sessions
    DROP COLUMN health_connect_id,
    DROP COLUMN source,
    DROP COLUMN source_activity;
