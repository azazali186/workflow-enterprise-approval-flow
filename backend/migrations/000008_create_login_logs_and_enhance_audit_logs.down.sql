-- Migration 000008: Drop login_logs and revert audit_logs

ALTER TABLE audit_logs DROP COLUMN IF EXISTS duration_ms;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS error_message;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS status;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS request_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS after_state;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS before_state;

DROP TABLE IF EXISTS login_logs;
