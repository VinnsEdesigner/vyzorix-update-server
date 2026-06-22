// Package security provides security-related utilities for authentication,
// authorization, password hashing, and input validation.
//
// This is the main entry point for the security package. All security-related
// functionality is organized into subpackages for better maintainability.
package security

import (
	"context"
	"time"

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

// JWTManager is the type alias for jwt.Manager.
type JWTManager = jwt.Manager

// OperatorClaims represents the JWT claims for an authenticated operator.
type OperatorClaims = jwt.OperatorClaims

// ErrInvalidToken is returned when a JWT token is invalid.
var ErrInvalidToken = jwt.ErrInvalidToken

// ErrExpiredToken is returned when a JWT token has expired.
var ErrExpiredToken = jwt.ErrExpiredToken

// NewJWTManager creates a new JWT manager.
func NewJWTManager(secret string, expiry time.Duration, issuer string) *jwt.Manager {
	return jwt.NewManager(secret, expiry, issuer)
}

// PasswordPolicy defines the requirements for a valid password.
type PasswordPolicy = password.Policy

// PasswordError represents validation failures for a password.
type PasswordError = password.Error

// DefaultPasswordPolicy is the standard password policy.
var DefaultPasswordPolicy = password.DefaultPolicy

// UserPasswordPolicy is a user-friendly password policy.
var UserPasswordPolicy = password.UserPolicy

// ValidatePassword checks a password against the given policy.
func ValidatePassword(pwd string, policy password.Policy) error {
	return password.Validate(pwd, policy)
}

// PasswordStrength returns a score from 0-5 based on password complexity.
func PasswordStrength(pwd string) int {
	return password.Strength(pwd)
}

// SessionManager handles encrypted session cookies.
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

// NewSessionManager creates a new session manager.
func NewSessionManager(secret string) *session.Manager {
	return session.NewManager(secret)
}

// HashOperatorID creates a hash of an operator ID.
func HashOperatorID(id string) string {
	return session.HashOperatorID(id)
}

// TOTPConfig holds TOTP configuration.
type TOTPConfig = totp.Config

// TOTP provides TOTP-based MFA.
type TOTP = totp.TOTP

const TOTPDigits = totp.Digits

// DefaultTOTPConfig returns the default TOTP configuration.
func DefaultTOTPConfig() totp.Config {
	return totp.DefaultConfig()
}

// GenerateSecret generates a new TOTP secret.
func GenerateSecret() (string, error) {
	return totp.GenerateSecret()
}

// NewTOTP creates a new TOTP instance.
func NewTOTP(secret string, cfg totp.Config) *totp.TOTP {
	return totp.New(secret, cfg)
}

// GenerateBackupCodes generates backup codes.
func GenerateBackupCodes(count int) ([]string, error) {
	return totp.GenerateBackupCodes(count)
}

// ValidateBackupCode validates a backup code.
func ValidateBackupCode(stored []string, code string) int {
	return totp.ValidateBackupCode(stored, code)
}

// RemoveBackupCode removes a backup code by index.
func RemoveBackupCode(codes []string, index int) []string {
	return totp.RemoveBackupCode(codes, index)
}

// RateLimitConfig holds rate limiter configuration.
type RateLimitConfig = ratelimit.Config

// RateLimiter provides HTTP rate limiting.
type RateLimiter = ratelimit.Limiter

// MultiWindowLimiter provides multi-window rate limiting.
type MultiWindowLimiter = ratelimit.MultiWindowLimiter

// AuthRateLimiter is a pre-configured rate limiter for auth endpoints.
var AuthRateLimiter = ratelimit.AuthLimiter

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(window time.Duration, maxRequests int) *ratelimit.Limiter {
	return ratelimit.New(window, maxRequests)
}

// NewMultiWindowLimiter creates a multi-window rate limiter.
func NewMultiWindowLimiter(limits map[string]struct {
	Window time.Duration
	Max    int
}) *ratelimit.MultiWindowLimiter {
	return ratelimit.NewMultiWindow(limits)
}

// LockoutConfig holds lockout configuration.
type LockoutConfig = lockout.Config

// AccountLockout provides account lockout functionality.
type AccountLockout = lockout.Handler

// LockoutReason describes why a session was revoked.
type LockoutReason = lockout.Reason

// ErrInvalidPassword is returned for invalid password attempts.
var ErrInvalidPassword = lockout.ErrInvalidPassword

// ErrTokenGenerationFailed is returned when token generation fails.
var ErrTokenGenerationFailed = lockout.ErrTokenGenerationFailed

// DefaultLockoutConfig returns the default lockout configuration.
func DefaultLockoutConfig() lockout.Config {
	return lockout.DefaultConfig()
}

// NewAccountLockout creates a new account lockout handler.
func NewAccountLockout(storage lockout.Storage, config lockout.Config) *lockout.Handler {
	return lockout.New(storage, config)
}

// FakeHash provides timing-safe comparison for non-existent passwords.
func FakeHash(a, b string) bool {
	return lockout.FakeHash(a, b)
}

// IsValidPassword checks if a password meets basic requirements.
func IsValidPassword(pwd string) bool {
	return lockout.IsValidPassword(pwd)
}

// GenerateFakeToken generates a fake token for timing attacks.
func GenerateFakeToken() string {
	return lockout.GenerateFakeToken()
}

// EmailValidator provides configurable email validation.
type EmailValidator = validate.Validator

// ValidationError represents a validation error.
type ValidationError = validate.Error

// EmailValidationResult contains detailed validation results.
type EmailValidationResult = validate.Result

// EmailValidationError represents a specific email validation error.
type EmailValidationError = validate.ValidationError

const (
	MaxEmailLength    = validate.MaxEmailLength
	MaxNameLength     = validate.MaxNameLength
	MaxPasswordLength = validate.MaxPasswordLength
	MinPasswordLength = validate.MinPasswordLength
	MaxDeviceIDLength = validate.MaxDeviceIDLength
	MaxCommandLength  = validate.MaxCommandLength
	MaxTokenLength    = validate.MaxTokenLength
)

// DefaultEmailValidator returns a validator for typical user registration.
var DefaultEmailValidator = validate.DefaultValidator()

// StrictEmailValidator returns a validator for high-security requirements.
var StrictEmailValidator = validate.StrictValidator()

// ValidateEmail validates an email address.
func ValidateEmail(email string) (string, error) {
	return validate.Email(email)
}

// ValidateEmailFull performs comprehensive email validation.
func ValidateEmailFull(email string, validator *validate.Validator) *validate.Result {
	return validate.EmailFull(email, validator)
}

// ValidateName validates a name.
func ValidateName(name string) (string, error) {
	return validate.Name(name)
}

// ValidatePasswordLength validates password length constraints.
func ValidatePasswordLength(pwd string) error {
	return validate.PasswordLength(pwd)
}

// ValidateDeviceID validates a device ID.
func ValidateDeviceID(id string) (string, error) {
	return validate.DeviceID(id)
}

// ValidateCommand validates a command string.
func ValidateCommand(cmd string) (string, error) {
	return validate.Command(cmd)
}

// ValidateToken validates a token string.
func ValidateToken(token string) (string, error) {
	return validate.Token(token)
}

// SanitizeString removes potentially dangerous characters.
func SanitizeString(s string, maxLen int) string {
	return validate.Sanitize(s, maxLen)
}

// ContainsInvalidUTF8 checks for invalid UTF-8.
func ContainsInvalidUTF8(s string) bool {
	return validate.ContainsInvalidUTF8(s)
}

// ContainsControlCharacters checks for Unicode control characters.
func ContainsControlCharacters(s string) bool {
	return validate.ContainsControlCharacters(s)
}

// ExtractDomain extracts the domain from an email.
func ExtractDomain(email string) string {
	return validate.ExtractDomain(email)
}

// IsEmailDomainDisposable checks if email domain is disposable.
func IsEmailDomainDisposable(email string) bool {
	return validate.IsDisposableDomain(email)
}

// ClearMXCache clears the DNS MX record cache.
func ClearMXCache() {
	validate.ClearMXCache()
}

// NormalizeEmailForComparison normalizes an email.
func NormalizeEmailForComparison(email string) string {
	return validate.NormalizeEmail(email)
}

// ValidateEmailURI validates a mailto: URI.
func ValidateEmailURI(uri string) error {
	return validate.EmailURI(uri)
}

// GitHubTokenResponse represents the token response from GitHub.
type GitHubTokenResponse = oauth.GitHubTokenResponse

// GitHubUserInfo represents the GitHub user profile.
type GitHubUserInfo = oauth.GitHubUserInfo

// GitHubEmailInfo represents a GitHub email address.
type GitHubEmailInfo = oauth.GitHubEmailInfo

// GitHubOAuthConfig holds GitHub OAuth configuration.
type GitHubOAuthConfig = oauth.GitHubConfig

// GoogleTokenVerifier verifies Google ID tokens.
type GoogleTokenVerifier = oauth.Verifier

// GoogleClaims represents the claims in a Google ID token.
type GoogleClaims = oauth.GoogleClaims

// GoogleUserInfo represents user info from Google's userinfo endpoint.
type GoogleUserInfo = oauth.GoogleUserInfo

var (
	ErrInvalidGoogleToken     = oauth.ErrInvalidGoogleToken
	ErrGoogleTokenExpired     = oauth.ErrGoogleTokenExpired
	ErrGoogleTokenBadIssuer   = oauth.ErrGoogleTokenBadIssuer
	ErrGoogleTokenBadAudience = oauth.ErrGoogleTokenBadAudience
)

// ExchangeGitHubCode exchanges an authorization code for a GitHub access token.
func ExchangeGitHubCode(ctx context.Context, code string, config oauth.GitHubConfig) (*oauth.GitHubTokenResponse, error) {
	return oauth.ExchangeGitHubCode(ctx, code, config)
}

// FetchGitHubUserProfile retrieves the GitHub user profile.
func FetchGitHubUserProfile(ctx context.Context, accessToken string) (*oauth.GitHubUserInfo, error) {
	return oauth.FetchGitHubUserProfile(ctx, accessToken)
}

// FetchGitHubEmails retrieves all email addresses for the GitHub user.
func FetchGitHubEmails(ctx context.Context, accessToken string) ([]oauth.GitHubEmailInfo, error) {
	return oauth.FetchGitHubEmails(ctx, accessToken)
}

// GetPrimaryEmail finds the primary verified email from the email list.
func GetPrimaryEmail(emails []oauth.GitHubEmailInfo) string {
	return oauth.GetPrimaryEmail(emails)
}

// NewGoogleTokenVerifier creates a new verifier for Google ID tokens.
func NewGoogleTokenVerifier(audience string) *oauth.Verifier {
	return oauth.NewGoogleVerifier(audience)
}

// DecodeGoogleIDToken verifies and decodes a Google ID token.
func DecodeGoogleIDToken(token, audience string) (*oauth.GoogleClaims, error) {
	return oauth.DecodeGoogleIDToken(token, audience)
}

// GetGoogleUserInfo fetches user info from Google's userinfo endpoint.
func GetGoogleUserInfo(ctx context.Context, accessToken string) (*oauth.GoogleUserInfo, error) {
	return oauth.GetGoogleUserInfo(ctx, accessToken)
}

// OriginValidator validates WebSocket origins.
type OriginValidator = origin.Validator

// NewOriginValidator creates a validator with allowed origins.
func NewOriginValidator(origins []string) *origin.Validator {
	return origin.NewValidator(origins)
}

// RequestSigner signs API requests.
type RequestSigner = request_signer.Signer

// NewRequestSigner creates a new request signer.
func NewRequestSigner(clientID, clientSecret string) (*request_signer.Signer, error) {
	return request_signer.New(clientID, clientSecret)
}

// RevocationReason describes why a session was revoked.
type RevocationReason = revocation.Reason

// RevocationEntry represents a revoked session entry.
type RevocationEntry = revocation.Entry

// RevocationList manages session revocation.
type RevocationList = revocation.List

const (
	RevokeReasonLogout         = revocation.ReasonLogout
	RevokeReasonPasswordChange = revocation.ReasonPasswordChange
	RevokeReasonAdmin          = revocation.ReasonAdmin
	RevokeReasonSecurity       = revocation.ReasonSecurity
	RevokeReasonExpired        = revocation.ReasonExpired
)

// NewRevocationList creates a new revocation list.
func NewRevocationList(maxEntries int, ttl time.Duration) *revocation.List {
	return revocation.New(maxEntries, ttl)
}

// DefaultRevocationList creates a revocation list with sensible defaults.
func DefaultRevocationList() *revocation.List {
	return revocation.Default()
}

// SecretStore manages encrypted storage of per-device command secrets.
type SecretStore = secretstore.Store

// ErrSecretStoreNotFound is returned when a secret is not found.
var ErrSecretStoreNotFound = secretstore.ErrNotFound

// NewSecretStore creates a new secret store.
func NewSecretStore(baseDir, masterKeyBase64 string) (*secretstore.Store, error) {
	return secretstore.New(baseDir, masterKeyBase64)
}
