package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles POST /v1/auth/login.
type LoginHandler struct {
	authService  *auth.AuthService
	presenter    *response.Presenter
	emailService *emailService.Service
	lockout      *LoginLockout
}

// LoginLockout provides login attempt tracking with device fingerprinting.
type LoginLockout struct {
	attempts map[string]*LoginAttemptInfo
	mu       sync.RWMutex
}

// LoginAttemptInfo tracks login attempts with device fingerprint.
type LoginAttemptInfo struct {
	FirstAt     time.Time
	LockedUntil *time.Time
	LastIP      string
	UserAgent   string
	Count       int
}

// NewLoginLockout creates a new login lockout tracker.
func NewLoginLockout() *LoginLockout {
	return &LoginLockout{
		attempts: make(map[string]*LoginAttemptInfo),
	}
}

// RecordFailed records a failed login attempt.
func (l *LoginLockout) RecordFailed(email, ip, userAgent string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	info, exists := l.attempts[email]
	now := time.Now()

	if !exists {
		l.attempts[email] = &LoginAttemptInfo{
			Count:     1,
			FirstAt:   now,
			LastIP:    ip,
			UserAgent: userAgent,
		}
		return true
	}

	// Reset if lockout expired
	if info.LockedUntil != nil && now.After(*info.LockedUntil) {
		info.Count = 1
		info.FirstAt = now
		info.LockedUntil = nil
		info.LastIP = ip
		info.UserAgent = userAgent
		return true
	}

	info.Count++
	info.LastIP = ip
	info.UserAgent = userAgent

	// Lock after 5 failed attempts
	if info.Count >= 5 {
		lockDuration := time.Hour
		lockedUntil := now.Add(lockDuration)
		info.LockedUntil = &lockedUntil
		return false // Now locked
	}

	return true
}

// IsLocked checks if an account is locked.
func (l *LoginLockout) IsLocked(email string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.attempts[email]
	if !exists || info.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*info.LockedUntil)
}

// Clear clears failed attempts for successful login.
func (l *LoginLockout) Clear(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, email)
}

// GetAttemptsRemaining returns remaining attempts before lockout.
func (l *LoginLockout) GetAttemptsRemaining(email string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.attempts[email]
	if !exists {
		return 5
	}
	remaining := 5 - info.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RetryAfter returns time until lockout expires.
func (l *LoginLockout) RetryAfter(email string) time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.attempts[email]
	if !exists || info.LockedUntil == nil {
		return 0
	}
	return time.Until(*info.LockedUntil)
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(authService *auth.AuthService, presenter *response.Presenter, emailService *emailService.Service) *LoginHandler {
	return &LoginHandler{
		authService:  authService,
		presenter:    presenter,
		emailService: emailService,
		lockout:      NewLoginLockout(),
	}
}

// Handle processes the login request.
func (h *LoginHandler) Handle(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Normalize email (trim whitespace and lowercase)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		h.presenter.BadRequest(c, "email and password required")
		return
	}

	// Get device fingerprint
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	deviceFingerprint := h.generateDeviceFingerprint(c, req.Email)

	// Check lockout BEFORE any expensive operations
	if h.lockout.IsLocked(req.Email) {
		retryAfter := h.lockout.RetryAfter(req.Email)
		h.presenter.OK(c, gin.H{"error": "account_locked", "message": "Too many failed attempts. Account temporarily locked.", "retry_after": retryAfter.Seconds(), "locked_until": time.Now().Add(retryAfter).Unix()})
		return
	}

	// Validate email format using enterprise-grade validator
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		h.presenter.Unauthorized(c, "invalid email or password")
		return
	}

	// Add request timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Attempt login with device fingerprint for hardened verification
	result, session, err := h.authService.LoginWithDevice(ctx, &req, &auth.DeviceInfo{
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
	})

	if err != nil {
		// Record failed attempt
		h.lockout.RecordFailed(req.Email, ipAddress, userAgent)

		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			h.presenter.Unauthorized(c, "invalid email or password")
		case errors.Is(err, application.ErrMFARequired):
			// Clear lockout on successful credential verification
			h.lockout.Clear(req.Email)
			h.presenter.OK(c, gin.H{
				"mfa_required": true,
				"operator_id":  result.OperatorID,
			})
		case errors.Is(err, application.ErrAccountLocked):
			h.presenter.OK(c, gin.H{"error": "account_locked", "message": "Account temporarily locked due to suspicious activity"})
		default:
			h.presenter.InternalError(c, "an error occurred")
		}

		return
	}

	// Clear failed attempts on successful login
	h.lockout.Clear(req.Email)

	// Send login notification for NEW device/IP (not for remembered devices)
	if h.authService.ShouldNotifyLogin(ctx, result.OperatorID, ipAddress, userAgent, deviceFingerprint) {
		go func() {
			if h.emailService != nil && result != nil {
				loginData := emailService.LoginNotificationData{
					OperatorName: result.Name,
					IPAddress:    ipAddress,
					UserAgent:    userAgent,
					Location:     "Unknown",
					Device:       userAgent,
					Timestamp:    time.Now().Format(time.RFC1123),
				}
				
				if err := h.emailService.SendNewLoginNotificationEmail(context.Background(), result.Email, loginData); err != nil {
					slog.Warn("failed to send login notification email",
						"operatorId", result.OperatorID,
						"email", result.Email,
						"error", err)
				}
			}
		}()
	}

	// Create session cookie with session ID (not operator ID)
	if session != nil && h.authService.GetSessionManager() != nil {
		cookie, err := h.authService.GetSessionManager().CreateCookie(session.ID)
		if err != nil {
			h.presenter.InternalError(c, "Failed to create session")
			return
		}
		h.presenter.SetSessionCookie(c, cookie)
	}

	h.presenter.OK(c, result)
}

