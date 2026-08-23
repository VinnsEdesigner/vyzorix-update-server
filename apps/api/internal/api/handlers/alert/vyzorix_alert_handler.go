// Package alert provides HTTP handlers for org-scoped alert rules.
package alert

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	alertapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
)

// Compile-time references to the openapi DTO types referenced by the swaggo
// annotations above. The annotations live in comments so the Go compiler
// can't see them; these vars make the dependency real so `go build` doesn't
// flag the openapi import as unused. swag init then parses the
// `@Success`/`@Param` annotations into swagger.json definitions.
var (
	_ openapi.AlertRule
	_ openapi.AlertRuleRequest
	_ openapi.AlertRuleListResult
	_ openapi.AlertHistoryEvent
	_ openapi.AlertHistoryResult
	_ openapi.AlertEvaluateResult
	_ openapi.DeletedResult
	_ openapi.ErrorResponse
)

// Handler processes alert rule CRUD and manual evaluation requests.
type Handler struct {
	service   *alertapp.Service
	evaluator *alertapp.Evaluator
}

// NewHandler creates a new alert Handler.
func NewHandler(service *alertapp.Service, evaluator *alertapp.Evaluator) *Handler {
	return &Handler{service: service, evaluator: evaluator}
}

// Service returns the underlying alert application service (used by GraphQL registration).
func (h *Handler) Service() *alertapp.Service {
	if h == nil {
		return nil
	}
	return h.service
}

type ruleRequest struct {
	Name                  string  `json:"name"`
	Metric                string  `json:"metric"`
	Condition             string  `json:"condition"`
	WebhookURL            string  `json:"webhook_url"`
	OnNoData              string  `json:"on_no_data"`
	OnError               string  `json:"on_error"`
	Threshold             float64 `json:"threshold"`
	ForSeconds            int     `json:"for_seconds"`
	NotifyIntervalSeconds int     `json:"notify_interval_seconds"`
	Enabled               bool    `json:"enabled"`
}

func (r *ruleRequest) toInput(orgID string) *alertapp.RuleInput {
	return &alertapp.RuleInput{
		OrgID:                 orgID,
		Name:                  r.Name,
		Metric:                alert.Metric(r.Metric),
		Condition:             alert.Condition(r.Condition),
		Threshold:             r.Threshold,
		ForSeconds:            r.ForSeconds,
		NotifyIntervalSeconds: r.NotifyIntervalSeconds,
		OnNoData:              alert.NoDataPolicy(r.OnNoData),
		OnError:               alert.ErrorPolicy(r.OnError),
		WebhookURL:            r.WebhookURL,
		Enabled:               r.Enabled,
	}
}

func ruleJSON(v *alertapp.RuleView) gin.H {
	rule := v.Rule
	instances := make([]gin.H, 0, len(v.Instances))
	for _, inst := range v.Instances {
		instances = append(instances, gin.H{
			"labels":       inst.Labels,
			"state":        string(inst.State),
			"value":        inst.Value,
			"evaluated_at": inst.EvaluatedAt,
		})
	}
	return gin.H{
		"id":                      rule.ID,
		"org_id":                  rule.OrgID,
		"name":                    rule.Name,
		"metric":                  string(rule.Metric),
		"condition":               string(rule.Condition),
		"threshold":               rule.Threshold,
		"for_seconds":             rule.ForSeconds,
		"notify_interval_seconds": rule.NotifyIntervalSeconds,
		"enabled":                 rule.Enabled,
		"webhook_url":             rule.WebhookURL,
		"created_at":              rule.CreatedAt,
		"updated_at":              rule.UpdatedAt,
		"on_no_data":              string(rule.OnNoData),
		"on_error":                string(rule.OnError),
		"instances":               instances,
	}
}

