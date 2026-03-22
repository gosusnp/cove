-- Copyright (c) 2026 Jimmy Ma
-- SPDX-License-Identifier: Elastic-2.0

-- Migrate program exercises stored in the programs.sets JSONB column:
-- rename weight_kg -> weight and add weight_unit = 'kg' for existing rows.
UPDATE cove.programs
SET sets = (
    SELECT jsonb_agg(
        jsonb_set(
            s,
            '{exercises}',
            (
                SELECT jsonb_agg(
                    CASE
                        WHEN e ? 'weight_kg' THEN
                            jsonb_set(
                                jsonb_set(e - 'weight_kg', '{weight}', e->'weight_kg'),
                                '{weight_unit}',
                                '"kg"'
                            )
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
