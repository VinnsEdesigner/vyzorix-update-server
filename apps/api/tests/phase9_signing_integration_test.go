package tests

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
)

// Phase 9: Signing integration test.
//
// This test starts the real server (assumed already running on localhost:3000),
// inserts a verified operator into the Turso DB, logs in to obtain a session
// signing key, then exercises the full signing pipeline:
//
//   1. Signed REST tenant request succeeds
//   2. Signed GraphQL HTTP request succeeds
//   3. Signed GraphQL batch request succeeds
//   4. API-key-authenticated request with signing succeeds
//   5. Unsigned request is rejected (no bypass)
//   6. Tampered signature is rejected
//   7. Old /v1/orgs/:orgId/graphql URL returns 404
//   8. New /:orgId/graphql URL works

const (
	phase9Email    = "phase9-signing@vyzorix.dev"
	phase9Password = "Phase9#Signing2026!"
	phase9Name     = "Phase9 Signing Operator"
)

// phase9SignRequest sets X-Vyzorix-* headers on req using the same canonical
// format as the server's Verifier.Verify.
func phase9SignRequest(req *http.Request, signingKey string, body []byte) {
	nonce := newUUID()
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())

	mac := hmac.New(sha512.New, []byte(signingKey))
	_, _ = mac.Write([]byte(req.Method + "\n" + req.URL.RequestURI() + "\n" + nonce + "\n" + ts + "\n"))
	_, _ = mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Vyzorix-Timestamp", ts)
	req.Header.Set("X-Vyzorix-Nonce", nonce)
	req.Header.Set("X-Vyzorix-Signature", sig)
}

// phase9DeriveAPIKeySecret mirrors the server's deriveAPIKeySigningSecret.
func phase9DeriveAPIKeySecret(fullKey string) string {
	h := sha512.Sum512([]byte(fullKey))
	return hex.EncodeToString(h[:])
}

