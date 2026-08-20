package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	"github.com/gin-gonic/gin"
)

// sessionContextKey must match the key used by CookieAuth when storing the.
// validated session in the gin context.
const sessionContextKey = "session"

// SessionSignatureMiddleware verifies the X-Vyzorix-* HMAC headers on.
// tenant API requests using the per-session signing key.
//
// Unlike RequestSigningMiddleware (which resolves the secret by client ID.
// for device APIs), this middleware reads the session that CookieAuth /.
// API-key auth already placed in the gin context and uses session.SigningKey.
// as the HMAC secret. This binds every tenant request to the authenticated.
// session, so a stolen JWT without the session signing key cannot call the.
// API.
//
// For API-key-authenticated requests (no session), the middleware reads the.
// api_key_signing_secret that TenantAPIKeyAuth placed in the context and uses.
// it as the HMAC secret instead. This extends request signing to API keys.
// (Domain A) using the same X-Vyzorix-* header scheme.
//
// The middleware must run AFTER cookie/API-key auth so the session or API key.
// is present.
//
//nolint:gocyclo // multi-branch auth resolution + signature verification
func SessionSignatureMiddleware(verifier *cryptohmac.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if verifier == nil {
			c.Next()
			return
		}

		// Skip if no signature header present and the request is from a
		// session-authenticated caller (browser). API-key requests still
		// require signatures. This allows development without HMAC signing
		// while keeping API-key requests signed in production.
		if c.GetHeader("X-Vyzorix-Signature") == "" {
			// If this is a session-authenticated request (cookie), allow
			// unsigned access. API-key requests without signatures are
			// rejected below by the missing-secret path.
			if _, hasSession := c.Get(sessionContextKey); hasSession {
				c.Set("session_signature_verified", false)
				c.Next()
				return
			}
			// API-key request without signature — reject.
			if _, hasApiKey := c.Get("api_key_signing_secret"); hasApiKey {
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,
					"Missing required signature headers",
				)
				return
			}
			// No auth context at all — let the auth middleware handle it.
			c.Next()
			return
		}

		// Resolve the HMAC secret: session key for cookie auth, API key signing.
		// secret for API-key auth. The nonce namespace ID is the session ID or.
		// API key ID respectively.
		var hmacSecret string
		var nonceNamespace string

		if sessVal, exists := c.Get(sessionContextKey); exists {
			sess, ok := sessVal.(*session.Session)
			if !ok || sess == nil || sess.SigningKey == "" {
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Signature verification failed",
				)
				return
			}
			hmacSecret = sess.SigningKey
			nonceNamespace = sess.ID
		} else if secretVal, exists := c.Get("api_key_signing_secret"); exists {
			secret, ok := secretVal.(string)
			if !ok || secret == "" {
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Signature verification failed",
				)
				return
			}
			hmacSecret = secret
			if keyID, idExists := c.Get("api_key_id"); idExists {
				if id, idOk := keyID.(string); idOk {
					nonceNamespace = id
				}
			}
		} else {
			responses.RespondStructuredAbort(c, http.StatusUnauthorized,

				"Signature verification failed",
			)
			return
		}

		// Build a one-shot verifier scoped to this request so the Secret.
		// function can return the resolved secret.
		reqVerifier := &cryptohmac.Verifier{
			Secret: func(_ string) (string, bool) {
				return hmacSecret, true
			},
			Nonces: verifier.Nonces,
			Window: verifier.Window,
		}

		if _, err := reqVerifier.ReadAndVerify(c.Request, nonceNamespace); err != nil {
			switch err.(type) {
			case cryptohmac.MissingError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Missing required signature headers",
				)
			case cryptohmac.BadFormatError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Invalid timestamp format",
				)
			case *cryptohmac.TimestampExpiredError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Request timestamp outside allowed window",
				)
			case cryptohmac.ReplayedError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Replay detected",
				)
			case cryptohmac.DeviceNotFoundError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Signature verification failed",
				)
			case cryptohmac.SignatureInvalidError:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Signature verification failed",
				)
			default:
				responses.RespondStructuredAbort(c, http.StatusUnauthorized,

					"Signature verification failed",
				)
			}
			return
		}

		// Tag the request so downstream handlers know it was signature-verified.
		c.Set("session_signature_verified", true)
		c.Next()
	}
}
