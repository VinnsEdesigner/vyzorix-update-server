package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	infraConfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// MiddlewareFactory creates and configures all middleware with their dependencies.
// This centralizes middleware creation and reduces coupling in server.go.
type MiddlewareFactory struct {
	clientService       ClientServiceProvider
	ipIntelligence      *IPIntelligence
	sessionManager      *infraauth.SessionManager
	authService         *appsvc.AuthService
	revocationList      *infraauth.RevocationList
	signatureVerifier   *SignatureVerifier
	turnstile           *TurnstileVerifier
	csrfProtector       *CSRFProtector
	log                 *slog.Logger
	lockout             *Lockout
	jwtSecret           string
	publicDir           string
	allowedOrigins      []string
	authRateLimitPerMin int
	rateLimitPerMinute  int
	hmacWindow          time.Duration
	enforceHMAC         bool
}

// ClientServiceProvider interface for client service dependency.
type ClientServiceProvider interface {
	GetHmacKey(ctx context.Context, clientID string) (string, bool)
}

// FactoryConfig holds configuration for the middleware factory.
type FactoryConfig struct {
	PublicDir        string
	JWTSecret        string
	AllowedOrigins   []string
	HMACWindow       time.Duration
	RateLimitPerMin  int
	AuthRateLimitMin int
	EnforceHMAC      bool
}

// NewMiddlewareFactory creates a new MiddlewareFactory with all dependencies.
func NewMiddlewareFactory(
	log *slog.Logger,
	sessionManager *infraauth.SessionManager,
	authService *appsvc.AuthService,
	clientService ClientServiceProvider,
	cfg FactoryConfig,
) *MiddlewareFactory {
	f := &MiddlewareFactory{
		log:                 log,
		sessionManager:      sessionManager,
		authService:         authService,
		clientService:       clientService,
		allowedOrigins:      cfg.AllowedOrigins,
		enforceHMAC:         cfg.EnforceHMAC,
		hmacWindow:          cfg.HMACWindow,
		publicDir:           cfg.PublicDir,
		jwtSecret:           cfg.JWTSecret,
		rateLimitPerMinute:  cfg.RateLimitPerMin,
		authRateLimitPerMin: cfg.AuthRateLimitMin,
	}

	// Initialize all middleware
	f.initSignatureVerifier()
	f.initIPIntelligence()
	f.initLockout()
	f.initCSRF()
	f.initTurnstile()
	f.initRevocationList()

	return f
}

// =============================================================================
// Core Middleware
// =============================================================================

// Recovery returns the panic recovery middleware.
func (f *MiddlewareFactory) Recovery() gin.HandlerFunc {
	return GinPanicRecovery(f.log)
}

// Logger returns the request logging middleware.
func (f *MiddlewareFactory) Logger() gin.HandlerFunc {
	return Logger(f.log)
}

// RequestID returns the request ID middleware.
func (f *MiddlewareFactory) RequestID() gin.HandlerFunc {
	return RequestIDMiddleware()
}

// BodySizeLimit returns the body size limit middleware.
func (f *MiddlewareFactory) BodySizeLimit() gin.HandlerFunc {
	return BodySizeLimit(DefaultBodySizeLimit)
}

// ErrorHandler returns the error handling middleware.
func (f *MiddlewareFactory) ErrorHandler() gin.HandlerFunc {
	return ErrorHandler(f.log)
}

// =============================================================================
// Security Middleware
// =============================================================================

// CORS returns the CORS middleware.
func (f *MiddlewareFactory) CORS() gin.HandlerFunc {
	return CORSHandler(f.allowedOrigins)
}

// SecurityHeaders returns security headers middleware.
func (f *MiddlewareFactory) SecurityHeaders() gin.HandlerFunc {
	return SecurityHeaders()
}

// SecurityHeadersRelaxed returns relaxed security headers for development.
func (f *MiddlewareFactory) SecurityHeadersRelaxed() gin.HandlerFunc {
	return SecurityHeadersRelaxed()
}

// DisableTrace returns middleware to disable TRACE method.
func (f *MiddlewareFactory) DisableTrace() gin.HandlerFunc {
	return DisableTrace()
}

// DisableConnect returns middleware to disable CONNECT method.
func (f *MiddlewareFactory) DisableConnect() gin.HandlerFunc {
	return DisableConnect()
}

// =============================================================================
// Rate Limiting
// =============================================================================

// RateLimiter returns the general rate limiter middleware.
func (f *MiddlewareFactory) RateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(f.rateLimitPerMinute, time.Minute)
	return limiter.Middleware()
}

// AuthRateLimiter returns the stricter rate limiter for auth endpoints.
func (f *MiddlewareFactory) AuthRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(f.authRateLimitPerMin, time.Minute)
	return limiter.Middleware()
}

// =============================================================================
// Authentication Middleware
// =============================================================================

// CookieAuth returns the cookie-based authentication middleware.
func (f *MiddlewareFactory) CookieAuth() gin.HandlerFunc {
	return NewCookieAuth(f.sessionManager, f.authService).Middleware()
}

// SignatureVerifier returns the request signature verification middleware.
func (f *MiddlewareFactory) SignatureVerifier() gin.HandlerFunc {
	if f.signatureVerifier == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return RequestSigningMiddleware(f.signatureVerifier)
}

// =============================================================================
// Protection Middleware
// =============================================================================

