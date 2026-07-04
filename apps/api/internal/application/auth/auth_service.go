// Package auth provides authentication and authorization services for the application.
// It handles user registration, login, session management, MFA, email verification,
// password reset, and OAuth authentication flows.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/email_verification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/refresh_token"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	infraSession "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
)

// UpdateOperatorRequest represents a request to update an operator.
type UpdateOperatorRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Role  *string `json:"role,omitempty"`
}

// =============================================================================
// Password Policy Delegation
// =============================================================================

// PasswordPolicy delegates to domain operator package.
type PasswordPolicy = operator.PasswordPolicy

// DefaultPasswordPolicy delegates to domain operator package.
var DefaultPasswordPolicy = operator.DefaultPasswordPolicy

// ValidatePassword delegates to domain operator package.
var ValidatePassword = operator.ValidatePassword

// PasswordStrength delegates to domain operator package.
var PasswordStrength = operator.Strength

// PasswordError delegates to domain operator package.
type PasswordError = operator.PasswordError

// =============================================================================
// Service Struct and Constructor
// =============================================================================

// AuthService handles authentication operations.
type AuthService struct {
	operatorRepo       operator.Repository
	sessionRepo        session.Repository
	emailVerifyRepo    email_verification.Repository
	passwordResetRepo  password_reset.Repository
	passwordHasher     PasswordHasher
	refreshTokenRepo   RefreshTokenRepository
	jwtManager         *infraauth.JWTManager
	sessionManager     *infraSession.Manager
	sessionTTL         time.Duration
	refreshTokenExpiry time.Duration
}

// RefreshTokenRepository interface for refresh token operations.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByID(ctx context.Context, id string) (*RefreshToken, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllForOperator(ctx context.Context, operatorID string) error
	CleanupExpired(ctx context.Context, olderThan time.Duration) (int, error)
}

// RefreshToken aliases domain refresh_token package.
type RefreshToken = refresh_token.RefreshToken

// NewAuthService creates a new AuthService.
func NewAuthService(
	operatorRepo operator.Repository,
	sessionRepo session.Repository,
	emailVerifyRepo email_verification.Repository,
	passwordResetRepo password_reset.Repository,
	passwordHasher PasswordHasher,
	sessionTTL time.Duration,
) *AuthService {
	return &AuthService{
		operatorRepo:      operatorRepo,
		sessionRepo:       sessionRepo,
		emailVerifyRepo:   emailVerifyRepo,
		passwordResetRepo: passwordResetRepo,
		passwordHasher:    passwordHasher,
		sessionTTL:        sessionTTL,
	}
}

// NewAuthServiceWithRefresh creates a new AuthService with refresh token support.
func NewAuthServiceWithRefresh(
	operatorRepo operator.Repository,
	sessionRepo session.Repository,
	emailVerifyRepo email_verification.Repository,
	passwordResetRepo password_reset.Repository,
	passwordHasher PasswordHasher,
	sessionTTL time.Duration,
	refreshTokenRepo RefreshTokenRepository,
	refreshTokenExpiry time.Duration,
	jwtManager *infraauth.JWTManager,
) *AuthService {
	return &AuthService{
		operatorRepo:       operatorRepo,
		sessionRepo:        sessionRepo,
		emailVerifyRepo:    emailVerifyRepo,
		passwordResetRepo:  passwordResetRepo,
		passwordHasher:     passwordHasher,
		sessionTTL:         sessionTTL,
		refreshTokenRepo:   refreshTokenRepo,
		refreshTokenExpiry: refreshTokenExpiry,
		jwtManager:         jwtManager,
	}
}

// SetJWTManager sets the JWT manager for the auth service.
func (s *AuthService) SetJWTManager(jwtManager *infraauth.JWTManager) {
	s.jwtManager = jwtManager
}

// SetSessionManager sets the session manager for the auth service.
func (s *AuthService) SetSessionManager(sessionManager *infraSession.Manager) {
	s.sessionManager = sessionManager
}

// GetSessionManager returns the session manager.
func (s *AuthService) GetSessionManager() *infraSession.Manager {
	return s.sessionManager
}

// =============================================================================
// Core Auth: Login, Register, Session Management, Password Change
// =============================================================================