// HandleWithTokens processes the login request and returns tokens for API clients.
// This endpoint is for non-browser clients that need JWT access tokens and refresh tokens.
// It does NOT set session cookies - only returns tokens in the response body.
func (h *LoginHandler) HandleWithTokens(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Normalize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		h.presenter.BadRequest(c, "email and password required")
		return
	}

	// Check lockout BEFORE any expensive operations
	if h.lockout.IsLocked(req.Email) {
		retryAfter := h.lockout.RetryAfter(req.Email)
		h.presenter.OK(c, gin.H{"error": "account_locked", "message": "Too many failed attempts. Account temporarily locked.", "retry_after": retryAfter.Seconds(), "locked_until": time.Now().Add(retryAfter).Unix()})
		return
	}

	// Validate email format
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		h.presenter.BadRequest(c, "Invalid email format")
		return
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Login with tokens
	result, err := h.authService.LoginWithTokens(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMFARequired):
			// MFA is required - return partial response with mfa_required flag
			// Get role from selected organization if available
			role := ""
			if result.SelectedOrganization != nil {
				role = result.SelectedOrganization.Role
			}
			h.presenter.OK(c, gin.H{
				"mfa_required": true,
				"operator_id":  result.OperatorID,
				"email":        result.Email,
				"name":         result.Name,
				"role":         role,
				"mfa_enabled": true,
			})
		case errors.Is(err, application.ErrInvalidCredentials):
			h.lockout.RecordFailed(req.Email, c.ClientIP(), c.GetHeader("User-Agent"))
			h.presenter.Unauthorized(c, "Invalid email or password")
		default:
			h.presenter.InternalError(c, "Login failed")
		}
		return
	}

	// Clear lockout on successful login
	h.lockout.Clear(req.Email)

	// Return tokens (no cookie set - this is for API clients)
	h.presenter.OK(c, result)
}

// generateDeviceFingerprint creates a fingerprint from request headers.
func (h *LoginHandler) generateDeviceFingerprint(c *gin.Context, email string) string {
	// Combine available signals for device fingerprint
	ua := c.GetHeader("User-Agent")
	acceptLang := c.GetHeader("Accept-Language")
	acceptEnc := c.GetHeader("Accept-Encoding")
	ip := c.ClientIP()

	// Use crypto/sha256 for consistent fingerprint
	combined := ip + "|" + email + "|" + ua + "|" + acceptLang + "|" + acceptEnc
	fingerprint := auth.HashFingerprint(combined)
	return fingerprint
}
