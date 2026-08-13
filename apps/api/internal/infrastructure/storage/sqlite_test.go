package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// TestOpenSQLite_LocalFile verifies the local SQLite backend opens, migrates,
// runs a round-trip query, and reports the correct backend metadata.
func TestOpenSQLite_LocalFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := DefaultConfig(filepath.Join(tmp, "test.db"))

	s, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if got := s.Backend(); got != BackendSQLite {
		t.Fatalf("Backend() = %q, want %q", got, BackendSQLite)
	}

	var n int
	if err := s.DB().QueryRow("SELECT 1").Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d, want 1", n)
	}

	info := s.Info()
	if info["backend"] != "sqlite" {
		t.Fatalf("Info backend = %v, want sqlite", info["backend"])
	}
	if _, ok := info["path"]; !ok {
		t.Fatalf("Info missing path for sqlite backend: %v", info)
	}
}

// TestOpenTurso_Remote runs against a real Turso libSQL endpoint when the
// TURSO_DB_URL and TURSO_AUTH_TOKEN env vars are present. It is skipped
// otherwise so local/dev runs without Turso credentials still pass.
func TestOpenTurso_Remote(t *testing.T) {
	url := os.Getenv("TURSO_DB_URL")
	if url == "" {
		t.Skip("TURSO_DB_URL not set; skipping live Turso test")
	}
	// Prefer a per-DB scoped token when present (the vyzor scope DB is
	// provisioned with its own token under TURSO_VYZOR_SCOPE_DB_TOKEN).
	// Fall back to the generic TURSO_AUTH_TOKEN otherwise.
	token := os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN")
	if token == "" {
		token = os.Getenv("TURSO_AUTH_TOKEN")
	}
	if token == "" {
		t.Skip("no Turso auth token set; skipping live Turso test")
	}

	cfg := DefaultTursoConfig(url, token)
	// Use a short health-check period so the goroutine starts quickly.
	cfg.HealthCheckPeriod = 500 * time.Millisecond

	s, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open turso: %v", err)
	}
	defer s.Close()

	if got := s.Backend(); got != BackendTurso {
		t.Fatalf("Backend() = %q, want %q", got, BackendTurso)
	}

	// Round-trip query through database/sql + libSQL driver.
	var n int
	if err := s.DB().QueryRow("SELECT 1").Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d, want 1", n)
	}

	// The migration ledger must exist on the remote DB after Open().
	var table string
	err = s.DB().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations' LIMIT 1",
	).Scan(&table)
	if err != nil {
		t.Fatalf("schema_migrations lookup: %v", err)
	}
	if table != "schema_migrations" {
		t.Fatalf("got %q, want schema_migrations", table)
	}

	// Info must never leak the auth token.
	info := s.Info()
	if info["backend"] != "turso" {
		t.Fatalf("Info backend = %v, want turso", info["backend"])
	}
	urlField, _ := info["url"].(string)
	if urlField == "" {
		t.Fatalf("Info missing url for turso backend: %v", info)
	}
	if fmt.Sprint(info["url"]) != redactTursoURL(url) {
		t.Fatalf("Info url not redacted: got %q", info["url"])
	}

	// Context-scoped ping must respect a deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.PingContext(ctx); err != nil {
		t.Fatalf("PingContext: %v", err)
	}
}

// TestOpenTurso_MissingCredentials verifies the turso backend fails fast with
// a clear error when credentials are absent, rather than panicking.
func TestOpenTurso_MissingCredentials(t *testing.T) {
	cases := []struct {
		name string
		cfg  *Config
	}{
		{"missing url", &Config{Backend: BackendTurso, TursoAuthToken: "tok"}},
		{"missing token", &Config{Backend: BackendTurso, TursoURL: "libsql://x.turso.io"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Open(c.cfg); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

// TestOpenSQLite_NilConfig guards the constructor.
func TestOpenSQLite_NilConfig(t *testing.T) {
	if _, err := Open(nil); err == nil {
		t.Fatalf("expected error for nil config, got nil")
	}
}

// TestSQLiteTxn_RoundTrip verifies the transaction manager commits on success
// and rolls back on error, using the local SQLite backend.
func TestSQLiteTxn_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	s, err := Open(DefaultConfig(filepath.Join(tmp, "tx.db")))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Reuse the migrations-provided settings table for a tx round-trip. The
	// body must execute against the transaction (pulled from ctx), not the
	// pool — otherwise the single-writer SQLite pool deadlocks.
	set := func(ctx context.Context) error {
		tx, ok := transaction.TxFromContext(ctx)
		if !ok {
			return fmt.Errorf("no tx in context")
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO settings(key,value,updated_at) VALUES(?,?,?)`,
			"k", "v", time.Now().UnixMilli())
		return err
	}
	if err := s.WithTx(context.Background(), set); err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	var v string
	if err := s.DB().QueryRow("SELECT value FROM settings WHERE key=?", "k").Scan(&v); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if v != "v" {
		t.Fatalf("got %q, want v", v)
	}

	// A failing tx body must roll back and surface the error.
	boom := func(ctx context.Context) error {
		return fmt.Errorf("intentional")
	}
	if err := s.WithTx(context.Background(), boom); err == nil {
		t.Fatalf("expected tx error, got nil")
	}
}
