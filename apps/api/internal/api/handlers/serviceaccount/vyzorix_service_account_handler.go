// Package serviceaccount provides HTTP handlers for org-scoped service accounts.
package serviceaccount

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	svcapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/serviceaccount"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.ServiceAccount
	_ openapi.ServiceAccountToken
	_ openapi.ServiceAccountListResult
	_ openapi.ServiceAccountTokenListResult
	_ openapi.ServiceAccountTokenCreated
	_ openapi.ServiceAccountTokenRotated
	_ openapi.CreateServiceAccountRequest
	_ openapi.CreateServiceAccountTokenRequest
	_ openapi.RotateServiceAccountTokenRequest
	_ openapi.DeletedResult
	_ openapi.RevokedResult
	_ openapi.ErrorResponse
)

// Handler processes service account CRUD and token operations.
type Handler struct {
	service *svcapp.Service
}

// NewHandler creates a new service account Handler.
func NewHandler(service *svcapp.Service) *Handler {
	return &Handler{service: service}
}

// Service returns the underlying application service (GraphQL wiring).
func (h *Handler) Service() *svcapp.Service { return h.service }

type createServiceAccountRequest struct {
	Name string `json:"name"`
}

type createTokenRequest struct {
	ExpiresAt *string  `json:"expires_at"`
	ServiceID string   `json:"service_id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
}

type rotateTokenRequest struct {
	ExpiresAt *string  `json:"expires_at"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
}

func serviceAccountJSON(sa *serviceaccount.ServiceAccount) gin.H {
	return gin.H{
		"id":         sa.ID,
		"org_id":     sa.OrgID,
		"name":       sa.Name,
		"enabled":    sa.Enabled,
		"created_at": sa.CreatedAt,
		"updated_at": sa.UpdatedAt,
	}
}

func tokenJSON(token *serviceaccount.Token) gin.H {
	return gin.H{
		"id":         token.ID,
		"service_id": token.ServiceID,
		"name":       token.Name,
		"key_prefix": token.KeyPrefix,
		"scopes":     token.Scopes,
		"valid":      token.Valid,
		"expires_at": token.ExpiresAt,
		"created_at": token.CreatedAt,
		"revoked_at": token.RevokedAt,
	}
}

func parseExpiresAt(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t
	}
	return nil
}

// List handles GET /v1/service-accounts.
// @Summary      List service accounts
// @Description  Returns all org-scoped service accounts.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.ServiceAccountListResult  "service accounts"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts [get]
func (h *Handler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	accounts, err := h.service.List(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list service accounts"))
		return
	}
	items := make([]gin.H, 0, len(accounts))
	for _, sa := range accounts {
		items = append(items, serviceAccountJSON(sa))
	}
	c.JSON(http.StatusOK, gin.H{"service_accounts": items})
}

// Create handles POST /v1/service-accounts.
// @Summary      Create service account
// @Description  Creates a new org-scoped service account.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.CreateServiceAccountRequest  true  "service account definition"
// @Success      201  {object}  openapi.ServiceAccount  "created service account"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts [post]
func (h *Handler) Create(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req createServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	sa, err := h.service.Create(c.Request.Context(), orgID, req.Name)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, serviceAccountJSON(sa))
}

// Delete handles DELETE /v1/service-accounts/:id.
// @Summary      Delete service account
// @Description  Removes a service account and its tokens.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "service account ID"
// @Success      200  {object}  openapi.DeletedResult  "deleted confirmation"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if err := h.service.Delete(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ListTokens handles GET /v1/service-accounts/:id/tokens.
// @Summary      List service account tokens
// @Description  Returns all tokens for a service account.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "service account ID"
// @Success      200  {object}  openapi.ServiceAccountTokenListResult  "tokens"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts/{id}/tokens [get]
func (h *Handler) ListTokens(c *gin.Context) {
	if _, err := h.service.Get(c.Request.Context(), middleware.GetOrganizationID(c), c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	tokens, err := h.service.ListTokens(c.Request.Context(), c.Param("id"))
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list tokens"))
		return
	}
	items := make([]gin.H, 0, len(tokens))
	for _, token := range tokens {
		items = append(items, tokenJSON(token))
	}
	c.JSON(http.StatusOK, gin.H{"tokens": items})
}

// CreateToken handles POST /v1/service-accounts/:id/tokens. Returns the full
// key once — never stored or returned again.
// @Summary      Create service account token
// @Description  Creates a new token for a service account. Returns the full key once.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "service account ID"
// @Param        body  body  openapi.CreateServiceAccountTokenRequest  true  "token definition"
// @Success      201  {object}  openapi.ServiceAccountTokenCreated  "created token with full key"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts/{id}/tokens [post]
func (h *Handler) CreateToken(c *gin.Context) {
	var req createTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	if req.ServiceID == "" {
		req.ServiceID = c.Param("id")
	}
	token, fullKey, err := h.service.CreateToken(c.Request.Context(), &svcapp.TokenInput{
		ServiceID: req.ServiceID,
		Name:      req.Name,
		Scopes:    req.Scopes,
		ExpiresAt: parseExpiresAt(req.ExpiresAt),
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response := tokenJSON(token)
	response["full_key"] = fullKey
	c.JSON(http.StatusCreated, response)
}

// RevokeToken handles DELETE /v1/service-accounts/:id/tokens/:token_id.
// @Summary      Revoke service account token
// @Description  Revokes a single token by ID.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "service account ID"
// @Param        token  path  string  true  "token ID"
// @Success      200  {object}  openapi.RevokedResult  "revoked confirmation"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts/{id}/tokens/{token} [delete]
func (h *Handler) RevokeToken(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if _, err := h.service.Get(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	if err := h.service.RevokeToken(c.Request.Context(), c.Param("token_id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": true})
}

// RotateToken handles POST /v1/service-accounts/:id/tokens/:token_id/rotate.
// @Summary      Rotate service account token
// @Description  Revokes the existing token and issues a new one. Returns the full key once.
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "service account ID"
// @Param        body  body  openapi.RotateServiceAccountTokenRequest  true  "new token definition"
// @Success      201  {object}  openapi.ServiceAccountTokenRotated  "rotated token with full key"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /service-accounts/{id}/rotate [post]
func (h *Handler) RotateToken(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if _, err := h.service.Get(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	var req rotateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	token, fullKey, err := h.service.RotateToken(c.Request.Context(), c.Param("token_id"), &svcapp.TokenInput{
		ServiceID: c.Param("id"),
		Name:      req.Name,
		Scopes:    req.Scopes,
		ExpiresAt: parseExpiresAt(req.ExpiresAt),
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	response := tokenJSON(token)
	response["full_key"] = fullKey
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, svcapp.ErrServiceAccountNotFound):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "service account not found"))
	case errors.Is(err, svcapp.ErrInvalidToken), errors.Is(err, svcapp.ErrInvalidServiceAccount), errors.Is(err, svcapp.ErrInvalidScope):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
	default:
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "service account operation failed"))
	}
}
