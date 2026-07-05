package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// OAuthState represents an OAuth state for CSRF protection.
type OAuthState struct {
	ExpiresAt   time.Time
	CreatedAt   time.Time
	ID          string
	State       string
	RedirectURL string
	Provider    string
}

// Ensure OAuthStateRepositoryImpl implements OAuthStateRepository interface.
var _ OAuthStateRepository = (*OAuthStateRepositoryImpl)(nil)

// OAuthStateRepository is the interface for OAuth state operations.
type OAuthStateRepository interface {
	Create(ctx context.Context, state, redirectURL, provider string) (string, error)
	Validate(ctx context.Context, state string) (redirectURL string, stateID string, err error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// OAuthStateRepositoryImpl handles OAuth state persistence.
type OAuthStateRepositoryImpl struct {
	db     *sql.DB
	encKey []byte // 32 bytes for AES-256
}

// NewOAuthStateRepository creates a new OAuthStateRepository.
func NewOAuthStateRepository(db *sql.DB, encKey string) (OAuthStateRepository, error) {
	key, err := hex.DecodeString(encKey)
	if err != nil || len(key) != 32 {
		// Generate a random key if not provided
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
	}
	return &OAuthStateRepositoryImpl{db: db, encKey: key}, nil
}

// Create stores a new OAuth state.
func (r *OAuthStateRepositoryImpl) Create(ctx context.Context, state, redirectURL, provider string) (string, error) {
	// Encrypt the state for storage
	encryptedState, err := r.encrypt(state)
	if err != nil {
		return "", err
	}

	now := time.Now()
	expiresAt := now.Add(10 * time.Minute) // State expires in 10 minutes

	// Generate unique ID
	idBytes := make([]byte, 16)
	if _, randErr := rand.Read(idBytes); randErr != nil {
		return "", randErr
	}
	id := hex.EncodeToString(idBytes)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO oauth_states (id, state, redirect_url, provider, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, encryptedState, redirectURL, provider, expiresAt.UnixMilli(), now.UnixMilli())
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetByState retrieves and validates an OAuth state.
func (r *OAuthStateRepositoryImpl) GetByState(ctx context.Context, state string) (*OAuthState, error) {
	var oauth OAuthState
	var encryptedState string
	var expiresAtMillis, createdAtMillis int64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, state, redirect_url, provider, expires_at, created_at
		FROM oauth_states
		WHERE expires_at > ?
	`, time.Now().UnixMilli()).Scan(&oauth.ID, &encryptedState, &oauth.RedirectURL, &oauth.Provider, &expiresAtMillis, &createdAtMillis)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("oauth state not found or expired")
		}
		return nil, err
	}

	oauth.ExpiresAt = time.UnixMilli(expiresAtMillis)
	oauth.CreatedAt = time.UnixMilli(createdAtMillis)

	// Decrypt the state
	decryptedState, err := r.decrypt(encryptedState)
	if err != nil {
		return nil, err
	}

	// Verify the state matches
	if decryptedState != state {
		return nil, errors.New("oauth state mismatch")
	}

	oauth.State = decryptedState
	return &oauth, nil
}

// Validate retrieves and validates an OAuth state for the OAuth handler.
// 8: This ensures the state was created by our server and hasn't been tampered with.
func (r *OAuthStateRepositoryImpl) Validate(ctx context.Context, state string) (redirectURL string, stateID string, err error) {
	oauth, err := r.GetByState(ctx, state)
	if err != nil {
		return "", "", err
	}
	return oauth.RedirectURL, oauth.ID, nil
}

// Delete removes an OAuth state.
func (r *OAuthStateRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_states WHERE id = ?`, id)
	return err
}

// DeleteExpired removes all expired OAuth states.
func (r *OAuthStateRepositoryImpl) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_states WHERE expires_at <= ?`, time.Now().UnixMilli())
	return err
}

// encrypt encrypts plaintext using AES-GCM.
func (r *OAuthStateRepositoryImpl) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(r.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-GCM.
func (r *OAuthStateRepositoryImpl) decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(r.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// SetupOAuthStateTable creates the oauth_states table if it doesn't exist.
// Deprecated: Use migrateCreateOAuthStates via runMigrations instead.
func SetupOAuthStateTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS oauth_states (
			id TEXT PRIMARY KEY,
			state TEXT NOT NULL,
			redirect_url TEXT NOT NULL,
			provider TEXT NOT NULL,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL
		)
	`)
	return err
}
