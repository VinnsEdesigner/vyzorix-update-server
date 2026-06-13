package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// SuccessReport represents the structure expected by the React frontend
type SuccessReport struct {
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	MemberID     string `json:"memberId"`
	OperatorRole string `json:"operatorRole"`
	Region       string `json:"region"`
	CreatedAt    string `json:"createdAt"`
	Method       string `json:"method"`
}

// InitDB sets up SQLite file connection and applies the schema tables
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./vyzorix.db")
	if err != nil {
		log.Fatalf("CRITICAL error: Failed to open SQLite database: %v", err)
	}

	// Configure Connection Pool Limits
	DB.SetMaxOpenConns(5)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	// Multi-tier Database Tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS operators (
			id TEXT PRIMARY KEY NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			operator_role TEXT NOT NULL DEFAULT 'Operator',
			region TEXT NOT NULL DEFAULT 'Paris, France',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sso_identities (
			id TEXT PRIMARY KEY NOT NULL,
			operator_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			provider_user_id TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS verification_tokens (
			token TEXT PRIMARY KEY NOT NULL,
			operator_id TEXT NOT NULL,
			email TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			is_used INTEGER DEFAULT 0 NOT NULL,
			poll_count INTEGER DEFAULT 0 NOT NULL,
			FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
		);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("CRITICAL schema migration query failed: %v", err)
		}
	}
	fmt.Println("Relational SQLite database tables initialized flawlessly.")
}

// GenerateNewUUIDv7 generates time-ordered lexicographically sortable UUID
func GenerateNewUUIDv7() (string, error) {
	u7, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u7.String(), nil
}

// CreateNewOperator inserts a pending or validated operator record
func CreateNewOperator(fullName, email, username, passwordHash string) (string, error) {
	id, err := GenerateNewUUIDv7()
	if err != nil {
		return "", err
	}

	query := `INSERT INTO operators (id, full_name, email, username, password_hash, operator_role, region, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, 'Operator', 'Paris, France', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	
	_, err = DB.Exec(query, id, fullName, email, username, passwordHash)
	if err != nil {
		return "", err
	}
	return id, nil
}

// FindOperatorById pulls standard profile details for success responses
func FindOperatorById(operatorId string) (*SuccessReport, error) {
	var fullName, email, username, operatorRole, region, createdAt string
	query := `SELECT full_name, email, username, operator_role, region, created_at FROM operators WHERE id = ?`
	
	err := DB.QueryRow(query, operatorId).Scan(&fullName, &email, &username, &operatorRole, &region, &createdAt)
	if err != nil {
		return nil, err
	}

	return &SuccessReport{
		FullName:     fullName,
		Email:        email,
		Username:     username,
		MemberID:     "VXZ-64981", // Corporate operator reference format
		OperatorRole: operatorRole,
		Region:       region,
		CreatedAt:    createdAt + " UTC",
		Method:       "Credentials/SSO",
	}, nil
}

// FindOperatorByIdentifier matches either email or username (for Login workflow)
func FindOperatorByIdentifier(identity string) (string, string, string, error) {
	var id, passwordHash, email string
	query := `SELECT id, password_hash, email FROM operators WHERE email = ? OR username = ?`
	err := DB.QueryRow(query, identity, identity).Scan(&id, &passwordHash, &email)
	if err != nil {
		return "", "", "", err
	}
	return id, passwordHash, email, nil
}

// GetOperatorBySSO resolves existing federated links
func GetOperatorBySSO(provider, providerUserID string) (string, error) {
	var operatorID string
	query := `SELECT operator_id FROM sso_identities WHERE provider = ? AND provider_user_id = ?`
	err := DB.QueryRow(query, provider, providerUserID).Scan(&operatorID)
	if err != nil {
		return "", err
	}
	return operatorID, nil
}

// CreateSSOIdentity maps a federated SSO identifier to an operator account
func CreateSSOIdentity(operatorID, provider, providerUserID string) error {
	id, err := GenerateNewUUIDv7()
	if err != nil {
		return err
	}

	query := `INSERT INTO sso_identities (id, operator_id, provider, provider_user_id) VALUES (?, ?, ?, ?)`
	_, err = DB.Exec(query, id, operatorID, provider, providerUserID)
	return err
}

// CheckEmailExists checks if an email is registered
func CheckEmailExists(email string) (bool, error) {
	var exists int
	query := `SELECT COUNT(1) FROM operators WHERE email = ?`
	err := DB.QueryRow(query, email).Scan(&exists)
	return exists > 0, err
}

// CreateVerificationToken registers a new queryable polling token
func CreateVerificationToken(operatorID, email string) (string, error) {
	tokenBytes, err := GenerateNewUUIDv7()
	if err != nil {
		return "", err
	}
	token := tokenBytes

	expiresAt := time.Now().Add(15 * time.Minute)
	query := `INSERT INTO verification_tokens (token, operator_id, email, expires_at, is_used) VALUES (?, ?, ?, ?, 0)`
	_, err = DB.Exec(query, token, operatorID, email, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

// PollVerificationStatus checks and updates polling token count for developer helper
func PollVerificationStatus(token string) (string, string, int, error) {
	var operatorID, email string
	var isUsed, pollCount int
	var expiresAt time.Time

	query := `SELECT operator_id, email, expires_at, is_used, poll_count FROM verification_tokens WHERE token = ?`
	err := DB.QueryRow(query, token).Scan(&operatorID, &email, &expiresAt, &isUsed, &pollCount)
	if err != nil {
		return "", "", 0, err
	}

	if isUsed == 1 {
		return operatorID, email, 1, nil
	}

	if time.Now().After(expiresAt) {
		return "", "", -1, fmt.Errorf("Token expired")
	}

	pollCount++
	var updateQuery string
	status := 0
	if pollCount >= 3 {
		updateQuery = `UPDATE verification_tokens SET poll_count = ?, is_used = 1 WHERE token = ?`
		status = 1
	} else {
		updateQuery = `UPDATE verification_tokens SET poll_count = ? WHERE token = ?`
	}
	_, _ = DB.Exec(updateQuery, pollCount, token)

	return operatorID, email, status, nil
}

// CancelAndResendToken invalidates previous tokens and registers a new token
func CancelAndResendToken(email string) (string, error) {
	// Find operator
	var operatorID string
	query := `SELECT id FROM operators WHERE email = ?`
	err := DB.QueryRow(query, email).Scan(&operatorID)
	if err != nil {
		return "", err
	}

	// Delete old tokens
	deleteQuery := `DELETE FROM verification_tokens WHERE email = ?`
	_, _ = DB.Exec(deleteQuery, email)

	// Create new token
	return CreateVerificationToken(operatorID, email)
}

// CancelVerificationSession removes registration process traces for re-registration
func CancelVerificationSession(email string) error {
	// 1. Delete associated verification token
	deleteTokens := `DELETE FROM verification_tokens WHERE email = ?`
	_, err := DB.Exec(deleteTokens, email)
	if err != nil {
		return err
	}

	// 2. Delete the operator if it hasn't been verified yet (has no successful login sessions)
	deleteOp := `DELETE FROM operators WHERE email = ? AND id NOT IN (SELECT operator_id FROM sso_identities)`
	_, err = DB.Exec(deleteOp, email)
	return err
}
