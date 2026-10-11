package admin

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var _ openapi.SupportBundleResult

type SupportBundleHandler struct {
	db *storage.SQLite
}

func NewSupportBundleHandler(db *storage.SQLite) *SupportBundleHandler {
	return &SupportBundleHandler{db: db}
}

// GetBundle handles GET /v1/admin/support-bundle.
// @Summary      Get support bundle
// @Description  Returns a diagnostic support bundle with runtime and DB stats (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.SupportBundleResult  "support bundle"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Router       /admin/support-bundle [get]
func (h *SupportBundleHandler) GetBundle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	name, _ := os.Hostname()
	bundle := openapi.SupportBundleResult{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Hostname:    name,
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		GoMaxProcs:  runtime.GOMAXPROCS(0),
		GoNumCpu:    runtime.NumCPU(),
	}

	if h.db != nil {
		var schemaVersion int
		_ = h.db.DB().QueryRowContext(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&schemaVersion)
		bundle.SchemaVersion = schemaVersion

		var deviceCount, orgCount, operatorCount int
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&deviceCount)
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM organizations WHERE is_active = 1`).Scan(&orgCount)
		_ = h.db.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM operators`).Scan(&operatorCount)
		bundle.DeviceCount = deviceCount
		bundle.OrgCount = orgCount
		bundle.OperatorCount = operatorCount
	}

	c.Header("Content-Disposition", "attachment; filename=vyzorix-support-bundle.json")
	c.JSON(http.StatusOK, bundle)
}
