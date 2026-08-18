package command

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
	domainerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeAudit captures CommandExecuted events so tests can assert on what the
// handler would persist to the audit trail.
type fakeAudit struct {
	events []audit.CommandExecutedEvent
}

func (f *fakeAudit) CommandExecuted(_ context.Context, e audit.CommandExecutedEvent) {
	f.events = append(f.events, e)
}

// fakeConfirmation is a configurable ConfirmationConsumer for tests. By
// default ConsumeForCommand succeeds (returns the profile); tests flip
// consumeErr to simulate invalid/expired/consumed/mismatched tokens.
type fakeConfirmation struct {
	consumeErr error
	calls      int
}

func (f *fakeConfirmation) ConsumeForCommand(_ *gin.Context, _, _, _, _ string) (*domaincommand.CommandRiskProfile, error) {
	f.calls++
	if f.consumeErr != nil {
		return nil, f.consumeErr
	}
	p := domaincommand.LookupRiskProfile("device.reboot")
	return &p, nil
}

// newHandlerWithFakes builds an ExecuteHandler wired with a fake audit logger,
// a real risk evaluator, and (optionally) a fake confirmation consumer.
func newHandlerWithFakes(t *testing.T, confirm ConfirmationConsumer) (*ExecuteHandler, *fakeAudit) {
	t.Helper()
	f := &fakeAudit{}
	h := &ExecuteHandler{
		riskEvaluator: domaincommand.NewRiskEvaluator(),
		audit:         f,
		confirmations: confirm,
	}
	return h, f
}

func newContextWithActor(op *operator.Operator) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/device/imei/command", nil)
	c.Set(middleware.ContextKeyOrganizationID, "org-1")
	if op != nil {
		c.Set(middleware.ContextKeyOperator, op)
	}
	return c, w
}

// withMFASession sets an MFA-verified session in the gin context, mirroring
// the cookie-auth middleware, so the handler can derive MFAVerified=true.
func withMFASession(c *gin.Context) {
	now := time.Now()
	c.Set(middleware.ContextKeySession, &session.Session{ID: "sess-1", MFAVerifiedAt: &now})
}

func TestAuthorizeCommand_AllowsLowRisk(t *testing.T) {
	h, aud := newHandlerWithFakes(t, nil)
	c, _ := newContextWithActor(&operator.Operator{ID: "op-1"})

	if !h.authorizeCommand(c, commandRequest{Command: domaincommand.TypeCheckUpdate}, "imei-1") {
		t.Fatal("expected low-risk command to be authorized")
	}
	if len(aud.events) != 0 {
		t.Errorf("expected no audit event on allow, got %d", len(aud.events))
	}
}

func TestAuthorizeCommand_BlocksHighRiskWithoutToken(t *testing.T) {
	h, aud := newHandlerWithFakes(t, &fakeConfirmation{})
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.reboot"}, "imei-1") {
		t.Fatal("expected high-risk command without token to be blocked")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
	if len(aud.events) != 1 {
		t.Fatalf("expected 1 blocked audit event, got %d", len(aud.events))
	}
	ev := aud.events[0]
	if ev.Result != audit.ResultBlocked {
		t.Errorf("audit result = %s, want blocked", ev.Result)
	}
	if ev.RiskTier != "high" {
		t.Errorf("audit risk_tier = %s, want high", ev.RiskTier)
	}
	if ev.Reason != "confirmation required" {
		t.Errorf("audit reason = %q, want 'confirmation required'", ev.Reason)
	}
	if ev.OperatorID != "op-1" {
		t.Errorf("audit operator_id = %q, want op-1", ev.OperatorID)
	}
}

func TestAuthorizeCommand_AllowsHighRiskWithValidToken(t *testing.T) {
	confirm := &fakeConfirmation{}
	h, aud := newHandlerWithFakes(t, confirm)
	c, _ := newContextWithActor(&operator.Operator{ID: "op-1"})

	if !h.authorizeCommand(c, commandRequest{Command: "device.reboot", ConfirmationToken: "tok-123"}, "imei-1") {
		t.Fatal("expected high-risk command with valid token to be authorized")
	}
	if confirm.calls != 1 {
		t.Errorf("expected confirmation consumer to be called once, got %d", confirm.calls)
	}
	if len(aud.events) != 0 {
		t.Errorf("expected no audit event on confirmed allow, got %d", len(aud.events))
	}
}

func TestAuthorizeCommand_BlocksHighRiskWhenConfirmationsDisabled(t *testing.T) {
	// A nil confirmation consumer means confirmations are disabled: even with a
	// token, the command must be blocked.
	h, aud := newHandlerWithFakes(t, nil)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.reboot", ConfirmationToken: "tok-123"}, "imei-1") {
		t.Fatal("expected high-risk command to be blocked when confirmations are disabled")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
	if len(aud.events) != 1 || aud.events[0].Reason != "confirmation required" {
		t.Fatalf("expected 1 blocked audit event, got %v", aud.events)
	}
}

