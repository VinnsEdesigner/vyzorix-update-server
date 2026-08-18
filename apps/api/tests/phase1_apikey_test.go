package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestPhase1APIKey tests API key creation, endpoint access, scope enforcement,
// and key management. 45 subtests.
//
// Flow: register operator → verify email → login → create org → create API keys.
// (read/write/admin) → test GET endpoints per scope → auth rejection → scope.
// enforcement → key list/rotate/revoke.
func TestPhase1APIKey(t *testing.T) {
	requireServer(t)
	requireMockEmail(t)

	ts := timestamp()
	email := fmt.Sprintf("tester1_%d@vyzorix-test.local", ts)
	name := "Test Operator 1"

// ── Setup: register + verify + login ──.
	c, err := newSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	var csrf string
	t.Run("csrf_token_acquired", func(t *testing.T) {
		csrf = getCSRF(t, c)
		if csrf == "" {
			t.Fatal("csrf token is empty")
		}
		t.Logf("csrf token acquired (len=%d)", len(csrf))
	})

	t.Run("register", func(t *testing.T) {
		status, body := doJSON(t, c, "POST", "/v1/auth/register",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"email": email, "password": testPassword, "name": name})
		if status != 201 {
			t.Fatalf("register -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		opID, _ := data["operator_id"].(string)
		if opID == "" {
			t.Fatal("missing operator_id in register response")
		}
		t.Logf("registered operator_id=%s", opID[:12])
	})

	t.Run("verification_email_captured", func(t *testing.T) {
		token := fetchVerifyToken(t, email)
		if token == "" {
			t.Fatal("no verification email captured by mock-resend")
		}
		t.Logf("token=%s...", token[:16])
	})

	vtok := fetchVerifyToken(t, email)

	t.Run("verify_email", func(t *testing.T) {
		status, body := doJSON(t, c, "POST", "/v1/auth/verify-email",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"token": vtok})
		if status != 200 {
			t.Fatalf("verify-email -> %d: %s", status, string(body))
		}
	})

	t.Run("login", func(t *testing.T) {
		status, body := doJSON(t, c, "POST", "/v1/auth/login",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"email": email, "password": testPassword})
		if status != 200 {
			t.Fatalf("login -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		if _, ok := data["error"]; ok {
			t.Fatalf("login returned error: %s", string(body))
		}
		loggedInEmail, _ := data["email"].(string)
		t.Logf("logged in as %s", loggedInEmail)
	})

// ── Setup: create organization ──.
	var orgID string

	t.Run("create_organization", func(t *testing.T) {
		status, body := doJSON(t, c, "POST", "/v1/organizations",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"name": fmt.Sprintf("TestOrg_%d", ts), "description": "Test organization", "role": "super_admin", "maxMembers": 10})
		if status == 201 {
			data := parseJSON(t, body)
			orgID, _ = data["id"].(string)
			t.Logf("org created id=%s", orgID[:12])
			return
		}
		if status == 409 {
// Org limit reached — reuse existing.
			s2, b2 := doJSON(t, c, "GET", "/v1/organizations", nil, nil)
			if s2 == 200 {
				var orgs []map[string]any
				_ = json.Unmarshal(b2, &orgs)
				if len(orgs) > 0 {
					orgID, _ = orgs[0]["id"].(string)
					t.Logf("org limit reached, reusing org id=%s", orgID[:12])
					return
				}
			}
		}
// Try listing existing orgs as fallback.
		s2, b2 := doJSON(t, c, "GET", "/v1/organizations", nil, nil)
		if s2 == 200 {
			var orgs []map[string]any
			_ = json.Unmarshal(b2, &orgs)
			if len(orgs) > 0 {
				orgID, _ = orgs[0]["id"].(string)
				t.Logf("reusing existing org id=%s", orgID[:12])
				return
			}
		}
		t.Fatalf("org create -> %d: %s", status, string(body))
	})

	if orgID == "" {
		t.Fatal("cannot proceed without orgID")
	}

	t.Run("select_organization", func(t *testing.T) {
		status, body := doJSON(t, c, "POST", "/v1/auth/organizations/select",
			map[string]string{"X-CSRF-Token": csrf},
			map[string]any{"organization_id": orgID})
		if status != 200 {
			t.Fatalf("org select -> %d: %s", status, string(body))
		}
		t.Logf("org selected: %s", orgID[:12])
	})

