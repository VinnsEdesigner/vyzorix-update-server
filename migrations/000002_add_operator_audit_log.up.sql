-- 000002_add_operator_audit_log.up.sql
-- Add audit log table for tracking operator actions
-- This migration is a template for future schema changes

-- Audit log table
CREATE TABLE IF NOT EXISTS operator_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operator_id TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    details TEXT,
    ip_address TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE SET NULL
);

-- Index for querying audit logs
CREATE INDEX IF NOT EXISTS idx_audit_operator ON operator_audit_log(operator_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON operator_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON operator_audit_log(created_at);