// Lockout returns the account lockout middleware.
func (f *MiddlewareFactory) Lockout() gin.HandlerFunc {
	if f.lockout == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return LockoutMiddleware(f.lockout)
}

// CSRF returns the CSRF protection middleware.
func (f *MiddlewareFactory) CSRF() gin.HandlerFunc {
	if f.csrfProtector == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return f.csrfProtector.Middleware()
}

// Turnstile returns the Turnstile bot protection middleware.
func (f *MiddlewareFactory) Turnstile() gin.HandlerFunc {
	if f.turnstile == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return TurnstileMiddleware(f.turnstile)
}

// PreventUserEnum returns middleware to prevent user enumeration attacks.
func (f *MiddlewareFactory) PreventUserEnum() gin.HandlerFunc {
	return PreventUserEnum()
}

// =============================================================================
// IP Intelligence
// =============================================================================

// IPIntelligence returns the IP intelligence middleware.
func (f *MiddlewareFactory) IPIntelligence() *IPIntelligence {
	return f.ipIntelligence
}

// =============================================================================
// SSR Proxy
// =============================================================================

// SSRProxy returns the SSR proxy middleware if configured.
func (f *MiddlewareFactory) SSRProxy(ssrConfig infraConfig.SSRConfig) gin.HandlerFunc {
	if !ssrConfig.EnableSSR || ssrConfig.SSRServerURL == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return SSRProxy(f.log, ssrConfig, f.publicDir, f.jwtSecret)
}

// =============================================================================
// Initialization Helpers
// =============================================================================

func (f *MiddlewareFactory) initSignatureVerifier() {
	if f.clientService == nil {
		return
	}

	signingConfig := LoadSigningConfig()

	// If EnforceHMAC is explicitly set in config, it overrides the default
	if !f.enforceHMAC && !signingConfig.Enabled {
		// Both say disabled - keep disabled
	} else if f.enforceHMAC {
		signingConfig.Enabled = true
	}

	hmacSecretFn := func(clientID string) (string, bool) {
		return f.clientService.GetHmacKey(context.Background(), clientID)
	}

	f.signatureVerifier = NewSignatureVerifier(signingConfig, hmacSecretFn)
}

func (f *MiddlewareFactory) initIPIntelligence() {
	config := LoadIPIntelligenceConfig()

	f.ipIntelligence = NewIPIntelligence(config)
	if f.ipIntelligence != nil {
		go f.ipIntelligence.StartCleanupRoutine(context.Background(), 5*time.Minute)
	}
}

func (f *MiddlewareFactory) initLockout() {
	lockoutConfig := LoadLockoutConfig()
	f.lockout = NewLockout(lockoutConfig)
}

func (f *MiddlewareFactory) initCSRF() {
	csrfConfig := DefaultCSRFConfig()
	f.csrfProtector = NewCSRFProtector(csrfConfig)
}

func (f *MiddlewareFactory) initTurnstile() {
	turnstileCfg := LoadTurnstileConfig()
	f.turnstile = NewTurnstileVerifier(turnstileCfg)
}

func (f *MiddlewareFactory) initRevocationList() {
	revocationConfig := LoadRevocationConfig()
	if revocationConfig.Enabled {
		f.revocationList = infraauth.DefaultRevocationList()
	}
}

// =============================================================================
// Legacy Support
// =============================================================================

// GetSignatureVerifier returns the signature verifier for direct use.
func (f *MiddlewareFactory) GetSignatureVerifier() *SignatureVerifier {
	return f.signatureVerifier
}

// GetIPIntelligence returns the IP intelligence instance.
func (f *MiddlewareFactory) GetIPIntelligence() *IPIntelligence {
	return f.ipIntelligence
}

// GetLockout returns the lockout instance.
func (f *MiddlewareFactory) GetLockout() *Lockout {
	return f.lockout
}

// GetCSRF returns the CSRF protector instance.
func (f *MiddlewareFactory) GetCSRF() *CSRFProtector {
	return f.csrfProtector
}

// GetTurnstile returns the Turnstile verifier instance.
func (f *MiddlewareFactory) GetTurnstile() *TurnstileVerifier {
	return f.turnstile
}

// GetRevocationList returns the revocation list.
func (f *MiddlewareFactory) GetRevocationList() *infraauth.RevocationList {
	return f.revocationList
}

// GetHmacVerifier returns the HMAC verifier for device command verification.
func (f *MiddlewareFactory) GetHmacVerifier() *cryptohmac.Verifier {
	hmacSecretFn := func(clientID string) (string, bool) {
		if f.clientService == nil {
			return "", false
		}

		return f.clientService.GetHmacKey(context.Background(), clientID)
	}

	return &cryptohmac.Verifier{
		Secret: hmacSecretFn,
		Nonces: cryptohmac.NewNonceCache(f.hmacWindow),
		Window: f.hmacWindow,
	}
}

// GetEncryptionKeyFn returns the encryption key function for response encryption.
func (f *MiddlewareFactory) GetEncryptionKeyFn() func(clientID string) ([]byte, bool) {
	return func(clientID string) ([]byte, bool) {
		secret, ok := f.clientService.GetHmacKey(context.Background(), clientID)
		if !ok || secret == "" {
			return nil, false
		}

		return cryptohmac.DeriveKey(secret), true
	}
}
