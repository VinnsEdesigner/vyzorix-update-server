package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// buildLoginResponse builds a LoginResponse with organization info.
func (s *AuthService) buildLoginResponse(op *operator.Operator) *dto.LoginResponse {
	// Get operator's organizations.
	orgs, err := s.GetOperatorOrganizations(context.Background(), op)
	if err != nil {
		// If we can't get organizations, indicate org selection is needed.
		return &dto.LoginResponse{
			OperatorID:           op.ID,
			Email:                op.Email,
			Name:                 op.Name,
			MFAEnabled:           op.MFAEnabled,
			NeedsOrganization:    true,
			Organizations:        nil,
			LastOrganizationID:   op.LastOrganizationID,
			SelectedOrganization: nil,
		}
	}

	// Determine if organization selection is required:.
	// - 0 memberships: needs organization (create/join).
	// - Multiple memberships without valid selected org: needs selection.
	needsOrg := len(orgs) == 0

	// Convert to dto.OrganizationInfo.
	dtoOrgs := make([]dto.OrganizationInfo, len(orgs))
	var selectedOrg *dto.OrganizationInfo
	for i, org := range orgs {
		dtoOrgs[i] = dto.OrganizationInfo{
			ID:   org.ID,
			Name: org.Name,
			Role: org.Role,
		}
		if org.ID == op.LastOrganizationID {
			selectedOrg = &dtoOrgs[i]
		}
	}

	// If LastOrganizationID is set but not found in active orgs,.
	// user has multiple orgs but the last one is invalid - needs selection.
	if op.LastOrganizationID != "" && selectedOrg == nil && len(orgs) > 1 {
		needsOrg = true
	}

	// If multiple orgs exist but none is selected, needs organization selection.
	if !needsOrg && len(orgs) > 1 && selectedOrg == nil {
		needsOrg = true
	}

	return &dto.LoginResponse{
		OperatorID:           op.ID,
		Email:                op.Email,
		Name:                 op.Name,
		MFAEnabled:           op.MFAEnabled,
		NeedsOrganization:    needsOrg,
		Organizations:        dtoOrgs,
		LastOrganizationID:   op.LastOrganizationID,
		SelectedOrganization: selectedOrg,
	}
}

// Login authenticates an operator and creates a session.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *session.Session, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			// Perform fake password hash to mitigate timing attacks.
			fakeHash := "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz"
			_ = s.passwordHasher.Verify(req.Password, fakeHash)
			return nil, nil, application.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Prevent nil pointer dereference if FindByEmail returns (nil, nil).
	if op == nil {
		// Perform fake password hash to mitigate timing attacks.
		fakeHash := "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz"
		_ = s.passwordHasher.Verify(req.Password, fakeHash)
		return nil, nil, application.ErrInvalidCredentials
	}

	if op.PasswordHash == "" {
		return nil, nil, application.ErrInvalidCredentials
	}

	// Verify password with proper error handling.
	if err = s.passwordHasher.Verify(req.Password, op.PasswordHash); err != nil {
		// Only return ErrInvalidCredentials for wrong password.
		// For other crypto errors, still return generic error to prevent info leak.
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" ||
			err.Error() == "crypto/scrypt: password hash does not match" ||
			err.Error() == "crypto/argon2: invalid hash" {
			return nil, nil, application.ErrInvalidCredentials
		}
		return nil, nil, application.ErrInvalidCredentials
	}

	// If MFA is required for this operator, enforce it.
	if op.MFARequired || op.HasMFA() {
		resp := s.buildLoginResponse(op)
		resp.MFAEnabled = true
		return resp, nil, application.ErrMFARequired
	}

	sess, err := s.CreateSession(ctx, op.ID)
	if err != nil {
		return nil, nil, err
	}

	// Auto-resolve and set organization for the session.
	// This enables single-org and last-used-org auto-selection.
	s.resolveAndSetOrganization(ctx, op, sess)

	return s.buildLoginResponse(op), sess, nil
}

// resolveAndSetOrganization resolves the organization for an operator and sets it on the session.
// It uses LastOrganizationID if valid, otherwise auto-selects single org.
func (s *AuthService) resolveAndSetOrganization(ctx context.Context, op *operator.Operator, sess *session.Session) {
	orgID, _, err := s.ResolveOrganizationForOperator(ctx, op)
	if err != nil {
		// No organization or selection required - do not set anything on session.
		return
	}

	// Valid org found - set on session and update LastOrganizationID.
	sess.SelectedOrganizationID = orgID
	_ = s.sessionRepo.UpdateOrganizationID(ctx, sess.ID, orgID)

	// Also update operators LastOrganizationID for persistence.
	if s.operatorRepo != nil && op.LastOrganizationID != orgID {
		op.LastOrganizationID = orgID
		if err := s.operatorRepo.Update(ctx, op); err != nil {
			s.logger.Warn("failed to update operator LastOrganizationID", "operatorID", op.ID, "error", err)
		}
	}
}