// Login authenticates an operator and creates a session.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *session.Session, error) {
	// Normalize email.
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find operator by email.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			// Deliberately run password check to prevent timing attacks.
			_ = s.passwordHasher.Verify(req.Password, "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz")
			return nil, nil, application.ErrInvalidCredentials
		}

		return nil, nil, err
	}

	// Check if this is a Google sign-in only account (no password).
	if op.PasswordHash == "" {
		return nil, nil, application.ErrInvalidCredentials
	}

	// Verify password.
	if err = s.passwordHasher.Verify(req.Password, op.PasswordHash); err != nil {
		return nil, nil, application.ErrInvalidCredentials
	}

	// Check if MFA is enabled.
	if op.HasMFA() {
		return &dto.LoginResponse{
			OperatorID: op.ID,
			Email:      op.Email,
			Name:       op.Name,
			Role:       string(op.Role),
			MFAEnabled: true,
		}, nil, application.ErrMFARequired
	}

	// Create session.
	sess, err := s.CreateSession(ctx, op.ID)
	if err != nil {
		return nil, nil, err
	}

	return &dto.LoginResponse{
		OperatorID: op.ID,
		Email:      op.Email,
		Name:       op.Name,
		Role:       string(op.Role),
		MFAEnabled: op.MFAEnabled,
	}, sess, nil
}

// Register creates a new operator.
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest, validatePassword bool) (*dto.RegisterResponse, error) {
	// Normalize email.
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	// Validate password if requested.
	if validatePassword {
		if err := ValidatePassword(req.Password, DefaultPasswordPolicy); err != nil {
			return nil, err
		}
	}

	// Check if user exists.
	existing, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, application.ErrUserExists
	}

	// Hash password.
	hash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Determine role (first operator becomes super admin).
	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	role := operator.RoleOperator
	if count == 0 {
		role = operator.RoleSuperAdmin
	}

	// Generate ID.
	id := shared.GenerateID()

	now := time.Now()
	op := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		PasswordHash:  hash,
		Role:          role,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: op.ID,
		Email:      op.Email,
		Name:       op.Name,
	}, nil
}

// RegisterAsSuperAdmin creates a new operator as super admin (internal use).
func (s *AuthService) RegisterAsSuperAdmin(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	// Check if user exists.
	existing, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, application.ErrUserExists
	}

	// Hash password.
	hash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate ID.
	id := shared.GenerateID()

	now := time.Now()
	op := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		PasswordHash:  hash,
		Role:          operator.RoleSuperAdmin,
		EmailVerified: true, // Super admin is pre-verified
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: op.ID,
		Email:      op.Email,
		Name:       op.Name,
	}, nil
}

// CreateSession creates a new session for an operator.
func (s *AuthService) CreateSession(ctx context.Context, operatorID string) (*session.Session, error) {
	id := shared.GenerateID()

	now := time.Now()
	sess := &session.Session{
		ID:         id,
		OperatorID: operatorID,
		ExpiresAt:  now.Add(s.sessionTTL),
		CreatedAt:  now,
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// Logout invalidates a session by adding it to the revocation list.
// This allows for server-side session invalidation while maintaining audit trail.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	// Add to revocation list for audit trail (allows checking "was this session revoked")
	// Ignoring error - revocation is best-effort, main operation is session deletion
	_ = s.sessionRepo.AddSessionRevocation(ctx, sessionID, "operator_logout")
	// Delete the session from active sessions
	return s.sessionRepo.Delete(ctx, sessionID)
}

// LogoutAll invalidates all sessions for an operator.
func (s *AuthService) LogoutAll(ctx context.Context, operatorID string) error {
	// Revoke all sessions and delete them
	if err := s.sessionRepo.RevokeAllOperatorSessions(ctx, operatorID); err != nil {
		return err
	}
	// Also revoke all refresh tokens for this operator
	// Ignoring error - revocation is best-effort, main operation succeeded
	_ = s.RevokeAllRefreshTokens(ctx, operatorID)

	return nil
}

// GetOperator retrieves an operator by ID.
func (s *AuthService) GetOperator(ctx context.Context, id string) (*operator.Operator, error) {
	return s.operatorRepo.FindByID(ctx, id)
}

// GetOperatorByEmail retrieves an operator by email.
func (s *AuthService) GetOperatorByEmail(ctx context.Context, email string) (*operator.Operator, error) {
	return s.operatorRepo.FindByEmail(ctx, email)
}

// GetSession retrieves a session by ID.
func (s *AuthService) GetSession(ctx context.Context, id string) (*session.Session, error) {
	return s.sessionRepo.FindByID(ctx, id)
}

// ValidateSession validates a session and returns the operator.
// It checks expiration and revocation status.
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*session.Session, *operator.Operator, error) {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		if err == session.ErrNotFound {
			return nil, nil, application.ErrUnauthorized
		}

		return nil, nil, err
	}

	// Check if session is expired.
	if sess.IsExpired() {
		// Clean up expired session.
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, nil, application.ErrTokenExpired
	}

	// Check if session has been revoked (server-side logout).
	revoked, err := s.sessionRepo.IsSessionRevoked(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if revoked {
		return nil, nil, application.ErrUnauthorized
	}

	op, err := s.operatorRepo.FindByID(ctx, sess.OperatorID)
	if err != nil {
		return nil, nil, err
	}

	return sess, op, nil
}

