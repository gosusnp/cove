-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Revert: rename weight -> weight_kg and remove weight_unit from exercises in the JSONB column.
UPDATE cove.programs
SET sets = (
    SELECT jsonb_agg(
        jsonb_set(
            s,
            '{exercises}',
            (
                SELECT jsonb_agg(
                    CASE
                        WHEN e ? 'weight' THEN
                            (e - 'weight' - 'weight_unit') || jsonb_build_object('weight_kg', e->'weight')
                        ELSE e
                    END
                )
                FROM jsonb_array_elements(s->'exercises') e
            )
        )
    )
    FROM jsonb_array_elements(sets) s
)
WHERE sets IS NOT NULL AND jsonb_array_length(sets) > 0;