// LoginWithTokens authenticates an operator and returns tokens for API clients.
// This method is used for non-browser clients that need JWT access tokens and refresh tokens.
func (s *AuthService) LoginWithTokens(ctx context.Context, req *dto.LoginRequest) (*dto.LoginWithTokensResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			// Perform fake password hash to mitigate timing attacks.
			fakeHash := "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz"
			_ = s.passwordHasher.Verify(req.Password, fakeHash)
			return nil, application.ErrInvalidCredentials
		}
		return nil, err
	}

	if op == nil {
		// Perform fake password hash to mitigate timing attacks.
		fakeHash := "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz"
		_ = s.passwordHasher.Verify(req.Password, fakeHash)
		return nil, application.ErrInvalidCredentials
	}

	if op.PasswordHash == "" {
		return nil, application.ErrInvalidCredentials
	}

	if err = s.passwordHasher.Verify(req.Password, op.PasswordHash); err != nil {
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" ||
			err.Error() == "crypto/scrypt: password hash does not match" ||
			err.Error() == "crypto/argon2: invalid hash" {
			return nil, application.ErrInvalidCredentials
		}
		return nil, application.ErrInvalidCredentials
	}

	// If MFA is required, return partial response indicating MFA is needed.
	if op.MFARequired || op.HasMFA() {
		resp := s.buildLoginWithTokensResponse(op)
		resp.MFAEnabled = true
		return resp, application.ErrMFARequired
	}

	// Create session.
	sess, err := s.CreateSession(ctx, op.ID)
	if err != nil {
		return nil, err
	}

	// Generate JWT access token.
	var accessToken string
	var expiresAt int64
	if s.jwtManager != nil {
		accessToken, _, err = s.jwtManager.Generate(op.ID, op.Email, op.Name, "")
		if err != nil {
			return nil, err
		}
		expiresAt = time.Now().Add(15 * time.Minute).Unix()
	}

	// Issue refresh token.
	var refreshTokenVal string
	if s.refreshTokenRepo != nil {
		refreshTokenVal, err = s.IssueRefreshToken(ctx, op.ID, sess.ID)
		if err != nil {
			return nil, err
		}
	}

	// Auto-resolve and set organization for the session.
	// This enables single-org and last-used-org auto-selection.
	s.resolveAndSetOrganization(ctx, op, sess)

	resp := s.buildLoginWithTokensResponse(op)
	resp.AccessToken = accessToken
	resp.RefreshToken = refreshTokenVal
	resp.ExpiresAt = expiresAt
	resp.SessionID = sess.ID
	resp.SigningKey = sess.SigningKey
	return resp, nil
}

// buildLoginWithTokensResponse builds a LoginWithTokensResponse with organization info.
func (s *AuthService) buildLoginWithTokensResponse(op *operator.Operator) *dto.LoginWithTokensResponse {
	// Get operator's organizations.
	orgs, err := s.GetOperatorOrganizations(context.Background(), op)
	if err != nil {
		// If we can't get organizations, indicate org selection is needed.
		return &dto.LoginWithTokensResponse{
			OperatorID:           op.ID,
			Email:                op.Email,
			Name:                 op.Name,
			MFAEnabled:           op.MFAEnabled,
			NeedsOrganization:    true,
			Organizations:        nil,
			LastOrganizationID:   op.LastOrganizationID,
			SelectedOrganization: nil,
		}
	}

	// Determine if organization selection is required:.
	// - 0 memberships: needs organization (create/join).
	// - Multiple memberships without valid selected org: needs selection.
	needsOrg := len(orgs) == 0

	// Convert to dto.OrganizationInfo.
	dtoOrgs := make([]dto.OrganizationInfo, len(orgs))
	var selectedOrg *dto.OrganizationInfo
	for i, org := range orgs {
		dtoOrgs[i] = dto.OrganizationInfo{
			ID:   org.ID,
			Name: org.Name,
			Role: org.Role,
		}
		if org.ID == op.LastOrganizationID {
			selectedOrg = &dtoOrgs[i]
		}
	}

	// If LastOrganizationID is set but not found in active orgs,.
	// user has multiple orgs but the last one is invalid - needs selection.
	if op.LastOrganizationID != "" && selectedOrg == nil && len(orgs) > 1 {
		needsOrg = true
	}

	// If multiple orgs exist but none is selected, needs organization selection.
	if !needsOrg && len(orgs) > 1 && selectedOrg == nil {
		needsOrg = true
	}

	return &dto.LoginWithTokensResponse{
		OperatorID:           op.ID,
		Email:                op.Email,
		Name:                 op.Name,
		MFAEnabled:           op.MFAEnabled,
		NeedsOrganization:    needsOrg,
		Organizations:        dtoOrgs,
		LastOrganizationID:   op.LastOrganizationID,
		SelectedOrganization: selectedOrg,
	}
}

