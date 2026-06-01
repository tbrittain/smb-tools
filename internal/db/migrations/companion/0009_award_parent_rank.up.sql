-- Add parent_award_id and runner_up_rank to awards.
--
-- parent_award_id: the top-level award this award is a runner-up slot for.
--   NULL for primary awards (MVP, Cy Young, All-Star, etc.)
--   Non-NULL for runner-up awards (MVP-2 → MVP, Cy Young-2 → Cy Young, etc.)
--
-- runner_up_rank: ordinal rank among runner-ups of the same parent.
--   1 = best runner-up (MVP-2), 2 = next (MVP-3), etc.
--   NULL for primary (non-runner-up) awards.

ALTER TABLE awards ADD COLUMN parent_award_id INTEGER REFERENCES awards(id);
ALTER TABLE awards ADD COLUMN runner_up_rank  INTEGER;

-- MVP runner-ups.
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'MVP'), runner_up_rank = 1 WHERE name = 'MVP-2';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'MVP'), runner_up_rank = 2 WHERE name = 'MVP-3';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'MVP'), runner_up_rank = 3 WHERE name = 'MVP-4';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'MVP'), runner_up_rank = 4 WHERE name = 'MVP-5';

-- Cy Young runner-ups.
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'Cy Young'), runner_up_rank = 1 WHERE name = 'Cy Young-2';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'Cy Young'), runner_up_rank = 2 WHERE name = 'Cy Young-3';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'Cy Young'), runner_up_rank = 3 WHERE name = 'Cy Young-4';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'Cy Young'), runner_up_rank = 4 WHERE name = 'Cy Young-5';

-- ROY runner-ups.
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'ROY'), runner_up_rank = 1 WHERE name = 'ROY-2';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'ROY'), runner_up_rank = 2 WHERE name = 'ROY-3';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'ROY'), runner_up_rank = 3 WHERE name = 'ROY-4';
UPDATE awards SET parent_award_id = (SELECT id FROM awards WHERE name = 'ROY'), runner_up_rank = 4 WHERE name = 'ROY-5';
