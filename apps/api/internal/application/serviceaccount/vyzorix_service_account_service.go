// Package serviceaccount provides service account and token management:
// create/list/delete, token rotation, secretscan, and service account auth.
package serviceaccount

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrServiceAccountNotFound = errors.New("service account not found")
	ErrInvalidToken           = errors.New("invalid token")
	ErrInvalidServiceAccount  = errors.New("invalid service account")
	ErrInvalidScope           = errors.New("invalid scope")
)

// TokenInput carries the mutable fields of a create/update/rotate request.
type TokenInput struct {
	ExpiresAt *time.Time
	ServiceID string
	Name      string
	Scopes    []string
}

// Service provides service account and token lifecycle management.
type Service struct {
	repo        serviceaccount.Repository
	tokens      serviceaccount.TokenRepository
	auditLogger *audit.Logger
	emailService *email.Service
	memberRepo   organization.MemberRepository
	maxTokens    int
}

// httpClient returns a bounded HTTP client for code search calls.
func (s *Service) httpClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// SetMaxTokens caps how many tokens a service account may hold (0 = unlimited).
func (s *Service) SetMaxTokens(limit int) {
	s.maxTokens = limit
}

// SetMemberRepository wires the member lookup for leak notification resolution.
func (s *Service) SetMemberRepository(repo organization.MemberRepository) {
	s.memberRepo = repo
}

// NewService creates a new Service.
func NewService(repo serviceaccount.Repository, tokens serviceaccount.TokenRepository) *Service {
	return &Service{repo: repo, tokens: tokens}
}

// SetAuditLogger wires the audit logger for lifecycle events.
func (s *Service) SetAuditLogger(logger *audit.Logger) {
	s.auditLogger = logger
}

// audit writes an audit entry when the logger is available.
func (s *Service) audit(ctx context.Context, action audit.Action, serviceID, tokenID string) {
	if s.auditLogger == nil {
		return
	}
	entry := &audit.Entry{
		Action:       action,
		ResourceType: "service_account",
		ResourceID:   serviceID,
		ActorType:    "operator",
	}
	if tokenID != "" {
		entry.Metadata = map[string]string{"token_id": tokenID}
	}
	s.auditLogger.LogEvent(ctx, entry)
}

// Create validates and persists a service account.
func (s *Service) Create(ctx context.Context, orgID, name string) (*serviceaccount.ServiceAccount, error) {
	sa := &serviceaccount.ServiceAccount{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      name,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := sa.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidServiceAccount, err)
	}
	if err := s.repo.Save(ctx, sa); err != nil {
		return nil, err
	}
	s.audit(ctx, audit.ActionServiceAccountCreated, sa.ID, "")
	return sa, nil
}

// List returns all service accounts of an org.
func (s *Service) List(ctx context.Context, orgID string) ([]*serviceaccount.ServiceAccount, error) {
	return s.repo.ListByOrg(ctx, orgID)
}