// Register creates a new operator.
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest, validatePassword bool) (*dto.RegisterResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	if validatePassword {
		if err := ValidatePassword(req.Password, DefaultPasswordPolicy); err != nil {
			return nil, fmt.Errorf("%w: %v", application.ErrInvalidInput, err)
		}
		// Check if password was found in known data breaches.
		breached, breachErr := infraauth.CheckPasswordBreached(req.Password)
		if breached {
			if breachErr != nil && errors.Is(breachErr, infraauth.ErrBreachCheckFailed) {
				return nil, application.ErrBreachCheckUnavailable
			}
			return nil, application.ErrPasswordBreached
		}
	}

	existing, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, application.ErrUserExists
	}

	hash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	id := shared.GenerateID()

	op := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		PasswordHash:  hash,
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: false,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		// Handle race condition: if UNIQUE constraint fails due to concurrent registration,.
		// return ErrUserExists instead of opaque database error.
		if errors.Is(err, operator.ErrEmailExists) {
			return nil, application.ErrUserExists
		}
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: id,
		Email:      email,
		Name:       name,
	}, nil
}

// RegisterAsSuperAdmin registers the first operator as super admin.
func (s *AuthService) RegisterAsSuperAdmin(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if err := ValidatePassword(req.Password, DefaultPasswordPolicy); err != nil {
		return nil, fmt.Errorf("%w: %v", application.ErrInvalidInput, err)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, application.ErrForbidden
	}

	hash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	id := shared.GenerateID()

	op := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		PasswordHash:  hash,
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: true,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: id,
		Email:      email,
		Name:       name,
	}, nil
}

// CreateSession creates a new session for an operator.
// If the operator has reached their max concurrent sessions limit,
// the oldest session is revoked before creating a new one.
func (s *AuthService) CreateSession(ctx context.Context, operatorID string) (*session.Session, error) {
	// Get operator security settings for session limit.
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	maxSessions := 5 // default.
	settings := op.SecuritySettings
	if settings.MaxConcurrentSessions > 0 {
		maxSessions = settings.MaxConcurrentSessions
	}

	// Count active sessions.
	activeSessions, err := s.sessionRepo.ListActiveByOperator(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// If at limit, revoke the oldest session before creating a new one.
	if len(activeSessions) >= maxSessions {
		oldest := activeSessions[0]
		if err := s.sessionRepo.AddSessionRevocation(ctx, oldest.ID, "max_sessions_reached"); err != nil {
			return nil, fmt.Errorf("failed to revoke oldest session: %w", err)
		}
		if err := s.sessionRepo.Delete(ctx, oldest.ID); err != nil {
			return nil, fmt.Errorf("failed to delete oldest session: %w", err)
		}
	}

	now := time.Now()
	id := shared.GenerateID()

	// Generate a per-session HMAC signing key (32 random bytes, hex-encoded).
	// The browser client receives this key on login and uses it to sign every
	// subsequent request. The server verifies the signature to confirm the
	// request originated from the authenticated client.
	signingKeyBytes := make([]byte, 32)
	if _, err := rand.Read(signingKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}
	signingKey := hex.EncodeToString(signingKeyBytes)

	sess := &session.Session{
		ID:         id,
		OperatorID: operatorID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.sessionTTL),
		SigningKey: signingKey,
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// Logout destroys a session and revokes associated refresh tokens.

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	// Get operator ID before deleting session for refresh token revocation.
	var operatorID string
	if sess, err := s.sessionRepo.FindByID(ctx, sessionID); err == nil && sess != nil {
		operatorID = sess.OperatorID
	}

	// Add to revocation list for audit - log error but continue.
	if err := s.sessionRepo.AddSessionRevocation(ctx, sessionID, "operator_logout"); err != nil {
		if s.logger != nil {
			s.logger.Warn("failed to add session revocation during logout",
				"sessionID", sessionID,
				"error", err)
		}
	}

	// Delete the session.
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return err
	}

	// Return error if revocation fails to ensure session is fully invalidated.
	if operatorID != "" {
		if err := s.RevokeAllRefreshTokens(ctx, operatorID); err != nil {
			if s.logger != nil {
				s.logger.Error("failed to revoke refresh tokens during logout - session may persist",
					"operatorID", operatorID,
					"error", err)
			}
			return fmt.Errorf("logout incomplete: failed to revoke refresh tokens: %w", err)
		}
	}

	return nil
}

