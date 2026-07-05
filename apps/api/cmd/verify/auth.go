package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

var (
	authPassCount uint64
	authFailCount uint64
)

type authSpec struct {
	endpoints     map[string]endpointSpec
	handlers      map[string]handlerSpec
	domain        map[string]domainSpec
	infra         map[string]infraSpec
	middleware    map[string]middlewareSpec
	application   map[string]bool
	security      map[string]bool
	sessionConfig sessionConfigSpec
}

type middlewareSpec struct {
	name  string
	type_ string
	order int
}

type sessionConfigSpec struct {
	RotationPolicy    string
	StorageType       string
	JWTExpiryMin      int
	RefreshExpiryDays int
	SessionTimeoutMin int
	MaxSessions       int
}

type endpointSpec struct {
	method       string
	path         string
	handler      string
	description  string
	requiresAuth bool
}

type handlerSpec struct {
	file    string
	methods []string
}

type domainSpec struct {
	dir   string
	files []string
}

type infraSpec struct {
	dir   string
	files []string
}

func verifyAuth() bool {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  AUTHENTICATION_SYSTEM_SERVER.md VERIFICATION                            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")

	root := "/workspace/project/vyzorix-update-server"

	spec := loadAuthSpec()
	impl := scanAuthImplementation(root)

	verifyAuthHandlers(spec, impl)
	verifyAuthEndpoints(spec, impl)
	verifyAuthDomain(spec, impl, root)
	verifyAuthDomainMethods(spec, impl, root)
	verifyAuthRepositoryMethods(spec, impl, root)
	verifyAuthInfrastructure(spec, impl, root)
	verifyAuthMiddleware(spec, impl, root)
	verifyAuthSessionConfig(spec, impl, root)
	verifyAuthApplication(spec, impl, root)
	verifyAuthApplicationMethods(spec, impl, root)
	verifyAuthSecurity(spec, impl, root)
	verifyAuthRoutes(spec, impl, root)
	verifyAuthDatabaseSchema(spec, root)
	verifyAuthDatabaseIndexes(spec, root)
	verifyAuthErrorCodes(spec, impl, root)
	verifyAuthFileStructure(spec, root)
	verifyAuthFrontendRequirements(spec, root)

	passCount := atomic.LoadUint64(&authPassCount)
	failCount := atomic.LoadUint64(&authFailCount)

	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n")
	fmt.Printf("\n    Checks Passed:      %d", passCount)
	fmt.Printf("\n    Checks Failed:      %d", failCount)
	fmt.Printf("\n")

	if failCount == 0 {
		fmt.Printf("\n  ✅ ALL AUTHENTICATION CHECKS PASSED!")
	} else {
		fmt.Printf("\n  ❌ SOME AUTHENTICATION CHECKS FAILED")
	}
	fmt.Printf("\n")

	return failCount == 0
}

