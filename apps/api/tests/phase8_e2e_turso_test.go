package tests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"

	gorillaws "github.com/gorilla/websocket"
)

// ─── Phase 8: Full end-to-end live server with real Turso DB ───.
//
// This test spins up the *real* Vyzorix API server via wire.Injector with the.
// actual Turso database (TURSO_DB_URL + TURSO_VYZOR_SCOPE_DB_TOKEN). It then:
//
//   1. Inserts a verified operator directly into the DB (skipping register/email).
//   2. Logs in via POST /v1/auth/login → obtains a session cookie.
//   3. Creates an organization via POST /v1/organizations.
//   4. Opens a GraphQL subscription WebSocket to /:org/graphql/ws with the cookie.
//   5. Subscribes to telemetryReceived, publishes telemetry via the hub, and.
//      verifies the subscriber receives the published event live.
//
// This is the most realistic test: it exercises the full middleware chain.
// (cookie auth → session validation → org membership → subscription handler)
// against a real remote database.

const phase8TestEmail = "phase8-e2e@vyzorix.dev"
const phase8TestPassword = "Phase8#E2E2026!"
const phase8TestName = "Phase8 E2E Operator"

// phase8State holds the live server and shared state.
type phase8State struct {
	server    *httptest.Server
	apiServer *api.Server
	db        *sql.DB
	storage   *storage.SQLite
	hub       interface {
		PublishTelemetry(string, string, interface{})
	}
	operatorID    string
	orgID         string
	sessionCookie *http.Cookie
	csrfToken     string
	signingKey    string // per-session HMAC key returned by login
	client        *http.Client // cookie jar tracks _csrf + vyz_session across requests
}

// newPhase8State builds a full API server wired to the real Turso DB,
// mirroring cmd/api/api_main.go exactly.
func newPhase8State(t *testing.T) *phase8State {
	t.Helper()

// Set required env vars for config.Load(). These must be set BEFORE.
	// config.Load() is called so the resolved backend is Turso.
	t.Setenv("TURSO_DB_URL", os.Getenv("TURSO_DB_URL"))
	t.Setenv("TURSO_AUTH_TOKEN", os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN"))
	t.Setenv("DATABASE_BACKEND", "turso")
	t.Setenv("DATABASE_URL", os.Getenv("TURSO_DB_URL")) // Required by validateRequiredSecrets.
	t.Setenv("JWT_SECRET", "phase8-test-jwt-secret-32chars-min!!")
	t.Setenv("SESSION_SECRET", "phase8-test-session-secret-32chars!")
	t.Setenv("API_KEY_test", "vxyz-phase8-test-key-1234567890")
	t.Setenv("ENABLE_GRAPHQL", "true")
	t.Setenv("SSR_ENABLED", "false")
	t.Setenv("NODE_ENV", "development")
	t.Setenv("ENFORCE_HMAC", "false")
	t.Setenv("CSRF_ENABLED", "false")
	t.Setenv("ALLOWED_ORIGINS", "*")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load failed: %v", err)
	}

	// Build the full server via wire dependency injection (real Turso DB).
	wiredServer, err := wire.Injector(cfg)
	if err != nil {
		t.Fatalf("wire.Injector failed: %v", err)
	}

	deps := wiredServer.Dependencies
	result := wiredServer.Result

	apiServer := api.NewServerWithDeps(&api.ServerConfigWithDeps{
		Config:          cfg,
		Log:             deps.Log,
		DB:              deps.DB,
		Engine:          result.Engine,
		Middleware:      result.MiddlewareSet,
		HandlerSet:      result.HandlerSet,
		SessionManager:  deps.SessionManager,
		Hub:             deps.Hub,
		AuditLogger:     deps.AuditLogger,
		UpdatesService:  deps.UpdatesService,
		APIKeyService:   deps.APIKeyService,
		DeviceRepo:      deps.DeviceRepo,
		IdempotencyRepo: deps.IdempotencyRepo,
	})

// Register GraphQL, mirroring api_main.go's wiring of GraphQL-specific.
	// services that are NOT part of the main wire graph.
	db := deps.DB.DB()
	deviceRepo := deps.DeviceService.DeviceRepo()
	commandRepo := deps.CommandService.CommandRepo()

	historyService := command.NewHistoryService(commandRepo, deviceRepo)
	logsRepo := storage.NewLogsRepository(db)
	metricsRepo := storage.NewMetricsRepository(db)
	logsSvc := logs.NewService(logsRepo, deps.Log)
	metricsSvc := appmetrics.NewService(metricsRepo, nil, nil)
	dashboardSvc := dashboard.NewService(deviceRepo, commandRepo, logsRepo)
	diagnosticsRepo := storage.NewDiagnosticsRepository(db)
	diagnosticsSvc := diagnosticsapp.NewService(diagnosticsRepo, deviceRepo, deps.Hub, cfg.DiagnosticsConfig)

	settingsService := auth.NewClientSettingsService(deps.OperatorRepo)
	notificationSvc := appoperator.NewNotificationService(deps.OperatorRepo)
	webhookClient := infrawebhook.NewClient(10 * time.Second)

	if regErr := apiServer.RegisterGraphQL(
		deps.DeviceService,
		deps.DeviceSettingsService,
		deps.CommandService,
		historyService,
		dashboardSvc,
		logsSvc,
		metricsSvc,
		deps.TelemetryRepo,
		logsRepo,
		metricsRepo,
		deps.Hub,
		deps.UpdatesService,
		diagnosticsSvc,
		deps.OperatorRepo,
		settingsService,
		notificationSvc,
		webhookClient,
		result.HandlerSet.OrgService,
		result.HandlerSet.OrgSettingsService,
		result.HandlerSet.MemberService,
		result.HandlerSet.InvitationService,
	); regErr != nil {
		t.Fatalf("RegisterGraphQL failed: %v", regErr)
	}

	ts := httptest.NewServer(apiServer.Routes())

// Build an HTTP client with a cookie jar so _csrf and vyz_session cookies.
	// are automatically retained across requests (required for double-submit CSRF).
	jar, err := newCookieJar()
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	state := &phase8State{
		server:    ts,
		apiServer: apiServer,
		db:        deps.DB.DB(),
		storage:   deps.DB,
		hub:       deps.Hub,
		client:    client,
	}

	return state
}

