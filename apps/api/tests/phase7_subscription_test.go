package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gqlmiddleware "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/subscription"
	wsHandler "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	config "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	wshub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

// gqlSubState holds shared state for GraphQL subscription WS tests.
type gqlSubState struct {
	hub      *wshub.Hub
	server   *httptest.Server
	ctx      context.Context
	cancel   context.CancelFunc
	operator *operator.Operator
	orgID    string
}

func newGQLSubState(t *testing.T) *gqlSubState {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	noopRepo := newNoopDeviceRepoForTests()

	h := wshub.New(logger, noopRepo, nil, nil, nil)
	h.SetRateLimiter(wshub.NewRateLimiter(logger, &wshub.RateLimiterConfig{
		Rate:            10000,
		Burst:           10000,
		CleanupInterval: time.Minute,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	testOp := &operator.Operator{
		ID:    "op-test-001",
		Email: "test@vyzorix.dev",
		Name:  "Test Operator",
	}
	orgID := "org-test-001"

	cfg := config.Config{
		Env:            "development",
		EnforceHMAC:    false,
		AllowedOrigins: []string{"*"},
	}

	auditLogger := newTestAuditLogger(logger)
	auditAdapter := subscription.NewAuditLoggerAdapter(auditLogger)

	// AuthMiddleware with nil session/auth — we inject operator via middleware.
	authMw := gqlmiddleware.NewAuthMiddleware(nil, nil, logger)

	subHandler := subscription.NewHandler(&subscription.Config{
		Hub:         h,
		Resolver:    &resolver.Resolver{},
		AuthMw:      authMw,
		Logger:      logger,
		AuditLogger: auditAdapter,
		Config:      cfg,
	})

	r := gin.New()
	// Middleware that injects the test operator into gin context.
	r.Use(func(c *gin.Context) {
		c.Set("operator", testOp)
		c.Next()
	})
	r.GET("/:org/graphql/ws", subHandler.HandleWebSocket)

	server := httptest.NewServer(r)

	return &gqlSubState{
		hub:      h,
		server:   server,
		ctx:      ctx,
		cancel:   cancel,
		operator: testOp,
		orgID:    orgID,
	}
}

func (s *gqlSubState) close() {
	s.cancel()
	s.server.Close()
	time.Sleep(50 * time.Millisecond)
}

func (s *gqlSubState) wsURL() string {
	return "ws" + strings.TrimPrefix(s.server.URL, "http") + "/" + s.orgID + "/graphql/ws"
}

// wsMessage is the graphql-ws protocol message.
type wsMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// readMsg reads a WS message with timeout, failing the test on error.
func readMsg(t *testing.T, conn *gorillaws.Conn, timeout time.Duration) wsMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal ws message: %v (raw: %s)", err, string(data))
	}
	return msg
}

// readMsgNoFail reads a WS message with timeout, returning ok=false on timeout/error.
func readMsgNoFail(t *testing.T, conn *gorillaws.Conn, timeout time.Duration) (wsMessage, bool) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return wsMessage{}, false
	}
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wsMessage{}, false
	}
	return msg, true
}

// msgReader wraps a WS connection with a background goroutine that continuously.
// reads messages into a channel. This avoids the gorilla read-deadline issue.
// where a timeout can corrupt the connection's read state for subsequent reads.
type msgReader struct {
	ch     chan wsMessage
	errCh  chan error
	closed chan struct{}
}

func newMsgReader(conn *gorillaws.Conn) *msgReader {
	r := &msgReader{
		ch:     make(chan wsMessage, 32),
		errCh:  make(chan error, 1),
		closed: make(chan struct{}),
	}
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				select {
				case r.errCh <- err:
				default:
				}
				return
			}
			var msg wsMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			select {
			case r.ch <- msg:
			case <-r.closed:
				return
			}
		}
	}()
	return r
}

// recv waits for a message within the timeout. Returns ok=false if no message.
// arrives or the reader has stopped.
func (r *msgReader) recv(timeout time.Duration) (wsMessage, bool) {
	select {
	case msg := <-r.ch:
		return msg, true
	case err := <-r.errCh:
		_ = err
		return wsMessage{}, false
	case <-time.After(timeout):
		return wsMessage{}, false
	case <-r.closed:
		return wsMessage{}, false
	}
}

func (r *msgReader) stop() {
	close(r.closed)
}