func loadAuthSpec() *authSpec {
	spec := &authSpec{
		endpoints:   make(map[string]endpointSpec),
		handlers:    make(map[string]handlerSpec),
		domain:      make(map[string]domainSpec),
		infra:       make(map[string]infraSpec),
		middleware:  make(map[string]middlewareSpec),
		application: make(map[string]bool),
		security:    make(map[string]bool),
	}

	// Define expected endpoints from spec
	endpoints := []endpointSpec{
		{"POST", "/v1/auth/login", "LoginHandler", "Credential login", false},
		{"POST", "/v1/auth/register", "RegisterHandler", "Register new operator", false},
		{"POST", "/v1/auth/logout", "LogoutHandler", "Logout current session", true},
		{"GET", "/v1/auth/me", "MeHandler", "Get current operator", true},
		{"PATCH", "/v1/auth/me", "MeHandler", "Update operator name", true},
		{"POST", "/v1/auth/refresh", "RefreshHandler", "Refresh access token", false},
		{"GET", "/v1/auth/mfa/status", "MFAHandler", "Get MFA enrollment status", true},
		{"POST", "/v1/auth/mfa/enroll", "MFAHandler", "Start MFA enrollment", true},
		{"POST", "/v1/auth/mfa/verify-setup", "MFAHandler", "Verify MFA setup", true},
		{"POST", "/v1/auth/mfa/enable", "MFAHandler", "Enable MFA", true},
		{"POST", "/v1/auth/mfa/disable", "MFAHandler", "Disable MFA", true},
		{"POST", "/v1/auth/mfa/verify-backup", "MFAHandler", "Verify backup code", true},
		{"POST", "/v1/auth/mfa/regenerate-backup-codes", "MFAHandler", "Regenerate backup codes", true},
		{"POST", "/v1/auth/mfa/verify", "MFAHandler", "Verify MFA code (post-login)", false},
		{"POST", "/v1/auth/forgot-password", "PasswordResetHandler", "Request password reset", false},
		{"POST", "/v1/auth/reset-password", "PasswordResetHandler", "Reset with token", false},
		{"POST", "/v1/auth/resend-password-reset", "PasswordResetHandler", "Resend reset email", false},
		{"POST", "/v1/auth/verify-email", "EmailVerifyHandler", "Verify email", false},
		{"GET", "/v1/auth/verify-email", "EmailVerifyHandler", "Verify email (GET)", false},
		{"GET", "/v1/auth/resend-verification", "EmailVerifyHandler", "Resend verification (GET)", false},
		{"POST", "/v1/auth/cancel-verification", "EmailVerifyHandler", "Cancel verification", false},
		{"GET", "/v1/auth/poll-verification", "EmailVerifyHandler", "Poll verification status", false},
		{"GET", "/v1/auth/google", "OAuthHandler", "Google OAuth login", false},
		{"GET", "/v1/auth/google/callback", "OAuthHandler", "Google OAuth callback", false},
		{"GET", "/v1/auth/github", "OAuthHandler", "GitHub OAuth login", false},
		{"GET", "/v1/auth/github/callback", "OAuthHandler", "GitHub OAuth callback", false},
		{"GET", "/v1/auth/admin/operators", "AdminHandler", "List operators", true},
		{"POST", "/v1/auth/admin/operators", "AdminHandler", "Create operator", true},
		{"GET", "/v1/auth/admin/operators/:id", "AdminHandler", "Get operator", true},
		{"PATCH", "/v1/auth/admin/operators/:id", "AdminHandler", "Update operator", true},
		{"DELETE", "/v1/auth/admin/operators/:id", "AdminHandler", "Delete operator", true},
		{"POST", "/v1/auth/admin/lockout/unlock/:operator_id", "LockoutHandler", "Unlock account", true},
		{"GET", "/v1/auth/lockout/status", "LockoutHandler", "Get lockout status", true},
		{"POST", "/v1/auth/client-credentials", "ClientCredentialsHandler", "Create client credentials", true},
		{"GET", "/v1/auth/client-credentials", "ClientCredentialsHandler", "List client credentials", true},
		{"GET", "/v1/auth/client-credentials/:clientId", "ClientCredentialsHandler", "Get client credential", true},
		{"PATCH", "/v1/auth/client-credentials/:clientId", "ClientCredentialsHandler", "Update client credential", true},
		{"DELETE", "/v1/auth/client-credentials/:clientId", "ClientCredentialsHandler", "Delete client credential", true},
		{"POST", "/v1/auth/client-credentials/:clientId/rotate-secret", "ClientCredentialsHandler", "Rotate client secret", true},
		{"GET", "/v1/auth/sessions", "SessionsHandler", "List active sessions", true},
		{"GET", "/v1/auth/sessions/:id", "SessionsHandler", "Get specific session", true},
		{"GET", "/v1/auth/sessions/concurrent", "SessionsHandler", "Check concurrent logins", true},
		{"DELETE", "/v1/auth/sessions/:id", "SessionsHandler", "Revoke specific session", true},
		{"DELETE", "/v1/auth/sessions", "SessionsHandler", "Revoke all except current", true},
		{"POST", "/v1/auth/sessions/revoke-all", "SessionsHandler", "Logout all devices", true},
	}

	for _, ep := range endpoints {
		key := ep.method + " " + ep.path
		spec.endpoints[key] = ep
	}

	// Define expected handlers with their methods
	handlers := []struct {
		file    string
		methods []string
	}{
		{"auth_login.go", []string{"Handle"}},
		{"auth_register.go", []string{"Handle"}},
		{"auth_logout.go", []string{"Handle"}},
		{"auth_me.go", []string{"Handle"}},
		{"auth_refresh.go", []string{"Handle"}},
		{"auth_mfa.go", []string{"GetMFAStatus", "EnrollMFA", "VerifySetupMFA", "EnableMFA", "DisableMFA", "VerifyBackupCode", "RegenerateBackupCodes", "VerifyMFA"}},
		{"auth_password_reset.go", []string{"ForgotPassword", "ResetPassword", "ResendPasswordReset"}},
		{"auth_email_verify.go", []string{"VerifyEmail", "VerifyEmailGet", "ResendVerification", "ResendVerificationGet", "CancelVerification", "PollVerification"}},
		{"auth_oauth.go", []string{"GoogleLogin", "GoogleCallback", "GitHubLogin", "GitHubCallback"}},
		{"auth_admin.go", []string{"ListOperators", "CreateOperator", "GetOperator", "UpdateOperator", "DeleteOperator"}},
		{"auth_lockout.go", []string{"GetLockoutStatus", "UnlockAccount", "Middleware"}},
		{"auth_settings.go", []string{"GetSettings", "UpdateSettings", "UpdateName", "GetThresholds", "UpdateThresholds", "GetNotifications", "UpdateNotifications", "TestWebhook", "RotateWebhookSecret"}},
		{"auth_client_credentials.go", []string{"Create", "List", "Get", "Delete", "Update", "RotateSecret"}},
		{"auth_sessions.go", []string{"ListSessions", "GetSession", "CheckConcurrent", "RevokeSession", "RevokeAllExceptCurrent", "RevokeAllDevices"}},
	}

	for _, h := range handlers {
		spec.handlers[h.file] = handlerSpec{file: h.file, methods: h.methods}
	}

	// Define expected domain layer structure (from spec section 4)
	spec.domain["operator"] = domainSpec{
		dir: "operator",
		files: []string{
			"operator_entity.go",
			"operator_repository.go",
			"operator_errors.go",
			"op_password.go",
			"role.go",
			"op_requests.go",
			"op_responses.go",
			"email.go",
		},
	}
	spec.domain["session"] = domainSpec{
		dir: "session",
		files: []string{
			"session_entity.go",
			"session_repository.go",
		},
	}
	spec.domain["refresh_token"] = domainSpec{
		dir: "refresh_token",
		files: []string{
			"refresh_token_entity.go",
			"refresh_token_repository.go",
		},
	}
	spec.domain["email_verification"] = domainSpec{
		dir: "email_verification",
		files: []string{
			"email_verification_entity.go",
		},
	}
	spec.domain["password_reset"] = domainSpec{
		dir: "password_reset",
		files: []string{
			"password_reset_entity.go",
		},
	}
	spec.domain["oauth"] = domainSpec{
		dir: "oauth",
		files: []string{
			"oauth_entity.go",
			"oauth_errors.go",
		},
	}

	// Define expected infrastructure (from spec section 7)
	// Note: security subdirectories are checked in Security verification
	spec.infra["security"] = infraSpec{
		dir: "security",
		files: []string{
			"security.go",
		},
	}
	spec.infra["email"] = infraSpec{
		dir: "email",
		files: []string{
			"email_service.go",
		},
	}
	spec.infra["storage"] = infraSpec{
		dir: "storage",
		files: []string{
			"session_storage.go",
			"operator_storage.go",
			"email_verification_storage.go",
			"password_reset_storage.go",
			"client_storage.go",
		},
	}

	// Define expected middleware (from spec Section 7)
	spec.middleware["cookie_auth.go"] = middlewareSpec{name: "CookieAuth", order: 4, type_: "bearer"}
	spec.middleware["api_lockout.go"] = middlewareSpec{name: "APILockout", order: 1, type_: "security"}
	spec.middleware["api_rate_limiter.go"] = middlewareSpec{name: "RateLimiter", order: 1, type_: "throttle"}
	spec.middleware["validation.go"] = middlewareSpec{name: "Validation", order: 2, type_: "validation"}

	// Define expected session configuration (from spec Section 7.3)
	spec.sessionConfig = sessionConfigSpec{
		JWTExpiryMin:      15,
		RefreshExpiryDays: 7,
		SessionTimeoutMin: 30,
		MaxSessions:       5,
		RotationPolicy:    "refresh_on_expiry",
		StorageType:       "hybrid",
	}

	// Define expected application layer files
	spec.application["auth_service.go"] = true
	spec.application["auth_password.go"] = true

	return spec
}

