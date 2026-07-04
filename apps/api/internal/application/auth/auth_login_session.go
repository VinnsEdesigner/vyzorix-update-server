package auth

import (
	"context"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// Login authenticates an operator and creates a session.
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *session.Session, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			_ = s.passwordHasher.Verify(req.Password, "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz")
			return nil, nil, application.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if op.PasswordHash == "" {
		return nil, nil, application.ErrInvalidCredentials
	}

	if err = s.passwordHasher.Verify(req.Password, op.PasswordHash); err != nil {
		return nil, nil, application.ErrInvalidCredentials
	}

	if op.HasMFA() {
		return &dto.LoginResponse{
			OperatorID: op.ID,
			Email:      op.Email,
			Name:       op.Name,
			Role:       string(op.Role),
			MFAEnabled: true,
		}, nil, application.ErrMFARequired
	}

	sess, err := s.CreateSession(ctx, op.ID)
	if err != nil {
		return nil, nil, err
	}

	return &dto.LoginResponse{
		OperatorID: op.ID,
		Email:      op.Email,
		Name:       op.Name,
		Role:       string(op.Role),
		MFAEnabled: op.MFAEnabled,
	}, sess, nil
}

// Register creates a new operator.
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest, validatePassword bool) (*dto.RegisterResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	if validatePassword {
		if err := ValidatePassword(req.Password, DefaultPasswordPolicy); err != nil {
			return nil, err
		}
		// CRITICAL-6: Check if password was found in known data breaches
		if breached, _ := infraauth.CheckPasswordBreached(req.Password); breached {
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

	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	role := operator.RoleOperator
	if count == 0 {
		role = operator.RoleSuperAdmin
	}

	now := time.Now()
	id := shared.GenerateID()

	op := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		PasswordHash:  hash,
		Role:          role,
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: false,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: id,
		Email:     email,
		Name:      name,
	}, nil
}

// RegisterAsSuperAdmin registers the first operator as super admin.
func (s *AuthService) RegisterAsSuperAdmin(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if err := ValidatePassword(req.Password, DefaultPasswordPolicy); err != nil {
		return nil, err
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
		Role:          operator.RoleSuperAdmin,
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: true,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		OperatorID: id,
		Email:     email,
		Name:      name,
	}, nil
}

// CreateSession creates a new session for an operator.
func (s *AuthService) CreateSession(ctx context.Context, operatorID string) (*session.Session, error) {
	now := time.Now()
	id := shared.GenerateID()

	sess := &session.Session{
		ID:         id,
		OperatorID: operatorID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.sessionTTL),
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// Logout destroys a session.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	_ = s.sessionRepo.AddSessionRevocation(ctx, sessionID, "operator_logout")
	return s.sessionRepo.Delete(ctx, sessionID)
}

// LogoutAll destroys all sessions for an operator.
func (s *AuthService) LogoutAll(ctx context.Context, operatorID string) error {
	if err := s.sessionRepo.RevokeAllOperatorSessions(ctx, operatorID); err != nil {
		return err
	}
	_ = s.RevokeAllRefreshTokens(ctx, operatorID)
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
// It checks expiration and revocation status.
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

	return sess, op, nil
}

// ChangePassword changes an operator's password.
func (s *AuthService) ChangePassword(ctx context.Context, operatorID, oldPassword, newPassword string) error {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}

	if err := s.passwordHasher.Verify(oldPassword, op.PasswordHash); err != nil {
		return application.ErrInvalidCredentials
	}

	if err := ValidatePassword(newPassword, DefaultPasswordPolicy); err != nil {
		return err
	}

	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	op.PasswordHash = hash
	op.UpdatedAt = time.Now()

	return s.operatorRepo.Update(ctx, op)
}