func (s *phase8State) close() {
	if s.server != nil {
		s.server.Close()
	}
}

// newCookieJar creates a net/http cookie jar for the test client.
func newCookieJar() (http.CookieJar, error) {
	return cookiejar.New(nil)
}

// fetchCSRFToken calls GET /v1/auth/csrf-token, stores the token, and relies on.
// the cookie jar to retain the _csrf cookie for the double-submit pattern.
func (s *phase8State) fetchCSRFToken(t *testing.T) {
	t.Helper()

	resp, err := s.client.Get(s.server.URL + "/v1/auth/csrf-token")
	if err != nil {
		t.Fatalf("csrf-token request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("csrf-token failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("csrf-token parse: %v", err)
	}
	if result.CSRFToken == "" {
		t.Fatalf("csrf-token response missing token: %s", string(body))
	}
	s.csrfToken = result.CSRFToken
}

// insertVerifiedOperator inserts an operator with email_verified=1 directly into.
// the Turso DB, skipping the register/email-verification flow. It also cleans up.
// any prior orgs/members/sessions for this email from earlier test runs so the.
// 2-active-org limit isn't hit.
func (s *phase8State) insertVerifiedOperator(t *testing.T) string {
	t.Helper()

	hasher := password.NewArgon2idHasher()
	hash, err := hasher.Hash(phase8TestPassword)
	if err != nil {
		t.Fatalf("password hash failed: %v", err)
	}

	// Clean up any prior operator with this email (and its orgs/members/sessions).
	s.execSQL(t, fmt.Sprintf(
		`DELETE FROM organization_members WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase8TestEmail))
	s.execSQL(t, fmt.Sprintf(
		`DELETE FROM organizations WHERE created_by IN (SELECT id FROM operators WHERE email = '%s')`, phase8TestEmail))
	s.execSQL(t, fmt.Sprintf(
		`DELETE FROM auth_sessions WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase8TestEmail))
	s.execSQL(t, fmt.Sprintf(
		`DELETE FROM operators WHERE email = '%s'`, phase8TestEmail))

	operatorID := "op-phase8-" + newUUID()
	nowMs := time.Now().UnixMilli()

	// Insert the verified operator.
	s.execSQL(t, fmt.Sprintf(
		`INSERT INTO operators (id, email, name, password_hash, role, mfa_enabled, email_verified, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', 'operator', 0, 1, %d, %d)`,
		operatorID, phase8TestEmail, phase8TestName, hash, nowMs, nowMs))

	s.operatorID = operatorID
	return operatorID
}

// execSQL runs a raw SQL statement against Turso via the HTTP API.
func (s *phase8State) execSQL(t *testing.T, statement string) {
	t.Helper()
	tursoExec(t, []map[string]any{{"q": statement}})
}

// login performs POST /v1/auth/login and extracts the session cookie.
// It first fetches a CSRF token (double-submit pattern) then sends it.
func (s *phase8State) login(t *testing.T) *http.Cookie {
	t.Helper()

	// Fetch CSRF token first (sets _csrf cookie in jar, returns token in body).
	s.fetchCSRFToken(t)

	body, _ := json.Marshal(map[string]string{
		"email":    phase8TestEmail,
		"password": phase8TestPassword,
	})

	req, err := http.NewRequest("POST", s.server.URL+"/v1/auth/login", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", s.csrfToken)

	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var loginResp map[string]any
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	// Capture the per-session signing key for signing subsequent tenant requests.
	if key, _ := loginResp["signing_key"].(string); key != "" {
		s.signingKey = key
	}

// The login endpoint sets a session cookie. The cookie jar retains it; also.
	// extract it explicitly for the WebSocket dial (which needs a raw header).
	var sessionCookie *http.Cookie
	u, _ := url.Parse(s.server.URL)
	for _, c := range s.client.Jar.Cookies(u) {
		if c.Name == "vyz_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("login did not set vyz_session cookie; cookies: %+v", s.client.Jar.Cookies(u))
	}

	s.sessionCookie = sessionCookie
	t.Logf("login successful: operator_id=%v, cookie set", loginResp["operator_id"])
	return sessionCookie
}

// createOrganization performs POST /v1/organizations and returns the org ID.
func (s *phase8State) createOrganization(t *testing.T) string {
	t.Helper()

	// Refresh CSRF token for the authenticated request (session-based validation).
	s.fetchCSRFToken(t)

	body, _ := json.Marshal(map[string]any{
		"name":        "Phase8 Test Org " + newUUID()[:8],
		"description": "E2E test organization for Phase 8",
		"role":        "super_admin",
	})

	req, err := http.NewRequest("POST", s.server.URL+"/v1/organizations", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("create org request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", s.csrfToken)
	if s.signingKey != "" {
		signRequest(req, s.signingKey, body)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatalf("create org failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var orgResp map[string]any
	if err := json.Unmarshal(respBody, &orgResp); err != nil {
		t.Fatalf("failed to parse org response: %v", err)
	}

	orgID, ok := orgResp["id"].(string)
	if !ok || orgID == "" {
		t.Fatalf("create org response missing id: %s", string(respBody))
	}

	s.orgID = orgID
	t.Logf("organization created: id=%s", orgID)
	return orgID
}

// dialSubWS opens a WebSocket to /:org/graphql/ws with the session cookie.
func (s *phase8State) dialSubWS(t *testing.T) *gorillaws.Conn {
	t.Helper()

	wsURL := strings.Replace(s.server.URL, "http://", "ws://", 1)
	wsURL = wsURL + "/" + s.orgID + "/graphql/ws"

	// Parse the server URL to get the host for the cookie.
	parsed, err := url.Parse(s.server.URL)
	if err != nil {
		t.Fatalf("url parse: %v", err)
	}

	dialer := gorillaws.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	header := http.Header{}
	header.Set("Cookie", fmt.Sprintf("vyz_session=%s", s.sessionCookie.Value))
	header.Set("Origin", parsed.Host)

	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("WS dial failed: %v, status: %d, location: %q, body: %s",
				err, resp.StatusCode, resp.Header.Get("Location"), string(body))
		}
		t.Fatalf("WS dial failed: %v", err)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	return conn
}

// TestPhase8_E2E_TursoLiveServer is the full end-to-end test:
// real Turso DB → login → create org → WS subscribe → publish → receive.
func TestPhase8_E2E_TursoLiveServer(t *testing.T) {
	// Skip if Turso credentials aren't available.
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	// 1. Insert a verified operator directly into the DB.
	operatorID := state.insertVerifiedOperator(t)
	t.Logf("step 1: verified operator inserted: %s", operatorID)

	// 2. Login to get a session cookie.
	cookie := state.login(t)
	t.Logf("step 2: login successful, cookie: %s...%s", cookie.Value[:8], cookie.Value[len(cookie.Value)-4:])

	// 3. Create an organization.
	orgID := state.createOrganization(t)
	t.Logf("step 3: organization created: %s", orgID)

	// 4. Open a GraphQL subscription WebSocket.
	conn := state.dialSubWS(t)
	defer conn.Close()
	t.Log("step 4: WebSocket connected")

	// Use the msgReader to avoid gorilla read-deadline corruption.
	reader := newMsgReader(conn)
	defer reader.stop()

	// 5. Send connection_init → expect connection_ack.
	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	ack, ok := reader.recv(5 * time.Second)
	if !ok || ack.Type != "connection_ack" {
		t.Fatalf("step 5: expected connection_ack, got ok=%v type=%s", ok, ack.Type)
	}
	t.Log("step 5: connection acknowledged")

	// 6. Subscribe to telemetryReceived for a specific device.
	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId riskScore } }",
		"variables": map[string]any{"deviceId": "IMEI_PHASE8_LIVE"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-phase8", Payload: subPayload})

	// Expect initial null "next" message.
	initNext, ok := reader.recv(5 * time.Second)
	if !ok || initNext.Type != "next" {
		t.Fatalf("step 6: expected initial next, got ok=%v type=%s", ok, initNext.Type)
	}
	t.Log("step 6: subscribed to telemetryReceived, initial next received")

	// 7. Publish telemetry via the hub → subscriber should receive it.
//    The hub is the same instance the server uses, so the publish goes.
	//    through the real subscription routing.
	state.hub.PublishTelemetry(operatorID, "IMEI_PHASE8_LIVE", map[string]any{
		"deviceId":  "IMEI_PHASE8_LIVE",
		"riskScore": 42,
		"timestamp": time.Now().Unix(),
	})

	// 8. Receive the published telemetry event.
	pubNext, ok := reader.recv(10 * time.Second)
	if !ok {
		t.Fatal("step 8: did not receive published telemetry event")
	}
	if pubNext.Type != "next" {
		t.Fatalf("step 8: expected next, got %s", pubNext.Type)
	}

	// Verify the payload contains the telemetry data.
	var payload struct {
		Data struct {
			TelemetryReceived struct {
				DeviceID  string `json:"deviceId"`
				RiskScore int    `json:"riskScore"`
			} `json:"telemetryReceived"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pubNext.Payload, &payload); err != nil {
		// Payload might be a raw JSON string; try generic parse.
		var raw map[string]any
		if err2 := json.Unmarshal(pubNext.Payload, &raw); err2 != nil {
			t.Fatalf("step 8: failed to parse payload: %v (raw: %s)", err, string(pubNext.Payload))
		}
		t.Logf("step 8: received telemetry event (raw): %s", string(pubNext.Payload))
	} else {
		if payload.Data.TelemetryReceived.DeviceID != "IMEI_PHASE8_LIVE" {
			t.Errorf("step 8: expected deviceId IMEI_PHASE8_LIVE, got %s", payload.Data.TelemetryReceived.DeviceID)
		}
		if payload.Data.TelemetryReceived.RiskScore != 42 {
			t.Errorf("step 8: expected riskScore 42, got %d", payload.Data.TelemetryReceived.RiskScore)
		}
		t.Logf("step 8: received telemetry: deviceId=%s riskScore=%d",
			payload.Data.TelemetryReceived.DeviceID, payload.Data.TelemetryReceived.RiskScore)
	}

	// 9. Send complete to unsubscribe.
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-phase8"})
	time.Sleep(300 * time.Millisecond)

	// 10. Publish another event → should NOT receive it (unsubscribed).
	state.hub.PublishTelemetry(operatorID, "IMEI_PHASE8_LIVE", map[string]any{"riskScore": 99})
	if _, ok := reader.recv(2 * time.Second); ok {
		t.Error("step 10: should not receive event after complete")
	}

	t.Log("=== Phase 8 E2E COMPLETE: Turso DB → login → create org → WS subscribe → publish → receive → complete ===")
}

// TestPhase8_E2E_DeviceUpdateSubscription tests the deviceUpdated subscription.
// type through the full live server.
func TestPhase8_E2E_DeviceUpdateSubscription(t *testing.T) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	operatorID := state.insertVerifiedOperator(t)
	state.login(t)
	state.createOrganization(t)

	conn := state.dialSubWS(t)
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected connection_ack")
	}

	subPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { deviceUpdated { id status } }",
		"variables": map[string]any{"deviceId": "IMEI_PHASE8_DEVICE"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-dev", Payload: subPayload})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next")
	}

	// Publish a device update.
	state.hub.(interface {
		PublishDeviceUpdate(string, string, interface{})
	}).PublishDeviceUpdate(operatorID, "IMEI_PHASE8_DEVICE", map[string]any{
		"id":     "IMEI_PHASE8_DEVICE",
		"status": "online",
	})

	msg, ok := reader.recv(10 * time.Second)
	if !ok {
		t.Fatal("did not receive device update event")
	}
	if msg.Type != "next" {
		t.Fatalf("expected next, got %s", msg.Type)
	}
	t.Logf("received device update event: %s", string(msg.Payload))

	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-dev"})
	t.Log("device update subscription E2E verified")
}

// TestPhase8_E2E_MultipleSubscriptionTypes tests subscribing to multiple.
// subscription types simultaneously on the live server.
func TestPhase8_E2E_MultipleSubscriptionTypes(t *testing.T) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	operatorID := state.insertVerifiedOperator(t)
	state.login(t)
	state.createOrganization(t)

	conn := state.dialSubWS(t)
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected connection_ack")
	}

	// Subscribe to telemetry.
	telemPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId } }",
		"variables": map[string]any{"deviceId": "IMEI_MULTI_TELEM"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-telem", Payload: telemPayload})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next for telemetry")
	}

	// Subscribe to command status.
	cmdPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { commandStatusChanged { dispatchId status } }",
		"variables": map[string]any{"dispatchId": "DISPATCH_MULTI_001"},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-cmd", Payload: cmdPayload})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next for command status")
	}

	// Publish telemetry → should arrive on sub-telem.
	state.hub.PublishTelemetry(operatorID, "IMEI_MULTI_TELEM", map[string]any{
		"deviceId": "IMEI_MULTI_TELEM",
	})
	msg, ok := reader.recv(10 * time.Second)
	if !ok || msg.Type != "next" {
		t.Fatalf("expected telemetry next, got ok=%v type=%s", ok, msg.Type)
	}
	t.Logf("telemetry event received: %s", string(msg.Payload))

	// Publish command status → should arrive on sub-cmd.
	state.hub.(interface {
		PublishCommandStatus(string, string, interface{})
	}).PublishCommandStatus(operatorID, "DISPATCH_MULTI_001", map[string]any{
		"dispatchId": "DISPATCH_MULTI_001",
		"status":     "delivered",
	})
	msg, ok = reader.recv(10 * time.Second)
	if !ok || msg.Type != "next" {
		t.Fatalf("expected command status next, got ok=%v type=%s", ok, msg.Type)
	}
	t.Logf("command status event received: %s", string(msg.Payload))

	// Complete both subscriptions.
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-telem"})
	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-cmd"})
	t.Log("multiple subscription types E2E verified")
}

// TestPhase8_E2E_OrgEventSubscription tests the organizationEvent subscription.
// through the full live server.
func TestPhase8_E2E_OrgEventSubscription(t *testing.T) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	state.insertVerifiedOperator(t)
	state.login(t)
	orgID := state.createOrganization(t)

	conn := state.dialSubWS(t)
	defer conn.Close()

	reader := newMsgReader(conn)
	defer reader.stop()

	sendMsg(t, conn, wsMessage{Type: "connection_init"})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected connection_ack")
	}

	orgPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { organizationEvent { type message } }",
		"variables": map[string]any{"orgId": orgID},
	})
	sendMsg(t, conn, wsMessage{Type: "subscribe", ID: "sub-org", Payload: orgPayload})
	if _, ok := reader.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next")
	}

	// Publish an org event.
	state.hub.(interface {
		PublishOrganizationEvent(string, interface{})
	}).PublishOrganizationEvent(orgID, map[string]any{
		"type":    "member_joined",
		"message": "A new member joined the organization",
	})

	msg, ok := reader.recv(10 * time.Second)
	if !ok || msg.Type != "next" {
		t.Fatalf("expected org event next, got ok=%v type=%s", ok, msg.Type)
	}
	t.Logf("org event received: %s", string(msg.Payload))

	sendMsg(t, conn, wsMessage{Type: "complete", ID: "sub-org"})
	t.Log("organization event subscription E2E verified")
}

// TestPhase8_E2E_ReconnectAfterDisconnect verifies that after disconnecting.
// and reconnecting, the subscription still works.
func TestPhase8_E2E_ReconnectAfterDisconnect(t *testing.T) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	operatorID := state.insertVerifiedOperator(t)
	state.login(t)
	state.createOrganization(t)

	// First connection.
	conn1 := state.dialSubWS(t)
	reader1 := newMsgReader(conn1)
	defer reader1.stop()

	sendMsg(t, conn1, wsMessage{Type: "connection_init"})
	if _, ok := reader1.recv(5 * time.Second); !ok {
		t.Fatal("expected connection_ack on first connection")
	}

	telemPayload, _ := json.Marshal(map[string]any{
		"query":     "subscription { telemetryReceived { deviceId } }",
		"variables": map[string]any{"deviceId": "IMEI_RECONNECT"},
	})
	sendMsg(t, conn1, wsMessage{Type: "subscribe", ID: "sub-recon", Payload: telemPayload})
	if _, ok := reader1.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next on first connection")
	}

	// Disconnect.
	conn1.Close()
	time.Sleep(500 * time.Millisecond)

	// Reconnect.
	conn2 := state.dialSubWS(t)
	defer conn2.Close()
	reader2 := newMsgReader(conn2)
	defer reader2.stop()

	sendMsg(t, conn2, wsMessage{Type: "connection_init"})
	if _, ok := reader2.recv(5 * time.Second); !ok {
		t.Fatal("expected connection_ack on reconnect")
	}

	sendMsg(t, conn2, wsMessage{Type: "subscribe", ID: "sub-recon2", Payload: telemPayload})
	if _, ok := reader2.recv(5 * time.Second); !ok {
		t.Fatal("expected initial next on reconnect")
	}

	// Publish → should receive on the new connection.
	state.hub.PublishTelemetry(operatorID, "IMEI_RECONNECT", map[string]any{
		"deviceId": "IMEI_RECONNECT",
	})
	msg, ok := reader2.recv(10 * time.Second)
	if !ok || msg.Type != "next" {
		t.Fatalf("expected telemetry next after reconnect, got ok=%v type=%s", ok, msg.Type)
	}
	t.Logf("telemetry received after reconnect: %s", string(msg.Payload))

	sendMsg(t, conn2, wsMessage{Type: "complete", ID: "sub-recon2"})
	t.Log("reconnect E2E verified")
}

// TestPhase8_E2E_UnauthenticatedWSRejected verifies that a WS connection.
// without a session cookie is rejected by the middleware.
func TestPhase8_E2E_UnauthenticatedWSRejected(t *testing.T) {
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set; skipping live Turso E2E test")
	}

	state := newPhase8State(t)
	defer state.close()

	state.insertVerifiedOperator(t)
	state.login(t)
	state.createOrganization(t)

	// Attempt to dial WITHOUT the session cookie.
	wsURL := strings.Replace(state.server.URL, "http://", "ws://", 1)
	wsURL = wsURL + "/" + state.orgID + "/graphql/ws"

	dialer := gorillaws.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, resp, err := dialer.Dial(wsURL, http.Header{})
	if err == nil {
		conn.Close()
		t.Fatal("expected WS dial to fail without session cookie, but it succeeded")
	}
	if resp != nil {
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
	t.Log("unauthenticated WS correctly rejected")
}
