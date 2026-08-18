package response

import (
	"context"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"

	"github.com/gin-gonic/gin"
)

// Presenter handles HTTP response formatting and infrastructure concerns.
type Presenter struct {
	authService    *auth.AuthService
	auditLogger    *audit.Logger
	ipIntelligence *middleware.IPIntelligence
}

// NewPresenter creates a new response presenter.
func NewPresenter(authService *auth.AuthService, auditLogger *audit.Logger, ipIntelligence *middleware.IPIntelligence) *Presenter {
	return &Presenter{
		authService:    authService,
		auditLogger:    auditLogger,
		ipIntelligence: ipIntelligence,
	}
}

// LoginSuccess handles successful login with audit and IP tracking.
func (p *Presenter) LoginSuccess(c *gin.Context, operatorID string) {
	if p.ipIntelligence != nil {
		p.ipIntelligence.RecordAuthSuccess(c)
	}

	if p.auditLogger != nil {
		// Extract the trace_id from the request context so the audit entry
		// correlates with the access log.
		traceID := ""
		if tid, ok := c.Get("trace_id"); ok {
			if id, ok := tid.(string); ok {
				traceID = id
			}
		}
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		operatorID := operatorID
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.LogEvent(ctx, &audit.Entry{
				OperatorID: operatorID,
				Action:     audit.ActionLoginSuccess,
				IPAddress:  ipAddress,
				UserAgent:  userAgent,
				Result:     audit.ResultSuccess,
				TraceID:    traceID,
				ActorType:  "operator",
			})
		}()
	}
}

// LoginFailure handles failed login with audit and IP tracking.
func (p *Presenter) LoginFailure(c *gin.Context, email, reason string) {
	if p.ipIntelligence != nil {
		p.ipIntelligence.RecordAuthFailure(c)
	}

	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.LoginFailed(ctx, email, c.ClientIP(), c.GetHeader("User-Agent"), reason)
		}()
	}
}

// LogoutSuccess logs a successful logout.
func (p *Presenter) LogoutSuccess(c *gin.Context, operatorID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.Logout(ctx, operatorID, c.ClientIP(), c.GetHeader("User-Agent"))
		}()
	}
}

// LogoutFailure logs a failed logout attempt.
func (p *Presenter) LogoutFailure(c *gin.Context, operatorID, reason string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.SessionRevoked(ctx, operatorID, c.ClientIP(), c.GetHeader("User-Agent"), reason)
		}()
	}
}

// RegisterSuccess logs a successful registration.
func (p *Presenter) RegisterSuccess(c *gin.Context, operatorID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.Register(ctx, operatorID, c.ClientIP(), c.GetHeader("User-Agent"))
		}()
	}
}

// PasswordChangeSuccess logs a password change.
func (p *Presenter) PasswordChangeSuccess(c *gin.Context, operatorID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.PasswordChange(ctx, operatorID, c.ClientIP(), c.GetHeader("User-Agent"))
		}()
	}
}

// AccountLocked logs an account lockout.
func (p *Presenter) AccountLocked(c *gin.Context, operatorID string, attempts int) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.AccountLocked(ctx, operatorID, c.ClientIP(), attempts)
		}()
	}
}

// APIClientCreated logs API client creation.
func (p *Presenter) APIClientCreated(c *gin.Context, operatorID, clientID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.APIClientCreated(ctx, operatorID, clientID, c.ClientIP())
		}()
	}
}

// APIClientRevoked logs API client revocation.
func (p *Presenter) APIClientRevoked(c *gin.Context, operatorID, clientID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.APIClientRevoked(ctx, operatorID, clientID, c.ClientIP())
		}()
	}
}

// APIClientSecretRotated logs API client secret rotation.
func (p *Presenter) APIClientSecretRotated(c *gin.Context, operatorID, clientID string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.APIClientSecretRotated(ctx, operatorID, clientID, c.ClientIP())
		}()
	}
}

// AdminAction logs an admin action.
func (p *Presenter) AdminAction(c *gin.Context, operatorID, action, resourceType, resourceID string, metadata map[string]string) {
	if p.auditLogger != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.AdminAction(ctx, operatorID, action, resourceType, resourceID, c.ClientIP(), metadata)
		}()
	}
}

// BadRequest sends a 400 response.
func (p *Presenter) BadRequest(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 response.
func (p *Presenter) Unauthorized(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 response.
func (p *Presenter) Forbidden(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusForbidden, message)
}

// NotFound sends a 404 response.
func (p *Presenter) NotFound(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusNotFound, message)
}

// Conflict sends a 409 response.
func (p *Presenter) Conflict(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusConflict, message)
}

// InternalError sends a 500 response.
func (p *Presenter) InternalError(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusInternalServerError, message)
}

// BadGateway sends a 502 response.
func (p *Presenter) BadGateway(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusBadGateway, message)
}

// NotImplemented sends a 501 response.
func (p *Presenter) NotImplemented(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusNotImplemented, message)
}

// ServiceUnavailable sends a 503 response.
func (p *Presenter) ServiceUnavailable(c *gin.Context, message string) {
	responses.RespondStructured(c, http.StatusServiceUnavailable, message)
}

// OK sends a 200 response with data.
func (p *Presenter) OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Created sends a 201 response with data.
func (p *Presenter) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// SetSessionCookie sets the session cookie on the response. Uses http.SetCookie
// so the SameSite attribute from the cookie is preserved (gin's c.SetCookie
// drops it).
func (p *Presenter) SetSessionCookie(c *gin.Context, cookie *http.Cookie) {
	if cookie != nil {
		http.SetCookie(c.Writer, cookie)
	}
}

// ClearSessionCookie clears the session cookie.
func (p *Presenter) ClearSessionCookie(c *gin.Context) {
	// Use the correct cookie name from middleware to match cookie_auth.go.
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(middleware.CookieName, "", -1, "/", "", false, true)
}
