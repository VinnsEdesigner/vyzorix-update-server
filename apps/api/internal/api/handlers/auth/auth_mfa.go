package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.MFAStatusResult
	_ openapi.MFAEnrollResult
	_ openapi.MFAVerifySetupRequest
	_ openapi.MFAEnableRequest
	_ openapi.MFAEnableResult
	_ openapi.MFADisableRequest
	_ openapi.MFABackupCodeRequest
	_ openapi.MFABackupCodeResult
	_ openapi.MFARegenerateResult
	_ openapi.MFAVerifyRequest
	_ openapi.MFAVerifyResult
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
)

// MFAHandler handles MFA endpoints.
type MFAHandler struct {
	authService  *appauth.AuthService
	operatorRepo operator.Repository
	presenter    *response.Presenter
	auditLogger  *audit.Logger
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(authService *appauth.AuthService, operatorRepo operator.Repository, presenter *response.Presenter) *MFAHandler {
	return &MFAHandler{authService: authService, operatorRepo: operatorRepo, presenter: presenter}
}

// SetAuditLogger sets the audit logger for MFA operations.
func (h *MFAHandler) SetAuditLogger(logger *audit.Logger) {
	h.auditLogger = logger
}

// getOperatorFromSession extracts operator ID from the authenticated session.
// See SettingsHandler.getOperatorFromSession: the cookieAuth middleware already.
// validated the (encrypted) session cookie and stored the operator in context,
// so we reuse that instead of passing the raw cookie ciphertext to ValidateSession.
func (h *MFAHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		return "", application.ErrUnauthorized
	}

	return op.ID, nil
}

// GetMFAStatus handles GET /v1/auth/mfa/status.
// @Summary      Get MFA status
// @Description  Returns whether MFA is enabled for the current operator
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.MFAStatusResult  "MFA status"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/status [get]
func (h *MFAHandler) GetMFAStatus(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, schema.MFAStatusResult{MFAEnabled: mfaEnabled})
}

// EnrollMFA handles POST /v1/auth/mfa/enroll.
// @Summary      Enroll MFA
// @Description  Generates a new TOTP MFA secret and enrollment URI
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.MFAEnrollResult  "MFA secret and URI"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/enroll [post]
func (h *MFAHandler) EnrollMFA(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Get operator for email.
	op, err := h.operatorRepo.FindByID(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	result, err := h.authService.EnrollMFA(c.Request.Context(), opID, op.Email)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{
		"secret": result.Secret,
		"uri":    result.URI,
	})
}

// VerifySetupMFA handles POST /v1/auth/mfa/verify-setup.
// @Summary      Verify MFA setup
// @Description  Verifies a TOTP code during MFA enrollment setup
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.MFAVerifySetupRequest  true  "TOTP code"
// @Success      200  {object}  openapi.SuccessResult  "verified"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated / invalid code"
// @Router       /auth/mfa/verify-setup [post]
func (h *MFAHandler) VerifySetupMFA(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Token string `json:"token"` // TOTP code to verify.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code.
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"verified": true})
}

// EnableMFA handles POST /v1/auth/mfa/enable.
// @Summary      Enable MFA
// @Description  Enables MFA after verifying a TOTP code and returns backup codes
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.MFAEnableRequest  true  "TOTP code"
// @Success      200  {object}  openapi.MFAEnableResult  "enabled with backup codes"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated / invalid code"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/enable [post]
func (h *MFAHandler) EnableMFA(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code first.
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Token)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Generate backup codes.
	backupCodes, err := infraauth.GenerateBackupCodes(8)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	binding := infraauth.CreateMFASecretBinding(opID, req.Token)

	// Enable MFA and save backup codes using UpdateOperatorMFA.
	err = h.operatorRepo.UpdateOperatorMFA(c.Request.Context(), opID, req.Token, binding.MAC, backupCodes)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"success": true, "backup_codes": backupCodes})
}

// DisableMFA handles POST /v1/auth/mfa/disable.
// @Summary      Disable MFA
// @Description  Disables MFA after verifying a TOTP code
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.MFADisableRequest  true  "TOTP code"
// @Success      200  {object}  openapi.SuccessResult  "disabled"
// @Failure      400  {object}  openapi.ErrorResponse  "code required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated / invalid code"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/disable [post]
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "code is required")
		return
	}

	if req.Code == "" {
		h.presenter.BadRequest(c, "MFA code is required to disable MFA")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify TOTP code before disabling MFA.
	_, err = h.authService.VerifyMFACode(c.Request.Context(), opID, req.Code)
	if err != nil {
		h.presenter.Unauthorized(c, "Invalid MFA code")
		return
	}

	err = h.authService.DisableMFA(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, schema.SuccessResult{Success: true})
}

