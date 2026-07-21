package auth

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/refresh_token"
)

// RefreshTokenResult holds the result of a refresh token rotation.
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	SessionID    string
}

// AccessTokenResult holds the result of generating an access token.
type AccessTokenResult struct {
	AccessToken string
	ExpiresAt   int64
}

// GenerateAccessToken generates a JWT access token for an operator.
func (s *AuthService) GenerateAccessToken(ctx context.Context, operatorID, email, name, role string) (*AccessTokenResult, error) {
	if s.jwtManager == nil {
		return nil, errors.New("JWT manager not configured")
	}

	accessToken, expiresAt, err := s.jwtManager.Generate(operatorID, email, name, role)
	if err != nil {
		return nil, err
	}

	return &AccessTokenResult{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt.Unix(),
	}, nil
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

	// Fetch operator details for the JWT claims
	op, err := s.operatorRepo.FindByID(ctx, existing.OperatorID)
	if err != nil {
		return nil, application.ErrUnauthorized
	}

	// Determine the role from organization membership.
	// Role is org-scoped, so we use the operator's LastOrganizationID to determine context.
	role := "operator"
	if s.memberRepo != nil && op.LastOrganizationID != "" {
		member, err := s.memberRepo.FindByOperatorAndOrg(ctx, op.ID, op.LastOrganizationID)
		if err == nil && member.IsActive() {
			role = string(member.Role)
		}
	}

	// Generate new JWT access token with full operator details.
	accessToken, expiresAt, err := s.jwtManager.Generate(
		op.ID,
		op.Email,
		op.Name,
		role,
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

	rt := &refresh_token.RefreshToken{
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