// ChangePassword changes an operator's password.
func (s *AuthService) ChangePassword(ctx context.Context, operatorID, oldPassword, newPassword string) error {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}

	// Verify old password.
	if err = s.passwordHasher.Verify(oldPassword, op.PasswordHash); err != nil {
		return application.ErrInvalidCredentials
	}

	// Validate new password.
	if err = ValidatePassword(newPassword, DefaultPasswordPolicy); err != nil {
		return err
	}

	// Hash new password.
	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	// Update password.
	return s.operatorRepo.UpdatePassword(ctx, operatorID, hash)
}

// =============================================================================
// MFA: Verification, Enrollment, Enable/Disable
// =============================================================================

// VerifyMFACode verifies an MFA code.
func (s *AuthService) VerifyMFACode(ctx context.Context, operatorID, code string) (*session.Session, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if !op.HasMFA() {
		return nil, application.ErrForbidden
	}

	// Verify TOTP code.
	totp := infraauth.NewTOTP(op.MFASecret, infraauth.DefaultTOTPConfig())
	if !totp.Verify(code) {
		return nil, application.ErrInvalidCredentials
	}

	return s.CreateSession(ctx, operatorID)
}

// MFAEnrollResult holds MFA enrollment data.
type MFAEnrollResult struct {
	Secret      string
	URI         string
	BackupCodes []string
}

// EnrollMFA generates a new MFA secret for enrollment.
func (s *AuthService) EnrollMFA(ctx context.Context, operatorID, email string) (*MFAEnrollResult, error) {
	// Generate TOTP secret.
	secret, err := infraauth.GenerateSecret()
	if err != nil {
		return nil, err
	}

	cfg := infraauth.DefaultTOTPConfig()
	cfg.AccountName = email
	totp := infraauth.NewTOTP(secret, cfg)

	// Generate backup codes.
	backupCodes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}

	return &MFAEnrollResult{
		Secret:      secret,
		URI:         totp.ProvisioningURI(),
		BackupCodes: backupCodes,
	}, nil
}

// EnableMFA enables MFA for an operator after verifying a code.
func (s *AuthService) EnableMFA(ctx context.Context, operatorID, secret string, backupCodes []string) error {
	return s.operatorRepo.UpdateMFA(ctx, operatorID, secret, true)
}

// DisableMFA disables MFA for an operator.
func (s *AuthService) DisableMFA(ctx context.Context, operatorID string) error {
	return s.operatorRepo.UpdateMFA(ctx, operatorID, "", false)
}

// GetMFAStatus returns MFA status for an operator.
func (s *AuthService) GetMFAStatus(ctx context.Context, operatorID string) (bool, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return false, err
	}

	return op.MFAEnabled, nil
}

// RegenerateBackupCodes generates and saves new backup codes for an operator.
func (s *AuthService) RegenerateBackupCodes(ctx context.Context, operatorID string) ([]string, error) {
	// Get current operator to get MFA secret
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if op.MFASecret == "" {
		return nil, application.ErrInvalidInput
	}

	// Generate new backup codes
	codes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}

	// Save to database
	err = s.operatorRepo.UpdateOperatorMFA(ctx, operatorID, op.MFASecret, codes)
	if err != nil {
		return nil, err
	}

	return codes, nil
}

// VerifyBackupCode verifies a backup code and removes it if valid.
func (s *AuthService) VerifyBackupCode(ctx context.Context, operatorID, code string) (bool, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return false, err
	}

	if len(op.BackupCodes) == 0 {
		return false, nil
	}

	idx := infraauth.ValidateBackupCode(op.BackupCodes, code)
	if idx < 0 {
		return false, nil
	}

	// Remove the used backup code
	remaining := infraauth.RemoveBackupCode(op.BackupCodes, idx)

	err = s.operatorRepo.UpdateOperatorMFA(ctx, operatorID, op.MFASecret, remaining)
	if err != nil {
		return false, err
	}

	return true, nil
}