// phase9Do sends a signed or unsigned request and returns status + body.
func phase9Do(c *http.Client, method, url string, headers map[string]string, body []byte, signingKey string) (int, []byte) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, []byte(err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if signingKey != "" {
		phase9SignRequest(req, signingKey, body)
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

func TestPhase9SigningIntegration(t *testing.T) {
	// Check that the server is running.
	requireServer(t)

	// Check Turso env vars.
	if os.Getenv("TURSO_DB_URL") == "" || os.Getenv("TURSO_VYZOR_SCOPE_DB_TOKEN") == "" {
		t.Skip("TURSO_DB_URL or TURSO_VYZOR_SCOPE_DB_TOKEN not set — skipping")
	}

	// ── Setup: insert verified operator ──
	hasher := password.NewArgon2idHasher()
	hash, err := hasher.Hash(phase9Password)
	if err != nil {
		t.Fatalf("password hash: %v", err)
	}

	// Clean up any prior operator.
	tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
		`DELETE FROM organization_members WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
	tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
		`DELETE FROM organizations WHERE created_by IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
	tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
		`DELETE FROM auth_sessions WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
	tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
		`DELETE FROM operators WHERE email = '%s'`, phase9Email)}})

	operatorID := "op-phase9-" + newUUID()
	nowMs := time.Now().UnixMilli()

	tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
		`INSERT INTO operators (id, email, name, password_hash, role, mfa_enabled, email_verified, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', 'operator', 0, 1, %d, %d)`,
		operatorID, phase9Email, phase9Name, hash, nowMs, nowMs)}})

	t.Cleanup(func() {
		tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
			`DELETE FROM organization_members WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
		tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
			`DELETE FROM organizations WHERE created_by IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
		tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
			`DELETE FROM auth_sessions WHERE operator_id IN (SELECT id FROM operators WHERE email = '%s')`, phase9Email)}})
		tursoExec(t, []map[string]any{{"q": fmt.Sprintf(
			`DELETE FROM operators WHERE email = '%s'`, phase9Email)}})
	})

	// ── Login ──
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	c := &http.Client{
		Timeout: 15 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Get CSRF token.
	csrfResp, err := c.Get(baseURL + "/v1/auth/csrf-token")
	if err != nil {
		t.Fatalf("csrf GET: %v", err)
	}
	var csrfData map[string]any
	_ = json.NewDecoder(csrfResp.Body).Decode(&csrfData)
	_ = csrfResp.Body.Close()
	csrf := csrfData["csrf_token"].(string)

	// Login.
	loginBody, _ := json.Marshal(map[string]string{
		"email":    phase9Email,
		"password": phase9Password,
	})
	status, body := phase9Do(c, "POST", baseURL+"/v1/auth/login",
		map[string]string{"X-CSRF-Token": csrf}, loginBody, "")
	if status != 200 {
		t.Fatalf("login -> %d: %s", status, string(body))
	}

	var loginResp map[string]any
	_ = json.Unmarshal(body, &loginResp)
	signingKey, _ := loginResp["signing_key"].(string)
	if signingKey == "" {
		t.Fatal("login response has no signing_key — server is not generating per-session signing keys")
	}
	t.Logf("✓ Login succeeded, signing key received (%d chars)", len(signingKey))

	// Create an organization (signed request to tenant group).
	orgBody, _ := json.Marshal(map[string]any{
		"name":        "Phase9 Test Org " + newUUID()[:8],
		"description": "Signing integration test org",
		"role":        "super_admin",
	})
	status, body = phase9Do(c, "POST", baseURL+"/v1/organizations",
		map[string]string{"X-CSRF-Token": csrf}, orgBody, signingKey)
	if status != 201 {
		t.Fatalf("create org -> %d: %s", status, string(body))
	}
	var orgResp map[string]any
	_ = json.Unmarshal(body, &orgResp)
	orgID, _ := orgResp["id"].(string)
	if orgID == "" {
		t.Fatalf("create org response has no id: %s", string(body))
	}
	t.Logf("✓ Organization created (signed REST): %s", orgID)

	// Select the organization.
	selectBody, _ := json.Marshal(map[string]string{"organization_id": orgID})
	status, body = phase9Do(c, "POST", baseURL+"/v1/auth/organizations/select",
		map[string]string{"X-CSRF-Token": csrf}, selectBody, signingKey)
	if status != 200 {
		t.Fatalf("org select -> %d: %s", status, string(body))
	}
	t.Logf("✓ Organization selected")

	// ── Test 1: Signed GraphQL HTTP request succeeds ──
	t.Run("signed_graphql_query", func(t *testing.T) {
		gqlBody, _ := json.Marshal(map[string]string{
			"query": `{ __schema { queryType { name } } }`,
		})
		status, body := phase9Do(c, "POST", baseURL+"/"+orgID+"/graphql",
			map[string]string{"X-CSRF-Token": csrf}, gqlBody, signingKey)
		if status != 200 {
			t.Fatalf("signed GraphQL -> %d: %s", status, string(body))
		}
		var result map[string]any
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("GraphQL response parse: %v (body: %s)", err, string(body))
		}
		if hasGraphQLErrors(result) {
			t.Fatalf("GraphQL errors: %s", string(body))
		}
		t.Logf("✓ Signed GraphQL query succeeded (introspection returned)")
	})

	// ── Test 2: Signed GraphQL batch request succeeds ──
	t.Run("signed_graphql_batch", func(t *testing.T) {
		queries := []map[string]any{
			{"query": `{ __schema { queryType { name } } }`},
			{"query": `{ __typename }`},
		}
		batchBody, _ := json.Marshal(queries)
		status, body := phase9Do(c, "POST", baseURL+"/"+orgID+"/graphql/batch",
			map[string]string{"X-CSRF-Token": csrf}, batchBody, signingKey)
		if status != 200 {
			t.Fatalf("signed batch -> %d: %s", status, string(body))
		}
		var results []map[string]any
		if err := json.Unmarshal(body, &results); err != nil {
			t.Fatalf("batch response parse: %v (body: %s)", err, string(body))
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 batch results, got %d: %s", len(results), string(body))
		}
		t.Logf("✓ Signed GraphQL batch succeeded (2 queries returned)")
	})

	// ── Test 3: Unsigned request is rejected (no bypass) ──
	t.Run("unsigned_request_rejected", func(t *testing.T) {
		gqlBody, _ := json.Marshal(map[string]string{
			"query": `{ __typename }`,
		})
		status, body := phase9Do(c, "POST", baseURL+"/"+orgID+"/graphql",
			map[string]string{"X-CSRF-Token": csrf}, gqlBody, "") // no signing key
		if status != 401 {
			t.Errorf("unsigned request -> %d (expected 401): %s", status, string(body))
			return
		}
		if !strings.Contains(string(body), "SIGN_001") && !strings.Contains(string(body), "Missing required signature") {
			t.Errorf("expected SIGN_001 error, got: %s", string(body))
			return
		}
		t.Logf("✓ Unsigned request correctly rejected (401 SIGN_001)")
	})

	// ── Test 4: Tampered signature is rejected ──
	t.Run("tampered_signature_rejected", func(t *testing.T) {
		gqlBody, _ := json.Marshal(map[string]string{
			"query": `{ __typename }`,
		})
		req, _ := http.NewRequest("POST", baseURL+"/"+orgID+"/graphql", bytes.NewReader(gqlBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrf)
		phase9SignRequest(req, signingKey, gqlBody)
		// Tamper: flip the last character of the signature.
		sig := req.Header.Get("X-Vyzorix-Signature")
		last := sig[len(sig)-1]
		if last == 'A' {
			sig = sig[:len(sig)-1] + "B"
		} else {
			sig = sig[:len(sig)-1] + "A"
		}
		req.Header.Set("X-Vyzorix-Signature", sig)

		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("tampered request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 401 {
			t.Errorf("tampered signature -> %d (expected 401): %s", resp.StatusCode, string(body))
			return
		}
		t.Logf("✓ Tampered signature correctly rejected (401)")
	})

	// ── Test 5: Old /v1/orgs/:orgId/graphql URL returns 404 ──
	t.Run("old_graphql_url_404", func(t *testing.T) {
		gqlBody, _ := json.Marshal(map[string]string{
			"query": `{ __typename }`,
		})
		status, body := phase9Do(c, "POST", baseURL+"/v1/orgs/"+orgID+"/graphql",
			map[string]string{"X-CSRF-Token": csrf}, gqlBody, signingKey)
		// The old URL either returns 404 (no SSR) or the SPA index (200 HTML).
		// Either way it must NOT return a valid GraphQL JSON response.
		if status == 200 && strings.Contains(strings.ToLower(string(body)), `"data"`) {
			t.Errorf("old URL served GraphQL (expected 404 or non-GraphQL): %s", string(body))
			return
		}
		t.Logf("✓ Old /v1/orgs/%s/graphql URL does not serve GraphQL (status %d)", orgID, status)
	})

	// ── Test 6: API key auth + signing ──
	t.Run("api_key_auth_signing", func(t *testing.T) {
		// Create an API key via the session.
		keyBody, _ := json.Marshal(map[string]string{
			"name":  "phase9-test-key",
			"scope": "read",
		})
		status, body := phase9Do(c, "POST", baseURL+"/v1/auth/api-keys",
			map[string]string{"X-CSRF-Token": csrf, "X-Organization-ID": orgID}, keyBody, signingKey)
		if status != 201 {
			t.Fatalf("create API key -> %d: %s", status, string(body))
		}
		var keyResp map[string]any
		_ = json.Unmarshal(body, &keyResp)
		apiKey, _ := keyResp["api_key"].(string)
		if apiKey == "" {
			t.Fatalf("no api_key in response: %s", string(body))
		}
		t.Logf("✓ API key created: %s...", apiKey[:min(16, len(apiKey))])

		// Use the API key with signing to query a tenant route.
		derivedSecret := phase9DeriveAPIKeySecret(apiKey)
		devicesBody, _ := json.Marshal(map[string]any{
			"page": 1,
			"limit": 10,
		})
		status, body = phase9Do(c, "POST", baseURL+"/v1/devices/list",
			map[string]string{"X-API-Key": apiKey, "X-Organization-ID": orgID}, devicesBody, derivedSecret)
		// The device list endpoint may return 200 or 404 depending on route;
		// the key test is that it's NOT 401 (which would mean signing failed).
		if status == 401 {
			errData := string(body)
			if strings.Contains(errData, "SIGN_") {
				t.Errorf("API key signing failed -> %d: %s", status, errData)
				return
			}
		}
		t.Logf("✓ API key + signing accepted (status %d)", status)
	})

	// ── Test 7: API key without signing is rejected ──
	t.Run("api_key_unsigned_rejected", func(t *testing.T) {
		// Create another API key.
		keyBody, _ := json.Marshal(map[string]string{
			"name":  "phase9-unsigned-key",
			"scope": "read",
		})
		status, body := phase9Do(c, "POST", baseURL+"/v1/auth/api-keys",
			map[string]string{"X-CSRF-Token": csrf, "X-Organization-ID": orgID}, keyBody, signingKey)
		if status != 201 {
			t.Fatalf("create API key -> %d: %s", status, string(body))
		}
		var keyResp map[string]any
		_ = json.Unmarshal(body, &keyResp)
		apiKey, _ := keyResp["api_key"].(string)

		// Use API key WITHOUT signing on the GraphQL endpoint (which has
		// the signing middleware in its chain). The signing middleware
		// must reject the unsigned request before it reaches the handler.
		gqlBody, _ := json.Marshal(map[string]string{
			"query": `{ __typename }`,
		})
		status, body = phase9Do(c, "POST", baseURL+"/"+orgID+"/graphql",
			map[string]string{"X-API-Key": apiKey, "X-Organization-ID": orgID}, gqlBody, "") // no signing
		if status != 401 {
			t.Errorf("unsigned API key -> %d (expected 401): %s", status, string(body))
			return
		}
		if !strings.Contains(string(body), "SIGN_001") && !strings.Contains(string(body), "Missing required signature") {
			t.Errorf("expected SIGN_001 error, got: %s", string(body))
			return
		}
		t.Logf("✓ Unsigned API key request correctly rejected (401 SIGN_001)")
	})
}
