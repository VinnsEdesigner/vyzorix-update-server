package auth

import (
"context"
"time"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/email_verification"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/refresh_token"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
infraSession "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
)

// PasswordPolicy delegates to domain operator package.
type PasswordPolicy = operator.PasswordPolicy

// DefaultPasswordPolicy delegates to domain operator package.
var DefaultPasswordPolicy = operator.DefaultPasswordPolicy

// ValidatePassword delegates to domain operator package.
var ValidatePassword = operator.ValidatePassword

// PasswordStrength delegates to domain operator package.
var PasswordStrength = operator.Strength

// PasswordError delegates to domain operator package.
type PasswordError = operator.PasswordError

// RefreshTokenRepository interface for refresh token operations.
type RefreshTokenRepository interface {
Create(ctx context.Context, rt *refresh_token.RefreshToken) error
FindByID(ctx context.Context, id string) (*refresh_token.RefreshToken, error)
FindByTokenHash(ctx context.Context, tokenHash string) (*refresh_token.RefreshToken, error)
Revoke(ctx context.Context, id string) error
RevokeByTokenHash(ctx context.Context, tokenHash string) error
RevokeAllForOperator(ctx context.Context, operatorID string) error
CleanupExpired(ctx context.Context) error
}

// SelectOrganizationResult contains the result of selecting an organization.
type SelectOrganizationResult struct {
Session          *session.Session
OrganizationID   string
OrganizationName string
Role            organization.OrganizationRole
}

// OrganizationInfo contains organization details for client response.
type OrganizationInfo struct {
ID   string `json:"id"`
Name string `json:"name"`
Role string `json:"role"`
}

// AuthService handles authentication operations.
type AuthService struct {
operatorRepo        operator.Repository
memberRepo         organization.MemberRepository
orgRepo            organization.OrganizationRepository
invitationRepo     organization.InvitationRepository
sessionRepo        session.Repository
emailVerifyRepo    email_verification.Repository
passwordResetRepo  password_reset.Repository
passwordHasher     PasswordHasher
refreshTokenRepo   RefreshTokenRepository
jwtManager         *infraauth.JWTManager
sessionManager     *infraSession.Manager
ldapConfig         *LDAPConfig
deviceStore        *DeviceStore
deviceRepo         device.Repository
sessionTTL         time.Duration
refreshTokenExpiry time.Duration
}

// NewAuthService creates a new AuthService.
func NewAuthService(
operatorRepo operator.Repository,
sessionRepo session.Repository,
emailVerifyRepo email_verification.Repository,
passwordResetRepo password_reset.Repository,
passwordHasher PasswordHasher,
sessionTTL time.Duration,
) *AuthService {
return &AuthService{
operatorRepo:      operatorRepo,
sessionRepo:        sessionRepo,
emailVerifyRepo:   emailVerifyRepo,
passwordResetRepo: passwordResetRepo,
passwordHasher:     passwordHasher,
sessionTTL:         sessionTTL,
}
}

// NewAuthServiceWithRefresh creates a new AuthService with refresh token support.
func NewAuthServiceWithRefresh(
operatorRepo operator.Repository,
sessionRepo session.Repository,
emailVerifyRepo email_verification.Repository,
passwordResetRepo password_reset.Repository,
passwordHasher PasswordHasher,
sessionTTL time.Duration,
refreshTokenRepo RefreshTokenRepository,
refreshTokenExpiry time.Duration,
jwtManager *infraauth.JWTManager,
ldapConfig *LDAPConfig,
) *AuthService {
return &AuthService{
operatorRepo:        operatorRepo,
sessionRepo:         sessionRepo,
emailVerifyRepo:     emailVerifyRepo,
passwordResetRepo:   passwordResetRepo,
passwordHasher:      passwordHasher,
sessionTTL:          sessionTTL,
refreshTokenRepo:    refreshTokenRepo,
refreshTokenExpiry:  refreshTokenExpiry,
jwtManager:          jwtManager,
ldapConfig:          ldapConfig,
}
}

// SetLDAPConfig sets the LDAP configuration for the auth service.
func (s *AuthService) SetLDAPConfig(cfg *LDAPConfig) {
s.ldapConfig = cfg
}

// SetJWTManager sets the JWT manager for the auth service.
func (s *AuthService) SetJWTManager(jwtManager *infraauth.JWTManager) {
s.jwtManager = jwtManager
}

// SetSessionManager sets the session manager for the auth service.
func (s *AuthService) SetSessionManager(sessionManager *infraSession.Manager) {
s.sessionManager = sessionManager
}

// SetInvitationRepository sets the invitation repository for operator deletion flow.
func (s *AuthService) SetInvitationRepository(invitationRepo organization.InvitationRepository) {
s.invitationRepo = invitationRepo
}

// SetDeviceRepository sets the device repository for operator deletion flow.
func (s *AuthService) SetDeviceRepository(deviceRepo device.Repository) {
s.deviceRepo = deviceRepo
}

// SetMemberRepository sets the member repository for organization membership checks.
func (s *AuthService) SetMemberRepository(memberRepo organization.MemberRepository) {
s.memberRepo = memberRepo
}

