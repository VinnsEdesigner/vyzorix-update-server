package confirmation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeRepo implements confirmation.Repository in memory for handler tests.
type fakeRepo struct {
	created []*confirmation.PendingConfirmation
}

func (r *fakeRepo) Create(_ context.Context, c *confirmation.PendingConfirmation) error {
	r.created = append(r.created, c)
	return nil
}
func (r *fakeRepo) Get(_ context.Context, _ string) (*confirmation.PendingConfirmation, error) {
	return nil, confirmation.ErrNotFound
}
func (r *fakeRepo) Consume(ctx context.Context, token string, at time.Time) (*confirmation.PendingConfirmation, error) {
	return nil, confirmation.ErrNotFound
}
func (r *fakeRepo) DeleteExpired(context.Context) (int64, error) { return 0, nil }

func newConfirmHandler(t *testing.T) (*Handler, *fakeRepo) {
	t.Helper()
	repo := &fakeRepo{}
	svc := appconfirmation.NewService(repo)
	return NewHandler(svc, nil, command.NewRiskEvaluator()), repo
}

func newRequestContext(op *operator.Operator, body string) (*gin.Context, *httptest.ResponseRecorder) {
	const imei = "imei-1"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost,
		"/v1/devices/"+imei+"/command/confirm", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextKeyOrganizationID, "org-1")
	c.Params = gin.Params{{Key: "imei", Value: imei}}
	if op != nil {
		c.Set(middleware.ContextKeyOperator, op)
	}
	return c, w
}

func TestRequestConfirmation_IssuesTokenForRiskyCommand(t *testing.T) {
	h, repo := newConfirmHandler(t)
	c, w := newRequestContext(&operator.Operator{ID: "op-1"}, `{"command":"device.reboot"}`)

	h.RequestConfirmation(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 confirmation created, got %d", len(repo.created))
	}
	pc := repo.created[0]
	if pc.OperatorID != "op-1" || pc.Command != "device.reboot" || pc.DeviceID != "imei-1" {
		t.Errorf("unexpected confirmation: %+v", pc)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["confirmation_token"] == "" || resp["confirmation_required"] != true {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestRequestConfirmation_NotRequiredForLowRisk(t *testing.T) {
	h, repo := newConfirmHandler(t)
	c, w := newRequestContext(&operator.Operator{ID: "op-1"}, `{"command":"CHECK_UPDATE"}`)

	h.RequestConfirmation(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if len(repo.created) != 0 {
		t.Errorf("expected no token for low-risk command, got %d created", len(repo.created))
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["confirmation_required"] != false {
		t.Errorf("expected confirmation_required=false, got %v", resp["confirmation_required"])
	}
}

func TestRequestConfirmation_RejectsMissingCommand(t *testing.T) {
	h, _ := newConfirmHandler(t)
	c, w := newRequestContext(&operator.Operator{ID: "op-1"}, `{}`)

	h.RequestConfirmation(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestRequestConfirmation_RejectsMissingOperator(t *testing.T) {
	h, _ := newConfirmHandler(t)
	c, w := newRequestContext(nil, `{"command":"device.reboot"}`)

	h.RequestConfirmation(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
