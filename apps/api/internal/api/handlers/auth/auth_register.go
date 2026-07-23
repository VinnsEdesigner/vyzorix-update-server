package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles POST /v1/auth/register.
type RegisterHandler struct {
	authService     *auth.AuthService
	emailSvc        *emailService.Service
	emailVerifyRepo *storage.EmailVerificationRepository
	presenter       *response.Presenter
	log             *slog.Logger
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(authService *auth.AuthService, emailSvc *emailService.Service, emailVerifyRepo *storage.EmailVerificationRepository, presenter *response.Presenter, log *slog.Logger) *RegisterHandler {
	if log == nil {
		log = slog.Default()
	}
	return &RegisterHandler{
		authService:     authService,
		emailSvc:        emailSvc,
		emailVerifyRepo: emailVerifyRepo,
		presenter:       presenter,
		log:             log,
	}
}

// Handle processes the register request.
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	
	if err := dto.ValidateRegisterRequest(&req); err != nil {
		h.presenter.BadRequest(c, err.Error())
		return
	}

	// Normalize email (trim whitespace and lowercase) - matches old behavior.
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Also normalize name.
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		h.presenter.BadRequest(c, "email, password, and name required")
		return
	}

	// Validate email format using enterprise-grade validator.
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Add request timeout - prevents hanging on slow DB queries (matches old behavior).
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

	// Log successful registration.
	h.presenter.RegisterSuccess(c, result.OperatorID)

	// Send verification email after successful registration.
	if err := h.sendVerificationEmail(c.Request.Context(), req.Email, req.Name, result.OperatorID); err != nil {
		h.presenter.InternalError(c, "Registration successful but failed to send verification email: "+err.Error())
		return
	}

	h.presenter.Created(c, result)
}

// sendVerificationEmail creates a verification token and sends the verification email synchronously.
// If email sending fails, the failure is recorded in the database so the operator can see it on the verification page.
func (h *RegisterHandler) sendVerificationEmail(ctx context.Context, email, name, operatorID string) error {
	if h.emailSvc == nil || !h.emailSvc.IsConfigured() {
		return nil
	}

	// Create verification token - MUST succeed for email verification to work.
	token, verificationID, err := h.authService.CreateEmailVerification(ctx, operatorID)
	if err != nil {
		return err
	}

	// Send email synchronously with a timeout.
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := h.emailSvc.SendVerificationEmail(sendCtx, email, name, token); err != nil {
		// Record the email failure in the database so the operator can see it.
		if h.emailVerifyRepo != nil {
			if markErr := h.emailVerifyRepo.MarkEmailFailed(ctx, verificationID, err.Error()); markErr != nil {
				h.log.Error("failed to record email failure",
					"verificationID", verificationID,
					"emailError", err.Error(),
					"markError", markErr,
				)
			}
		}
		h.log.Error("failed to send verification email",
			"email", email,
			"operatorID", operatorID,
			"error", err,
		)
		return err
	}

	// Record successful email delivery.
	if h.emailVerifyRepo != nil {
		if markErr := h.emailVerifyRepo.MarkEmailSent(ctx, verificationID, time.Now().UTC()); markErr != nil {
			h.log.Error("failed to record email success",
				"verificationID", verificationID,
				"markError", markErr,
			)
		}
	}

	return nil
}