// LogoutAll destroys all sessions for an operator.

func (s *AuthService) LogoutAll(ctx context.Context, operatorID string) error {
	if err := s.sessionRepo.RevokeAllOperatorSessions(ctx, operatorID); err != nil {
		return err
	}

	if err := s.RevokeAllRefreshTokens(ctx, operatorID); err != nil {
		if s.logger != nil {
			s.logger.Error("failed to revoke refresh tokens during logout all - sessions may persist",
				"operatorID", operatorID,
				"error", err)
		}
		return fmt.Errorf("logout incomplete: failed to revoke refresh tokens: %w", err)
	}
	return nil
}

// GetOperator retrieves an operator by ID.
func (s *AuthService) GetOperator(ctx context.Context, id string) (*operator.Operator, error) {
	return s.operatorRepo.FindByID(ctx, id)
}

// GetOperatorByEmail retrieves an operator by email.
func (s *AuthService) GetOperatorByEmail(ctx context.Context, email string) (*operator.Operator, error) {
	return s.operatorRepo.FindByEmail(ctx, email)
}

// GetSession retrieves a session by ID.
func (s *AuthService) GetSession(ctx context.Context, id string) (*session.Session, error) {
	return s.sessionRepo.FindByID(ctx, id)
}

// ValidateSession validates a session and returns the operator.
// It checks expiration, revocation status, and MFA requirements.
// It also loads memberships so that operator.GetMembership() works correctly.
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*session.Session, *operator.Operator, error) {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		if err == session.ErrNotFound {
			return nil, nil, application.ErrUnauthorized
		}
		return nil, nil, err
	}

	// Check if session is expired.
	if sess.IsExpired() {
		// Clean up expired session.
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, nil, application.ErrTokenExpired
	}

	// Check if session has been revoked (server-side logout).
	revoked, err := s.sessionRepo.IsSessionRevoked(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if revoked {
		return nil, nil, application.ErrUnauthorized
	}

	op, err := s.operatorRepo.FindByID(ctx, sess.OperatorID)
	if err != nil {
		return nil, nil, err
	}

	// MFA enforcement: If operator has MFA enabled and MFA was enabled at a specific time,.
	// reject sessions that were not MFA-verified after MFA was enabled.
	// This prevents pre-MFA sessions from bypassing MFA requirements.
	if op.MFAEnabled && op.MFAEnabledAt != nil {
		// Session must have been MFA-verified after MFA was enabled.
		if sess.MFAVerifiedAt == nil || sess.MFAVerifiedAt.Before(*op.MFAEnabledAt) {
			// Clean up the session and require re-authentication with MFA.
			_ = s.sessionRepo.Delete(ctx, sessionID)
			return nil, nil, application.ErrUnauthorized
		}
	}

	// Load memberships so that operator.GetMembership() works correctly.
	if s.memberRepo != nil {
		memberships, err := s.memberRepo.ListByOperator(ctx, op.ID)
		if err != nil {
			return nil, nil, err
		}
		op.Memberships = memberships
	}

	return sess, op, nil
}

// ChangePassword changes an operator's password.
func (s *AuthService) ChangePassword(ctx context.Context, operatorID, oldPassword, newPassword string) error {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}

	if verifyErr := s.passwordHasher.Verify(oldPassword, op.PasswordHash); verifyErr != nil {
		return application.ErrInvalidCredentials
	}

	if validateErr := ValidatePassword(newPassword, DefaultPasswordPolicy); validateErr != nil {
		return fmt.Errorf("%w: %v", application.ErrInvalidInput, validateErr)
	}

	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	op.PasswordHash = hash
	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return err
	}

	// Invalidate all sessions and refresh tokens for security.
	_ = s.LogoutAll(ctx, operatorID)
	_ = s.RevokeAllRefreshTokens(ctx, operatorID)

	return nil
}