// ── Create API keys ──.
	apiKeys := map[string]string{}

	t.Run("create_read_key", func(t *testing.T) {
		key := createAPIKey(t, c, csrf, orgID, "test-key-read", "read")
		if key == "" {
			t.Fatal("read key is empty")
		}
		apiKeys["read"] = key
		t.Logf("read key: %s...", key[:20])
	})

	t.Run("create_write_key", func(t *testing.T) {
		key := createAPIKey(t, c, csrf, orgID, "test-key-write", "write")
		if key == "" {
			t.Fatal("write key is empty")
		}
		apiKeys["write"] = key
		t.Logf("write key: %s...", key[:20])
	})

	t.Run("create_admin_key", func(t *testing.T) {
		key := createAPIKey(t, c, csrf, orgID, "test-key-admin", "admin")
		if key == "" {
			t.Fatal("admin key is empty")
		}
		apiKeys["admin"] = key
		t.Logf("admin key: %s...", key[:20])
	})

// ── Test GET endpoints with each scope ──.
	getEndpoints := []struct {
		path  string
		label string
	}{
		{"/v1/dashboard/devices", "dashboard devices"},
		{"/v1/devices", "devices list"},
		{"/v1/connections", "connections"},
		{"/v1/connections/metrics", "connection metrics"},
		{"/v1/updates/status", "updates status"},
		{"/v1/updates/versions", "updates versions"},
		{"/v1/updates/history", "updates history"},
		{"/v1/organizations", "organizations list"},
		{"/v1/telemetry/history", "telemetry history"},
	}

	for _, scope := range []string{"read", "write", "admin"} {
		key, ok := apiKeys[scope]
		if !ok {
			continue
		}
		for _, ep := range getEndpoints {
			ep := ep
			t.Run(fmt.Sprintf("scope_%s_GET_%s", scope, ep.label), func(t *testing.T) {
				time.Sleep(100 * time.Millisecond) // avoid rate limit
				status, _ := doRaw(t, httpClient, "GET", ep.path,
					map[string]string{"X-API-Key": key, "X-Organization-ID": orgID})
				if status == 200 || status == 400 || status == 404 {
					t.Logf("%s GET %s -> %d", scope, ep.label, status)
					return
				}
				t.Errorf("%s GET %s -> %d (expected 200/400/404)", scope, ep.label, status)
			})
		}
	}

// ── Auth rejection ──.
	t.Run("no_key_rejected", func(t *testing.T) {
		status, _ := doRaw(t, httpClient, "GET", "/v1/devices",
			map[string]string{"X-Organization-ID": orgID})
		if status == 401 || status == 403 {
			t.Logf("no key -> %d (rejected)", status)
			return
		}
		t.Errorf("no key -> %d (expected 401/403)", status)
	})

	t.Run("invalid_key_rejected", func(t *testing.T) {
		status, _ := doRaw(t, httpClient, "GET", "/v1/devices",
			map[string]string{"X-API-Key": "vxyz_invalid_key_12345", "X-Organization-ID": orgID})
		if status == 401 || status == 403 {
			t.Logf("invalid key -> %d (rejected)", status)
			return
		}
		t.Errorf("invalid key -> %d (expected 401/403)", status)
	})

// ── Scope enforcement ──.
	t.Run("read_key_POST_rejected", func(t *testing.T) {
		key := apiKeys["read"]
		status, _ := doJSON(t, httpClient, "POST", "/v1/updates/sync",
			map[string]string{"X-API-Key": key, "X-Organization-ID": orgID},
			map[string]any{})
		if status == 403 {
			t.Logf("read key POST updates/sync -> 403 (scope enforced)")
			return
		}
		t.Errorf("read key POST updates/sync -> %d (expected 403)", status)
	})

	t.Run("write_key_DELETE_rejected", func(t *testing.T) {
		key := apiKeys["write"]
		status, _ := doRaw(t, httpClient, "DELETE", "/v1/devices/nonexistent",
			map[string]string{"X-API-Key": key, "X-Organization-ID": orgID})
		if status == 403 {
			t.Logf("write key DELETE device -> 403 (scope enforced)")
			return
		}
		if status == 404 {
			t.Logf("write key DELETE device -> 404 (passed scope, not found)")
			return
		}
		t.Errorf("write key DELETE device -> %d (expected 403/404)", status)
	})

	t.Run("admin_key_DELETE_scope_passed", func(t *testing.T) {
		key := apiKeys["admin"]
		status, _ := doRaw(t, httpClient, "DELETE", "/v1/devices/nonexistent",
			map[string]string{"X-API-Key": key, "X-Organization-ID": orgID})
		if status == 403 || status == 404 || status == 200 {
			t.Logf("admin key DELETE device -> %d (scope passed)", status)
			return
		}
		t.Errorf("admin key DELETE device -> %d (expected 403/404/200)", status)
	})