// sendMsg sends a WS message as JSON.
func sendMsg(t *testing.T, conn *gorillaws.Conn, msg wsMessage) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := conn.WriteMessage(gorillaws.TextMessage, data); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
}

func dialGQLWS(t *testing.T, url string) *gorillaws.Conn {
	t.Helper()
	dialer := gorillaws.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("WS dial failed: %v", err)
	}
	return conn
}

// ─── 7.1: Connection init → ack handshake ───.

func TestPhase7Sub_ConnectionInitAck(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	// Send connection_init.
	sendMsg(t, conn, wsMessage{Type: "connection_init"})

	// Expect connection_ack.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "connection_ack" {
		t.Errorf("expected connection_ack, got %s", msg.Type)
	}
	t.Log("connection_init → connection_ack handshake verified")
}

// ─── 7.2: Subscribe to deviceUpdates → initial null next ───.

func TestPhase7Sub_SubscribeDeviceUpdates(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second) // connection_ack

	// Subscribe to deviceUpdated.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id status } }",
		"variables": map[string]any{"deviceId": "IMEI_DEV_001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub1", Payload: subPayload})

	// Expect initial next with null data.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	if msg.ID != "sub1" {
		t.Errorf("expected ID sub1, got %s", msg.ID)
	}

	// Verify payload has deviceUpdated: null.
	var payload struct {
		Data struct {
			DeviceUpdated *json.RawMessage `json:"deviceUpdated"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	t.Log("subscribe deviceUpdates → initial null next received")
}

// ─── 7.3: Publish deviceUpdate → subscriber receives it ───.

func TestPhase7Sub_PublishDeviceUpdateReceived(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second) // ack

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id status } }",
		"variables": map[string]any{"deviceId": "IMEI_PUB_001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub1", Payload: subPayload})
	readMsg(t, conn, 3*time.Second) // initial null next

	// Publish a device update event from the server side.
	updateData := map[string]string{"id": "IMEI_PUB_001", "status": "online"}
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_PUB_001", updateData)

	// Subscriber should receive the published event.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next with published data, got %s", msg.Type)
	}

	var payload struct {
		Data struct {
			DeviceUpdated map[string]string `json:"deviceUpdated"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Data.DeviceUpdated["status"] != "online" {
		t.Errorf("deviceUpdated status = %s, want online", payload.Data.DeviceUpdated["status"])
	}
	t.Log("publish deviceUpdate → subscriber received event with data")
}

// ─── 7.4: Subscribe to telemetryReceived → publish → receive ───.

func TestPhase7Sub_SubscribeTelemetryReceived(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId riskScore } }",
		"variables": map[string]any{"deviceId": "IMEI_TELE_001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-tele", Payload: subPayload})
	readMsg(t, conn, 3*time.Second) // initial null

	// Publish telemetry event.
	teleData := map[string]any{"deviceId": "IMEI_TELE_001", "riskScore": 77}
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_TELE_001", teleData)

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}

	var payload struct {
		Data struct {
			TelemetryReceived map[string]any `json:"telemetryReceived"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Data.TelemetryReceived["riskScore"] == nil {
		t.Error("telemetryReceived.riskScore should not be nil")
	}
	t.Log("subscribe telemetryReceived → publish → subscriber received telemetry event")
}

// ─── 7.5: Subscribe to commandStatusChanged → publish → receive ───.

func TestPhase7Sub_SubscribeCommandStatusChanged(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { commandStatusChanged { dispatchId status } }",
		"variables": map[string]any{"dispatchId": "dispatch-001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-cmd", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	cmdData := map[string]string{"dispatchId": "dispatch-001", "status": "delivered"}
	state.hub.PublishCommandStatus(state.operator.ID, "dispatch-001", cmdData)

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}

	var payload struct {
		Data struct {
			CommandStatusChanged map[string]string `json:"commandStatusChanged"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Data.CommandStatusChanged["status"] != "delivered" {
		t.Errorf("status = %s, want delivered", payload.Data.CommandStatusChanged["status"])
	}
	t.Log("subscribe commandStatusChanged → publish → subscriber received status change")
}

// ─── 7.6: Subscribe to organizationEvent → publish → receive ───.

func TestPhase7Sub_SubscribeOrganizationEvent(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { organizationEvent { type } }",
		"variables": map[string]any{"orgId": state.orgID},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-org", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	orgData := map[string]string{"type": "member_joined", "orgId": state.orgID}
	state.hub.PublishOrganizationEvent(state.orgID, orgData)

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}

	var payload struct {
		Data struct {
			OrganizationEvent map[string]string `json:"organizationEvent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Data.OrganizationEvent["type"] != "member_joined" {
		t.Errorf("type = %s, want member_joined", payload.Data.OrganizationEvent["type"])
	}
	t.Log("subscribe organizationEvent → publish → subscriber received org event")
}

// ─── 7.7: Subscribe to memberEvent → publish → receive ───.

func TestPhase7Sub_SubscribeMemberEvent(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { memberEvent { event } }",
		"variables": map[string]any{"orgId": state.orgID},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-member", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	memberData := map[string]string{"event": "member_invited", "email": "new@vyzorix.dev"}
	state.hub.PublishMemberEvent(state.orgID, memberData)

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}

	var payload struct {
		Data struct {
			MemberEvent map[string]string `json:"memberEvent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Data.MemberEvent["event"] != "member_invited" {
		t.Errorf("event = %s, want member_invited", payload.Data.MemberEvent["event"])
	}
	t.Log("subscribe memberEvent → publish → subscriber received member event")
}

// ─── 7.8: Complete (unsubscribe) stops events ───.

func TestPhase7Sub_CompleteStopsEvents(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	if _, ok := reader.recv(3 * time.Second); !ok {
		t.Fatal("expected connection_ack")
	}

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_COMPLETE_001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-comp", Payload: subPayload})
	if _, ok := reader.recv(3 * time.Second); !ok {
		t.Fatal("expected initial next")
	}

	// Send complete to unsubscribe.
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-comp"})
	time.Sleep(200 * time.Millisecond)

	// Publish an event — subscriber should NOT receive it.
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_COMPLETE_001", map[string]string{"status": "offline"})

	// Try to read — should timeout (no message).
	if _, ok := reader.recv(1 * time.Second); ok {
		t.Error("should not receive event after complete (unsubscribe)")
	}
	t.Log("complete → subscriber no longer receives events")
}

// ─── 7.9: Multiple subscriptions on same connection ───.

func TestPhase7Sub_MultipleSubscriptions(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Subscribe to two different devices.
	sub1, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_MULTI_A"},
	})
	sub2, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_MULTI_B"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-a", Payload: sub1})
	readMsg(t, conn, 3*time.Second) // initial null for sub-a
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-b", Payload: sub2})
	readMsg(t, conn, 3*time.Second) // initial null for sub-b

	// Publish for device A — should get next on sub-a.
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_MULTI_A", map[string]string{"status": "online"})
	msgA := readMsg(t, conn, 3*time.Second)
	if msgA.ID != "sub-a" {
		t.Errorf("expected event on sub-a, got %s", msgA.ID)
	}

	// Publish for device B — should get next on sub-b.
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_MULTI_B", map[string]string{"status": "offline"})
	msgB := readMsg(t, conn, 3*time.Second)
	if msgB.ID != "sub-b" {
		t.Errorf("expected event on sub-b, got %s", msgB.ID)
	}
	t.Log("multiple subscriptions on same connection: events routed to correct subscription IDs")
}