// =============================================================================
// Email Verification: Verify, Resend, Create, Poll, Cancel
// =============================================================================

// VerifyEmailResult holds the result of email verification.
type VerifyEmailResult struct {
	Email    string
	Verified bool
}

// VerifyEmail verifies an email using a token.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*VerifyEmailResult, error) {
	// Hash the token for lookup (using SHA-256 per original HashToken).
	tokenHash := hashTokenSha256(token)

	// Look up token in database.
	ev, err := s.emailVerifyRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == email_verification.ErrNotFound {
			return &VerifyEmailResult{Verified: false}, nil
		}

		return nil, err
	}

	// Check if expired.
	if ev.IsExpired() {
		return &VerifyEmailResult{Verified: false}, nil
	}

	// Mark email as verified.
	if err := s.operatorRepo.UpdateEmailVerified(ctx, ev.OperatorID, true); err != nil {
		return nil, err
	}

	// Delete the verification token (single use).
	_ = s.emailVerifyRepo.Delete(ctx, ev.ID)

	// Get operator email.
	op, _ := s.operatorRepo.FindByID(ctx, ev.OperatorID)
	email := ""

	if op != nil {
		email = op.Email
	}

	return &VerifyEmailResult{Verified: true, Email: email}, nil
}

// ResendVerification sends a new verification email.
func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Find operator.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			// Return success anyway to prevent enumeration.
			return nil
		}

		return err
	}

	// Check if already verified.
	if op.EmailVerified {
		return application.ErrEmailAlreadyVerified
	}

	// Delete old verification tokens.
	_ = s.emailVerifyRepo.DeleteByOperator(ctx, op.ID)

	// Note: Email sending is handled by the caller (handler) after calling this method.

	return nil
}

// CreateEmailVerification creates a new email verification token and returns the raw token.
func (s *AuthService) CreateEmailVerification(ctx context.Context, operatorID string) (string, error) {
	// Generate token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	tokenHash := hashTokenSha256(token)

	// Generate ID.
	id := shared.GenerateID()

	// Create verification record.
	ev := &email_verification.EmailVerification{
		ID:         id,
		OperatorID: operatorID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour), // 24 hour expiry
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.emailVerifyRepo.Create(ctx, ev); err != nil {
		return "", err
	}

	return token, nil
}

// PollVerification checks the status of a verification token.
func (s *AuthService) PollVerification(ctx context.Context, token string) (string, string, error) {
	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up in database.
	ev, err := s.emailVerifyRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == email_verification.ErrNotFound {
			return "invalid", "", nil
		}

		return "", "", err
	}

	// Check if expired.
	if ev.IsExpired() {
		// Delete expired token.
		_ = s.emailVerifyRepo.Delete(ctx, ev.ID)
		return "expired", "", nil
	}

	// Get operator to check verification status.
	op, err := s.operatorRepo.FindByID(ctx, ev.OperatorID)
	if err != nil || op == nil {
		//nolint:nilerr // Operator not found or error, still waiting
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

	// Find operator.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil
		}

		return err
	}

	// Delete email verifications for this operator.
	return s.emailVerifyRepo.DeleteByOperator(ctx, op.ID)
}

// =============================================================================
// Password Reset: Generate, Validate, Reset
// =============================================================================

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	Token     string
	ExpiresAt time.Time
	Email     string
}

// GeneratePasswordResetToken generates a password reset token.
func (s *AuthService) GeneratePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if operator exists.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			// Return success anyway to prevent email enumeration.
			return "", nil
		}

		return "", err
	}

	// Delete old tokens for this operator.
	_ = s.passwordResetRepo.DeleteByOperator(ctx, op.ID)

	// Generate token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	tokenHash := hashTokenSha256(token)

	// Generate ID.
	id := shared.GenerateID()

	// Create reset token record.
	prt := &password_reset.PasswordResetToken{
		ID:         id,
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(15 * time.Minute),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.passwordResetRepo.Create(ctx, prt); err != nil {
		return "", err
	}

	return token, nil
}

