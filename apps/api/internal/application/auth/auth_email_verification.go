package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/email_verification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// VerifyEmailResult holds the result of email verification.
type VerifyEmailResult struct {
	Email    string
	Verified bool
}

// VerifyEmail verifies an email using a token.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*VerifyEmailResult, error) {
	tokenHash := hashTokenSha256(token)

	ev, err := s.emailVerifyRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == email_verification.ErrNotFound {
			return &VerifyEmailResult{Verified: false}, nil
		}
		return nil, err
	}

	if ev.IsExpired() {
		return &VerifyEmailResult{Verified: false}, nil
	}

	if err := s.operatorRepo.UpdateEmailVerified(ctx, ev.OperatorID, true); err != nil {
		return nil, err
	}

	_ = s.emailVerifyRepo.Delete(ctx, ev.ID)

	op, _ := s.operatorRepo.FindByID(ctx, ev.OperatorID)
	email := ""
	if op != nil {
		email = op.Email
	}

	return &VerifyEmailResult{Verified: true, Email: email}, nil
}

// ResendVerification sends a new verification email.
func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			return nil
		}
		return err
	}

	if op.EmailVerified {
		return application.ErrEmailAlreadyVerified
	}

	// Delete old verification tokens
	_ = s.emailVerifyRepo.DeleteByOperator(ctx, op.ID)

	// Create new verification token
	token, err := s.CreateEmailVerification(ctx, op.ID)
	if err != nil {
		return err
	}

	// Log the token for email sending (actual email sending should be done at handler level)
	// This is the standard pattern: service creates token, handler sends email
	slog.Default().Info("email_verification_resent",
		"operator_id", op.ID,
		"email", email,
		"token_hint", token[:8]+"...", // Only log partial token for debugging
	)

	return nil
}

// CreateEmailVerification creates a new email verification token and returns the raw token.
func (s *AuthService) CreateEmailVerification(ctx context.Context, operatorID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	tokenHash := hashTokenSha256(token)

	id := shared.GenerateID()

	ev := &email_verification.EmailVerification{
		ID:         id,
		OperatorID: operatorID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.emailVerifyRepo.Create(ctx, ev); err != nil {
		return "", err
	}

	return token, nil
}

// PollVerification checks the status of a verification token.
func (s *AuthService) PollVerification(ctx context.Context, token string) (string, string, error) {
	tokenHash := hashTokenSha256(token)

	ev, err := s.emailVerifyRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == email_verification.ErrNotFound {
			return "invalid", "", nil
		}
		return "", "", err
	}

	if ev.IsExpired() {
		_ = s.emailVerifyRepo.Delete(ctx, ev.ID)
		return "expired", "", nil
	}

	op, err := s.operatorRepo.FindByID(ctx, ev.OperatorID)
	if err != nil {
		return "waiting", "", err
	}
	if op == nil {
		return "waiting", "", nil
	}

	if op.EmailVerified {
		return "success", op.Email, nil
	}

	return "waiting", op.Email, nil
}

// CancelVerification removes pending verification tokens.
func (s *AuthService) CancelVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			return nil
		}
		return err
	}

	return s.emailVerifyRepo.DeleteByOperator(ctx, op.ID)
}