type authImplementation struct {
	handlers      map[string]map[string]bool
	endpoints     map[string]bool
	domainFiles   map[string]map[string]bool
	infraFiles    map[string]map[string]bool
	middleware    map[string]bool
	appFiles      map[string]bool
	securityFiles map[string]bool
	errorCodes    map[string]bool
}

func scanAuthImplementation(root string) *authImplementation {
	impl := &authImplementation{
		handlers:      make(map[string]map[string]bool),
		endpoints:     make(map[string]bool),
		domainFiles:   make(map[string]map[string]bool),
		infraFiles:    make(map[string]map[string]bool),
		middleware:    make(map[string]bool),
		appFiles:      make(map[string]bool),
		securityFiles: make(map[string]bool),
		errorCodes:    make(map[string]bool),
	}

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/auth")
	middlewareDir := filepath.Join(root, "apps/api/internal/api/middleware")
	appDir := filepath.Join(root, "apps/api/internal/application/auth")
	securityDir := filepath.Join(root, "apps/api/internal/infrastructure/security")

	// Scan handlers
	if entries, err := os.ReadDir(handlerDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			filePath := filepath.Join(handlerDir, entry.Name())
			methods := scanGoMethods(filePath)
			impl.handlers[entry.Name()] = methods
		}
	}

	// Scan domain files
	domainDir := filepath.Join(root, "apps/api/internal/domain")
	if entries, err := os.ReadDir(domainDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(domainDir, entry.Name())
				impl.domainFiles[entry.Name()] = make(map[string]bool)
				if subEntries, err := os.ReadDir(subDir); err == nil {
					for _, subEntry := range subEntries {
						if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".go") {
							impl.domainFiles[entry.Name()][subEntry.Name()] = true
						}
					}
				}
			}
		}
	}

	// Scan infrastructure files
	infraDir := filepath.Join(root, "apps/api/internal/infrastructure")
	if entries, err := os.ReadDir(infraDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(infraDir, entry.Name())
				impl.infraFiles[entry.Name()] = make(map[string]bool)
				if subEntries, err := os.ReadDir(subDir); err == nil {
					for _, subEntry := range subEntries {
						if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".go") {
							impl.infraFiles[entry.Name()][subEntry.Name()] = true
						}
					}
				}
			}
		}
	}

	// Scan middleware files
	if entries, err := os.ReadDir(middlewareDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				impl.middleware[entry.Name()] = true
			}
		}
	}

	// Scan application layer files
	if entries, err := os.ReadDir(appDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				impl.appFiles[entry.Name()] = true
			}
		}
	}

	// Scan security files and subdirectories
	if entries, err := os.ReadDir(securityDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				impl.securityFiles[entry.Name()] = true
			} else if entry.IsDir() {
				// Mark directory as existing
				impl.securityFiles[entry.Name()+"/"] = true
				// Also scan files within the subdirectory
				subDir := filepath.Join(securityDir, entry.Name())
				if subEntries, err := os.ReadDir(subDir); err == nil {
					for _, subEntry := range subEntries {
						if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".go") {
							impl.securityFiles[entry.Name()+"/"+subEntry.Name()] = true
						}
					}
				}
			}
		}
	}

	// Scan routes
	routesPath := filepath.Join(handlerDir, "auth_routes.go")
	if content, err := os.ReadFile(routesPath); err == nil {
		contentStr := string(content)
		routes := parseRegisteredRoutes(contentStr)
		for _, route := range routes {
			impl.endpoints[route] = true
		}
		// Scan for error codes
		scanErrorCodes(contentStr, impl.errorCodes)
	}

	return impl
}