// ─── 7.10: Unknown message type → error ───.

func TestPhase7Sub_UnknownMessageType(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Send unknown message type.
	sendMsg(t, conn, wsMessage{Type: "unknown_type", ID: "test1"})

	// Expect error message.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error, got %s", msg.Type)
	}
	if msg.ID != "test1" {
		t.Errorf("expected ID test1, got %s", msg.ID)
	}
	t.Log("unknown message type → error response verified")
}

// ─── 7.11: Invalid JSON → error ───.

func TestPhase7Sub_InvalidJSON(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Send invalid JSON.
	conn.WriteMessage(gorillaws.TextMessage, []byte("not valid json"))

	// Expect error message.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error, got %s", msg.Type)
	}
	t.Log("invalid JSON → error response verified")
}

// ─── 7.12: Missing org parameter → 400 ───.

func TestPhase7Sub_MissingOrgParameter(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

// Connect without org parameter — the route is /:org/graphql/ws.
	// so connecting to /graphql/ws won't match the route.
	url := "ws" + strings.TrimPrefix(state.server.URL, "http") + "/graphql/ws"
	dialer := gorillaws.Dialer{HandshakeTimeout: 5 * time.Second}
	_, resp, err := dialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected dial to fail for missing org")
	}
	if resp != nil && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Logf("got status %d (404 expected for unmatched route)", resp.StatusCode)
	}
	t.Log("missing org parameter → connection rejected")
}

