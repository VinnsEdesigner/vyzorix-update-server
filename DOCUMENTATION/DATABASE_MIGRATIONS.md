# Database Migrations

Migrations run at startup. The server won't start if a migration fails. There are 59 migrations as of this writing, numbered 1 through 59.

## How they run

The `runMigrations` function in `internal/infrastructure/storage/sqlite.go` does this:

1. Creates the `schema_migrations` table if it doesn't exist
2. Reads the highest applied version
3. For each pending migration, in order:
   - Begins a transaction
   - Runs the migration's `Apply(tx)` function (which takes a `*sql.Tx`, not `*sql.DB`)
   - Records the version in `schema_migrations` (same transaction)
   - Commits
4. If any step fails, the transaction rolls back

This is important: each migration is atomic. If a migration creates a temp table, copies data, and renames — but the rename fails — the temp table creation and data copy roll back. The database is left exactly as it was before the migration started. No half-rebuilt schema.

## Why transactions matter

Before this fix, migrations ran with auto-commit. Each SQL statement committed independently. The table-rebuild migrations (049, 050, 052, 056) do `CREATE _new → INSERT INTO _new SELECT → DROP old → RENAME _new to old`. If step 3 failed, steps 1 and 2 were already committed. On restart, the migration tried to re-run (version wasn't recorded), but `CREATE TABLE IF NOT EXISTS _new` succeeded (already exists), and the copy step might fail or duplicate data. The server couldn't start. The database was effectively bricked.

The fix wraps each migration in `tx.Begin()` / `tx.Commit()`. SQLite WAL mode supports transactional DDL, so `CREATE`, `DROP`, and `ALTER TABLE RENAME` all roll back on failure.

## Migration functions

Each migration is a function with the signature `func(tx *sql.Tx) error`. They're registered in a slice in `sqlite.go`:

```go
var migrations = []Migration{
    {Apply: migrateCreateDevices, Name: "create_devices_table", Version: 1},
    // ... 57 more ...
    {Apply: migrateAuditLogsChangeTrackingColumns, Name: "add_audit_logs_change_tracking_columns", Version: 59},
}
```

Most migration functions live in separate files (`internal/infrastructure/storage/0XX_*.go`). A few live directly in `sqlite.go`.

## The table-rebuild pattern

Some migrations need to change column types or constraints that SQLite's `ALTER TABLE` doesn't support. They use the rebuild pattern:

1. Drop any leftover `_new` table from a previous failed run (`DROP TABLE IF EXISTS _new`)
2. Create `_new` with the desired schema
3. Copy data from the old table (`INSERT INTO _new SELECT ... FROM old`)
4. Drop the old table
5. Rename `_new` to the old table name
6. Recreate indexes

All six steps run inside a transaction. If step 4 fails, steps 1-3 roll back and the old table is untouched.

## Idempotent column additions

Migrations that add columns use `ALTER TABLE ADD COLUMN` with error checking:

```go
for _, col := range []string{"trace_id", "risk_tier"} {
    if _, err := tx.ExecContext(ctx, `ALTER TABLE audit_logs ADD COLUMN ` + col + ` TEXT`); err != nil {
        if !isDuplicateColumnErr(err) {
            return err
        }
    }
}
```

`isDuplicateColumnErr` checks for SQLite's "duplicate column name" error, which means the column already exists. This makes the migration idempotent — safe to re-run.

## Migrations added in this work

| # | Name | What it does |
|---|------|-------------|
| 57 | `add_audit_logs_risk_columns` | Adds `trace_id` and `risk_tier` to `audit_logs` |
| 58 | `create_command_confirmations_table` | Creates the `command_confirmations` table for confirmation tokens |
| 59 | `add_audit_logs_change_tracking_columns` | Adds `actor_type`, `actor_email`, `old_value`, `new_value` to `audit_logs` |
