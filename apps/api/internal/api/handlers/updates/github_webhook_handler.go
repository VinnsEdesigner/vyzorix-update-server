package updates

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/gin-gonic/gin"
)

// GitHubWebhookHandler handles GitHub webhook events for auto-syncing.
type GitHubWebhookHandler struct {
	service      *updates.Service
	webhookSecret string
	auditLogger  *audit.Logger
	log          *slog.Logger
}

// NewGitHubWebhookHandler creates a new GitHub webhook handler.
func NewGitHubWebhookHandler(service *updates.Service, webhookSecret string, auditLogger *audit.Logger, log *slog.Logger) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		service:      service,
		webhookSecret: webhookSecret,
		auditLogger:  auditLogger,
		log:          log,
	}
}

// HandleWebhook handles POST /v1/updates/webhook/github.
// GitHub sends release events to this endpoint for automatic syncing.
func (h *GitHubWebhookHandler) HandleWebhook(c *gin.Context) {
	// Read body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "failed to read request body"})
		return
	}

	// Verify signature if webhook secret is configured
	if h.webhookSecret != "" {
		if !h.verifySignature(c, body) {
			h.log.Warn("invalid webhook signature", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid signature"})
			return
		}
	}

	// Check event type
	event := c.GetHeader("X-GitHub-Event")
	if event == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "missing GitHub-Event header"})
		return
	}

	h.log.Info("GitHub webhook received", "event", event, "ip", c.ClientIP())

	// Only process release events
	if event != "release" && event != "push" {
		h.log.Info("ignoring non-release/push event", "event", event)
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "message": "event type not processed"})
		return
	}

	// Audit log webhook receipt using AdminAction for flexibility
	if h.auditLogger != nil {
		h.auditLogger.AdminAction(
			c.Request.Context(),
			"github-webhook",
			"webhook_received",
			"github",
			event,
			c.ClientIP(),
			map[string]string{"event": event},
		)
	}

	// Trigger sync
	result, err := h.service.SyncFromGitHub(c.Request.Context())
	if err != nil {
		h.log.Error("webhook-triggered sync failed", "error", err)
		if h.auditLogger != nil {
			h.auditLogger.UpdateSyncFailed(
				c.Request.Context(),
				"github-webhook",
				c.ClientIP(),
				"",
				err.Error(),
			)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "sync_failed",
			"message": "failed to sync from GitHub",
		})
		return
	}

	// Audit log successful sync
	if h.auditLogger != nil {
		h.auditLogger.AdminAction(
			c.Request.Context(),
			"github-webhook",
			"webhook_sync_completed",
			"github",
			"",
			c.ClientIP(),
			map[string]string{"versions_found": fmt.Sprintf("%d", result.VersionsFound)},
		)
	}

	h.log.Info("webhook sync completed", "versions", result.VersionsFound)

	c.JSON(http.StatusOK, gin.H{
		"status":         "synced",
		"versions_found": result.VersionsFound,
		"message":        result.Message,
	})
}

// verifySignature verifies the GitHub webhook signature using HMAC-SHA256.
// GitHub signs payloads using HMAC-SHA256 with the webhook secret.
func (h *GitHubWebhookHandler) verifySignature(c *gin.Context, body []byte) bool {
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		// Also check X-Hub-Signature for older webhooks (HMAC-SHA1)
		signature = c.GetHeader("X-Hub-Signature")
		if signature == "" {
			return false
		}
		// For SHA1, we still use HMAC-SHA256 comparison (just prefix check)
		if !strings.HasPrefix(signature, "sha1=") {
			return false
		}
		// Compare using HMAC-SHA1
		mac := hmac.New(sha256.New, []byte(h.webhookSecret))
		mac.Write(body)
		expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
		return hmac.Equal([]byte(signature), []byte(expected))
	}

	// SHA256 signature
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison
	return hmac.Equal([]byte(signature), []byte(expected))
}

// GetWebhookInfo returns webhook endpoint information (non-sensitive).
func (h *GitHubWebhookHandler) GetWebhookInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"webhook_enabled":   h.webhookSecret != "",
		"supported_events":   []string{"release", "push"},
	})
}
