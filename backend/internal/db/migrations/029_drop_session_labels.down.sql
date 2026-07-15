-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

ALTER TABLE cove.workout_sessions
    ADD COLUMN labels TEXT[] NOT NULL DEFAULT '{}';