// ValidatePasswordResetToken validates a password reset token.
func (s *AuthService) ValidatePasswordResetToken(ctx context.Context, token, email string) error {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up token.
	prt, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == password_reset.ErrNotFound {
			return application.ErrInvalidInput
		}

		return err
	}

	// Check if expired.
	if prt.IsExpired() {
		return application.ErrTokenExpired
	}

	// Check if already used.
	if prt.IsUsed() {
		return password_reset.ErrUsed
	}

	// Verify operator email matches.
	op, err := s.operatorRepo.FindByID(ctx, prt.OperatorID)
	if err != nil {
		return err
	}

	if strings.ToLower(op.Email) != email {
		return application.ErrInvalidInput
	}

	return nil
}

// ResetPassword resets a password using a token.
func (s *AuthService) ResetPassword(ctx context.Context, token, _email string, newPassword string) error {
	// Note: email parameter kept for API compatibility but operator is derived from token.
	// Validate password.
	if err := ValidatePassword(newPassword, DefaultPasswordPolicy); err != nil {
		return application.ErrInvalidInput
	}

	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up token.
	prt, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Check if expired.
	if prt.IsExpired() {
		return application.ErrTokenExpired
	}

	// Check if already used.
	if prt.IsUsed() {
		return password_reset.ErrUsed
	}

	// Get operator.
	op, err := s.operatorRepo.FindByID(ctx, prt.OperatorID)
	if err != nil {
		return err
	}

	// Hash new password.
	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	// Update password.
	if err := s.operatorRepo.UpdatePassword(ctx, op.ID, hash); err != nil {
		return err
	}

	// Mark token as used.
	_ = s.passwordResetRepo.MarkUsed(ctx, prt.ID)

	// Revoke all sessions for this operator.
	_ = s.sessionRepo.DeleteByOperatorID(ctx, op.ID)

	return nil
}

// =============================================================================
// Resend Rate Limiting
// =============================================================================

const (
	// MaxResendCount is the maximum number of resend attempts before lockout.
	MaxResendCount = 6
	// ResendLockoutDuration is the duration of the lockout after exceeding max attempts.
	ResendLockoutDuration = 5 * time.Hour
	// ResendCooldownBase is the base cooldown in seconds (30 seconds per attempt after first).
	ResendCooldownBase = 30
)

// ResendRateLimitResult holds the result of a resend rate limit check.
type ResendRateLimitResult struct {
	LockedUntil *time.Time
	RetryAfter  int
	Allowed     bool
}

// GetResendTracker retrieves the resend tracker for an email.
func (s *AuthService) GetResendTracker(ctx context.Context, email string) (*password_reset.ResendTracker, error) {
	emailHash := hashEmailForTracker(email)

	tracker, err := s.passwordResetRepo.GetResendTracker(ctx, emailHash)
	if err != nil {
		if err == password_reset.ErrNotFound {
			//nolint:nilnil
			return nil, nil // No tracker means no attempts yet
		}

		return nil, err
	}

	return tracker, nil
}

// CheckResendRateLimit checks if a resend is allowed for the given email.
func (s *AuthService) CheckResendRateLimit(ctx context.Context, email string) (*ResendRateLimitResult, error) {
	tracker, err := s.GetResendTracker(ctx, email)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// First resend or no tracker - always allowed.
	if tracker == nil {
		return &ResendRateLimitResult{Allowed: true}, nil
	}

	// Check if currently locked out.
	if tracker.IsLockedOut() {
		return &ResendRateLimitResult{
			Allowed:     false,
			LockedUntil: tracker.LockoutUntil,
		}, nil
	}

	// Calculate required cooldown based on resend count.
	// First resend (count=1): 0 cooldown
	// Second resend (count=2): 30 seconds cooldown
	// Third resend (count=3): 60 seconds cooldown
	// etc.
	if tracker.ResendCount > 1 {
		requiredCooldown := (tracker.ResendCount - 1) * ResendCooldownBase
		timeSinceLastResend := now.Sub(tracker.LastResendAt).Seconds()

		if timeSinceLastResend < float64(requiredCooldown) {
			retryAfter := requiredCooldown - int(timeSinceLastResend)

			return &ResendRateLimitResult{
				Allowed:    false,
				RetryAfter: retryAfter,
			}, nil
		}
	}

	return &ResendRateLimitResult{Allowed: true}, nil
}

