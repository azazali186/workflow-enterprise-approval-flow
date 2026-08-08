-- Migration 000011: applications carry a human-readable title (submitted by
-- the client and used in list views); workflow_steps gets supporting indexes.

ALTER TABLE applications ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow_order
    ON workflow_steps (workflow_id, step_order);
