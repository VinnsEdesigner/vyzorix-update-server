package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/crypto/argon2"
)

const nullStr = "NULL"

func hashArgon2id(password string) string {
	var memory uint32 = 64 * 1024
	var iterations uint32 = 3
	var parallelism uint8 = 4
	var saltLength uint32 = 16
	var keyLength uint32 = 32
	salt := make([]byte, saltLength)
	_, _ = rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	encSalt := base64.RawStdEncoding.EncodeToString(salt)
	encHash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iterations, parallelism, encSalt, encHash)
}

func main() {
	url := os.Getenv("TURSO_DB_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")
	if token == "" {
		token = os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN")
	}
	if url == "" || token == "" {
		fmt.Println("missing TURSO_DB_URL or token")
		os.Exit(1)
	}
	dsn := url + "?authToken=" + token
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		fmt.Println("open error:", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		fmt.Println("ping error:", err)
		os.Exit(1)
	}
	fmt.Println("connected to Turso")

	email := "Vinns@vyzorix.local"
	password := "KamauV#2026"

	var existing string
	_ = db.QueryRowContext(ctx, "SELECT id FROM operators WHERE email = ?", email).Scan(&existing)
	if existing != "" {
		hash := hashArgon2id(password)
		_, err = db.ExecContext(ctx, "UPDATE operators SET password_hash = ?, email_verified = 1, role = 'super_admin' WHERE id = ?", hash, existing)
		if err != nil {
			fmt.Println("update error:", err)
			os.Exit(1)
		}
		fmt.Println("updated existing operator:", existing)
		os.Exit(0)
	}

	id := uuid.New().String()
	hash := hashArgon2id(password)
	now := time.Now().UnixMilli()
	_, err = db.ExecContext(ctx, `INSERT INTO operators (id, email, name, password_hash, google_id, github_id,
		mfa_secret, mfa_secret_mac, mfa_enabled, email_verified, created_at, updated_at, fcm_token, last_organization_id, role)
		VALUES (?, ?, ?, ?,
			`+nullStr+`, `+nullStr+`,
			`+nullStr+`, `+nullStr+`, 0, 1, ?, ?,
			`+nullStr+`, `+nullStr+`, 'super_admin')`,
		id, email, "Test Bot", hash, now, now)
	if err != nil {
		fmt.Println("insert error:", err)
		os.Exit(1)
	}
	fmt.Printf("seeded super_admin operator id: %s email: %s password: [REDACTED]\n", id, email)
}