// UpdateResendTracker updates the resend tracker after a resend attempt.
// This should be called after successfully sending a reset email.
func (s *AuthService) UpdateResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)
	now := time.Now().UTC()

	// Get existing tracker.
	tracker, err := s.passwordResetRepo.GetResendTracker(ctx, emailHash)
	if err != nil && err != password_reset.ErrNotFound {
		return err
	}

	var newCount int
	var lockoutUntil *time.Time

	if tracker == nil {
		newCount = 1
	} else {
		newCount, lockoutUntil = s.calculateResendCount(tracker)
	}

	// Generate ID for tracker.
	id := shared.GenerateID()

	newTracker := &password_reset.ResendTracker{
		ID:           id,
		EmailHash:    emailHash,
		ResendCount:  newCount,
		LastResendAt: now,
		LockoutUntil: lockoutUntil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// If tracker exists, preserve created_at.
	if tracker != nil && !tracker.IsLockedOut() {
		newTracker.ID = tracker.ID
		newTracker.CreatedAt = tracker.CreatedAt
	}

	return s.passwordResetRepo.UpsertResendTracker(ctx, newTracker)
}

func (s *AuthService) calculateResendCount(tracker *password_reset.ResendTracker) (int, *time.Time) {
	if tracker.IsLockedOut() {
		return 1, nil
	}
	newCount := tracker.ResendCount + 1
	if newCount > MaxResendCount {
		lockout := time.Now().UTC().Add(ResendLockoutDuration)
		return tracker.ResendCount, &lockout
	}
	return newCount, nil
}

// DeleteResendTracker removes the resend tracker for an email (e.g., after successful reset).
func (s *AuthService) DeleteResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)
	return s.passwordResetRepo.DeleteResendTracker(ctx, emailHash)
}

// hashEmailForTracker creates a SHA256 hash of the lowercased email for tracker lookups.
func hashEmailForTracker(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	hash := sha256.Sum256([]byte(normalized))

	return hex.EncodeToString(hash[:])
}

// =============================================================================
// OAuth: Google, GitHub Authentication
// =============================================================================

// OAuthResult holds the result of an OAuth operation.
type OAuthResult struct {
	Operator *operator.Operator
	IsNew    bool
}

// FindOrCreateGoogleOperator finds or creates an operator by Google ID.
func (s *AuthService) FindOrCreateGoogleOperator(ctx context.Context, googleID, email, name string) (*OAuthResult, error) {
	// Try to find by Google ID.
	op, err := s.operatorRepo.FindByGoogleID(ctx, googleID)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Try to find by email and link Google account.
	op, err = s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		// Link existing account to Google.
		if err = s.operatorRepo.UpdateGoogleID(ctx, op.ID, googleID); err != nil {
			return nil, err
		}

		op.GoogleID = googleID

		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Create new operator.
	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	role := operator.RoleOperator
	if count == 0 {
		role = operator.RoleSuperAdmin
	}

	id := shared.GenerateID()

	now := time.Now()
	newOp := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		GoogleID:      googleID,
		Role:          role,
		EmailVerified: true, // Google verifies email
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, newOp); err != nil {
		return nil, err
	}

	return &OAuthResult{Operator: newOp, IsNew: true}, nil
}

// FindOrCreateGitHubOperator finds or creates an operator by GitHub ID.
func (s *AuthService) FindOrCreateGitHubOperator(ctx context.Context, githubID, email, name string) (*OAuthResult, error) {
	// Try to find by GitHub ID.
	op, err := s.operatorRepo.FindByGitHubID(ctx, githubID)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Try to find by email and link GitHub account.
	op, err = s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		// Link existing account to GitHub.
		if err = s.operatorRepo.UpdateGitHubID(ctx, op.ID, githubID); err != nil {
			return nil, err
		}

		op.GitHubID = githubID

		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Create new operator.
	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	role := operator.RoleOperator
	if count == 0 {
		role = operator.RoleSuperAdmin
	}

	id := shared.GenerateID()

	now := time.Now()
	newOp := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		GitHubID:      githubID,
		Role:          role,
		EmailVerified: false, // GitHub may not verify email
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, newOp); err != nil {
		return nil, err
	}

	return &OAuthResult{Operator: newOp, IsNew: true}, nil
}

// =============================================================================
// Settings & Admin: Update, Reset, List Operators
// =============================================================================

// UpdateSettingsRequest is the payload for updating operator settings.
type UpdateSettingsRequest struct {
	Name       *string                  `json:"name,omitempty"`
	Thresholds *operator.Thresholds     `json:"thresholds,omitempty"`
	Client     *operator.ClientSettings `json:"client,omitempty"`
	Reset      bool                     `json:"reset,omitempty"`
}

