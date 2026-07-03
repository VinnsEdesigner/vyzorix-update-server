// Package verify provides server-side verification for AUTHENTICATION_SYSTEM_SERVER.md
// This script verifies ALL server-side requirements from the authentication system specification
// including endpoints, handlers, error handling, security, file names, and database schema.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyAuthSecurity verifies ALL server-side requirements from AUTHENTICATION_SYSTEM_SERVER.md
func verifyAuthSecurity() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  AUTHENTICATION_SYSTEM_SERVER.md - SERVER-SIDE VERIFICATION          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// SECTION 2: Current Server Structure
	// =========================================================================
	fmt.Println("📋 SECTION 2: CURRENT SERVER STRUCTURE")
	fmt.Println(strings.Repeat("─", 75))

	// 2.1 Existing Auth Handlers
	fmt.Println("--- 2.1 Existing Auth Handlers ---")
	authHandlers := []struct {
		id      string
		file    string
		methods []string
	}{
		{"H-AUTH1", "auth_login.go", []string{"Handle"}},
		{"H-AUTH2", "auth_register.go", []string{"Handle"}},
		{"H-AUTH3", "auth_logout.go", []string{"Handle"}},
		{"H-AUTH4", "auth_me.go", []string{"Handle"}},
		{"H-AUTH5", "auth_mfa.go", []string{"GetMFAStatus", "EnrollMFA", "VerifySetupMFA", "EnableMFA", "DisableMFA", "VerifyBackupCode", "RegenerateBackupCodes"}},
		{"H-AUTH6", "auth_oauth.go", []string{"GoogleLogin", "GoogleCallback", "GitHubLogin", "GitHubCallback"}},
		{"H-AUTH7", "auth_password_reset.go", []string{"ForgotPassword", "ResetPassword", "ResendPasswordReset"}},
		{"H-AUTH8", "auth_email_verify.go", []string{"VerifyEmail", "ResendVerification", "CancelVerification", "PollVerification"}},
		{"H-AUTH9", "auth_settings.go", []string{"UpdateName", "UpdateSettings"}},
		{"H-AUTH10", "auth_admin.go", []string{"ListOperators", "CreateOperator", "GetOperator", "UpdateOperator", "DeleteOperator"}},
		{"H-AUTH11", "auth_lockout.go", []string{"GetLockoutStatus", "UnlockAccount"}},
	}

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/auth/")
	for _, h := range authHandlers {
		path := filepath.Join(handlerDir, h.file)
		if content, err := os.ReadFile(path); err == nil {
			contentStr := string(content)
			allFound := true
			missingMethods := []string{}
			for _, method := range h.methods {
				if !strings.Contains(contentStr, method) {
					allFound = false
					missingMethods = append(missingMethods, method)
				}
			}
			if allFound {
				fmt.Printf("  ✅ %s  %-25s (%v)\n", h.id, h.file, h.methods)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %-25s (MISSING: %v)\n", h.id, h.file, missingMethods)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %-25s (FILE NOT FOUND)\n", h.id, h.file)
			failed++
		}
	}

	// 2.2 Domain Layer Entities
	fmt.Println("\n--- 2.2 Domain Layer Entities ---")
	domainEntities := []struct {
		id   string
		dir  string
		file string
	}{
		// operator domain
		{"DOM-E1", "operator", "operator_entity.go"},
		{"DOM-E2", "operator", "operator_repository.go"},
		{"DOM-E3", "operator", "operator_errors.go"},
		{"DOM-E4", "operator", "operator_password.go"},
		{"DOM-E5", "operator", "operator_role.go"},
		{"DOM-E6", "operator", "operator_requests.go"},
		{"DOM-E7", "operator", "operator_responses.go"},
		{"DOM-E8", "operator", "operator_email.go"},
		// session domain
		{"DOM-E9", "session", "session_entity.go"},
		{"DOM-E10", "session", "session_repository.go"},
		// refresh_token domain
		{"DOM-E11", "refresh_token", "refresh_token_entity.go"},
		{"DOM-E12", "refresh_token", "refresh_token_repository.go"},
		// email_verification domain
		{"DOM-E13", "email_verification", "email_verification_entity.go"},
		{"DOM-E14", "email_verification", "email_verification_repository.go"},
		{"DOM-E15", "email_verification", "email_verification_requests.go"},
		{"DOM-E16", "email_verification", "email_verification_responses.go"},
		// password_reset domain
		{"DOM-E17", "password_reset", "password_reset_entity.go"},
		{"DOM-E18", "password_reset", "password_reset_repository.go"},
		{"DOM-E19", "password_reset", "password_reset_requests.go"},
		{"DOM-E20", "password_reset", "password_reset_responses.go"},
		// oauth domain
		{"DOM-E21", "oauth", "oauth_entity.go"},
		{"DOM-E22", "oauth", "oauth_errors.go"},
	}

	domainDir := filepath.Join(root, "apps/api/internal/domain/")
	for _, d := range domainEntities {
		path := filepath.Join(domainDir, d.dir, d.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  domain/%s/%s\n", d.id, d.dir, d.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  domain/%s/%s (MISSING)\n", d.id, d.dir, d.file)
			failed++
		}
	}

	// 2.3 Infrastructure Components
	fmt.Println("\n--- 2.3 Infrastructure Components ---")
	infraComponents := []struct {
		id   string
		dir  string
		file string
	}{
		{"INF-C1", "security", "jwt.go"},
		{"INF-C2", "security", "password.go"},
		{"INF-C3", "security", "google.go"},
		{"INF-C4", "security/session", "session_manager.go"},
		{"INF-C5", "security/session", "session.go"},
		{"INF-C6", "email", "email_service.go"},
		{"INF-C7", "storage", "auth_storage.go"},
		{"INF-C8", "storage", "session_storage.go"},
		{"INF-C9", "storage", "operator_storage.go"},
	}

	infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
	for _, i := range infraComponents {
		path := filepath.Join(infraDir, i.dir, i.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  infrastructure/%s/%s\n", i.id, i.dir, i.file)
			passed++
		} else {
			// Check without extension
			pathExt := path + ".go"
			if _, err := os.Stat(pathExt); err == nil {
				fmt.Printf("  ✅ %s  infrastructure/%s/%s.go\n", i.id, i.dir, i.file)
				passed++
			} else {
				fmt.Printf("  ❌ %s  infrastructure/%s/%s (MISSING)\n", i.id, i.dir, i.file)
				failed++
			}
		}
	}

	// 2.4 Middleware Components
	fmt.Println("\n--- 2.4 Middleware Components ---")
	middlewareComponents := []struct {
		id   string
		file string
	}{
		{"MID-C1", "auth.go"},
		{"MID-C2", "cookie_auth.go"},
		{"MID-C3", "lockout.go"},
		{"MID-C4", "rate_limit.go"},
		{"MID-C5", "validation.go"},
	}

	middlewareDir := filepath.Join(root, "apps/api/internal/api/middleware/")
	for _, m := range middlewareComponents {
		path := filepath.Join(middlewareDir, m.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  api/middleware/%s\n", m.id, m.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  api/middleware/%s (MISSING)\n", m.id, m.file)
			failed++
		}
	}

	// 2.5 Response Components
	fmt.Println("\n--- 2.5 Response Components ---")
	responsesPath := filepath.Join(root, "apps/api/internal/api/responses/presenter.go")
	if _, err := os.Stat(responsesPath); err == nil {
		fmt.Printf("  ✅ RESP-1  api/responses/presenter.go\n")
		passed++
	} else {
		fmt.Printf("  ❌ RESP-1  api/responses/presenter.go (MISSING)\n")
		failed++
	}

	// =========================================================================
	// SECTION 3: Required Endpoints
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REQUIRED ENDPOINTS")
	fmt.Println(strings.Repeat("─", 75))

	// 3.1 Authentication Endpoints
	fmt.Println("--- 3.1 Authentication Endpoints ---")
	authEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
		file    string
	}{
		{"AE-1", "POST", "/v1/auth/login", "Handle", "auth_login.go"},
		{"AE-2", "POST", "/v1/auth/register", "Handle", "auth_register.go"},
		{"AE-3", "POST", "/v1/auth/logout", "Handle", "auth_logout.go"},
		{"AE-4", "GET", "/v1/auth/me", "Handle", "auth_me.go"},
		{"AE-5", "PATCH", "/v1/auth/me", "Handle", "auth_me.go"},
		{"AE-6", "POST", "/v1/auth/refresh", "Handle", "auth_refresh.go"},
	}

	for _, ep := range authEndpoints {
		path := filepath.Join(handlerDir, ep.file)
		if content, err := os.ReadFile(path); err == nil {
			if strings.Contains(string(content), ep.handler) {
				fmt.Printf("  ✅ %s  %-6s %-35s (%s in %s)\n", ep.id, ep.method, ep.path, ep.handler, ep.file)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING in %s)\n", ep.id, ep.method, ep.path, ep.handler, ep.file)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %-6s %-35s (FILE NOT FOUND: %s)\n", ep.id, ep.method, ep.path, ep.file)
			failed++
		}
	}

	// 3.2 MFA Endpoints
	fmt.Println("\n--- 3.2 MFA Endpoints ---")
	mfaEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"MFA-E1", "GET", "/v1/auth/mfa/status", "GetMFAStatus"},
		{"MFA-E2", "POST", "/v1/auth/mfa/enroll", "EnrollMFA"},
		{"MFA-E3", "POST", "/v1/auth/mfa/verify-setup", "VerifySetupMFA"},
		{"MFA-E4", "POST", "/v1/auth/mfa/verify", "VerifyMFA"},
		{"MFA-E5", "POST", "/v1/auth/mfa/enable", "EnableMFA"},
		{"MFA-E6", "POST", "/v1/auth/mfa/disable", "DisableMFA"},
		{"MFA-E7", "POST", "/v1/auth/mfa/verify-backup", "VerifyBackupCode"},
		{"MFA-E8", "POST", "/v1/auth/mfa/regenerate-backup-codes", "RegenerateBackupCodes"},
	}

	mfaPath := filepath.Join(handlerDir, "auth_mfa.go")
	mfaContent, _ := os.ReadFile(mfaPath)
	mfaContentStr := string(mfaContent)

	// Also check for separate MFA verify handler
	mfaVerifyPath := filepath.Join(handlerDir, "auth_mfa_verify.go")
	mfaVerifyExists := false
	if content, err := os.ReadFile(mfaVerifyPath); err == nil {
		mfaVerifyExists = strings.Contains(string(content), "VerifyMFA") || strings.Contains(string(content), "MFAVerify")
	}

	for _, ep := range mfaEndpoints {
		handlerPattern := "func (h *MFAHandler) " + ep.handler
		found := false
		if ep.handler == "VerifyMFA" {
			found = mfaVerifyExists || strings.Contains(mfaContentStr, handlerPattern)
		} else {
			found = strings.Contains(mfaContentStr, handlerPattern)
		}
		if found {
			fmt.Printf("  ✅ %s  %-6s %-35s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// 3.3 Password Reset Endpoints
	fmt.Println("\n--- 3.3 Password Reset Endpoints ---")
	passwordResetEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"PR-E1", "POST", "/v1/auth/forgot-password", "ForgotPassword"},
		{"PR-E2", "POST", "/v1/auth/reset-password", "ResetPassword"},
		{"PR-E3", "POST", "/v1/auth/resend-password-reset", "ResendPasswordReset"},
	}

	prPath := filepath.Join(handlerDir, "auth_password_reset.go")
	if content, err := os.ReadFile(prPath); err == nil {
		contentStr := string(content)
		for _, ep := range passwordResetEndpoints {
			handlerPattern := "func (h *PasswordResetHandler) " + ep.handler
			if strings.Contains(contentStr, handlerPattern) {
				fmt.Printf("  ✅ %s  %-6s %-40s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %-6s %-40s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
				failed++
			}
		}
	} else {
		for _, ep := range passwordResetEndpoints {
			fmt.Printf("  ⚠️  %s  %-6s %-40s (FILE NOT FOUND)\n", ep.id, ep.method, ep.path)
			failed++
		}
	}

	// 3.4 Email Verification Endpoints
	fmt.Println("\n--- 3.4 Email Verification Endpoints ---")
	emailVerifyEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"EV-E1", "POST", "/v1/auth/verify-email", "VerifyEmail"},
		{"EV-E2", "POST", "/v1/auth/resend-verification", "ResendVerification"},
		{"EV-E3", "POST", "/v1/auth/cancel-verification", "CancelVerification"},
		{"EV-E4", "GET", "/v1/auth/poll-verification", "PollVerification"},
	}

	evPath := filepath.Join(handlerDir, "auth_email_verify.go")
	if content, err := os.ReadFile(evPath); err == nil {
		contentStr := string(content)
		for _, ep := range emailVerifyEndpoints {
			handlerPattern := "func (h *EmailVerifyHandler) " + ep.handler
			if strings.Contains(contentStr, handlerPattern) {
				fmt.Printf("  ✅ %s  %-6s %-40s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %-6s %-40s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
				failed++
			}
		}
	} else {
		for _, ep := range emailVerifyEndpoints {
			fmt.Printf("  ⚠️  %s  %-6s %-40s (FILE NOT FOUND)\n", ep.id, ep.method, ep.path)
			failed++
		}
	}

	// 3.5 OAuth Endpoints
	fmt.Println("\n--- 3.5 OAuth Endpoints ---")
	oauthEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"OAUTH-E1", "GET", "/v1/auth/google", "GoogleLogin"},
		{"OAUTH-E2", "GET", "/v1/auth/google/callback", "GoogleCallback"},
		{"OAUTH-E3", "GET", "/v1/auth/github", "GitHubLogin"},
		{"OAUTH-E4", "GET", "/v1/auth/github/callback", "GitHubCallback"},
	}

	oauthPath := filepath.Join(handlerDir, "auth_oauth.go")
	oauthContent, _ := os.ReadFile(oauthPath)
	oauthContentStr := string(oauthContent)
	for _, ep := range oauthEndpoints {
		handlerPattern := "func (h *OAuthHandler) " + ep.handler
		if strings.Contains(oauthContentStr, handlerPattern) {
			fmt.Printf("  ✅ %s  %-6s %-35s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// 3.6 Admin Endpoints
	fmt.Println("\n--- 3.6 Admin Endpoints ---")
	adminEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"ADMIN-E1", "GET", "/v1/auth/admin/operators", "ListOperators"},
		{"ADMIN-E2", "POST", "/v1/auth/admin/operators", "CreateOperator"},
		{"ADMIN-E3", "GET", "/v1/auth/admin/operators/:id", "GetOperator"},
		{"ADMIN-E4", "PATCH", "/v1/auth/admin/operators/:id", "UpdateOperator"},
		{"ADMIN-E5", "DELETE", "/v1/auth/admin/operators/:id", "DeleteOperator"},
		{"ADMIN-E6", "GET", "/v1/auth/admin/lockout/:operator_id", "GetLockoutStatus"},
		{"ADMIN-E7", "POST", "/v1/auth/admin/lockout/:operator_id/unlock", "UnlockAccount"},
	}

	adminPath := filepath.Join(handlerDir, "auth_admin.go")
	adminContent, _ := os.ReadFile(adminPath)
	adminContentStr := string(adminContent)
	for _, ep := range adminEndpoints {
		handlerPattern := "func (h *AdminHandler) " + ep.handler
		if strings.Contains(adminContentStr, handlerPattern) {
			fmt.Printf("  ✅ %s  %-6s %-45s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-45s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// =========================================================================
	// SECTION 4: Domain Layer
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: DOMAIN LAYER STRUCTURES")
	fmt.Println(strings.Repeat("─", 75))

	domainChecks := []struct {
		id          string
		description string
		checkFile   string
		check       string
	}{
		// Operator domain
		{"D-CHK1", "Operator struct with ID, Email, Name, Role", "operator/operator_entity.go", "type Operator struct"},
		{"D-CHK2", "Operator role constants", "operator/operator_role.go", "Role"},
		{"D-CHK3", "Operator repository interface", "operator/operator_repository.go", "Repository"},
		{"D-CHK4", "Password hashing methods", "operator/operator_password.go", "HashPassword"},
		{"D-CHK5", "Login request type", "operator/operator_requests.go", "LoginRequest"},
		// Session domain
		{"D-CHK6", "Session struct", "session/session_entity.go", "type Session struct"},
		{"D-CHK7", "Session repository interface", "session/session_repository.go", "Repository"},
		// Refresh token domain
		{"D-CHK8", "RefreshToken struct", "refresh_token/refresh_token_entity.go", "type RefreshToken struct"},
		{"D-CHK9", "RefreshToken repository interface", "refresh_token/refresh_token_repository.go", "Repository"},
		// Email verification domain
		{"D-CHK10", "EmailVerification struct", "email_verification/email_verification_entity.go", "type EmailVerification struct"},
		// Password reset domain
		{"D-CHK11", "PasswordReset struct", "password_reset/password_reset_entity.go", "type PasswordReset struct"},
		// OAuth domain
		{"D-CHK12", "OAuthProvider enum", "oauth/oauth_entity.go", "OAuthProvider"},
	}

	for _, d := range domainChecks {
		path := filepath.Join(domainDir, d.checkFile)
		if content, err := os.ReadFile(path); err == nil {
			if strings.Contains(string(content), d.check) {
				fmt.Printf("  ✅ %s  %s\n", d.id, d.description)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %s (MISSING in %s)\n", d.id, d.description, d.checkFile)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %s (FILE NOT FOUND: %s)\n", d.id, d.description, d.checkFile)
			failed++
		}
	}

	// =========================================================================
	// SECTION 5: Application Layer - Auth Service
	// =========================================================================
	fmt.Println("\n📋 SECTION 5: APPLICATION LAYER - Auth Service")
	fmt.Println(strings.Repeat("─", 75))

	authServiceMethods := []struct {
		id     string
		method string
		file   string
	}{
		{"AS-1", "Login", "auth_service.go"},
		{"AS-2", "Register", "auth_service.go"},
		{"AS-3", "Logout", "auth_service.go"},
		{"AS-4", "CreateSession", "auth_service.go"},
		{"AS-5", "VerifyPassword", "auth_service.go"},
		{"AS-6", "HashPassword", "auth_service.go"},
		{"AS-7", "GetMFAStatus", "auth_service.go"},
		{"AS-8", "EnrollMFA", "auth_service.go"},
		{"AS-9", "VerifyMFAEnrollment", "auth_service.go"},
		{"AS-10", "EnableMFA", "auth_service.go"},
		{"AS-11", "DisableMFA", "auth_service.go"},
		{"AS-12", "VerifyMFACode", "auth_service.go"},
		{"AS-13", "RegenerateBackupCodes", "auth_service.go"},
		{"AS-14", "ChangePassword", "auth_service.go"},
		{"AS-15", "ResetPassword", "auth_service.go"},
		{"AS-16", "InitiatePasswordReset", "auth_service.go"},
		{"AS-17", "HandleGoogleCallback", "auth_service.go"},
		{"AS-18", "HandleGitHubCallback", "auth_service.go"},
		{"AS-19", "RotateRefreshToken", "auth_service.go"},
		{"AS-20", "IssueRefreshToken", "auth_service.go"},
		{"AS-21", "RevokeAllRefreshTokens", "auth_service.go"},
		{"AS-22", "VerifyEmail", "auth_service.go"},
		{"AS-23", "InitiateEmailVerification", "auth_service.go"},
	}

	appDir := filepath.Join(root, "apps/api/internal/application/auth/")
	for _, s := range authServiceMethods {
		path := filepath.Join(appDir, s.file)
		if content, err := os.ReadFile(path); err == nil {
			contentStr := string(content)
			handlerPattern := "func (s *AuthService) " + s.method
			if strings.Contains(contentStr, handlerPattern) {
				fmt.Printf("  ✅ %s  AuthService.%s()\n", s.id, s.method)
				passed++
			} else {
				fmt.Printf("  ❌ %s  AuthService.%s() (MISSING)\n", s.id, s.method)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %s (FILE NOT FOUND)\n", s.id, s.file)
			failed++
		}
	}

	// =========================================================================
	// SECTION 6: Handler Specifications
	// =========================================================================
	fmt.Println("\n📋 SECTION 6: HANDLER SPECIFICATIONS")
	fmt.Println(strings.Repeat("─", 75))

	// 6.1 Login Handler
	fmt.Println("--- 6.1 Login Handler (POST /v1/auth/login) ---")
	loginChecks := []struct {
		id    string
		check string
	}{
		{"LH-1", "Parse LoginRequest"},
		{"LH-2", "Call authService.Login()"},
		{"LH-3", "ErrMFARequired"},
		{"LH-4", "Set session cookie"},
		{"LH-5", "Return LoginResponse"},
	}
	loginPath := filepath.Join(handlerDir, "auth_login.go")
	loginContent, _ := os.ReadFile(loginPath)
	for _, c := range loginChecks {
		if strings.Contains(string(loginContent), c.check) {
			fmt.Printf("  ✅ %s  %s\n", c.id, c.check)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", c.id, c.check)
			failed++
		}
	}

	// 6.2 MFA Verify Handler (POST /v1/auth/mfa/verify)
	fmt.Println("\n--- 6.2 MFA Verify Handler (POST /v1/auth/mfa/verify) ---")
	mfaVerifyChecks := []struct {
		id    string
		check string
	}{
		{"MV-1", "Parse MFAVerifyRequest"},
		{"MV-2", "Validate code format"},
		{"MV-3", "Call authService.VerifyMFACode()"},
		{"MV-4", "Create session"},
		{"MV-5", "Set session cookie"},
		{"MV-6", "Return MFAVerifyResponse"},
		{"MV-7", "access_token"},
		{"MV-8", "refresh_token"},
		{"MV-9", "expires_at"},
	}
	for _, c := range mfaVerifyChecks {
		if mfaVerifyExists || strings.Contains(string(mfaContent), c.check) {
			fmt.Printf("  ✅ %s  %s\n", c.id, c.check)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", c.id, c.check)
			failed++
		}
	}

	// 6.3 Refresh Token Handler (POST /v1/auth/refresh)
	fmt.Println("\n--- 6.3 Refresh Token Handler (POST /v1/auth/refresh) ---")
	refreshChecks := []struct {
		id    string
		check string
	}{
		{"RT-1", "Parse RefreshTokenRequest"},
		{"RT-2", "Validate refresh token"},
		{"RT-3", "Call authService.RotateRefreshToken()"},
		{"RT-4", "Refresh token rotation"},
		{"RT-5", "Return new tokens"},
	}
	refreshPath := filepath.Join(handlerDir, "auth_refresh.go")
	if content, err := os.ReadFile(refreshPath); err == nil {
		contentStr := string(content)
		for _, c := range refreshChecks {
			if strings.Contains(contentStr, c.check) {
				fmt.Printf("  ✅ %s  %s\n", c.id, c.check)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %s (MISSING)\n", c.id, c.check)
				failed++
			}
		}
	} else {
		for _, c := range refreshChecks {
			fmt.Printf("  ⚠️  %s  %s (FILE NOT FOUND)\n", c.id, c.check)
			failed++
		}
	}

	// 6.4 Register Handler
	fmt.Println("\n--- 6.4 Register Handler (POST /v1/auth/register) ---")
	registerChecks := []struct {
		id    string
		check string
	}{
		{"RH-1", "Parse RegisterRequest"},
		{"RH-2", "Validate email format"},
		{"RH-3", "Validate password strength"},
		{"RH-4", "Call authService.Register()"},
		{"RH-5", "Email verification sent"},
		{"RH-6", "Return RegisterResponse"},
	}
	registerPath := filepath.Join(handlerDir, "auth_register.go")
	registerContent, _ := os.ReadFile(registerPath)
	for _, c := range registerChecks {
		if strings.Contains(string(registerContent), c.check) {
			fmt.Printf("  ✅ %s  %s\n", c.id, c.check)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", c.id, c.check)
			failed++
		}
	}

	// 6.5 OAuth Handler
	fmt.Println("\n--- 6.5 OAuth Handler ---")
	oauthChecks := []struct {
		id    string
		check string
	}{
		{"OH-1", "GoogleLogin handler"},
		{"OH-2", "GoogleCallback handler"},
		{"OH-3", "GitHubLogin handler"},
		{"OH-4", "GitHubCallback handler"},
		{"OH-5", "OAuth state validation"},
		{"OH-6", "JWT token generation"},
	}
	for _, c := range oauthChecks {
		if strings.Contains(oauthContentStr, c.check) {
			fmt.Printf("  ✅ %s  %s\n", c.id, c.check)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", c.id, c.check)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: Infrastructure Requirements
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: INFRASTRUCTURE REQUIREMENTS")
	fmt.Println(strings.Repeat("─", 75))

	infraChecks := []struct {
		id          string
		description string
		checkFile   string
		check       string
	}{
		{"INF-1", "JWTManager with GenerateToken, ValidateToken", "security/jwt.go", "JWTManager"},
		{"INF-2", "SessionManager with Create, Get, Delete", "security/session/session_manager.go", "SessionManager"},
		{"INF-3", "PasswordHasher with Argon2id", "security/password.go", "Argon2"},
		{"INF-4", "GoogleVerifier for OAuth", "security/google.go", "GoogleVerifier"},
		{"INF-5", "RateLimiter middleware", "middleware/rate_limit.go", "RateLimiter"},
		{"INF-6", "Cookie authentication middleware", "middleware/cookie_auth.go", "CookieAuth"},
		{"INF-7", "Lockout middleware", "middleware/lockout.go", "Lockout"},
		{"INF-8", "Email service", "email/email_service.go", "email"},
	}

	for _, i := range infraChecks {
		var path string
		if strings.Contains(i.checkFile, "middleware") {
			path = filepath.Join(middlewareDir, strings.Replace(i.checkFile, "middleware/", "", 1))
		} else {
			path = filepath.Join(infraDir, i.checkFile)
		}
		if content, err := os.ReadFile(path); err == nil {
			if strings.Contains(string(content), i.check) {
				fmt.Printf("  ✅ %s  %s\n", i.id, i.description)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %s (MISSING in %s)\n", i.id, i.description, i.checkFile)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %s (FILE NOT FOUND: %s)\n", i.id, i.description, i.checkFile)
			failed++
		}
	}

	// =========================================================================
	// SECTION 8: Database Schema
	// =========================================================================
	fmt.Println("\n📋 SECTION 8: DATABASE SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	dbTables := []struct {
		id   string
		name string
	}{
		{"DB-T1", "operators"},
		{"DB-T2", "sessions"},
		{"DB-T3", "refresh_tokens"},
		{"DB-T4", "email_verifications"},
		{"DB-T5", "password_resets"},
	}

	schemaDir := filepath.Join(root, "supabase/migrations/")
	for _, t := range dbTables {
		found := false
		if entries, err := os.ReadDir(schemaDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					content, _ := os.ReadFile(filepath.Join(schemaDir, e.Name()))
					if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), t.name) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s table\n", t.id, t.name)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s table (MISSING)\n", t.id, t.name)
			failed++
		}
	}

	// Required Indexes
	fmt.Println("\n--- Required Indexes ---")
	indexes := []struct {
		id    string
		index string
	}{
		{"IDX-1", "idx_operators_email"},
		{"IDX-2", "idx_operators_google_id"},
		{"IDX-3", "idx_operators_github_id"},
		{"IDX-4", "idx_sessions_operator_id"},
		{"IDX-5", "idx_refresh_tokens_token_hash"},
		{"IDX-6", "idx_refresh_tokens_operator_id"},
		{"IDX-7", "idx_email_verifications_operator_id"},
		{"IDX-8", "idx_email_verifications_token_hash"},
		{"IDX-9", "idx_password_resets_operator_id"},
		{"IDX-10", "idx_password_resets_token_hash"},
	}

	for _, idx := range indexes {
		found := false
		if entries, err := os.ReadDir(schemaDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					content, _ := os.ReadFile(filepath.Join(schemaDir, e.Name()))
					if strings.Contains(string(content), idx.index) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s\n", idx.id, idx.index)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", idx.id, idx.index)
			failed++
		}
	}

	// =========================================================================
	// SECTION 9: Error Handling
	// =========================================================================
	fmt.Println("\n📋 SECTION 9: ERROR HANDLING")
	fmt.Println(strings.Repeat("─", 75))

	errorCodes := []struct {
		id   string
		code string
		http string
	}{
		{"ERR-1", "invalid_credentials", "401"},
		{"ERR-2", "mfa_required", "403"},
		{"ERR-3", "mfa_invalid", "401"},
		{"ERR-4", "token_expired", "410"},
		{"ERR-5", "token_invalid", "400"},
		{"ERR-6", "email_exists", "409"},
		{"ERR-7", "rate_limited", "429"},
		{"ERR-8", "account_locked", "423"},
	}

	errorsPath := filepath.Join(domainDir, "operator/operator_errors.go")
	if content, err := os.ReadFile(errorsPath); err == nil {
		contentStr := string(content)
		for _, e := range errorCodes {
			if strings.Contains(contentStr, e.code) || strings.Contains(contentStr, strings.ToUpper(e.code)) {
				fmt.Printf("  ✅ %s  %s (%s)\n", e.id, e.code, e.http)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %s (%s MISSING)\n", e.id, e.code, e.http)
				failed++
			}
		}
	} else {
		for _, e := range errorCodes {
			fmt.Printf("  ⚠️  %s  %s (FILE NOT FOUND)\n", e.id, e.code)
			failed++
		}
	}

	// Error response format
	fmt.Println("\n--- Error Response Format ---")
	errorFormatChecks := []struct {
		id    string
		check string
	}{
		{"EF-1", `"error":`},
		{"EF-2", `"message":`},
		{"EF-3", `"details":`},
	}
	for _, ef := range errorFormatChecks {
		found := false
		if entries, err := os.ReadDir(handlerDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					content, _ := os.ReadFile(filepath.Join(handlerDir, e.Name()))
					if strings.Contains(string(content), ef.check) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  Error format: %s\n", ef.id, ef.check)
			passed++
		} else {
			fmt.Printf("  ❌ %s  Error format: %s (MISSING)\n", ef.id, ef.check)
			failed++
		}
	}

	// =========================================================================
	// SECTION 10: Security Requirements
	// =========================================================================
	fmt.Println("\n📋 SECTION 10: SECURITY REQUIREMENTS")
	fmt.Println(strings.Repeat("─", 75))

	securityChecks := []struct {
		id          string
		description string
		checkFile   string
		check       string
	}{
		{"SEC-1", "Password hashing with Argon2id", "security/password.go", "Argon2"},
		{"SEC-2", "JWT token expiration", "security/jwt.go", "ExpiresAt"},
		{"SEC-3", "Refresh token rotation", "application/auth/auth_service.go", "RotateRefreshToken"},
		{"SEC-4", "Rate limiting on auth endpoints", "middleware/rate_limit.go", "RateLimiter"},
		{"SEC-5", "Account lockout after failed attempts", "middleware/lockout.go", "Lockout"},
		{"SEC-6", "Secure cookie settings", "middleware/cookie_auth.go", "Secure"},
	}

	for _, sec := range securityChecks {
		var path string
		if strings.Contains(sec.checkFile, "middleware") {
			path = filepath.Join(middlewareDir, strings.Replace(sec.checkFile, "middleware/", "", 1))
		} else if strings.Contains(sec.checkFile, "application") {
			path = filepath.Join(root, "apps/api/internal/", sec.checkFile)
		} else {
			path = filepath.Join(infraDir, sec.checkFile)
		}
		found := false
		if content, err := os.ReadFile(path); err == nil {
			found = strings.Contains(string(content), sec.check)
		}
		if found {
			fmt.Printf("  ✅ %s  %s\n", sec.id, sec.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", sec.id, sec.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 11: API Contract Reference
	// =========================================================================
	fmt.Println("\n📋 SECTION 11: API CONTRACT REFERENCE")
	fmt.Println(strings.Repeat("─", 75))

	apiContractChecks := []struct {
		id          string
		endpoint    string
		description string
		checkFile   string
		check       string
	}{
		{"AC-1", "POST /v1/auth/login", "LoginRequest with email and password", "auth_login.go", "LoginRequest"},
		{"AC-2", "POST /v1/auth/login", "LoginResponse with operator_id, email, name, role, mfa_enabled", "auth_login.go", "LoginResponse"},
		{"AC-3", "POST /v1/auth/mfa/verify", "MFAVerifyRequest with operator_id and code", "auth_mfa.go", "MFAVerifyRequest"},
		{"AC-4", "POST /v1/auth/mfa/verify", "MFAVerifyResponse with success, token, session", "auth_mfa.go", "MFAVerifyResponse"},
		{"AC-5", "POST /v1/auth/refresh", "RefreshTokenRequest with refresh_token", "auth_refresh.go", "RefreshTokenRequest"},
		{"AC-6", "POST /v1/auth/refresh", "RefreshTokenResponse with access_token, refresh_token, expires_at", "auth_refresh.go", "RefreshTokenResponse"},
		{"AC-7", "POST /v1/auth/forgot-password", "ForgotPasswordRequest with email", "auth_password_reset.go", "ForgotPasswordRequest"},
		{"AC-8", "POST /v1/auth/reset-password", "ResetPasswordRequest with token and new_password", "auth_password_reset.go", "ResetPasswordRequest"},
	}

	for _, ac := range apiContractChecks {
		path := filepath.Join(handlerDir, ac.checkFile)
		if content, err := os.ReadFile(path); err == nil {
			if strings.Contains(string(content), ac.check) {
				fmt.Printf("  ✅ %s  %s: %s\n", ac.id, ac.endpoint, ac.description)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %s: %s (MISSING)\n", ac.id, ac.endpoint, ac.description)
				failed++
			}
		} else {
			fmt.Printf("  ⚠️  %s  %s: %s (FILE NOT FOUND)\n", ac.id, ac.endpoint, ac.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 12: File Structure Verification
	// =========================================================================
	fmt.Println("\n📋 SECTION 12: FILE STRUCTURE VERIFICATION")
	fmt.Println(strings.Repeat("─", 75))

	fileStructureChecks := []struct {
		id   string
		file string
	}{
		// Handler files
		{"FS-H1", "api/handlers/auth/auth_login.go"},
		{"FS-H2", "api/handlers/auth/auth_register.go"},
		{"FS-H3", "api/handlers/auth/auth_logout.go"},
		{"FS-H4", "api/handlers/auth/auth_me.go"},
		{"FS-H5", "api/handlers/auth/auth_mfa.go"},
		{"FS-H6", "api/handlers/auth/auth_oauth.go"},
		{"FS-H7", "api/handlers/auth/auth_password_reset.go"},
		{"FS-H8", "api/handlers/auth/auth_email_verify.go"},
		{"FS-H9", "api/handlers/auth/auth_settings.go"},
		{"FS-H10", "api/handlers/auth/auth_admin.go"},
		{"FS-H11", "api/handlers/auth/auth_lockout.go"},
		{"FS-H12", "api/handlers/auth/auth_routes.go"},
		{"FS-H13", "api/handlers/auth/auth_refresh.go"},
		// Domain files
		{"FS-D1", "domain/operator/operator_entity.go"},
		{"FS-D2", "domain/operator/operator_repository.go"},
		{"FS-D3", "domain/operator/operator_errors.go"},
		{"FS-D4", "domain/operator/operator_password.go"},
		{"FS-D5", "domain/operator/operator_role.go"},
		{"FS-D6", "domain/session/session_entity.go"},
		{"FS-D7", "domain/session/session_repository.go"},
		{"FS-D8", "domain/refresh_token/refresh_token_entity.go"},
		{"FS-D9", "domain/refresh_token/refresh_token_repository.go"},
		{"FS-D10", "domain/email_verification/email_verification_entity.go"},
		{"FS-D11", "domain/password_reset/password_reset_entity.go"},
		{"FS-D12", "domain/oauth/oauth_entity.go"},
		// Application files
		{"FS-A1", "application/auth/auth_service.go"},
		{"FS-A2", "application/auth/auth_password.go"},
		// Infrastructure files
		{"FS-I1", "infrastructure/security/jwt.go"},
		{"FS-I2", "infrastructure/security/password.go"},
		{"FS-I3", "infrastructure/security/google.go"},
		{"FS-I4", "infrastructure/email/email_service.go"},
		// Middleware files
		{"FS-M1", "api/middleware/auth.go"},
		{"FS-M2", "api/middleware/cookie_auth.go"},
		{"FS-M3", "api/middleware/lockout.go"},
		{"FS-M4", "api/middleware/rate_limit.go"},
		{"FS-M5", "api/middleware/validation.go"},
		// Response files
		{"FS-R1", "api/responses/presenter.go"},
	}

	for _, f := range fileStructureChecks {
		path := filepath.Join(root, "apps/api/internal/", f.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  %s\n", f.id, f.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", f.id, f.file)
			failed++
		}
	}

	// =========================================================================
	// SECTION 13: Routes Registration
	// =========================================================================
	fmt.Println("\n📋 SECTION 13: ROUTES REGISTRATION")
	fmt.Println(strings.Repeat("─", 75))

	routeChecks := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"R-1", "POST", "/v1/auth/login", "LoginHandler"},
		{"R-2", "POST", "/v1/auth/register", "RegisterHandler"},
		{"R-3", "POST", "/v1/auth/logout", "LogoutHandler"},
		{"R-4", "GET", "/v1/auth/me", "MeHandler"},
		{"R-5", "PATCH", "/v1/auth/me", "MeHandler"},
		{"R-6", "POST", "/v1/auth/refresh", "RefreshHandler"},
		{"R-7", "GET", "/v1/auth/mfa/status", "MFAHandler"},
		{"R-8", "POST", "/v1/auth/mfa/enroll", "MFAHandler"},
		{"R-9", "POST", "/v1/auth/mfa/verify", "MFAHandler"},
		{"R-10", "POST", "/v1/auth/mfa/enable", "MFAHandler"},
		{"R-11", "POST", "/v1/auth/mfa/disable", "MFAHandler"},
		{"R-12", "POST", "/v1/auth/forgot-password", "PasswordResetHandler"},
		{"R-13", "POST", "/v1/auth/reset-password", "PasswordResetHandler"},
		{"R-14", "POST", "/v1/auth/resend-password-reset", "PasswordResetHandler"},
		{"R-15", "GET", "/v1/auth/google", "OAuthHandler"},
		{"R-16", "GET", "/v1/auth/google/callback", "OAuthHandler"},
		{"R-17", "GET", "/v1/auth/github", "OAuthHandler"},
		{"R-18", "GET", "/v1/auth/github/callback", "OAuthHandler"},
		{"R-19", "GET", "/v1/auth/admin/operators", "AdminHandler"},
		{"R-20", "POST", "/v1/auth/admin/operators", "AdminHandler"},
		{"R-21", "GET", "/v1/auth/admin/operators/:id", "AdminHandler"},
		{"R-22", "PATCH", "/v1/auth/admin/operators/:id", "AdminHandler"},
		{"R-23", "DELETE", "/v1/auth/admin/operators/:id", "AdminHandler"},
		{"R-24", "GET", "/v1/auth/admin/lockout/:operator_id", "AdminHandler"},
		{"R-25", "POST", "/v1/auth/admin/lockout/:operator_id/unlock", "AdminHandler"},
	}

	routesPath := filepath.Join(handlerDir, "auth_routes.go")
	if routesContent, err := os.ReadFile(routesPath); err == nil {
		routesContentStr := string(routesContent)
		for _, r := range routeChecks {
			if strings.Contains(routesContentStr, r.handler) && strings.Contains(routesContentStr, r.method) {
				fmt.Printf("  ✅ %s  %-6s %-45s -> %s\n", r.id, r.method, r.path, r.handler)
				passed++
			} else {
				fmt.Printf("  ❌ %s  %-6s %-45s -> %s (NOT REGISTERED)\n", r.id, r.method, r.path, r.handler)
				failed++
			}
		}
	} else {
		for _, r := range routeChecks {
			fmt.Printf("  ⚠️  %s  %-6s %-45s (ROUTES FILE NOT FOUND)\n", r.id, r.method, r.path)
			failed++
		}
	}

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println(strings.Repeat("═", 75))
	fmt.Printf("AUTHENTICATION SYSTEM SERVER: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
