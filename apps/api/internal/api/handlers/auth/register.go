package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles POST /v1/auth/register.
type RegisterHandler struct {
	authService *auth.AuthService
	emailSvc    *emailService.Service
	presenter   *response.Presenter
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(authService *auth.AuthService, emailSvc *emailService.Service, presenter *response.Presenter) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
		emailSvc:    emailSvc,
		presenter:   presenter,
	}
}

// Handle processes the register request.
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Normalize email (trim whitespace and lowercase) - matches old behavior
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Also normalize name
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		h.presenter.BadRequest(c, "email, password, and name required")
		return
	}

	// Validate email format using enterprise-grade validator
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Add request timeout - prevents hanging on slow DB queries (matches old behavior)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := h.authService.Register(ctx, &req, true)
	if err != nil {
		if errors.Is(err, application.ErrUserExists) {
			h.presenter.Conflict(c, "an account with this email already exists")
			return
		}
		if errors.Is(err, application.ErrInvalidInput) {
			h.presenter.BadRequest(c, "password does not meet requirements")
			return
		}
		h.presenter.InternalError(c, "an error occurred")
		return
	}

	// Log successful registration
	h.presenter.RegisterSuccess(c, result.OperatorID)

	// Send verification email after successful registration
	if err := h.sendVerificationEmail(c.Request.Context(), req.Email, req.Name, result.OperatorID); err != nil {
		h.presenter.InternalError(c, "Registration successful but failed to send verification email")
		return
	}

	h.presenter.Created(c, result)
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
