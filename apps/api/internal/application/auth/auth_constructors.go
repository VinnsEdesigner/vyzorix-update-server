package auth

import (
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/email_verification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/refresh_token"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	infraSession "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
)

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

// RefreshTokenRepository interface for refresh token operations.
type RefreshTokenRepository interface {
	Create(ctx interface{}, rt *RefreshToken) error
	FindByID(ctx interface{}, id string) (*RefreshToken, error)
	FindByTokenHash(ctx interface{}, tokenHash string) (*RefreshToken, error)
	Revoke(ctx interface{}, id string) error
	RevokeByTokenHash(ctx interface{}, tokenHash string) error
	RevokeAllForOperator(ctx interface{}, operatorID string) error
	CleanupExpired(ctx interface{}, olderThan time.Duration) (int, error)
}

// RefreshToken aliases domain refresh_token package.
type RefreshToken = refresh_token.RefreshToken

// AuthService handles authentication operations.
type AuthService struct {
	operatorRepo       operator.Repository
	sessionRepo        session.Repository
	emailVerifyRepo    email_verification.Repository
	passwordResetRepo password_reset.Repository
	passwordHasher    PasswordHasher
	refreshTokenRepo   RefreshTokenRepository
	jwtManager        *infraauth.JWTManager
	sessionManager    *infraSession.Manager
	sessionTTL         time.Duration
	refreshTokenExpiry time.Duration
	ldapConfig        *LDAPConfig
	deviceStore        *DeviceStore // In-memory device fingerprint store
}

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
	ldapConfig *LDAPConfig,
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
		ldapConfig:         ldapConfig,
	}
}

// SetLDAPConfig sets the LDAP configuration for the auth service.
func (s *AuthService) SetLDAPConfig(cfg *LDAPConfig) {
	s.ldapConfig = cfg
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
