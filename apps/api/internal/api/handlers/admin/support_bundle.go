package admin

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

type SupportBundleHandler struct {
	db *storage.SQLite
}

func NewSupportBundleHandler(db *storage.SQLite) *SupportBundleHandler {
	return &SupportBundleHandler{db: db}
}

func (h *SupportBundleHandler) GetBundle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	name, _ := os.Hostname()
	bundle := gin.H{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"hostname":     name,
		"go_version":   runtime.Version(),
		"goroutines":   runtime.NumGoroutine(),
		"go_max_procs": runtime.GOMAXPROCS(0),
		"go_num_cpu":   runtime.NumCPU(),
	}

	if h.db != nil {
		var schemaVersion int
		_ = h.db.DB().QueryRowContext(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&schemaVersion)
		bundle["schema_version"] = schemaVersion

		var deviceCount, orgCount, operatorCount int
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&deviceCount)
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM organizations WHERE is_active = 1`).Scan(&orgCount)
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM operators`).Scan(&operatorCount)
		bundle["device_count"] = deviceCount
		bundle["org_count"] = orgCount
		bundle["operator_count"] = operatorCount
	}

	c.Header("Content-Disposition", "attachment; filename=vyzorix-support-bundle.json")
	c.JSON(http.StatusOK, bundle)
}