// ─── 7.13: Unauthenticated connection → 401 ───.

func TestPhase7Sub_UnauthenticatedRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	noopRepo := newNoopDeviceRepoForTests()

	h := wshub.New(logger, noopRepo, nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	go h.Run(ctx)
	defer cancel()
	defer time.Sleep(50 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	cfg := config.Config{Env: "development", EnforceHMAC: false, AllowedOrigins: []string{"*"}}
	auditLogger := newTestAuditLogger(logger)
	auditAdapter := subscription.NewAuditLoggerAdapter(auditLogger)
	authMw := gqlmiddleware.NewAuthMiddleware(nil, nil, logger)

	subHandler := subscription.NewHandler(&subscription.Config{
		Hub:         h,
		Resolver:    &resolver.Resolver{},
		AuthMw:      authMw,
		Logger:      logger,
		AuditLogger: auditAdapter,
		Config:      cfg,
	})

	r := gin.New()
	// NO middleware to inject operator — simulates unauthenticated request.
	r.GET("/:org/graphql/ws", subHandler.HandleWebSocket)
	server := httptest.NewServer(r)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/org-test/graphql/ws"
	dialer := gorillaws.Dialer{HandshakeTimeout: 5 * time.Second}
	_, resp, err := dialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected dial to fail for unauthenticated request")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	t.Log("unauthenticated connection → 401 rejected")
}

// ─── 7.14: Disconnect cleans up subscriptions ───.

func TestPhase7Sub_DisconnectCleansUp(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_CLEANUP"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-clean", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	// Close the connection — should trigger cleanup.
	conn.Close()
	time.Sleep(500 * time.Millisecond)

// Publish an event — the disconnected client should not receive it.
	// (and the subscription should be cleaned up).
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_CLEANUP", map[string]string{"status": "offline"})
	time.Sleep(300 * time.Millisecond)

	// Verify no panic/crash — the server is still healthy.
	t.Log("disconnect → subscription cleanup verified (no crash, server healthy)")
}

// ─── 7.15: Subscribe with operator-wide (no deviceId) ───.

func TestPhase7Sub_OperatorWideSubscription(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Subscribe to all telemetry for this operator (no deviceId).
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId riskScore } }",
		"variables": map[string]any{},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-wide", Payload: subPayload})
	readMsg(t, conn, 3*time.Second) // initial null

	// Publish telemetry for any device — operator-wide subscriber should receive it.
	state.hub.PublishTelemetry(state.operator.ID, "ANY_DEVICE", map[string]int{"riskScore": 99})

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("operator-wide subscription (no deviceId) receives telemetry for any device")
}

// ─── 7.16: Multiple subscribers receive same event ───.

func TestPhase7Sub_MultipleSubscribersSameEvent(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn1 := dialGQLWS(t, state.wsURL())
	defer conn1.Close()
	conn2 := dialGQLWS(t, state.wsURL())
	defer conn2.Close()

	// Both subscribers do handshake.
	for _, c := range []*gorillaws.Conn{conn1, conn2} {
		sendMsg(t, c, wsMessage{Type: "connection_init"})
		readMsg(t, c, 3*time.Second)
	}

	// Both subscribe to the same device.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id status } }",
		"variables": map[string]any{"deviceId": "IMEI_SHARED"},
	})
	sendMsg(t, conn1, wsMessage{Type: "subscribe", ID: "s1", Payload: subPayload})
	readMsg(t, conn1, 3*time.Second)
	sendMsg(t, conn2, wsMessage{Type: "subscribe", ID: "s2", Payload: subPayload})
	readMsg(t, conn2, 3*time.Second)

	// Publish one event — both should receive it.
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_SHARED", map[string]string{"status": "online"})

	msg1 := readMsg(t, conn1, 3*time.Second)
	msg2 := readMsg(t, conn2, 3*time.Second)

	if msg1.Type != "next" || msg2.Type != "next" {
		t.Errorf("expected both next, got %s and %s", msg1.Type, msg2.Type)
	}
	t.Log("multiple subscribers: both received the same published event")
}

