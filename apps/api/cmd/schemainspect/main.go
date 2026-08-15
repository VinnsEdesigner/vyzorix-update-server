package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	dsn := buildDSN()
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		fmt.Println("open err:", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables := []string{"api_clients", "api_keys"}
	for _, t := range tables {
		fmt.Printf("=== %s ===\n", t)
		if err := inspectTable(ctx, db, t); err != nil {
			fmt.Println("  err:", err)
			continue
		}
		var n int
		_ = db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&n)
		fmt.Printf("  rows: %d\n", n)
	}
}

func inspectTable(ctx context.Context, db *sql.DB, table string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		fmt.Printf("  %d: %s %s (notnull=%d pk=%d)\n", cid, name, ctype, notnull, pk)
	}
	return rows.Err()
}

func buildDSN() string {
	url := os.Getenv("TURSO_DB_URL")
	tok := os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN")
	if tok == "" {
		tok = os.Getenv("TURSO_AUTH_TOKEN")
	}
	return url + "?authToken=" + tok
}
