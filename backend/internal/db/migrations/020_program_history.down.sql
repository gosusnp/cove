-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

DROP TRIGGER IF EXISTS program_history ON cove.programs;
DROP FUNCTION IF EXISTS cove.program_history_trigger();
DROP TABLE IF EXISTS cove.program_versions;