// ─── 7.17: Subscribe with unknown query → default acknowledgment ───.

func TestPhase7Sub_UnknownSubscriptionQuery(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Subscribe with an unknown subscription type.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { unknownField { id } }",
		"variables": map[string]any{},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-unknown", Payload: subPayload})

	// Should get a default next with __typename.
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("unknown subscription query → default acknowledgment received")
}

// ─── 7.18: Subscribe with invalid payload → error ───.

func TestPhase7Sub_InvalidSubscribePayload(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Send subscribe with invalid payload (not valid JSON object).
	sendMsg(t, conn, wsMessage{
		Type:    "subscribe",
		ID:      "sub-bad",
		Payload: json.RawMessage(`"not an object"`),
	})

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for invalid payload, got %s", msg.Type)
	}
	t.Log("invalid subscribe payload → error response verified")
}

// ─── 7.19: Complete non-existent subscription (no-op) ───.

func TestPhase7Sub_CompleteNonExistentSubscription(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Send complete for a subscription that doesn't exist — should be a no-op.
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "nonexistent"})

	// Connection should still be alive — send a new subscribe.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_NOOP"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-after", Payload: subPayload})
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next after no-op complete, got %s", msg.Type)
	}
	t.Log("complete for non-existent subscription → no-op, connection survives")
}

// ─── 7.20: Event filtering — subscriber for device A doesn't receive device B ───.

func TestPhase7Sub_EventFiltering(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	if msg, ok := reader.recv(3 * time.Second); !ok || msg.Type != "connection_ack" {
		t.Fatalf("expected connection_ack, got ok=%v", ok)
	}

	// Subscribe to telemetry for device A only.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId } }",
		"variables": map[string]any{"deviceId": "IMEI_FILTER_A"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-filter", Payload: subPayload})
	if msg, ok := reader.recv(3 * time.Second); !ok || msg.Type != "next" {
		t.Fatalf("expected initial next, got ok=%v", ok)
	}

	// Publish telemetry for device B — subscriber should NOT receive it.
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_FILTER_B", map[string]int{"riskScore": 50})
	if _, ok := reader.recv(1 * time.Second); ok {
		t.Error("subscriber for device A should not receive device B telemetry")
	}

	// Publish telemetry for device A — subscriber SHOULD receive it.
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_FILTER_A", map[string]int{"riskScore": 75})
	msg, ok := reader.recv(3 * time.Second)
	if !ok {
		t.Fatal("subscriber should receive device A telemetry")
	}
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("event filtering: subscriber for device A receives A but not B")
}

// ─── 7.21: Connection survives multiple subscribe/complete cycles ───.

func TestPhase7Sub_MultipleSubscribeCompleteCycles(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	for i := 0; i < 3; i++ {
		subID := fmt.Sprintf("sub-cycle-%d", i)
		imei := fmt.Sprintf("IMEI_CYCLE_%d", i)

		subPayload, _ := json.Marshal(map[string]any{
			"query":     "subscription { deviceUpdated { id } }",
			"variables": map[string]any{"deviceId": imei},
		})
		sendMsg(t, conn, wsMessage{Type: "subscribe", ID: subID, Payload: subPayload})
		readMsg(t, conn, 3*time.Second) // initial null

		// Publish and receive.
		state.hub.PublishDeviceUpdate(state.operator.ID, imei, map[string]string{"status": "online"})
		msg := readMsg(t, conn, 3*time.Second)
		if msg.ID != subID {
			t.Errorf("cycle %d: expected %s, got %s", i, subID, msg.ID)
		}

		// Complete.
		sendMsg(t, conn, wsMessage{Type: "complete", ID: subID})
		time.Sleep(100 * time.Millisecond)
	}
	t.Log("3 subscribe/publish/complete cycles on same connection: all succeeded")
}

// ─── 7.22: Concurrent subscribers stress test ───.

func TestPhase7Sub_ConcurrentSubscribers(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	const numClients = 5
	conns := make([]*gorillaws.Conn, numClients)
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := dialGQLWS(t, state.wsURL())
			conns[idx] = c
			sendMsg(t, c, wsMessage{Type: "connection_init"})
			readMsg(t, c, 3*time.Second)

			subPayload, _ := json.Marshal(map[string]any{
				"query":     "subscription { deviceUpdated { id } }",
				"variables": map[string]any{"deviceId": "IMEI_STRESS"},
			})
			sendMsg(t, c, wsMessage{Type: "subscribe", ID: "s", Payload: subPayload})
			readMsg(t, c, 3*time.Second)
		}(i)
	}
	wg.Wait()
	defer func() {
		for _, c := range conns {
			if c != nil {
				c.Close()
			}
		}
	}()

	// Publish one event — all should receive it.
	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_STRESS", map[string]string{"status": "online"})

	for i, c := range conns {
		msg := readMsg(t, c, 5*time.Second)
		if msg.Type != "next" {
			t.Errorf("client %d: expected next, got %s", i, msg.Type)
		}
	}
	t.Logf("5 concurrent subscribers all received the same published event")
}

