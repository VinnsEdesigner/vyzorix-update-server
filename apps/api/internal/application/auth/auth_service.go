// Package auth provides authentication and authorization services for the application.
// Functions have been split into multiple files:.
// - auth_constructors.go: AuthService struct, constructors, and password policy.
// - auth_helpers.go: Helper functions (hashTokenSha256, hashEmailForTracker).
// - auth_login_session.go: Login, Register, Session management.
// - auth_totp_mfa.go: TOTP MFA operations.
// - auth_email_verification.go: Email verification flows.
// - auth_password_reset.go: Password reset flows.
// - auth_google_oauth.go: Google OAuth.
// - auth_github_oauth.go: GitHub OAuth.
// - auth_operator_settings.go: Operator settings.
// - auth_operator_admin.go: Admin operator management.
// - auth_refresh_token.go: Refresh token rotation.
package auth
