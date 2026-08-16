package tests

import (
	"bytes"
	"context"
	crypto_rand "crypto/rand"
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"

	_ "github.com/mattn/go-sqlite3"
)

// Shared constants for integration tests.
const (
	baseURL  = "http://localhost:3000"
	mockURL  = "http://127.0.0.1:1080"
	stateDir = "/tmp/vyzorix-go-test-state.json"
	dataDir  = "/tmp/vyzorix-data"
)

// testPassword is the shared password for all test operators.
const testPassword = "TestPass#2026!"

// signingKeys maps a session HTTP client to the per-session HMAC signing key
// returned by the login response. doJSON / doRaw / graphql / graphqlBatch look
// up the key from this map and sign requests automatically, matching the
// server's SessionSignatureMiddleware. API-key-authenticated requests are
// signed with a secret derived from the full API key value (see
// deriveAPIKeySigningSecret in the keys service).
var signingKeys sync.Map // map[*http.Client]string

// TestState holds cross-phase shared state, persisted to stateDir.
type TestState struct {
	Email      string            `json:"email"`
	Password   string            `json:"password"`
	Name       string            `json:"name"`
	OrgID      string            `json:"org_id"`
	OperatorID string            `json:"operator_id"`
	APIKeys    map[string]string `json:"api_keys"`
	CSRF       string            `json:"csrf"`
	Cookies    map[string]string `json:"cookies"`
	PushID     string            `json:"push_id"`
	DeviceIMEI string            `json:"device_imei"`
	DeviceID   string            `json:"device_id"`
	Version    string            `json:"version"`
}

// saveState persists TestState to the shared file.
func saveState(s *TestState) {
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(stateDir, data, 0644)
}

// loadState reads TestState from the shared file.
func loadState() *TestState {
	data, err := os.ReadFile(stateDir)
	if err != nil {
		return &TestState{}
	}
	var s TestState
	_ = json.Unmarshal(data, &s)
	return &s
}

// httpClient is the shared HTTP client (no cookie jar — sessions manage their own).
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// newSession creates an HTTP client with a cookie jar for session-based flows.
func newSession() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout: 15 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

// getCSRF fetches a CSRF token and returns it. The _csrf cookie is stored in
// the client's jar automatically. Retries on 429 rate-limit.
func getCSRF(t *testing.T, c *http.Client) string {
	t.Helper()
	for attempt := 0; attempt < 5; attempt++ {
		resp, err := c.Get(baseURL + "/v1/auth/csrf-token")
		if err != nil {
			t.Fatalf("csrf-token GET: %v", err)
		}
		if resp.StatusCode == 429 {
			_ = resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("csrf-token GET status %d: %s", resp.StatusCode, string(body))
		}
		var data map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatalf("csrf-token decode: %v", err)
		}
		token, _ := data["csrf_token"].(string)
		return token
	}
	t.Fatal("csrf-token GET: rate limited after 5 retries")
	return ""
}

// signRequest sets the X-Vyzorix-Timestamp, X-Vyzorix-Nonce, and
// X-Vyzorix-Signature headers on req using the same HMAC-SHA512 scheme as the
// server's Verifier.Verify: canonical = METHOD\nPATH\nNONCE\nTIMESTAMP_MS\nBODY.
func signRequest(req *http.Request, signingKey string, body []byte) {
	nonce := newUUID()
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	mac := hmac.New(sha512.New, []byte(signingKey))
	_, _ = mac.Write([]byte(req.Method + "\n" + req.URL.RequestURI() + "\n" + nonce + "\n" + ts + "\n"))
	_, _ = mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Vyzorix-Timestamp", ts)
	req.Header.Set("X-Vyzorix-Nonce", nonce)
	req.Header.Set("X-Vyzorix-Signature", sig)
}

// resolveSigningKey determines the HMAC signing key for an outgoing request.
// For API-key-authenticated requests (X-API-Key header present), the secret is
// derived from the full API key value via SHA-512 hex — matching the server's
// deriveAPIKeySigningSecret. For session-authenticated requests, the key is
// looked up from the signingKeys registry (populated by login). Returns "" if
// no signing key is available, which means the request is sent unsigned.
func resolveSigningKey(req *http.Request, c *http.Client) string {
	if apiKey := req.Header.Get("X-API-Key"); apiKey != "" {
		h := sha512.Sum512([]byte(apiKey))
		return hex.EncodeToString(h[:])
	}
	if key, ok := signingKeys.Load(c); ok {
		return key.(string)
	}
	return ""
}

