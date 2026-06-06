-- 000001_init_schema.down.sql
-- Rollback initial database schema

DROP INDEX IF EXISTS idx_devices_online;
DROP INDEX IF EXISTS idx_telemetry_received;
DROP INDEX IF EXISTS idx_telemetry_device;
DROP INDEX IF EXISTS idx_commands_status;
DROP INDEX IF EXISTS idx_commands_device;
DROP INDEX IF EXISTS idx_auth_sessions_token;
DROP INDEX IF EXISTS idx_auth_sessions_operator;

DROP TABLE IF EXISTS telemetry;
DROP TABLE IF EXISTS commands;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS operators;