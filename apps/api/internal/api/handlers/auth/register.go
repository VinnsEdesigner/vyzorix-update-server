package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles POST /v1/auth/register.
type RegisterHandler struct {
	authService *auth.AuthService
	emailSvc   *emailService.Service
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(authService *auth.AuthService, emailSvc *emailService.Service) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
		emailSvc:   emailSvc,
	}
}

// Handle processes the register request.
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "email, password, and name required"})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req, true)
	if err != nil {
		if errors.Is(err, application.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user_exists", "message": "an account with this email already exists"})
			return
		}
		if errors.Is(err, application.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_password", "message": "password does not meet requirements"})
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
