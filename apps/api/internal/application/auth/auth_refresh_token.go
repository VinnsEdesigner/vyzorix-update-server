package auth

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
)

// RefreshTokenResult holds the result of a refresh token rotation.
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	SessionID    string
}

// RotateRefreshToken rotates a refresh token, revoking the old one and issuing a new one.
func (s *AuthService) RotateRefreshToken(ctx context.Context, oldRefreshToken string) (*RefreshTokenResult, error) {
	// Check if refresh token repository is configured
	if s.refreshTokenRepo == nil {
		return nil, application.ErrUnauthorized
	}

	// Check if JWT manager is configured
	if s.jwtManager == nil {
		return nil, application.ErrUnauthorized
	}

	// Hash the incoming token
	tokenHash := hashTokenSha256(oldRefreshToken)

	// Find the existing refresh token
	existing, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, application.ErrUnauthorized
	}

	// 10: Check if token was used from a different operator (cross-account theft!)
	if existing.IsRevoked {
		// Token was already used after rotation - potential theft!
		// Revoke ALL refresh tokens for the original operator (force re-login)
		_ = s.refreshTokenRepo.RevokeAllForOperator(ctx, existing.OperatorID)
		return nil, application.ErrUnauthorized
	}

	// Check if expired
	if existing.IsExpired() {
		return nil, application.ErrTokenExpired
	}

	// Generate new tokens
	accessToken, expiresAt, err := s.jwtManager.Generate(
		existing.OperatorID,
		"", // Email not needed for access token
		"", // Name not needed
		"", // Role not needed
	)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	_ = s.refreshTokenRepo.Revoke(ctx, existing.ID)

	// Issue new refresh token
	newRefreshToken, err := s.IssueRefreshToken(ctx, existing.OperatorID, existing.SessionID)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		SessionID:    existing.SessionID,
	}, nil
}

// IssueRefreshToken issues a new refresh token.
func (s *AuthService) IssueRefreshToken(ctx context.Context, operatorID, sessionID string) (string, error) {
	if s.refreshTokenRepo == nil {
		return "", errors.New("refresh token repository not configured")
	}

	token, err := shared.GenerateToken()
	if err != nil {
		return "", err
	}

	tokenHash := hashTokenSha256(token)
	id := shared.GenerateID()

	// Default to 7 days if refreshTokenExpiry not set
	expiry := s.refreshTokenExpiry
	if expiry == 0 {
		expiry = 7 * 24 * time.Hour
	}

	rt := &RefreshToken{
		ID:         id,
		OperatorID: operatorID,
		SessionID:  sessionID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(expiry),
		CreatedAt:  time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", err
	}

	return token, nil
}

// RevokeAllRefreshTokens revokes all refresh tokens for an operator.
func (s *AuthService) RevokeAllRefreshTokens(ctx context.Context, operatorID string) error {
	return s.refreshTokenRepo.RevokeAllForOperator(ctx, operatorID)
}