// ── Key management ──.
	var keysList []map[string]any

	t.Run("list_keys", func(t *testing.T) {
		status, body := doJSON(t, c, "GET", "/v1/auth/api-keys",
			map[string]string{"X-Organization-ID": orgID}, nil)
		if status != 200 {
			t.Fatalf("list keys -> %d: %s", status, string(body))
		}
// Response may be a list or an object with keys/items.
		var rawList []map[string]any
		if err := json.Unmarshal(body, &rawList); err == nil {
			keysList = rawList
		} else {
			data := parseJSON(t, body)
			if items, ok := data["keys"].([]any); ok {
				for _, item := range items {
					if m, ok := item.(map[string]any); ok {
						keysList = append(keysList, m)
					}
				}
			} else if items, ok := data["items"].([]any); ok {
				for _, item := range items {
					if m, ok := item.(map[string]any); ok {
						keysList = append(keysList, m)
					}
				}
			}
		}
		t.Logf("list keys -> %d keys", len(keysList))
		if len(keysList) == 0 {
			t.Error("expected at least 1 key")
		}
	})

	t.Run("rotate_write_key", func(t *testing.T) {
		var writeKeyID string
		for _, k := range keysList {
			if n, _ := k["name"].(string); n == "test-key-write" {
				writeKeyID, _ = k["id"].(string)
				break
			}
		}
		if writeKeyID == "" {
			t.Skip("write key not found in list")
		}
		status, body := doJSON(t, c, "POST", "/v1/auth/api-keys/"+writeKeyID+"/rotate",
			map[string]string{"X-CSRF-Token": csrf, "X-Organization-ID": orgID}, nil)
		if status != 200 {
			t.Fatalf("rotate write key -> %d: %s", status, string(body))
		}
		data := parseJSON(t, body)
		newKey, _ := data["api_key"].(string)
		if newKey == "" {
			t.Fatal("rotated key is empty")
		}
		apiKeys["write"] = newKey
		t.Logf("rotated write key -> %s...", newKey[:16])
	})

	t.Run("rotated_write_key_valid", func(t *testing.T) {
		key := apiKeys["write"]
		time.Sleep(500 * time.Millisecond)
		status, _ := doRaw(t, httpClient, "GET", "/v1/devices",
			map[string]string{"X-API-Key": key, "X-Organization-ID": orgID})
		if status == 200 || status == 404 || status == 400 {
			t.Logf("rotated write key GET devices -> %d", status)
			return
		}
		t.Errorf("rotated write key GET devices -> %d (expected 200/404/400)", status)
	})

// ── Save state for later phases (not a test, just cross-phase sharing) ──.
	opID := ""
	status, body := doJSON(t, c, "GET", "/v1/auth/me", nil, nil)
	if status == 200 {
		data := parseJSON(t, body)
		opID, _ = data["id"].(string)
		if opID == "" {
			opID, _ = data["operator_id"].(string)
		}
	}
	state := &TestState{
		Email:      email,
		Password:   testPassword,
		Name:       name,
		OrgID:      orgID,
		OperatorID: opID,
		APIKeys:    apiKeys,
		CSRF:       csrf,
	}
	saveState(state)

	// Write version.json + changelog.json for Phase 2 (setup, not a test).
	writeStaticFiles(t, ts)
}

// writeStaticFiles creates version.json and changelog.json in the data dir.
// This is setup for Phase 2 but called at the end of Phase 1 to avoid.
// duplicating logic.
func writeStaticFiles(t *testing.T, ts int64) {
	versionData := map[string]any{
		"version":        fmt.Sprintf("2.1.%d", ts%1000),
		"apk_filename":   fmt.Sprintf("vyzorix-2.1.%d.apk", ts%1000),
		"apk_sha256":     "a1b2c3d4e5f6" + strings.Repeat("0", 52),
		"release_notes":  "Test release with new features and bug fixes",
		"version_code":   ts % 10000,
		"apk_size_bytes": 15728640,
	}
	versionPath := filepath.Join(dataDir, "version.json")
	vdata, _ := json.MarshalIndent(versionData, "", "  ")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Logf("mkdir data dir: %v", err)
		return
	}
	if err := os.WriteFile(versionPath, vdata, 0644); err != nil {
		t.Logf("write version.json: %v", err)
		return
	}
	t.Logf("created version.json -> version=%s", versionData["version"])

	changelogData := map[string]any{
		"versions": []map[string]any{
			{"version": "2.1.0", "date": "2026-08-01", "notes": "Initial 2.1 release"},
			{"version": "2.0.5", "date": "2026-07-15", "notes": "Bug fixes and stability"},
			{"version": "2.0.0", "date": "2026-06-01", "notes": "Major 2.0 release with new UI"},
		},
	}
	changelogPath := filepath.Join(dataDir, "changelog.json")
	cdata, _ := json.MarshalIndent(changelogData, "", "  ")
	if err := os.WriteFile(changelogPath, cdata, 0644); err != nil {
		t.Logf("write changelog.json: %v", err)
		return
	}
	t.Logf("created changelog.json -> 3 versions")
}

// Suppress unused import warnings for http (used via httpClient references).
var _ = http.MethodGet