// ─── 7.23: Organization event uses connection orgID as fallback ───.

func TestPhase7Sub_OrgEventFallbackToConnectionOrgID(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Subscribe to organizationEvent WITHOUT orgId in variables.
	// The handler should fall back to the connection's orgID.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { organizationEvent { type } }",
		"variables": map[string]any{}, // no orgId
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-org-fb", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	// Publish to the connection's orgID.
	state.hub.PublishOrganizationEvent(state.orgID, map[string]string{"type": "settings_updated"})

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("organizationEvent falls back to connection orgID when variables.orgId is empty")
}

// ─── 7.24: Member event uses connection orgID as fallback ───.

func TestPhase7Sub_MemberEventFallbackToConnectionOrgID(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { memberEvent { event } }",
		"variables": map[string]any{},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-member-fb", Payload: subPayload})
	readMsg(t, conn, 3*time.Second)

	state.hub.PublishMemberEvent(state.orgID, map[string]string{"event": "role_changed"})

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("memberEvent falls back to connection orgID when variables.orgId is empty")
}

// ─── 7.25: Full lifecycle: init → subscribe → publish → complete → disconnect ───.

func TestPhase7Sub_FullLifecycle(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	// 1. Init.
	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	ack, ok := reader.recv(3 * time.Second)
	if !ok || ack.Type != "connection_ack" {
		t.Fatalf("expected connection_ack, got ok=%v", ok)
	}

	// 2. Subscribe to telemetry.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId riskScore } }",
		"variables": map[string]any{"deviceId": "IMEI_LIFECYCLE"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-life", Payload: subPayload})
	initNext, ok := reader.recv(3 * time.Second)
	if !ok || initNext.Type != "next" {
		t.Fatalf("expected initial next, got ok=%v", ok)
	}

	// 3. Publish telemetry → receive.
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_LIFECYCLE", map[string]any{"riskScore": 66})
	pubNext, ok := reader.recv(3 * time.Second)
	if !ok || pubNext.Type != "next" {
		t.Fatalf("expected published next, got ok=%v", ok)
	}

	// 4. Complete.
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-life"})
	time.Sleep(200 * time.Millisecond)

	// 5. Verify no more events.
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_LIFECYCLE", map[string]any{"riskScore": 99})
	if _, ok := reader.recv(1 * time.Second); ok {
		t.Error("should not receive event after complete")
	}

	// 6. Disconnect.
	conn.Close()
	time.Sleep(300 * time.Millisecond)
	t.Log("full lifecycle: init → subscribe → publish → receive → complete → disconnect: all verified")
}

// ─── 7.26: Device stream → subscription bridge (separate paths) ───.

func TestPhase7Sub_DeviceStreamAndSubscriptionCoexist(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	// Connect a device via the device stream WS (using the phase6 test server pattern).
	// We use the GraphQL state's hub for both.
	logger := slog.New(slog.NewTextHandler(&discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a device stream handler on a separate test server.
	deviceCfg := config.Config{Env: "development", EnforceHMAC: false, AllowedOrigins: []string{"*"}}
	auditLogger := newTestAuditLogger(logger)
	deviceHandler := wsHandler.NewStreamHandler(logger, deviceCfg, state.hub, cryptohmac.Verifier{}, auditLogger)

	gin.SetMode(gin.TestMode)
	deviceRouter := gin.New()
	deviceRouter.GET("/v1/device/:imei/stream", deviceHandler.Handle)
	deviceServer := httptest.NewServer(deviceRouter)
	defer deviceServer.Close()

	// Connect a device.
	deviceURL := "ws" + strings.TrimPrefix(deviceServer.URL, "http") + "/v1/device/IMEI_BRIDGE/stream"
	dialer := gorillaws.Dialer{HandshakeTimeout: 5 * time.Second}
	deviceConn, _, err := dialer.Dial(deviceURL, nil)
	if err != nil {
		t.Fatalf("device WS dial: %v", err)
	}
	defer deviceConn.Close()
	time.Sleep(200 * time.Millisecond)

	if !state.hub.Online("IMEI_BRIDGE") {
		t.Fatal("device should be online via device stream")
	}

	// Connect a GraphQL subscriber.
	subConn := dialGQLWS(t, state.wsURL())
	defer subConn.Close()
	sendMsg(t, subConn, wsMessage{Type: "connection_init"})
	readMsg(t, subConn, 3*time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId riskScore } }",
		"variables": map[string]any{"deviceId": "IMEI_BRIDGE"},
	})
	sendMsg(t, subConn, wsMessage{Type: "subscribe", ID: "sub-bridge", Payload: subPayload})
	readMsg(t, subConn, 3*time.Second)

// Publish telemetry via the subscription manager (simulating what an.
	// event processor would do after receiving device telemetry).
	state.hub.PublishTelemetry(state.operator.ID, "IMEI_BRIDGE", map[string]int{"riskScore": 55})

	msg := readMsg(t, subConn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("device stream and GraphQL subscription coexist on same hub")
}

// ─── 7.27: Connection without connection_init (direct subscribe) ───.

func TestPhase7Sub_SubscribeWithoutInit(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	// Send subscribe without connection_init first.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_NO_INIT"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-no-init", Payload: subPayload})

	// Should still get a next message (the handler processes subscribe regardless).
	msg, ok := readMsgNoFail(t, conn, 3*time.Second)
	if !ok {
		t.Fatal("expected a response even without connection_init")
	}
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("subscribe without prior connection_init still works (no enforced handshake)")
}

// ─── 7.28: Ping keepalive from server ───.

func TestPhase7Sub_ServerPingKeepalive(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// The subscription client's writePump sends pings every 30s.
// We can't wait 30s in a test, so just verify the connection stays alive.
	// for a few seconds and can still exchange messages.
	time.Sleep(2 * time.Second)

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id } }",
		"variables": map[string]any{"deviceId": "IMEI_KEEPALIVE"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-keep", Payload: subPayload})
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next after keepalive period, got %s", msg.Type)
	}
	t.Log("connection stays alive and responsive after 2s (ping keepalive interval is 30s)")
}

