-- 000002_add_operator_audit_log.down.sql
-- Rollback: Remove audit log table

DROP INDEX IF EXISTS idx_audit_created;
DROP INDEX IF EXISTS idx_audit_action;
DROP INDEX IF EXISTS idx_audit_operator;

DROP TABLE IF EXISTS operator_audit_log;