// List handles GET /v1/alerts/rules.
// @Summary      List alert rules
// @Description  Returns all org-scoped alert rules with their current instance states
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.AlertRuleListResult  "alert rules with instances"
// @Failure      400  {object}  openapi.ErrorResponse  "org context missing"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules [get]
func (h *Handler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	views, err := h.service.ListRules(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list alert rules"))
		return
	}
	rules := make([]gin.H, 0, len(views))
	for _, v := range views {
		rules = append(rules, ruleJSON(v))
	}
	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// Create handles POST /v1/alerts/rules.
// @Summary      Create alert rule
// @Description  Creates a new org-scoped alert rule. Validated before persistence
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.AlertRuleRequest  true  "rule definition"
// @Success      201  {object}  openapi.AlertRule  "created rule with instances"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules [post]
func (h *Handler) Create(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req ruleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	rule, err := h.service.CreateRule(c.Request.Context(), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	view, err := h.service.GetRule(c.Request.Context(), orgID, rule.ID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to load created rule"))
		return
	}
	c.JSON(http.StatusCreated, ruleJSON(view))
}

// Get handles GET /v1/alerts/rules/:id.
// @Summary      Get alert rule
// @Description  Returns one alert rule with its current instance states
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "rule ID"
// @Success      200  {object}  openapi.AlertRule  "rule with instances"
// @Failure      400  {object}  openapi.ErrorResponse  "not found / forbidden"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	view, err := h.service.GetRule(c.Request.Context(), orgID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruleJSON(view))
}

// Update handles PATCH /v1/alerts/rules/:id.
// @Summary      Update alert rule
// @Description  Replaces a rule's mutable fields. Disabling clears its instances
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "rule ID"
// @Param        body  body  openapi.AlertRuleRequest  true  "rule definition"
// @Success      200  {object}  openapi.AlertRule  "updated rule with instances"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req ruleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	rule, err := h.service.UpdateRule(c.Request.Context(), orgID, c.Param("id"), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	view, err := h.service.GetRule(c.Request.Context(), orgID, rule.ID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to load updated rule"))
		return
	}
	c.JSON(http.StatusOK, ruleJSON(view))
}

// Delete handles DELETE /v1/alerts/rules/:id.
// @Summary      Delete alert rule
// @Description  Removes the rule and its instances
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "rule ID"
// @Success      200  {object}  openapi.DeletedResult  "deleted confirmation"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if err := h.service.DeleteRule(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// History handles GET /v1/alerts/history (org-wide) and
// GET /v1/alerts/rules/:id/history (single rule).
// @Summary      Alert history
// @Description  Transition events for an org, optionally narrowed to one rule
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  false  "rule ID filter"
// @Param        limit  query  int  false  "event limit (default 200)"
// @Success      200  {object}  openapi.AlertHistoryResult  "event entries"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /alerts/rules/{id}/history [get]
func (h *Handler) History(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	ruleID := c.Param("id")
	limit := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := h.service.History(c.Request.Context(), orgID, ruleID, limit)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to load alert history"))
		return
	}
	items := make([]gin.H, 0, len(events))
	for _, evt := range events {
		items = append(items, gin.H{
			"id":         evt.ID,
			"rule_id":    evt.RuleID,
			"from_state": string(evt.FromState),
			"to_state":   string(evt.ToState),
			"value":      evt.Value,
			"created_at": evt.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"events": items})
}

// Evaluate handles POST /v1/alerts/rules/:id/evaluate — a manual, on-demand
// evaluation useful for testing rule definitions before they go live.
// @Summary      Manually evaluate a rule
// @Description  Triggers evaluation on the rule's metric immediately
// @Tags         alerts
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "rule ID"
// @Success      200  {object}  openapi.AlertEvaluateResult  "evaluation result (transitioned index)"
// @Failure      500  {object}  openapi.ErrorResponse  "evaluation failed"
// @Router       /alerts/rules/{id}/evaluate [post]
func (h *Handler) Evaluate(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	view, err := h.service.GetRule(c.Request.Context(), orgID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	transitioned, err := h.evaluator.EvaluateRule(c.Request.Context(), view.Rule.ID, time.Now())
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "evaluation failed"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"rule_id": view.Rule.ID, "transitioned": transitioned})
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, alertapp.ErrRuleNotFound):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "alert rule not found"))
	case errors.Is(err, alertapp.ErrInvalidRule):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
	default:
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "alert rule operation failed"))
	}
}
