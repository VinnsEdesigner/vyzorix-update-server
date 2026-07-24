// Package appcheck provides Firebase App Check verification for device attestation.
package appcheck

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/appcheck"
	"google.golang.org/api/option"
)

var (
	// ErrAppCheckDisabled is returned when App Check is not configured.
	ErrAppCheckDisabled = errors.New("app check not configured: FIREBASE_CREDENTIALS is empty")
	// ErrInvalidToken is returned when the App Check token is invalid.
	ErrInvalidToken = errors.New("invalid app check token")
	// ErrMissingToken is returned when the App Check token is missing.
	ErrMissingToken = errors.New("missing X-Firebase-AppCheck header")
)

// Verifier provides Firebase App Check token verification.
type Verifier struct {
	client  *appcheck.Client
	log     *slog.Logger
	appID   string
	enabled bool
}

// NewVerifier creates a new App Check verifier using Firebase credentials.
func NewVerifier(log *slog.Logger, rawCredentials, expectedAppID string) (*Verifier, error) {
	v := &Verifier{log: log}

	if rawCredentials == "" {
		log.Warn("app check disabled; FIREBASE_CREDENTIALS not configured")
		return v, nil
	}

	// Write credentials to a temporary file and use WithAuthCredentialsFile.
	// This is the recommended approach to avoid the deprecated WithCredentialsJSON.
	// by explicitly specifying the credential type as ServiceAccount.
	tmpFile, err := os.CreateTemp("", "appcheck-credentials-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp credentials file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write credentials and close file before using it.
	if _, writeErr := tmpFile.WriteString(rawCredentials); writeErr != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to write credentials: %w", writeErr)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to close temp credentials file: %w", closeErr)
	}

	// Use WithAuthCredentialsFile with explicit ServiceAccount type.
	// This validates that the credentials are actually a service account.
	creds := option.WithAuthCredentialsFile(option.ServiceAccount, tmpPath)

	app, err := firebase.NewApp(context.Background(), nil, creds)
	if err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	client, err := app.AppCheck(context.Background())
	if err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	// Clean up temp file after Firebase app is initialized.
	_ = os.Remove(tmpPath)

	v.client = client
	v.appID = expectedAppID
	v.enabled = true
	v.log.Info("firebase app check initialized", "app_id", expectedAppID)

	return v, nil
}

// Enabled returns whether App Check verification is available.
func (v *Verifier) Enabled() bool {
	return v != nil && v.enabled && v.client != nil
}

// VerifyToken verifies the App Check token from the X-Firebase-AppCheck header.
// Returns the decoded token on success, or an error on failure.
func (v *Verifier) VerifyToken(ctx context.Context, token string) (*appcheck.DecodedAppCheckToken, error) {
	if !v.Enabled() {
		return nil, ErrAppCheckDisabled
	}

	if token == "" {
		return nil, ErrMissingToken
	}

	decoded, err := v.client.VerifyToken(token)
	if err != nil {
		v.log.Warn("app check token verification failed", "error", err)
		return nil, ErrInvalidToken
	}

	// Optionally verify the app ID matches.
	// This ensures the token was issued to the expected application.
	if v.appID != "" && decoded.AppID != v.appID {
		v.log.Warn("app check token app ID mismatch",
			"expected", v.appID,
			"got", decoded.AppID,
		)
		return nil, ErrInvalidToken
	}

	return decoded, nil
}
