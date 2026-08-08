-- Migration 000010 (down): remove the columns added by the up migration.

DROP INDEX IF EXISTS idx_audit_logs_status;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS duration_ms;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS error_message;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS status;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS actor_email;
