-- 000001_init_schema.up.sql
-- Initial database schema for Vyzorix Update Server

-- Operators table (dashboard users)
CREATE TABLE IF NOT EXISTS operators (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT,
    role TEXT NOT NULL DEFAULT 'operator',
    email_verified INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Auth sessions table (JWT tracking)
CREATE TABLE IF NOT EXISTS auth_sessions (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

-- Email verifications table
CREATE TABLE IF NOT EXISTS email_verifications (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

-- Password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

-- Devices table (Android devices)
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    firebase_install_id TEXT NOT NULL,
    fcm_token TEXT,
    app_version TEXT,
    device_class TEXT,
    command_secret TEXT NOT NULL,
    command_secret_hash TEXT,
    online INTEGER NOT NULL DEFAULT 0,
    registered_at INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

-- Commands table (device commands)
CREATE TABLE IF NOT EXISTS commands (
    dispatch_id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    command TEXT NOT NULL,
    args TEXT,
    delivery TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    delivered_at INTEGER,
    status TEXT NOT NULL DEFAULT 'pending',
    wake_sent INTEGER,
    wake_error TEXT,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- Telemetry table (device telemetry data)
CREATE TABLE IF NOT EXISTS telemetry (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    received_at INTEGER NOT NULL,
    payload TEXT,
    risk_score INTEGER,
    buffer_level INTEGER,
    thermal_temp REAL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_auth_sessions_operator ON auth_sessions(operator_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_token ON auth_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_commands_device ON commands(device_id);
CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);
CREATE INDEX IF NOT EXISTS idx_telemetry_device ON telemetry(device_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_received ON telemetry(received_at);
CREATE INDEX IF NOT EXISTS idx_devices_online ON devices(online);