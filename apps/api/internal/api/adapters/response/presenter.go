package response

import (
	"context"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
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
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			p.auditLogger.LoginSuccess(ctx, operatorID, c.ClientIP(), c.GetHeader("User-Agent"))
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
	c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": message})
}

// Unauthorized sends a 401 response.
func (p *Presenter) Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": message})
}

// Forbidden sends a 403 response.
func (p *Presenter) Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": message})
}

// NotFound sends a 404 response.
func (p *Presenter) NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": message})
}

// Conflict sends a 409 response.
func (p *Presenter) Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{"error": "conflict", "message": message})
}

// InternalError sends a 500 response.
func (p *Presenter) InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": message})
}

// OK sends a 200 response with data.
func (p *Presenter) OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Created sends a 201 response with data.
func (p *Presenter) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// SetSessionCookie sets the session cookie on the response.
func (p *Presenter) SetSessionCookie(c *gin.Context, cookie *http.Cookie) {
	if cookie != nil {
		c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	}
}

// ClearSessionCookie clears the session cookie.
func (p *Presenter) ClearSessionCookie(c *gin.Context) {
	c.SetCookie("session_id", "", -1, "/", "", false, true)
}
