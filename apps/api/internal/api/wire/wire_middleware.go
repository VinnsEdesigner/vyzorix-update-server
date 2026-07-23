// Package wire provides dependency injection utilities for the API server.
package wire

import (
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"log/slog"
)

// MiddlewareConfig contains all dependencies needed by middleware factory.
type MiddlewareConfig struct {
	Log              *slog.Logger
	SessionManager   *infraauth.SessionManager
	AuthService      *auth.AuthService
	ClientService    *client.Service
	PublicDir        string
	JWTSecret        string
	AllowedOrigins   []string
	HMACWindow       time.Duration
	RateLimitPerMin  int
	AuthRateLimitMin int
	EnforceHMAC      bool
	APIKeyService    *keys.APIKeyService
	AuditLogger      *audit.Logger
}

// MiddlewareSet contains all middleware instances.
type MiddlewareSet struct {
	Factory           *middleware.MiddlewareFactory
	CookieAuth        *middleware.CookieAuth
	SignatureVerifier *middleware.SignatureVerifier
	Lockout           *middleware.Lockout
	CSRFProtector     *middleware.CSRFProtector
	TurnstileVerifier *middleware.TurnstileVerifier
	RevocationList    *infraauth.RevocationList
	IPIntelligence    *middleware.IPIntelligence
	HmacVerifier      *cryptohmac.Verifier
	EncryptKeyFn      func(clientID string) ([]byte, bool)
	RateLimiter       *middleware.RateLimiter
	AuthLimiter       *middleware.RateLimiter
	TenantAPIKeyAuth  *middleware.TenantAPIKeyAuth
	APIKeyRateLimiter *middleware.InMemoryRateLimiter
}

// WireMiddleware creates and wires all middleware instances.
func WireMiddleware(cfg MiddlewareConfig) *MiddlewareSet {
	ms := &MiddlewareSet{}

	// Create MiddlewareFactory to centralize all middleware creation.
	ms.Factory = middleware.NewMiddlewareFactory(
		cfg.Log,
		cfg.SessionManager,
		cfg.AuthService,
		cfg.ClientService,
		middleware.FactoryConfig{
			AllowedOrigins:   cfg.AllowedOrigins,
			EnforceHMAC:      cfg.EnforceHMAC,
			HMACWindow:       cfg.HMACWindow,
			PublicDir:        cfg.PublicDir,
			JWTSecret:        cfg.JWTSecret,
			RateLimitPerMin:  cfg.RateLimitPerMin,
			AuthRateLimitMin: cfg.AuthRateLimitMin,
		},
	)

	// Get middleware instances for server use.
	ms.HmacVerifier = ms.Factory.GetHmacVerifier()
	ms.EncryptKeyFn = ms.Factory.GetEncryptionKeyFn()
	ms.CookieAuth = middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService)
	ms.Lockout = ms.Factory.GetLockout()
	ms.CSRFProtector = ms.Factory.GetCSRF()
	ms.TurnstileVerifier = ms.Factory.GetTurnstile()
	ms.RevocationList = ms.Factory.GetRevocationList()
	ms.IPIntelligence = ms.Factory.IPIntelligence()
	ms.SignatureVerifier = ms.Factory.GetSignatureVerifier()

	// Wire API key middleware if service is available.
	if cfg.APIKeyService != nil && cfg.AuditLogger != nil {
		ms.TenantAPIKeyAuth = middleware.NewTenantAPIKeyAuth(cfg.APIKeyService, cfg.AuditLogger)
		ms.APIKeyRateLimiter = middleware.NewInMemoryRateLimiter(100, time.Minute)
	}

	return ms
}
