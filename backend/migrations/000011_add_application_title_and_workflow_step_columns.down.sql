-- Migration 000011 (down)

DROP INDEX IF EXISTS idx_workflow_steps_workflow_order;
ALTER TABLE applications DROP COLUMN IF EXISTS title;
