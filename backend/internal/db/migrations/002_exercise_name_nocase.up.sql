-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

CREATE UNIQUE INDEX exercises_name_nocase ON exercises (name COLLATE NOCASE);