// SetOrganizationRepository sets the organization repository.
func (s *AuthService) SetOrganizationRepository(orgRepo organization.OrganizationRepository) {
s.orgRepo = orgRepo
}

// GetSessionManager returns the session manager.
func (s *AuthService) GetSessionManager() *infraSession.Manager {
return s.sessionManager
}

// ResolveOrganizationForOperator resolves the appropriate organization for an operator.
// Priority: LastOrganizationID (if valid) > Single org > Multiple orgs (require selection)
//
// Returns:
// - orgID, orgName, nil: Single org or auto-selected org
// - "", "", ErrNoOrganization: No memberships
// - "", "", ErrOrgSelectionRequired: Multiple orgs require selection
func (s *AuthService) ResolveOrganizationForOperator(ctx context.Context, op *operator.Operator) (string, string, error) {
// Load memberships if not already loaded
if len(op.Memberships) == 0 && s.memberRepo != nil {
members, err := s.memberRepo.ListByOperator(ctx, op.ID)
if err != nil {
return "", "", err
}
op.Memberships = members
}

// Filter to active memberships only
var activeMemberships []*organization.OrganizationMember
for _, m := range op.Memberships {
if m.IsActive() {
activeMemberships = append(activeMemberships, m)
}
}

if len(activeMemberships) == 0 {
return "", "", application.ErrNoOrganization
}

// Single active org - auto-select
if len(activeMemberships) == 1 {
m := activeMemberships[0]
orgName := m.OrganizationID
if s.orgRepo != nil {
if org, err := s.orgRepo.FindByID(ctx, m.OrganizationID); err == nil {
orgName = org.Name
}
}
return m.OrganizationID, orgName, nil
}

// Multiple active orgs - check LastOrganizationID first
if op.LastOrganizationID != "" {
// Check if LastOrganizationID is still a valid active membership
for _, m := range activeMemberships {
if m.OrganizationID == op.LastOrganizationID {
orgName := m.OrganizationID
if s.orgRepo != nil {
if org, err := s.orgRepo.FindByID(ctx, m.OrganizationID); err == nil {
orgName = org.Name
}
}
return m.OrganizationID, orgName, nil
}
}
}

// Multiple orgs, no valid last org - require selection
return "", "", application.ErrOrgSelectionRequired
}

// SelectOrganization switches the session's organization context to the specified organization.
// This allows operators with multiple organization memberships to switch between them.
// Also updates the operator's LastOrganizationID for auto-selection on next login.
func (s *AuthService) SelectOrganization(ctx context.Context, operatorID, sessionID, organizationID string) (*SelectOrganizationResult, error) {
// Validate operator has membership in the target organization
member, err := s.ValidateOrganizationMembership(ctx, operatorID, organizationID)
if err != nil {
return nil, err
}

// Verify the session belongs to this operator
sess, err := s.sessionRepo.FindByID(ctx, sessionID)
if err != nil {
return nil, application.ErrUnauthorized
}
if sess.OperatorID != operatorID {
return nil, application.ErrUnauthorized
}

// Update session with new organization
if err := s.sessionRepo.UpdateOrganizationID(ctx, sessionID, organizationID); err != nil {
return nil, err
}

// Update operator's LastOrganizationID for auto-selection on next login
if s.operatorRepo != nil {
op, err := s.operatorRepo.FindByID(ctx, operatorID)
if err == nil && op != nil {
op.LastOrganizationID = organizationID
_ = s.operatorRepo.Update(ctx, op)
}
}

// Get organization name
orgName := organizationID
if s.orgRepo != nil {
if org, err := s.orgRepo.FindByID(ctx, organizationID); err == nil {
orgName = org.Name
}
}

// Update session object for return
sess.SelectedOrganizationID = organizationID

return &SelectOrganizationResult{
Session:          sess,
OrganizationID:   organizationID,
OrganizationName: orgName,
Role:            member.Role,
}, nil
}

// ValidateOrganizationMembership checks if an operator is a member of an organization.
func (s *AuthService) ValidateOrganizationMembership(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
if s.memberRepo == nil {
return nil, application.ErrForbidden
}

member, err := s.memberRepo.FindByOperatorAndOrg(ctx, operatorID, orgID)
if err != nil {
if err == organization.ErrMemberNotFound {
return nil, application.ErrForbidden
}
return nil, err
}

if !member.IsActive() {
return nil, application.ErrForbidden
}

return member, nil
}

// GetOperatorOrganizations returns all organizations for an operator with their roles.
func (s *AuthService) GetOperatorOrganizations(ctx context.Context, op *operator.Operator) ([]OrganizationInfo, error) {
// Load memberships if not already loaded
if len(op.Memberships) == 0 && s.memberRepo != nil {
members, err := s.memberRepo.ListByOperator(ctx, op.ID)
if err != nil {
return nil, err
}
op.Memberships = members
}

var result []OrganizationInfo
for _, m := range op.Memberships {
if !m.IsActive() {
continue
}

orgInfo := OrganizationInfo{
ID:   m.OrganizationID,
Role: string(m.Role),
}

// Get org name if repo available
if s.orgRepo != nil {
if org, err := s.orgRepo.FindByID(ctx, m.OrganizationID); err == nil {
orgInfo.Name = org.Name
}
}

result = append(result, orgInfo)
}

return result, nil
}