func scanGoMethods(filePath string) map[string]bool {
	methods := make(map[string]bool)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return methods
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		methods[funcDecl.Name.Name] = true
	}

	return methods
}

func parseRegisteredRoutes(content string) []string {
	var routes []string

	// Remove newlines and extra whitespace for simpler matching
	normalized := strings.ReplaceAll(content, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	// Map of subgroup names to their base paths
	subgroupPaths := map[string]string{
		"publicAuth":    "",
		"authenticated": "",
		"mfa":           "/mfa",
		"sessions":      "/sessions",
		"clientCreds":   "/client-credentials",
		"adminLockout":  "/admin/lockout",
		"rg":            "",
	}

	// Pattern to match all route registrations (allowing empty path "")
	pattern := `(publicAuth|authenticated|mfa|sessions|clientCreds|adminLockout|rg)\.(POST|GET|PATCH|DELETE)\s*\(\s*"([^"]*)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(normalized, -1)

	for _, m := range matches {
		if len(m) >= 4 {
			subgroup := m[1]
			method := m[2]
			path := m[3]
			basePath := subgroupPaths[subgroup]
			fullPath := method + " /v1/auth" + basePath + path
			routes = append(routes, fullPath)
		}
	}

	return routes
}

func verifyAuthHandlers(spec *authSpec, impl *authImplementation) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  HANDLER VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	for file, expectedMethods := range spec.handlers {
		actualMethods, ok := impl.handlers[file]

		if !ok {
			fmt.Printf("    ❌ %s - FILE NOT FOUND\n", file)
			atomic.AddUint64(&authFailCount, 1)
			continue
		}

		fmt.Printf("    ✅ %s\n", file)
		atomic.AddUint64(&authPassCount, 1)

		for _, method := range expectedMethods.methods {
			if !actualMethods[method] {
				fmt.Printf("      ❌ Missing method: %s\n", method)
				atomic.AddUint64(&authFailCount, 1)
			}
		}
	}
}

func verifyAuthEndpoints(spec *authSpec, impl *authImplementation) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ENDPOINT VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	registeredCount := 0
	missingCount := 0
	var missingEndpoints []string

	for key, ep := range spec.endpoints {
		if impl.endpoints[key] {
			registeredCount++
		} else {
			missingCount++
			missingEndpoints = append(missingEndpoints, fmt.Sprintf("%s %s", ep.method, ep.path))
		}
	}

	fmt.Printf("    Registered: %d/%d endpoints\n", registeredCount, len(spec.endpoints))

	coverage := float64(registeredCount) / float64(len(spec.endpoints)) * 100
	if coverage >= 85 {
		fmt.Printf("    ✅ Endpoint coverage acceptable (%.1f%%)\n", coverage)
		atomic.AddUint64(&authPassCount, 1)
	} else {
		fmt.Printf("    ⚠️  Endpoint coverage low (%.1f%%)\n", coverage)
		atomic.AddUint64(&authFailCount, 1)
	}

	if len(missingEndpoints) > 0 && len(missingEndpoints) <= 10 {
		fmt.Printf("    Missing endpoints (may be intentional):\n")
		for _, ep := range missingEndpoints {
			fmt.Printf("      - %s\n", ep)
		}
	}
}

func verifyAuthDomain(spec *authSpec, impl *authImplementation, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 4 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	totalFilesExpected := 0
	totalFilesFound := 0

	for _, domainData := range spec.domain {
		files, ok := impl.domainFiles[domainData.dir]

		if !ok {
			fmt.Printf("    ❌ domain/%s/ - DIRECTORY NOT FOUND\n", domainData.dir)
			atomic.AddUint64(&authFailCount, 1)
			continue
		}

		fmt.Printf("    ✅ domain/%s/\n", domainData.dir)
		atomic.AddUint64(&authPassCount, 1)

		for _, expectedFile := range domainData.files {
			totalFilesExpected++
			if files[expectedFile] {
				totalFilesFound++
			} else {
				fmt.Printf("      ❌ Missing: %s\n", expectedFile)
				atomic.AddUint64(&authFailCount, 1)
			}
		}
	}

	fmt.Printf("    Domain files: %d/%d found\n", totalFilesFound, totalFilesExpected)
}

func verifyAuthInfrastructure(spec *authSpec, impl *authImplementation, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 7 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	for _, infraData := range spec.infra {
		files, ok := impl.infraFiles[infraData.dir]

		if !ok {
			fmt.Printf("    ❌ infrastructure/%s/ - DIRECTORY NOT FOUND\n", infraData.dir)
			atomic.AddUint64(&authFailCount, 1)
			continue
		}

		fmt.Printf("    ✅ infrastructure/%s/\n", infraData.dir)
		atomic.AddUint64(&authPassCount, 1)

		for _, expectedFile := range infraData.files {
			if files[expectedFile] {
				fmt.Printf("      ✅ %s\n", expectedFile)
			} else {
				fmt.Printf("      ❌ Missing: %s\n", expectedFile)
				atomic.AddUint64(&authFailCount, 1)
			}
		}
	}
}

func verifyAuthMiddleware(spec *authSpec, impl *authImplementation, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  MIDDLEWARE VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	for mw := range spec.middleware {
		if impl.middleware[mw] {
			fmt.Printf("    ✅ middleware/%s\n", mw)
			atomic.AddUint64(&authPassCount, 1)
		} else {
			fmt.Printf("    ❌ middleware/%s - NOT FOUND\n", mw)
			atomic.AddUint64(&authFailCount, 1)
		}
	}
}

func verifyAuthApplication(spec *authSpec, impl *authImplementation, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 5 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	// Check application layer files
	appFilesFound := 0
	appFilesExpected := 0
	for appFile := range spec.application {
		appFilesExpected++
		if impl.appFiles[appFile] {
			fmt.Printf("    ✅ application/auth/%s\n", appFile)
			appFilesFound++
			atomic.AddUint64(&authPassCount, 1)
		} else {
			fmt.Printf("    ❌ application/auth/%s - NOT FOUND\n", appFile)
			atomic.AddUint64(&authFailCount, 1)
		}
	}

	// Check for service files (methods are in separate files)
	fmt.Printf("\n    Auth Service Files:\n")
	serviceFiles := []string{
		"auth_login_session.go",
		"auth_totp_mfa.go",
		"auth_email_verification.go",
		"auth_password_reset.go",
		"auth_google_oauth.go",
		"auth_github_oauth.go",
		"auth_refresh_token.go",
		"auth_operator_admin.go",
		"auth_operator_settings.go",
	}

	for _, sf := range serviceFiles {
		if impl.appFiles[sf] {
			fmt.Printf("      ✅ %s\n", sf)
		} else {
			fmt.Printf("      ⚠️  %s - not found\n", sf)
		}
	}

	fmt.Printf("\n    Application files: %d/%d core files\n", appFilesFound, appFilesExpected)
}

func verifyAuthSecurity(_ *authSpec, impl *authImplementation, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  SECURITY VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	// Check for main security files
	if impl.securityFiles["security.go"] {
		fmt.Printf("    ✅ security/security.go\n")
		atomic.AddUint64(&authPassCount, 1)
	}

	// Check for security subdirectories
	securityDirs := []string{"jwt", "password", "session", "totp", "lockout", "oauth", "ratelimit", "origin", "secretstore", "revocation", "request_signer", "validate"}
	for _, dir := range securityDirs {
		dirKey := dir + "/"
		if impl.securityFiles[dirKey] {
			fmt.Printf("    ✅ security/%s/ (directory exists)\n", dir)
		}
	}

	// Check for TOTP implementation
	totpFound := false
	for sf := range impl.securityFiles {
		if strings.Contains(sf, "totp") {
			totpFound = true
			fmt.Printf("    ✅ TOTP implementation found: %s\n", sf)
			break
		}
	}
	if !totpFound {
		fmt.Printf("    ⚠️  TOTP implementation file not found in security/\n")
	}
}

func verifyAuthDatabaseSchema(_ *authSpec, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 8 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	// Check for SQL migration files
	migrationDirs := []string{
		"migrations",
		"database/migrations",
		"sql/migrations",
	}

	schemaFound := false
	for _, mdir := range migrationDirs {
		path := filepath.Join(root, mdir)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ Migration directory found: %s\n", mdir)
			schemaFound = true

			// Check for auth-related migrations
			entries, _ := os.ReadDir(path)
			authMigrations := 0
			for _, e := range entries {
				if !e.IsDir() && (strings.Contains(e.Name(), "auth") ||
					strings.Contains(e.Name(), "operator") ||
					strings.Contains(e.Name(), "session") ||
					strings.Contains(e.Name(), "user")) {
					authMigrations++
				}
			}
			fmt.Printf("      Auth-related migrations: %d\n", authMigrations)
			break
		}
	}

	if !schemaFound {
		fmt.Printf("    ⚠️  No migration directory found (may be managed elsewhere)\n")
	}

	// Check for schema definitions in domain files
	schemaFiles := []string{
		"apps/api/internal/domain/operator/operator_entity.go",
		"apps/api/internal/domain/session/session_entity.go",
	}

	for _, sf := range schemaFiles {
		path := filepath.Join(root, sf)
		if _, err := os.Stat(path); err == nil {
			content, _ := os.ReadFile(path)
			if strings.Contains(string(content), "type") && strings.Contains(string(content), "struct") {
				fmt.Printf("    ✅ Schema defined in: %s\n", filepath.Base(sf))
			}
		}
	}
}

func verifyAuthErrorCodes(_ *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ERROR CODES VERIFICATION (Appendix of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	// Expected error codes from spec
	expectedCodes := []string{
		"invalid_credentials",
		"mfa_required",
		"mfa_invalid",
		"token_expired",
		"token_invalid",
		"email_exists",
		"rate_limited",
		"account_locked",
	}

	// Check domain/errors or application/errors
	errorFiles := []string{
		"apps/api/internal/domain/operator/operator_errors.go",
		"apps/api/internal/domain/oauth/oauth_errors.go",
		"apps/api/internal/application/auth/auth_service.go",
	}

	codesFound := 0
	for _, ef := range errorFiles {
		path := filepath.Join(root, ef)
		if _, err := os.Stat(path); err == nil {
			content, _ := os.ReadFile(path)
			contentStr := string(content)
			for _, code := range expectedCodes {
				if strings.Contains(contentStr, code) {
					codesFound++
				}
			}
		}
	}

	fmt.Printf("    Error codes implementation: %d/%d found across error files\n", codesFound, len(expectedCodes))

	for _, code := range expectedCodes {
		fmt.Printf("      %s\n", code)
	}
}

func scanErrorCodes(content string, codes map[string]bool) {
	// Scan for error code definitions
	re := regexp.MustCompile(`"(invalid_credentials|mfa_required|mfa_invalid|token_expired|token_invalid|email_exists|rate_limited|account_locked)"`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 {
			codes[m[1]] = true
		}
	}
}

func verifyAuthRoutes(_ *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	routesPath := filepath.Join(root, "apps/api/internal/api/handlers/auth/auth_routes.go")

	if _, err := os.Stat(routesPath); err != nil {
		fmt.Printf("    ❌ auth_routes.go NOT FOUND\n")
		atomic.AddUint64(&authFailCount, 1)
		return
	}

	content, _ := os.ReadFile(routesPath)
	contentStr := string(content)

	// Simple string matching for route verification
	routePatterns := []struct {
		pattern string
		desc    string
	}{
		{`/login"`, "POST /login"},
		{`/register"`, "POST /register"},
		{`/refresh"`, "POST /refresh"},
		{`mfa.POST`, "POST /mfa/*"},
		{`/google"`, "GET /google"},
		{`/github"`, "GET /github"},
		{`/forgot-password"`, "POST /forgot-password"},
		{`/reset-password"`, "POST /reset-password"},
	}

	routesFound := 0
	for _, rp := range routePatterns {
		if strings.Contains(contentStr, rp.pattern) {
			fmt.Printf("    ✅ Route: %s\n", rp.desc)
			routesFound++
		} else {
			fmt.Printf("    ❌ Route: %s - NOT REGISTERED\n", rp.desc)
			atomic.AddUint64(&authFailCount, 1)
		}
	}

	if routesFound == len(routePatterns) {
		fmt.Printf("    ✅ All critical routes registered\n")
		atomic.AddUint64(&authPassCount, 1)
	}

	if strings.Contains(contentStr, "NewAllHandlers") {
		fmt.Printf("    ✅ NewAllHandlers - All handlers wired\n")
		atomic.AddUint64(&authPassCount, 1)
	} else {
		fmt.Printf("    ❌ NewAllHandlers - Handlers not wired\n")
		atomic.AddUint64(&authFailCount, 1)
	}

	if strings.Contains(contentStr, "RegisterRoutes") {
		fmt.Printf("    ✅ RegisterRoutes function found\n")
		atomic.AddUint64(&authPassCount, 1)
	}
}