// Get returns one service account.
func (s *Service) Get(ctx context.Context, orgID, id string) (*serviceaccount.ServiceAccount, error) {
	sa, err := s.getScoped(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

// Delete removes a service account and revokes all its tokens.
func (s *Service) Delete(ctx context.Context, orgID, id string) error {
	if _, err := s.getScoped(ctx, orgID, id); err != nil {
		return err
	}
	tokens, err := s.tokens.ListByService(ctx, id)
	if err != nil {
		return err
	}
	for _, token := range tokens {
		_ = s.tokens.Revoke(ctx, token.ID, time.Now())
	}
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrServiceAccountNotFound
	}
	s.audit(ctx, audit.ActionServiceAccountDeleted, id, "")
	return nil
}

// CreateToken generates a new token, returning the full key only at creation.
func (s *Service) CreateToken(ctx context.Context, in *TokenInput) (*serviceaccount.Token, string, error) {
	if _, err := s.getScoped(ctx, "", in.ServiceID); err != nil {
		return nil, "", err
	}
	if s.maxTokens > 0 {
		tokens, err := s.tokens.ListByService(ctx, in.ServiceID)
		if err != nil {
			return nil, "", err
		}
		if len(tokens) >= s.maxTokens {
			return nil, "", errors.New("service account token quota exceeded")
		}
	}
	if len(in.Scopes) == 0 {
		in.Scopes = []string{"read"}
	}
	for _, scope := range in.Scopes {
		if !serviceaccount.Scope(scope).Valid() {
			return nil, "", fmt.Errorf("%w: %q", ErrInvalidScope, scope)
		}
	}

	key, err := generateKey()
	if err != nil {
		return nil, "", err
	}
	keyHash := hashKey(key)
	prefix := keyPrefix(key)

	token := &serviceaccount.Token{
		ID:        uuid.New(),
		ServiceID: in.ServiceID,
		Name:      in.Name,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Scopes:    in.Scopes,
		Valid:     true,
		ExpiresAt: in.ExpiresAt,
		CreatedAt: time.Now(),
	}
	if err := token.Validate(); err != nil {
		return nil, "", errors.Join(ErrInvalidToken, err)
	}
	if err := s.tokens.Save(ctx, token); err != nil {
		return nil, "", err
	}
	metrics.Get().RecordServiceAccountToken(in.ServiceID, "created")
	s.audit(ctx, audit.ActionServiceAccountTokenCreated, in.ServiceID, token.ID)
	return token, key, nil
}

// ListTokens returns tokens of a service account.
func (s *Service) ListTokens(ctx context.Context, serviceID string) ([]*serviceaccount.Token, error) {
	return s.tokens.ListByService(ctx, serviceID)
}

// RevokeToken revokes a token.
func (s *Service) RevokeToken(ctx context.Context, tokenID string) error {
	if _, err := s.tokens.GetByID(ctx, tokenID); err != nil {
		return err
	}
	return s.tokens.Revoke(ctx, tokenID, time.Now())
}

// RotateToken revokes the old token and creates a new one in its place.
// Returns the new token and the new full key (only available once).
func (s *Service) RotateToken(ctx context.Context, tokenID string, in *TokenInput) (*serviceaccount.Token, string, error) {
	old, err := s.tokens.GetByID(ctx, tokenID)
	if err != nil {
		return nil, "", err
	}
	if !old.IsUsable() {
		return nil, "", errors.New("cannot rotate an expired or revoked token")
	}

	if in.ServiceID == "" {
		in.ServiceID = old.ServiceID
	}
	if len(in.Scopes) == 0 {
		in.Scopes = old.Scopes
	}
	if in.Name == "" {
		in.Name = old.Name + "-rotated"
	}
	newToken, key, err := s.CreateToken(ctx, in)
	if err != nil {
		return nil, "", err
	}
	if err := s.tokens.Revoke(ctx, tokenID, time.Now()); err != nil {
		return nil, "", err
	}
	metrics.Get().RecordServiceAccountToken(in.ServiceID, "rotated")
	s.audit(ctx, audit.ActionServiceAccountTokenRotated, in.ServiceID, newToken.ID)
	return newToken, key, nil
}

// Authenticate validates a presented key against stored hashes. Returns the
// token when the key is valid and usable.
func (s *Service) Authenticate(ctx context.Context, key string) (*serviceaccount.Token, error) {
	keyHash := hashKey(key)
	token, err := s.tokens.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, errors.Join(ErrInvalidToken, err)
	}
	if !token.IsUsable() {
		return nil, ErrInvalidToken
	}
	return token, nil
}

// GetOrgID returns the org ID of the service account that owns a token (used
// by auth middleware to set org context from service account tokens).
func (s *Service) GetOrgID(ctx context.Context) string {
	// Fallback: the auth middleware already resolved the token and doesn't pass
	// the token ID here, so this is reserved for future use. Callers that know
	// the service account ID should query directly.
	return ""
}

// GetOrgIDByServiceID returns the org ID for a service account ID (the exact
// lookup the middleware needs).
func (s *Service) GetOrgIDByServiceID(ctx context.Context, serviceID string) string {
	if s.repo == nil {
		return ""
	}
	sa, err := s.repo.GetByID(ctx, serviceID)
	if err != nil {
		return ""
	}
	return sa.OrgID
}

// ListAllOrgs returns the distinct org IDs holding service accounts (used by
// the leak worker's cross-org scan).
func (s *Service) ListAllOrgs(ctx context.Context) ([]string, error) {
	if s.repo == nil {
		return nil, nil
	}
	// Repository-level query joining distinct org_id values.
	// Implemented on the storage side when invoked by the leak worker.
	return s.repo.ListAllOrgs(ctx)
}

// ScanForLeaks scans public GitHub and GitLab code search for tokens
// matching each service account of an org. Search by prefix only (no full
// keys in flight).
func (s *Service) ScanForLeaks(ctx context.Context, orgID string) (int, error) {
	accounts, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return 0, err
	}
	found := 0
	for _, sa := range accounts {
		tokens, err := s.tokens.ListByService(ctx, sa.ID)
		if err != nil {
			continue
		}
		for _, token := range tokens {
			if !token.IsUsable() {
				continue
			}
			leaks, err := s.searchCode(ctx, token.KeyPrefix)
			if err != nil {
				continue
			}
			found += leaks
			if leaks > 0 {
				metrics.Get().RecordServiceAccountToken(orgID, "leak_detected")
				if s.emailService != nil && s.memberRepo != nil {
					members, _ := s.memberRepo.FindActiveByOrganization(ctx, orgID)
					for _, m := range members {
						if m.Role == organization.RoleAdmin || m.Role == organization.RoleSuperAdmin {
							_ = s.emailService.SendLeakAlertEmail(ctx, m.OperatorEmail, sa.Name, token.KeyPrefix, leaks)
						}
					}
				}
			}
		}
	}
	return found, nil
}

