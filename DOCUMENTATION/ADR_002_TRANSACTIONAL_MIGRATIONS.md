# ADR-002: Transactional DDL Migrations in SQLite

**Status**: Accepted  
**Date**: 2026-08-18  

## Context

The migration system in `internal/infrastructure/storage/sqlite.go` originally ran each migration with auto-commit. The code literally had a comment: "Run migration directly (SQLite auto-commits each statement)." Each SQL statement inside a migration committed independently.

This caused a critical bug: the database could get "bricked" — stuck in a half-rebuilt schema state that prevented the server from starting.

### How the bricking happened

Three migrations (049, 050, 056) use the table-rebuild pattern to change column types or constraints that SQLite's `ALTER TABLE` doesn't support:

```
1. CREATE TABLE _new (...)
2. INSERT INTO _new SELECT ... FROM old
3. DROP TABLE old
4. ALTER TABLE _new RENAME TO old
```

With auto-commit, if step 2 or 3 failed (e.g., a CHECK constraint violation on existing data, or a column mismatch), steps 1 and 2 were already committed to the database. The `_new` table existed, data was copied into it, but the rename never happened.

On restart, the migration tried to re-run (the version wasn't recorded because the migration failed). `CREATE TABLE IF NOT EXISTS _new` succeeded (table already exists). `INSERT INTO _new SELECT ... FROM old` either failed again (same data issue) or succeeded (duplicating data). The migration couldn't complete. The version was never recorded. Every restart attempted the same migration and failed. The server couldn't start.

The only fix was to manually delete the `_new` table in the database file or reset the database entirely.

### Why this happened

SQLite is often thought of as not supporting transactional DDL. This is wrong — SQLite has supported transactional DDL (CREATE, DROP, ALTER inside a BEGIN/COMMIT) since version 3.25 (2018). In WAL mode (which Vyzorix uses), DDL statements are fully transactional. They roll back on failure just like DML.

The original code simply didn't use transactions for migrations. It wasn't a SQLite limitation — it was an implementation choice that didn't account for the table-rebuild pattern's failure modes.

## Decision

Wrap each migration in a database transaction:

```go
tx, err := db.Begin()
if err != nil {
    return fmt.Errorf("migration %d failed to begin tx: %w", m.Version, err)
}

if err := m.Apply(tx); err != nil {
    _ = tx.Rollback()
    return fmt.Errorf("migration %d failed: %w", m.Version, m.Name, err)
}

if err := setVersionTx(tx, m.Version); err != nil {
    _ = tx.Rollback()
    return fmt.Errorf("failed to set version: %w", err)
}

if err := tx.Commit(); err != nil {
    return fmt.Errorf("migration %d failed to commit: %w", m.Version, m.Name, err)
}
```

Each migration function now takes a `*sql.Tx` instead of `*sql.DB`. All 33 migration functions were converted from `func(db *sql.DB) error` to `func(tx *sql.Tx) error`.

Additionally, each table-rebuild migration now starts with:

```go
_, _ = tx.ExecContext(ctx, "DROP TABLE IF EXISTS _new")
```

This cleans up any leftover `_new` table from a previous failed run, making the migration safe to retry.

### Why not use a migration framework?

We considered golang-migrate, goose, and atlas. Reasons we didn't:

1. **Existing migration system works.** The 33 existing migrations were already written as Go functions. Porting them to a framework's format (SQL files, embedded migrations) would be a large refactor with no functional gain.

2. **The fix is small.** Adding `tx.Begin()` / `tx.Commit()` around each migration is a 10-line change. The migration functions already used `db.ExecContext` — changing `db` to `tx` was mechanical (a sed across 33 files).

3. **SQLite-specific behavior.** Frameworks often assume PostgreSQL/MySQL behavior where DDL is transactional by default. SQLite's transactional DDL depends on WAL mode, which is configured at the connection level. The migration code knows about WAL; a framework might not.

4. **Idempotency patterns.** The existing migrations use Go logic (checking column existence, branching) that's hard to express in pure SQL. A framework would need "up" and "down" SQL files, which don't map to the existing Go-based migrations.

### Why not just fix the table-rebuild migrations?

We could have added error handling to each rebuild migration (drop `_new` on failure, check for `_new` existence before creating). But that's treating the symptom — each migration would need its own error handling, and future migrations could repeat the mistake.

The transactional approach treats the root cause: partial migrations should never commit. It's a one-time fix that protects all existing and future migrations.

## Consequences

- **Migrations are atomic.** If a migration fails at any point, the entire migration rolls back. The schema is left exactly as it was before the migration started. No `_new` tables, no partial data copies, no half-renamed tables.

- **The server can always start.** If a migration fails, the old schema is intact. The server starts with the previous version's schema. The failed migration can be fixed and re-run.

- **Version recording is atomic with the migration.** The `schema_migrations` insert happens in the same transaction as the migration. If the migration succeeds but the version insert fails (extremely unlikely), the migration rolls back too. No "migration applied but version not recorded" state.

- **All migration functions take `*sql.Tx`.** The `Migration.Apply` field changed from `func(*sql.DB) error` to `func(*sql.Tx) error`. This is a compile-time guarantee that migrations can't accidentally use the raw `*sql.DB` (which would auto-commit).

- **`createMigrationsTable` and `getCurrentVersion` still use `*sql.DB`.** These run before the migration loop and don't need to be transactional (they're idempotent reads + a CREATE IF NOT EXISTS).