func TestAuthorizeCommand_BlocksHighRiskWithExpiredToken(t *testing.T) {
	confirm := &fakeConfirmation{consumeErr: domainconfirmation.ErrExpired}
	h, aud := newHandlerWithFakes(t, confirm)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.reboot", ConfirmationToken: "tok-expired"}, "imei-1") {
		t.Fatal("expected expired token to be rejected")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
	if len(aud.events) != 1 || aud.events[0].Reason != "confirmation required" {
		t.Fatalf("expected blocked audit with reason 'confirmation required', got %v", aud.events)
	}
}

func TestAuthorizeCommand_BlocksHighRiskWithMismatchedToken(t *testing.T) {
	confirm := &fakeConfirmation{consumeErr: domainconfirmation.ErrMismatch}
	h, _ := newHandlerWithFakes(t, confirm)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.reboot", ConfirmationToken: "tok-other"}, "imei-1") {
		t.Fatal("expected mismatched token to be rejected")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
}

func TestAuthorizeCommand_BlocksCriticalWithoutMFAEvenWithToken(t *testing.T) {
	// Critical-tier requires MFA; a token alone is insufficient without an
	// MFA-verified session.
	confirm := &fakeConfirmation{}
	h, aud := newHandlerWithFakes(t, confirm)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.factory_reset", ConfirmationToken: "tok-1"}, "imei-1") {
		t.Fatal("expected critical command without MFA to be blocked")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
	if len(aud.events) != 1 || aud.events[0].RiskTier != "critical" {
		t.Fatalf("expected 1 blocked critical audit event, got %v", aud.events)
	}
}

func TestAuthorizeCommand_AllowsCriticalWithMFAAndValidToken(t *testing.T) {
	// With an MFA-verified session AND a valid confirmation token, the
	// evaluator returns Allow for a critical command and the handler proceeds.
	confirm := &fakeConfirmation{}
	h, aud := newHandlerWithFakes(t, confirm)
	c, _ := newContextWithActor(&operator.Operator{ID: "op-1"})
	withMFASession(c)

	if !h.authorizeCommand(c, commandRequest{Command: "device.factory_reset", ConfirmationToken: "tok-1"}, "imei-1") {
		t.Fatal("expected critical command with MFA + valid token to be authorized")
	}
	if confirm.calls != 1 {
		t.Errorf("expected confirmation consumer to be called once, got %d", confirm.calls)
	}
	if len(aud.events) != 0 {
		t.Errorf("expected no audit event on allowed critical, got %d", len(aud.events))
	}
}

func TestAuthorizeCommand_BlocksCriticalWithMFAButNoToken(t *testing.T) {
	confirm := &fakeConfirmation{}
	h, aud := newHandlerWithFakes(t, confirm)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})
	withMFASession(c)

	if h.authorizeCommand(c, commandRequest{Command: "device.factory_reset"}, "imei-1") {
		t.Fatal("expected critical command with MFA but no token to be blocked")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
	if len(aud.events) != 1 || aud.events[0].Reason != "confirmation required" {
		t.Fatalf("expected blocked audit for missing token, got %v", aud.events)
	}
}

// A generic consume error (not a known sentinel) still results in a 425 block.
func TestAuthorizeCommand_BlocksOnUnknownConsumeError(t *testing.T) {
	confirm := &fakeConfirmation{consumeErr: errors.New("db unavailable")}
	h, _ := newHandlerWithFakes(t, confirm)
	c, w := newContextWithActor(&operator.Operator{ID: "op-1"})

	if h.authorizeCommand(c, commandRequest{Command: "device.reboot", ConfirmationToken: "tok-1"}, "imei-1") {
		t.Fatal("expected unknown consume error to block")
	}
	if w.Code != http.StatusTooEarly {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooEarly)
	}
}

func TestValidateCommandRequest_ReturnsFieldDetails(t *testing.T) {
	h, _ := newHandlerWithFakes(t, nil)
	verr := h.validateCommandRequest("", "", "")
	if verr == nil {
		t.Fatal("expected a validation error for empty imei/command")
	}
	details, ok := domainerrors.ValidationDetailsOf(verr)
	if !ok {
		t.Fatalf("expected a ValidationError, got %T: %v", verr, verr)
	}
	// Empty imei + empty command must each produce a field detail.
	fields := map[string]bool{}
	for _, d := range details {
		fields[d.Field] = true
	}
	if !fields["deviceId"] {
		t.Errorf("expected a deviceId detail, got %+v", details)
	}
	if !fields["command"] {
		t.Errorf("expected a command detail, got %+v", details)
	}
}

func TestValidateCommandRequest_NilWhenValid(t *testing.T) {
	h, _ := newHandlerWithFakes(t, nil)
	if verr := h.validateCommandRequest("imei-1", domaincommand.TypeCheckUpdate, ""); verr != nil {
		t.Fatalf("expected nil for valid input, got %v", verr)
	}
}