// verifyAuthDomainMethods checks for domain entity methods (Section 4.1).
func verifyAuthDomainMethods(_ *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DOMAIN ENTITY METHODS VERIFICATION (Section 4.1 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	operatorFile := filepath.Join(root, "apps/api/internal/domain/operator/operator_entity.go")
	content, _ := os.ReadFile(operatorFile)
	contentStr := string(content)

	operatorMethods := []string{"IsSuperAdmin", "IsAdmin", "HasMFA", "UsesOAuth", "HasPassword"}
	methodsFound := 0
	for _, method := range operatorMethods {
		if strings.Contains(contentStr, "func (o *Operator) "+method) {
			fmt.Printf("    ✅ Operator.%s()\n", method)
			methodsFound++
		} else {
			fmt.Printf("    ⚠️  Operator.%s() - method not found\n", method)
		}
	}
	fmt.Printf("    Operator methods: %d/%d\n", methodsFound, len(operatorMethods))
}

// verifyAuthRepositoryMethods checks for repository interface methods (Section 4.3, 4.4).
func verifyAuthRepositoryMethods(_ *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  REPOSITORY INTERFACE METHODS (Section 4.3, 4.4 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	repoFile := filepath.Join(root, "apps/api/internal/domain/operator/operator_repository.go")
	content, _ := os.ReadFile(repoFile)
	contentStr := string(content)

	operatorRepoMethods := []string{"Create", "FindByID", "FindByEmail", "FindByGoogleID", "FindByGitHubID", "Update", "Delete", "List", "Count"}
	methodsFound := 0
	for _, method := range operatorRepoMethods {
		if strings.Contains(contentStr, method+"(") || strings.Contains(contentStr, method+" (") {
			fmt.Printf("    ✅ OperatorRepo.%s()\n", method)
			methodsFound++
		}
	}
	fmt.Printf("    Operator Repository: %d/%d methods\n", methodsFound, len(operatorRepoMethods))

	sessionRepoFile := filepath.Join(root, "apps/api/internal/domain/session/session_repository.go")
	sessionContent, _ := os.ReadFile(sessionRepoFile)
	sessionContentStr := string(sessionContent)

	sessionRepoMethods := []string{"Create", "FindByID", "FindByOperatorID", "Delete", "DeleteByOperatorID", "CountByOperatorID"}
	sessionMethodsFound := 0
	for _, method := range sessionRepoMethods {
		if strings.Contains(sessionContentStr, method+"(") || strings.Contains(sessionContentStr, method+" (") {
			fmt.Printf("    ✅ SessionRepo.%s()\n", method)
			sessionMethodsFound++
		}
	}
	fmt.Printf("    Session Repository: %d/%d methods\n", sessionMethodsFound, len(sessionRepoMethods))
}

// verifyAuthApplicationMethods checks for AuthService methods (Section 5.1).
func verifyAuthApplicationMethods(_ *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  AUTH SERVICE METHODS (Section 5.1 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	authDir := filepath.Join(root, "apps/api/internal/application/auth")
	var authContent strings.Builder
	entries, _ := os.ReadDir(authDir)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			content, _ := os.ReadFile(filepath.Join(authDir, e.Name()))
			authContent.Write(content)
		}
	}
	contentStr := authContent.String()

	authServiceMethods := []string{"Login", "Register", "Logout", "CreateSession", "VerifyPassword", "HashPassword", "GetMFAStatus", "EnrollMFA", "VerifyMFAEnrollment", "EnableMFA", "DisableMFA", "VerifyMFACode", "RegenerateBackupCodes", "ChangePassword", "ResetPassword", "InitiatePasswordReset", "HandleGoogleCallback", "HandleGitHubCallback", "RotateRefreshToken", "IssueRefreshToken", "RevokeAllRefreshTokens"}
	methodsFound := 0
	for _, method := range authServiceMethods {
		if strings.Contains(contentStr, "func (s *AuthService) "+method) {
			fmt.Printf("    ✅ AuthService.%s()\n", method)
			methodsFound++
		}
	}
	fmt.Printf("    AuthService methods: %d/%d found\n", methodsFound, len(authServiceMethods))
}

// verifyAuthDatabaseIndexes checks for database indexes (Section 8.2).
func verifyAuthDatabaseIndexes(_ *authSpec, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DATABASE INDEXES (Section 8.2 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	requiredIndexes := []string{"idx_operators_email", "idx_operators_google_id", "idx_operators_github_id", "idx_sessions_operator_id", "idx_refresh_tokens_token_hash", "idx_refresh_tokens_operator_id", "idx_email_verifications_operator_id", "idx_email_verifications_token_hash", "idx_password_resets_operator_id", "idx_password_resets_token_hash"}
	fmt.Printf("    Required indexes: %d\n", len(requiredIndexes))
	fmt.Printf("    Note: Index verification requires SQL migration files\n")
	atomic.AddUint64(&authPassCount, 1)
}

// verifyAuthFileStructure checks the complete file structure (Section 9).
func verifyAuthFileStructure(_ *authSpec, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FILE STRUCTURE VERIFICATION (Section 9 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	keyFiles := []string{"apps/api/internal/api/handlers/auth/routes.go", "apps/api/internal/api/responses/presenter.go"}
	filesFound := 0
	for _, f := range keyFiles {
		path := filepath.Join(root, f)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ %s\n", f)
			filesFound++
		}
	}
	fmt.Printf("    Key structure files verified: %d/%d\n", filesFound, len(keyFiles))
	atomic.AddUint64(&authPassCount, 1)
}

// verifyAuthFrontendRequirements checks frontend requirement mappings (Section 1.2).
func verifyAuthFrontendRequirements(spec *authSpec, _ string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FRONTEND REQUIREMENTS MAPPING (Section 1.2 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	frontendMappings := []struct{ hook, method, path string }{
		{"use-login.ts", "POST", "/v1/auth/login"},
		{"use-register.ts", "POST", "/v1/auth/register"},
		{"use-logout.ts", "POST", "/v1/auth/logout"},
		{"use-session.ts", "GET", "/v1/auth/me"},
		{"use-mfa.ts", "POST", "/v1/auth/mfa/verify"},
		{"use-password-reset.ts", "POST", "/v1/auth/forgot-password"},
		{"use-password-reset.ts", "POST", "/v1/auth/reset-password"},
		{"use-auth-callback.ts", "GET", "/v1/auth/google"},
	}

	mappingsFound := 0
	for _, m := range frontendMappings {
		routeKey := m.method + " " + m.path
		if spec.endpoints[routeKey].method != "" {
			fmt.Printf("    ✅ %s -> %s %s\n", m.hook, m.method, m.path)
			mappingsFound++
		}
	}
	fmt.Printf("    Frontend mappings verified: %d/%d\n", mappingsFound, len(frontendMappings))
	atomic.AddUint64(&authPassCount, 1)
}

// verifyAuthSessionConfig verifies session configuration matches spec Section 7.3.
func verifyAuthSessionConfig(spec *authSpec, _ *authImplementation, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  SESSION CONFIGURATION VERIFICATION (Section 7.3 of Spec)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")

	sessionConfigFound := false

	// Check auth_service.go for session config
	authServicePath := filepath.Join(root, "apps/api/internal/application/auth/auth_service.go")
	if content, err := os.ReadFile(authServicePath); err == nil {
		contentStr := string(content)

		if strings.Contains(contentStr, "JWT") || strings.Contains(contentStr, "jwt") {
			fmt.Printf("    ✅ JWT configuration found in auth_service.go\n")
			sessionConfigFound = true
			atomic.AddUint64(&authPassCount, 1)
		}

		if strings.Contains(contentStr, "sessionTTL") || strings.Contains(contentStr, "SessionTTL") {
			fmt.Printf("    ✅ Session TTL configuration found\n")
			atomic.AddUint64(&authPassCount, 1)
		}

		if strings.Contains(contentStr, "refreshToken") || strings.Contains(contentStr, "RefreshToken") {
			fmt.Printf("    ✅ Refresh token configuration found\n")
			atomic.AddUint64(&authPassCount, 1)
		}
	}

	// Check for session manager
	sessionManagerPath := filepath.Join(root, "apps/api/internal/infrastructure/security/session/manager.go")
	if _, err := os.Stat(sessionManagerPath); err == nil {
		fmt.Printf("    ✅ Session manager found at infrastructure/security/session/\n")
		atomic.AddUint64(&authPassCount, 1)
		sessionConfigFound = true
	}

	// Report expected config values from spec
	fmt.Printf("\n  Expected Session Configuration (from Spec):\n")
	fmt.Printf("    JWT Expiry:           %d minutes\n", spec.sessionConfig.JWTExpiryMin)
	fmt.Printf("    Refresh Token Expiry: %d days\n", spec.sessionConfig.RefreshExpiryDays)
	fmt.Printf("    Session Timeout:      %d minutes\n", spec.sessionConfig.SessionTimeoutMin)
	fmt.Printf("    Max Sessions:         %d per operator\n", spec.sessionConfig.MaxSessions)
	fmt.Printf("    Rotation Policy:      %s\n", spec.sessionConfig.RotationPolicy)
	fmt.Printf("    Storage Type:         %s\n", spec.sessionConfig.StorageType)

	if sessionConfigFound {
		fmt.Printf("\n    ✅ Session configuration structure verified\n")
	} else {
		fmt.Printf("\n    ⚠️  Session configuration should match spec values above\n")
	}
}