// UpdateOperatorName updates the name for an operator.
func (s *AuthService) UpdateOperatorName(ctx context.Context, operatorID, name string) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrUnauthorized
		}

		return nil, err
	}

	op.Name = name
	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// UpdateSettings updates operator settings (name, thresholds, client settings).
func (s *AuthService) UpdateSettings(ctx context.Context, operatorID string, req *UpdateSettingsRequest) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrUnauthorized
		}

		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, application.ErrInvalidInput
		}

		op.Name = name
		if err := s.operatorRepo.Update(ctx, op); err != nil {
			return nil, err
		}
	}

	// Update thresholds in operator_settings table
	if req.Thresholds != nil {
		if err := s.operatorRepo.UpdateThresholds(ctx, operatorID, *req.Thresholds); err != nil {
			return nil, err
		}
		op.Thresholds = *req.Thresholds
	}

	// Update client settings in operator_settings table with validation
	if req.Client != nil {
		// Validate client settings
		if req.Client.RequestTimeoutMs < 500 || req.Client.RequestTimeoutMs > 60000 {
			return nil, fmt.Errorf("requestTimeoutMs must be between 500 and 60000")
		}
		if req.Client.LogBufferLimit < 50 || req.Client.LogBufferLimit > 5000 {
			return nil, fmt.Errorf("logBufferLimit must be between 50 and 5000")
		}
		if req.Client.SignalHistoryLimit < 30 || req.Client.SignalHistoryLimit > 2000 {
			return nil, fmt.Errorf("signalHistoryLimit must be between 30 and 2000")
		}

		if err := s.operatorRepo.UpdateClientSettings(ctx, operatorID, *req.Client); err != nil {
			return nil, err
		}
		op.ClientSettings = *req.Client
	}

	op.UpdatedAt = time.Now()

	return op, nil
}

// ResetSettings resets operator settings to defaults.
func (s *AuthService) ResetSettings(ctx context.Context, operatorID string) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrUnauthorized
		}

		return nil, err
	}

	// Reset thresholds to defaults in operator_settings table
	defaultThresholds := operator.Thresholds{
		RiskWarn:    70,
		RiskCrit:    85,
		ThermalWarn: 45,
		ThermalCrit: 50,
		BufferWarn:  30,
		BufferCrit:  15,
	}
	if err := s.operatorRepo.UpdateThresholds(ctx, operatorID, defaultThresholds); err != nil {
		return nil, err
	}

	// Reset client settings to defaults in operator_settings table
	defaultClientSettings := operator.ClientSettings{
		ServerURL:         "",
		DeviceID:          "",
		RequestTimeoutMs:  8000,
		AutoReconnect:     true,
		StrictHmac:       false,
		LogBufferLimit:    500,
		SignalHistoryLimit: 240,
		NotificationsEnabled: true,
	}
	if err := s.operatorRepo.UpdateClientSettings(ctx, operatorID, defaultClientSettings); err != nil {
		return nil, err
	}

	// Reset notifications to defaults
	defaultNotifications := operator.DefaultNotificationSettings()
	if err := s.operatorRepo.UpdateNotifications(ctx, operatorID, defaultNotifications); err != nil {
		return nil, err
	}

	op.Thresholds = defaultThresholds
	op.ClientSettings = defaultClientSettings
	op.UpdatedAt = time.Now()

	return op, nil
}

// ListAllOperators returns all operators (admin use).
func (s *AuthService) ListAllOperators(ctx context.Context, limit, offset int) ([]dto.OperatorListResponse, int, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	operators, total, err := s.operatorRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	response := make([]dto.OperatorListResponse, len(operators))
	for i, op := range operators {
		response[i] = dto.OperatorListResponse{
			ID:            op.ID,
			Email:         op.Email,
			Name:          op.Name,
			Role:          string(op.Role),
			MFAEnabled:    op.MFASecret != "",
			EmailVerified: op.EmailVerified,
			CreatedAt:     op.CreatedAt.UnixMilli(),
		}
	}

	return response, total, nil
}

