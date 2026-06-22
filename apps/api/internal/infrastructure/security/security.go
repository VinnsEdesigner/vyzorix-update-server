// Package security provides security-related utilities for authentication,
// authorization, password hashing, and input validation.
//
// This is the main entry point for the security package. All security-related
// functionality is organized into subpackages for better maintainability.
//
// Subpackages:
//   - jwt: JWT token management
//   - password: Password hashing and validation
//   - session: Session cookie management
//   - totp: TOTP-based MFA
//   - ratelimit: HTTP rate limiting
//   - lockout: Account lockout for brute force protection
//   - validate: Input validation utilities
//   - oauth: OAuth utilities (GitHub, Google)
//   - origin: WebSocket origin validation
//   - request_signer: API request signing
//   - revocation: Session revocation
//   - secretstore: Encrypted secret storage
package security

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/jwt"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/lockout"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/oauth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/origin"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/ratelimit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/request_signer"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/revocation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/secretstore"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/totp"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/validate"
)

// Re-export types and functions from subpackages for convenience

// JWT types.
type (
	JWTManager     = jwt.Manager
	OperatorClaims = jwt.OperatorClaims
)

// JWT errors.
var (
	ErrInvalidToken = jwt.ErrInvalidToken
	ErrExpiredToken = jwt.ErrExpiredToken
)

// Password types.
type (
	PasswordPolicy = password.Policy
	PasswordError  = password.Error
)

var (
	DefaultPasswordPolicy = password.DefaultPolicy
	UserPasswordPolicy    = password.UserPolicy
)

// Session types.
type SessionManager = session.Manager

const (
	CookieName   = session.CookieName
	CookieMaxAge = session.CookieMaxAge
	CookiePath   = session.CookiePath
)

var (
	ErrInvalidCookie    = session.ErrInvalidCookie
	ErrExpiredCookie    = session.ErrExpiredCookie
	ErrDecryptionFailed = session.ErrDecryptionFailed
)

// TOTP types.
type (
	TOTPConfig = totp.Config
	TOTP       = totp.TOTP
)

const TOTPDigits = totp.Digits

// RateLimiter types.
type (
	RateLimitConfig    = ratelimit.Config
	RateLimiter        = ratelimit.Limiter
	MultiWindowLimiter = ratelimit.MultiWindowLimiter
)

// Lockout types.
type (
	LockoutConfig  = lockout.Config
	AccountLockout = lockout.Handler
	LockoutReason  = lockout.Reason
)

// Lockout errors.
var (
	ErrInvalidPassword       = lockout.ErrInvalidPassword
	ErrTokenGenerationFailed = lockout.ErrTokenGenerationFailed
)

// Validate types.
type (
	EmailValidator        = validate.Validator
	ValidationError       = validate.Error
	EmailValidationResult = validate.Result
	EmailValidationError  = validate.ValidationError
)

const (
	MaxEmailLength    = validate.MaxEmailLength
	MaxNameLength     = validate.MaxNameLength
	MaxPasswordLength = validate.MaxPasswordLength
	MinPasswordLength = validate.MinPasswordLength
	MaxDeviceIDLength = validate.MaxDeviceIDLength
	MaxCommandLength  = validate.MaxCommandLength
	MaxTokenLength    = validate.MaxTokenLength
)

var (
	DefaultEmailValidator = validate.DefaultValidator()
	StrictEmailValidator  = validate.StrictValidator()
)

// OAuth types.
type (
	GitHubTokenResponse = oauth.GitHubTokenResponse
	GitHubUserInfo      = oauth.GitHubUserInfo
	GitHubEmailInfo     = oauth.GitHubEmailInfo
	GitHubOAuthConfig   = oauth.GitHubConfig
)

// Google OAuth types.
type (
	GoogleTokenVerifier = oauth.Verifier
	GoogleClaims        = oauth.GoogleClaims
	GoogleUserInfo      = oauth.GoogleUserInfo
)

var (
	ErrInvalidGoogleToken     = oauth.ErrInvalidGoogleToken
	ErrGoogleTokenExpired     = oauth.ErrGoogleTokenExpired
	ErrGoogleTokenBadIssuer   = oauth.ErrGoogleTokenBadIssuer
	ErrGoogleTokenBadAudience = oauth.ErrGoogleTokenBadAudience
)

// Origin types.
type OriginValidator = origin.Validator

// RequestSigner types.
type RequestSigner = request_signer.Signer

// Revocation types.
type (
	RevocationReason = revocation.Reason
	RevocationEntry  = revocation.Entry
	RevocationList   = revocation.List
)

const (
	RevokeReasonLogout         = revocation.ReasonLogout
	RevokeReasonPasswordChange = revocation.ReasonPasswordChange
	RevokeReasonAdmin          = revocation.ReasonAdmin
	RevokeReasonSecurity       = revocation.ReasonSecurity
	RevokeReasonExpired        = revocation.ReasonExpired
)

// SecretStore types.
type SecretStore = secretstore.Store

var ErrSecretStoreNotFound = secretstore.ErrNotFound

// AuthRateLimiter is a pre-configured rate limiter for auth endpoints.
var AuthRateLimiter = ratelimit.AuthLimiter