// doJSON sends a request with the given method, path, headers, and JSON body.
// Returns the status code and the raw response body. Retries on 429.
func doJSON(t *testing.T, c *http.Client, method, path string, headers map[string]string, body any) (int, []byte) {
	t.Helper()
	var bodyBytes []byte
	var reqBody io.Reader
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}
	for attempt := 0; attempt < 5; attempt++ {
		req, err := http.NewRequest(method, baseURL+path, reqBody)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if key := resolveSigningKey(req, c); key != "" {
			signRequest(req, key, bodyBytes)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("request %s %s: %v", method, path, err)
		}
		if resp.StatusCode == 429 {
			_ = resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, respBody
	}
	t.Fatalf("request %s %s: rate limited after 5 retries", method, path)
	return 0, nil
}

// doRaw sends a request with the given method, path, and headers (no body).
// Returns the status code and the raw response body. Retries on 429.
func doRaw(t *testing.T, c *http.Client, method, path string, headers map[string]string) (int, []byte) {
	t.Helper()
	for attempt := 0; attempt < 5; attempt++ {
		req, err := http.NewRequest(method, baseURL+path, nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if key := resolveSigningKey(req, c); key != "" {
			signRequest(req, key, nil)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("request %s %s: %v", method, path, err)
		}
		if resp.StatusCode == 429 {
			_ = resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, respBody
	}
	t.Fatalf("request %s %s: rate limited after 5 retries", method, path)
	return 0, nil
}

// parseJSON unmarshals body into a map[string]any.
func parseJSON(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("parse JSON: %v (body: %s)", err, string(body))
	}
	return m
}

// requireServer checks that the test server is running; skips the test if not.
func requireServer(t *testing.T) {
	t.Helper()
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(baseURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("server not running at %s — skipping integration test", baseURL)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
}

// requireMockEmail checks that the mock email server is running.
func requireMockEmail(t *testing.T) {
	t.Helper()
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(mockURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("mock email server not running at %s — skipping", mockURL)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
}

// emailLog represents an entry from the mock-resend /emails/ endpoint.
type emailLog struct {
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	HTML        string   `json:"html"`
	VerifyToken string   `json:"verify_token"`
}

// fetchEmails retrieves all emails from the mock-resend server.
func fetchEmails(t *testing.T) []emailLog {
	t.Helper()
	resp, err := httpClient.Get(mockURL + "/emails/")
	if err != nil {
		t.Fatalf("fetch emails: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var emails []emailLog
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		t.Fatalf("decode emails: %v", err)
	}
	return emails
}

// fetchVerifyToken polls the mock email server for a verification token
// matching the given email address.
func fetchVerifyToken(t *testing.T, email string) string {
	t.Helper()
	for i := 0; i < 15; i++ {
		emails := fetchEmails(t)
		for j := len(emails) - 1; j >= 0; j-- {
			e := emails[j]
			for _, to := range e.To {
				if to == email {
					if e.VerifyToken != "" {
						return e.VerifyToken
					}
					re := regexp.MustCompile(`[?&]token=([^&"'\s]+)`)
					if m := re.FindStringSubmatch(e.HTML); len(m) >= 2 {
						return m[1]
					}
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

// fetchInviteEmail polls for an invitation email matching the given address.
func fetchInviteEmail(t *testing.T, email string) *emailLog {
	t.Helper()
	for i := 0; i < 10; i++ {
		emails := fetchEmails(t)
		for j := len(emails) - 1; j >= 0; j-- {
			e := emails[j]
			for _, to := range e.To {
				if to == email && strings.Contains(strings.ToLower(e.Subject), "invit") {
					return &e
				}
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

// extractInviteToken extracts the path-based invite token from email HTML.
func extractInviteToken(html string) string {
	re := regexp.MustCompile(`/invite/([a-zA-Z0-9\-_]+)/accept`)
	if m := re.FindStringSubmatch(html); len(m) >= 2 {
		return m[1]
	}
	re = regexp.MustCompile(`/invite/([a-zA-Z0-9\-_]+)`)
	if m := re.FindStringSubmatch(html); len(m) >= 2 {
		return m[1]
	}
	return ""
}

// registerAndVerify registers a new operator and verifies the email via the
// mock-resend token. Returns the authenticated session client and CSRF token.
func registerAndVerify(t *testing.T, email, name string) (*http.Client, string) {
	t.Helper()
	c, err := newSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	csrf := getCSRF(t, c)
	status, body := doJSON(t, c, "POST", "/v1/auth/register",
		map[string]string{"X-CSRF-Token": csrf},
		map[string]any{"email": email, "password": testPassword, "name": name})
	if status != 201 && status != 200 {
		t.Fatalf("register %s -> %d: %s", email, status, string(body))
	}
	time.Sleep(1500 * time.Millisecond)
	token := fetchVerifyToken(t, email)
	if token == "" {
		t.Fatalf("no verify token captured for %s", email)
	}
	status, body = doJSON(t, c, "POST", "/v1/auth/verify-email",
		map[string]string{"X-CSRF-Token": csrf},
		map[string]any{"token": token})
	if status != 200 {
		t.Fatalf("verify-email -> %d: %s", status, string(body))
	}
	return c, csrf
}

// login creates a new session, logs in with the given credentials, selects the
// org, and returns the session client + CSRF token.
func login(t *testing.T, email, orgID string) (*http.Client, string) {
	t.Helper()
	c, err := newSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	csrf := getCSRF(t, c)
	status, body := doJSON(t, c, "POST", "/v1/auth/login",
		map[string]string{"X-CSRF-Token": csrf},
		map[string]any{"email": email, "password": testPassword})
	if status != 200 {
		t.Fatalf("login %s -> %d: %s", email, status, string(body))
	}
	// Capture the per-session signing key returned by the login response so
	// subsequent requests can be signed with X-Vyzorix-* headers.
	loginData := parseJSON(t, body)
	if signingKey, _ := loginData["signing_key"].(string); signingKey != "" {
		signingKeys.Store(c, signingKey)
	}
	if orgID != "" {
		status, body = doJSON(t, c, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"organization_id": orgID})
		if status != 200 {
			t.Fatalf("org select -> %d: %s", status, string(body))
		}
	}
	return c, csrf
}

// createAPIKey creates an API key with the given scope and returns the full key.
func createAPIKey(t *testing.T, c *http.Client, csrf, orgID, name, scope string) string {
	t.Helper()
	status, body := doJSON(t, c, "POST", "/v1/auth/api-keys",
		map[string]string{"X-CSRF-Token": csrf, "X-Organization-ID": orgID},
		map[string]any{"name": name, "scope": scope})
	if status != 201 {
		t.Fatalf("create %s key -> %d: %s", scope, status, string(body))
	}
	data := parseJSON(t, body)
	key, _ := data["api_key"].(string)
	return key
}

// graphql sends a GraphQL query to /:org/graphql using the session client.
func graphql(t *testing.T, c *http.Client, orgID, query string) (int, map[string]any) {
	t.Helper()
	payload := map[string]any{"query": query}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL+"/"+orgID+"/graphql", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new graphql request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := resolveSigningKey(req, c); key != "" {
		signRequest(req, key, data)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("graphql request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(respBody, &result)
	return resp.StatusCode, result
}

// graphqlRaw sends a GraphQL query with custom headers (no session).
func graphqlRaw(t *testing.T, orgID, query string, headers map[string]string) (int, map[string]any, string) {
	t.Helper()
	payload := map[string]any{"query": query}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL+"/"+orgID+"/graphql", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new graphql request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("graphql request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(respBody, &result)
	return resp.StatusCode, result, string(respBody)
}

// graphqlBatch sends a batch of GraphQL queries.
func graphqlBatch(t *testing.T, c *http.Client, orgID string, queries []map[string]any) (int, []map[string]any) {
	t.Helper()
	data, _ := json.Marshal(queries)
	req, err := http.NewRequest("POST", baseURL+"/"+orgID+"/graphql/batch", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new batch request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := resolveSigningKey(req, c); key != "" {
		signRequest(req, key, data)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("batch request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	var result []map[string]any
	_ = json.Unmarshal(respBody, &result)
	return resp.StatusCode, result
}

// hasGraphQLErrors checks if a GraphQL response contains errors.
func hasGraphQLErrors(data map[string]any) bool {
	errs, ok := data["errors"].([]any)
	return ok && len(errs) > 0
}

// tursoExec executes SQL statements against Turso via the HTTP API.
func tursoExec(t *testing.T, statements []map[string]any) {
	t.Helper()
	tursoURL := strings.Replace(os.Getenv("TURSO_DB_URL"), "libsql://", "https://", 1)
	tursoToken := os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN")
	payload, _ := json.Marshal(map[string]any{"statements": statements})
	req, err := http.NewRequest("POST", tursoURL, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("turso request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tursoToken)
	req.Header.Set("Content-Type", "application/json")
	c := &http.Client{Timeout: 20 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("turso exec: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		// Check if it's a UNIQUE constraint (idempotent insert).
		bodyStr := string(respBody)
		if strings.Contains(bodyStr, "UNIQUE") || strings.Contains(strings.ToLower(bodyStr), "constraint") {
			return
		}
		t.Fatalf("turso exec status %d: %s", resp.StatusCode, bodyStr)
	}
}

// newUUID generates a UUID v4 string.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = crypto_rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// timestamp returns the current Unix timestamp.
func timestamp() int64 {
	return time.Now().Unix()
}

// timestampMs returns the current Unix timestamp in milliseconds.
func timestampMs() int64 {
	return time.Now().UnixMilli()
}

// discardWriter is an io.Writer that discards all output (for test loggers).
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// noopDeviceRepoForTests is a minimal device.Repository implementation for WS tests.
type noopDeviceRepoForTests struct {
	online map[string]bool
	mu     sync.Mutex
}

func newNoopDeviceRepoForTests() *noopDeviceRepoForTests {
	return &noopDeviceRepoForTests{online: make(map[string]bool)}
}

func (r *noopDeviceRepoForTests) SetOnline(_ context.Context, id string, online bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.online[id] = online
	return nil
}
func (r *noopDeviceRepoForTests) FindByID(context.Context, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByDeviceID(context.Context, device.DeviceID) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIMEI(context.Context, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByFirebaseInstallID(context.Context, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIDAndOperator(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIDAndOperatorID(context.Context, device.DeviceID, device.OperatorID) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIMEIAndOperator(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIMEIAndOrganization(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) FindByIDAndOrganization(context.Context, string, string) (*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) Create(context.Context, *device.Device) error            { return nil }
func (r *noopDeviceRepoForTests) Update(context.Context, *device.Device) error            { return nil }
func (r *noopDeviceRepoForTests) Delete(context.Context, string) error                    { return nil }
func (r *noopDeviceRepoForTests) DeleteByDeviceID(context.Context, device.DeviceID) error { return nil }
func (r *noopDeviceRepoForTests) UpdateFCMToken(context.Context, string, string) error    { return nil }
func (r *noopDeviceRepoForTests) SetOnlineByDeviceID(context.Context, device.DeviceID, bool) error {
	return nil
}
func (r *noopDeviceRepoForTests) UpdateLastSeen(context.Context, string) error        { return nil }
func (r *noopDeviceRepoForTests) Touch(context.Context, string) error                 { return nil }
func (r *noopDeviceRepoForTests) SetSecretHash(context.Context, string, string) error { return nil }
func (r *noopDeviceRepoForTests) GetSecretHash(context.Context, string) (string, error) {
	return "", nil
}
func (r *noopDeviceRepoForTests) HashAllSecrets(context.Context) (int, error) { return 0, nil }
func (r *noopDeviceRepoForTests) List(context.Context, string, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepoForTests) ListByOperator(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) ListByOperatorID(context.Context, device.OperatorID) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) ListByOrganization(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) ListByOrganizationPaginated(context.Context, string, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepoForTests) Count(context.Context, string) (int, error) { return 0, nil }
func (r *noopDeviceRepoForTests) CountByOperator(context.Context, string) (int, error) {
	return 0, nil
}
func (r *noopDeviceRepoForTests) CountByOrganization(context.Context, string) (int, error) {
	return 0, nil
}
func (r *noopDeviceRepoForTests) SoftDelete(context.Context, string, int64, int64) error { return nil }
func (r *noopDeviceRepoForTests) SoftDeleteByIMEI(context.Context, string, int64, int64) error {
	return nil
}
func (r *noopDeviceRepoForTests) ListActive(context.Context, int, int) ([]*device.Device, int, error) {
	return nil, 0, nil
}
func (r *noopDeviceRepoForTests) ListActiveByOperator(context.Context, string) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) ListPending(context.Context) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) ListPendingByOperator(context.Context, device.OperatorID) ([]*device.Device, error) {
	return nil, nil
}
func (r *noopDeviceRepoForTests) DeleteScheduled(context.Context) (int, error) { return 0, nil }
func (r *noopDeviceRepoForTests) SoftDeleteByOrganization(context.Context, string, int64, int64) (int, error) {
	return 0, nil
}

// Compile-time check.
var _ device.Repository = (*noopDeviceRepoForTests)(nil)

// newTestAuditLogger creates an audit.Logger backed by an in-memory SQLite DB
// so audit logging doesn't panic during WS integration tests.
func newTestAuditLogger(log *slog.Logger) *audit.Logger {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(fmt.Sprintf("failed to open in-memory sqlite for audit: %v", err))
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		operator_id TEXT,
		action TEXT NOT NULL,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		resource_type TEXT,
		resource_id TEXT,
		metadata TEXT,
		result TEXT,
		created_at INTEGER NOT NULL
	)`)
	if err != nil {
		panic(fmt.Sprintf("failed to create audit_logs table: %v", err))
	}
	repo := audit.NewRepository(db)
	return audit.NewLogger(repo, log, audit.LoggerConfig{}, nil)
}