// CreateOperator creates a new operator (admin only).
func (s *AuthService) CreateOperator(ctx context.Context, req *dto.RegisterRequest) (*operator.Operator, error) {
	// Check if email already exists
	existing, err := s.operatorRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, application.ErrEmailExists
	}

	// Hash password
	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Determine role
	role := operator.RoleOperator
	if req.Role != "" {
		role = operator.OperatorRole(req.Role)
	}

	now := time.Now()
	id := shared.GenerateID()
	op := &operator.Operator{
		ID:           id,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// GetOperatorByID retrieves an operator by ID.
func (s *AuthService) GetOperatorByID(ctx context.Context, id string) (*operator.Operator, error) {
	return s.operatorRepo.FindByID(ctx, id)
}

// VerifyJWT verifies a JWT token and returns the claims.
func (s *AuthService) VerifyJWT(token string) (*infraauth.OperatorClaims, error) {
	if s.jwtManager == nil {
		return nil, infraauth.ErrInvalidToken
	}

	return s.jwtManager.Verify(token)
}

// UpdateOperator updates an existing operator (admin only).
func (s *AuthService) UpdateOperator(ctx context.Context, operatorID string, req *UpdateOperatorRequest) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return nil, application.ErrOperatorNotFound
		}

		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, application.ErrInvalidInput
		}
		op.Name = name
	}

	if req.Email != nil {
		// Check if new email already exists
		existing, err := s.operatorRepo.FindByEmail(ctx, *req.Email)
		if err != nil && err != operator.ErrNotFound {
			return nil, err
		}

		if existing != nil && existing.ID != operatorID {
			return nil, application.ErrEmailExists
		}

		op.Email = *req.Email
	}

	if req.Role != nil {
		role := operator.OperatorRole(*req.Role)
		if !role.IsValid() {
			return nil, application.ErrInvalidInput
		}
		op.Role = role
	}

	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// DeleteOperator deletes an operator (admin only).
func (s *AuthService) DeleteOperator(ctx context.Context, operatorID string) error {
	// Check if operator exists
	_, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return application.ErrOperatorNotFound
		}

		return err
	}

	return s.operatorRepo.Delete(ctx, operatorID)
}

// =============================================================================
// Helpers
// =============================================================================

// hashTokenSha256 creates a SHA-256 hash of a token (matches original HashToken).
func hashTokenSha256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// =============================================================================
// Refresh Token Rotation
// =============================================================================

// RefreshTokenResult holds the result of a refresh token rotation.
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	SessionID    string
}

// RotateRefreshToken rotates a refresh token, revoking the old one and issuing a new one.
// This is called during token refresh to implement refresh token rotation infraauth.
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

	// Check if revoked
	if existing.IsRevoked {
		// Token was already used after rotation - potential theft!
		// Revoke ALL refresh tokens for this operator (force re-login)
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

	// Create new session
	sess, err := s.CreateSession(ctx, existing.OperatorID)
	if err != nil {
		return nil, err
	}

	// Create new refresh token
	newRefreshToken, err := shared.GenerateToken()
	if err != nil {
		return nil, err
	}

	newRT := &RefreshToken{
		ID:           shared.GenerateID(),
		TokenHash:    hashTokenSha256(newRefreshToken),
		OperatorID:   existing.OperatorID,
		SessionID:    sess.ID,
		ExpiresAt:    time.Now().Add(s.refreshTokenExpiry),
		CreatedAt:    time.Now(),
		ReplacedByID: existing.ID, // Link to old token for audit
	}

	if err := s.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return nil, err
	}

	// Revoke old refresh token
	_ = s.refreshTokenRepo.Revoke(ctx, existing.ID)

	return &RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		SessionID:    sess.ID,
	}, nil
}

// IssueRefreshToken issues a new refresh token for a session.
// This is called during login to issue the initial refresh token.
func (s *AuthService) IssueRefreshToken(ctx context.Context, operatorID, sessionID string) (string, error) {
	if s.refreshTokenRepo == nil {
		return "", nil // Refresh tokens not configured
	}

	refreshToken, err := shared.GenerateToken()
	if err != nil {
		return "", err
	}

	rt := &RefreshToken{
		ID:         shared.GenerateID(),
		TokenHash:  hashTokenSha256(refreshToken),
		OperatorID: operatorID,
		SessionID:  sessionID,
		ExpiresAt:  time.Now().Add(s.refreshTokenExpiry),
		CreatedAt:  time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", err
	}

	return refreshToken, nil
}

// RevokeAllRefreshTokens revokes all refresh tokens for an operator.
func (s *AuthService) RevokeAllRefreshTokens(ctx context.Context, operatorID string) error {
	if s.refreshTokenRepo == nil {
		return nil
	}

	return s.refreshTokenRepo.RevokeAllForOperator(ctx, operatorID)
}