// VerifyBackupCode handles POST /v1/auth/mfa/verify-backup.
// @Summary      Verify MFA backup code
// @Description  Verifies a backup code for an operator with MFA enabled
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.MFABackupCodeRequest  true  "backup code"
// @Success      200  {object}  openapi.MFABackupCodeResult  "valid"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated / invalid code"
// @Failure      403  {object}  openapi.ErrorResponse  "MFA not enabled"
// @Router       /auth/mfa/verify-backup [post]
func (h *MFAHandler) VerifyBackupCode(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "")
		return
	}

	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Verify the operator has MFA enabled.
	mfaEnabled, err := h.authService.GetMFAStatus(c.Request.Context(), opID)
	if err != nil || !mfaEnabled {
		h.presenter.Forbidden(c, "")
		return
	}

	// Verify as backup code using service method.
	valid, err := h.authService.VerifyBackupCode(c.Request.Context(), opID, req.Code)
	if err != nil || !valid {
		h.presenter.Unauthorized(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"valid": true})
}

// RegenerateBackupCodes handles POST /v1/auth/mfa/regenerate-backup-codes.
// @Summary      Regenerate MFA backup codes
// @Description  Generates and persists a new set of MFA backup codes
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.MFARegenerateResult  "new backup codes"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/regenerate-backup-codes [post]
func (h *MFAHandler) RegenerateBackupCodes(c *gin.Context) {
	opID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Generate and persist new backup codes via service.
	backupCodes, err := h.authService.RegenerateBackupCodes(c.Request.Context(), opID)
	if err != nil {
		h.presenter.InternalError(c, "")
		return
	}

	h.presenter.OK(c, gin.H{"backup_codes": backupCodes})
}

// VerifyMFA handles POST /v1/auth/mfa/verify - Main MFA verification during login.
// @Summary      Verify MFA (login)
// @Description  Completes login by verifying an MFA code and creating a session with tokens
// @Tags         mfa
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.MFAVerifyRequest  true  "operator_id + MFA code"
// @Success      200  {object}  openapi.MFAVerifyResult  "session with tokens"
// @Failure      400  {object}  openapi.ErrorResponse  "operator_id and code required"
// @Failure      401  {object}  openapi.ErrorResponse  "invalid MFA code / operator invalid"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/mfa/verify [post]
func (h *MFAHandler) VerifyMFA(c *gin.Context) {
	var req struct {
		OperatorID string `json:"operator_id" binding:"required"`
		Code       string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "operator_id and code are required")
		return
	}

	// 2: Log MFA verify attempt.
	if h.auditLogger != nil {
		h.auditLogger.MFAVerifyAttempt(c.Request.Context(), req.OperatorID, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	// Verify the MFA code first.
	session, err := h.authService.VerifyMFACode(c.Request.Context(), req.OperatorID, req.Code)
	if err != nil {
		// 2: Log failed MFA attempt.
		if h.auditLogger != nil {
			h.auditLogger.MFAVerifyFailed(c.Request.Context(), req.OperatorID, c.ClientIP())
		}
		h.presenter.Unauthorized(c, "Invalid MFA code")
		return
	}

	// 2: Re-validate operator state before session creation.
	// This prevents race conditions where operator could be:.
	// - Deleted between MFA verify and session creation.
	// - MFA disabled.
	// - Role changed.
	op, err := h.operatorRepo.FindByID(c.Request.Context(), req.OperatorID)
	if err != nil {
		h.presenter.Unauthorized(c, "Operator not found or invalid")
		return
	}

	// Verify operator is still valid (not deleted, has required fields).
	if !op.IsValid() {
		h.presenter.Unauthorized(c, "Operator is invalid")
		return
	}

	// 2 END: Operator state validated, safe to create session.

	// Create session cookie with session ID (critical - must not fail silently).
	if h.authService.GetSessionManager() != nil {
		cookie, cookieErr := h.authService.GetSessionManager().CreateCookie(session.ID)
		if cookieErr != nil {
			h.presenter.InternalError(c, "Failed to create session")
			return
		}
		h.presenter.SetSessionCookie(c, cookie)
	}

	// Issue refresh token and access token for API clients.
	var refreshToken string
	var accessToken string
	var expiresAt int64
	// Get role from operator's membership in their last organization.
	role := ""
	if m := op.GetMembership(op.LastOrganizationID); m != nil {
		role = string(m.Role)
	}
	if h.authService != nil {
		refreshToken, err = h.authService.IssueRefreshToken(c.Request.Context(), req.OperatorID, session.ID)
		if err != nil {
			h.presenter.InternalError(c, "Failed to issue refresh token")
			return
		}

		// Generate proper JWT access token.
		tokenResult, tokenErr := h.authService.GenerateAccessToken(c.Request.Context(), op.ID, op.Email, op.Name, role)
		if tokenErr != nil {
			h.presenter.InternalError(c, "Failed to generate access token")
			return
		}
		accessToken = tokenResult.AccessToken
		expiresAt = tokenResult.ExpiresAt
	}

	// Return success with session info and tokens.
	// 2: Log MFA verify success.
	if h.auditLogger != nil {
		h.auditLogger.MFAVerifySuccess(c.Request.Context(), req.OperatorID, session.ID, c.ClientIP())
	}

	h.presenter.OK(c, gin.H{
		"success":       true,
		"session_id":    session.ID,
		"access_token":  accessToken, // Proper JWT access token.
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
		"signing_key":   session.SigningKey,
		"operator": gin.H{
			"id":          op.ID,
			"email":       op.Email,
			"name":        op.Name,
			"role":        role,
			"mfa_enabled": op.MFAEnabled,
		},
	})
}