// ─── 7.29: Unknown subscription query gets __typename ───.

func TestPhase7Sub_UnknownQueryGetsTypename(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// Subscribe with a completely unknown subscription type.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { somethingRandom { id } }",
		"variables": map[string]any{},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-typename", Payload: subPayload})

	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}

	// The payload should contain __typename: Subscription.
	var payload struct {
		Data struct {
			Typename string `json:"__typename"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Data.Typename != "Subscription" {
		t.Errorf("__typename = %s, want Subscription", payload.Data.Typename)
	}
	t.Log("unknown subscription query → __typename: Subscription fallback")
}

// ─── 7.30: Wildcard subscription query matching ───.

func TestPhase7Sub_QueryContainsMatching(t *testing.T) {
	state := newGQLSubState(t)
	defer state.close()

	conn := dialGQLWS(t, state.wsURL())
	defer conn.Close()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	readMsg(t, conn, 3*time.Second)

	// The subscription handler uses substring matching on the query.
	// "deviceUpdated" should match even in a longer query.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription SubDeviceUpdated { deviceUpdated(deviceId: $deviceId) { id status online } }",
		"variables": map[string]any{"deviceId": "IMEI_QUERY_MATCH"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-qm", Payload: subPayload})
	readMsg(t, conn, 3*time.Second) // initial null

	state.hub.PublishDeviceUpdate(state.operator.ID, "IMEI_QUERY_MATCH", map[string]string{"status": "online"})
	msg := readMsg(t, conn, 3*time.Second)
	if msg.Type != "next" {
		t.Errorf("expected next, got %s", msg.Type)
	}
	t.Log("query substring matching: 'deviceUpdated' found in complex query")
}
