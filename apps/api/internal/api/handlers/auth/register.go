package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	securityAuth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles POST /v1/auth/register.
type RegisterHandler struct {
	authService *auth.AuthService
	emailSvc    *emailService.Service
	auditLogger *audit.Logger
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(authService *auth.AuthService, emailSvc *emailService.Service, auditLogger *audit.Logger) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
		emailSvc:    emailSvc,
		auditLogger: auditLogger,
	}
}

// Handle processes the register request.
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request"})
		return
	}

	// Normalize email (trim whitespace and lowercase) - matches old behavior
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Also normalize name
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "email, password, and name required"})
		return
	}

	// Validate email format using enterprise-grade validator
	if _, err := securityAuth.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request"})
		return
	}

	// Add request timeout - prevents hanging on slow DB queries (matches old behavior)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := h.authService.Register(ctx, &req, true)
	if err != nil {
		if errors.Is(err, application.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "conflict", "message": "an account with this email already exists"})
			return
		}
		if errors.Is(err, application.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "password does not meet requirements"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	// Send verification email after successful registration
	if err := h.sendVerificationEmail(c.Request.Context(), req.Email, req.Name, result.OperatorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Registration successful but failed to send verification email",
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// sendVerificationEmail creates a verification token and sends the verification email.
func (h *RegisterHandler) sendVerificationEmail(ctx context.Context, email, name, operatorID string) error {
	if h.emailSvc == nil || !h.emailSvc.IsConfigured() {
		return nil
	}

	// Create verification token - MUST succeed for email verification to work
	token, err := h.authService.CreateEmailVerification(ctx, operatorID)
	if err != nil {
		return err
	}

	// Send email asynchronously (email send failure is non-fatal)
	go func() {
		_ = h.emailSvc.SendVerificationEmail(context.Background(), email, name, token)
	}()

	return nil
}