// searchCode queries the public code search APIs of GitHub and GitLab for
// occurrences of a token prefix. Returns the leak count or an error on
// both-API failure.
func (s *Service) searchCode(ctx context.Context, prefix string) (int, error) {
	total := 0
	var lastErr error

	if gh, err := s.githubCodeSearch(ctx, prefix); err == nil {
		total += gh
	} else {
		lastErr = err
	}
	if gl, err := s.gitlabCodeSearch(ctx, prefix); err == nil {
		total += gl
	} else if lastErr == nil {
		lastErr = err
	}
	if lastErr != nil && total == 0 {
		return 0, lastErr
	}
	return total, nil
}

//nolint:gosec
func (s *Service) githubCodeSearch(ctx context.Context, query string) (int, error) {
	url := "https://api.github.com/search/code?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github search status %d", resp.StatusCode)
	}

	var body struct {
		TotalCount int `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	return body.TotalCount, nil
}

//nolint:gosec
func (s *Service) gitlabCodeSearch(ctx context.Context, query string) (int, error) {
	base := os.Getenv("GITLAB_API_URL")
	if base == "" {
		base = "https://gitlab.com/api/v4"
	}
	url := base + "/search?scope=code&search=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("gitlab search status %d", resp.StatusCode)
	}

	var body []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	return len(body), nil
}

// Secretscan compares an outbound payload against the service account's
// token prefixes. When a token prefix appears in the payload, the account is
// notified to org admins via email; returns the number of matched leaks.
func (s *Service) Secretscan(ctx context.Context, orgID string, payload []byte) (int, error) {
	if len(payload) == 0 || orgID == "" {
		return 0, nil
	}
	accounts, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return 0, err
	}
	payloadStr := string(payload)
	leaks := 0
	var leaked *serviceaccount.Token
	for _, sa := range accounts {
		tokens, err := s.tokens.ListByService(ctx, sa.ID)
		if err != nil {
			continue
		}
		for _, token := range tokens {
			if token.IsUsable() && strings.Contains(payloadStr, token.KeyPrefix) {
				leaks++
				leaked = token
			}
		}
	}
	if leaks > 0 {
		metrics.Get().RecordServiceAccountToken(orgID, "leak_detected")
		if s.emailService != nil && s.memberRepo != nil {
			members, err := s.memberRepo.FindActiveByOrganization(ctx, orgID)
			if err == nil {
				account, _ := s.repo.GetByID(ctx, leaked.ServiceID)
				for _, m := range members {
					if m.Role == organization.RoleAdmin || m.Role == organization.RoleSuperAdmin {
						_ = s.emailService.SendLeakAlertEmail(ctx, m.OperatorEmail, account.Name, leaked.KeyPrefix, leaks)
					}
				}
			}
		}
	}
	return leaks, nil
}

func (s *Service) getScoped(ctx context.Context, orgID, id string) (*serviceaccount.ServiceAccount, error) {
	sa, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, serviceaccount.ErrNotFound) {
		return nil, ErrServiceAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	if orgID != "" && sa.OrgID != orgID {
		return nil, ErrServiceAccountNotFound
	}
	return sa, nil
}

func generateKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "vxyz_sa_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashKey(key string) string {
	salt := []byte("vyzorix-service-account-v1")
	return base64.RawURLEncoding.EncodeToString(argon2.IDKey([]byte(key), salt, 1, 65536, 4, 32))
}

func keyPrefix(key string) string {
	if len(key) < 16 {
		return key
	}
	return key[:16]
}

// HasScope reports whether the token's scopes include the required scope.
// read < write < admin: admin satisfies every lower scope.
func HasScope(token *serviceaccount.Token, required serviceaccount.Scope) bool {
	highest := false
	for _, s := range token.Scopes {
		switch serviceaccount.Scope(s) {
		case serviceaccount.ScopeAdmin:
			highest = true
		case serviceaccount.ScopeWrite:
			if required == serviceaccount.ScopeWrite || required == serviceaccount.ScopeRead {
				highest = true
			}
		case serviceaccount.ScopeRead:
			if required == serviceaccount.ScopeRead {
				highest = true
			}
		}
	}
	return highest
}

// ScopeLabel returns the highest privilege level as a label for metrics.
func ScopeLabel(token *serviceaccount.Token) string {
	for _, s := range token.Scopes {
		if s == string(serviceaccount.ScopeAdmin) {
			return "admin"
		}
	}
	for _, s := range token.Scopes {
		if s == string(serviceaccount.ScopeWrite) {
			return "write"
		}
	}
	return "read"
}

// ExtractKey extracts the service account key from the Authorization header.
// Supports "Bearer vxyz_sa_..." or "Authorization: vxyz_sa_..." forms.
func ExtractKey(authHeader string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "vxyz_sa_") {
		return rest
	}
	return ""
}
