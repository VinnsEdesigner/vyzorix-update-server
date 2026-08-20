// Package serviceaccount provides non-human identities for automation:
// named, org-scoped service accounts with scoped tokens. They carry the same
// scope model as API keys (read/write/admin) but are not attached to an
// operator, so key rotation and audit stay separate from human auth.
package serviceaccount

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when a service account is not found.
var ErrNotFound = errors.New("service account not found")

// Scope defines the permission level for a service account token.
type Scope string

const (
	ScopeRead  Scope = "read"
	ScopeWrite Scope = "write"
	ScopeAdmin Scope = "admin"
)

// Valid reports whether the scope is known.
func (s Scope) Valid() bool {
	switch s {
	case ScopeRead, ScopeWrite, ScopeAdmin:
		return true
	}
	return false
}

// ServiceAccount is a named, org-scoped identity for automation.
type ServiceAccount struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string
	OrgID     string
	Name      string
	Enabled   bool
}

// Validate checks the service account is well-formed.
func (sa *ServiceAccount) Validate() error {
	if strings.TrimSpace(sa.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(sa.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

// Token is one API token of a service account. The full key is only returned
// at creation; subsequent lookups see only the prefix + hash.
type Token struct {
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	RevokedAt  *time.Time
	LastUsedAt *time.Time
	ID         string
	ServiceID  string
	Name       string
	KeyHash    string
	KeyPrefix  string
	Scopes     []string
	Valid      bool
}

// IsExpired reports whether the token is past its expiry.
func (t *Token) IsExpired() bool {
	return t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt)
}

// IsUsable reports whether the token can authenticate.
func (t *Token) IsUsable() bool {
	return t.Valid && !t.IsExpired() && t.RevokedAt == nil
}

// Validate checks the token is well-formed.
func (t *Token) Validate() error {
	if strings.TrimSpace(t.ServiceID) == "" {
		return errors.New("service_id is required")
	}
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("name is required")
	}
	if len(t.Scopes) == 0 {
		return errors.New("scopes are required")
	}
	for _, scope := range t.Scopes {
		if !Scope(scope).Valid() {
			return fmt.Errorf("invalid scope %q", scope)
		}
	}
	return nil
}

// Repository persists service accounts.
type Repository interface {
	Save(ctx context.Context, sa *ServiceAccount) error
	GetByID(ctx context.Context, id string) (*ServiceAccount, error)
	ListByOrg(ctx context.Context, orgID string) ([]*ServiceAccount, error)
	ListAllOrgs(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// TokenRepository persists service account tokens.
type TokenRepository interface {
	Save(ctx context.Context, token *Token) error
	GetByID(ctx context.Context, id string) (*Token, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*Token, error)
	GetByPrefix(ctx context.Context, prefix string) (*Token, error)
	ListByService(ctx context.Context, serviceID string) ([]*Token, error)
	ListExpired(ctx context.Context, now time.Time) ([]*Token, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
	TouchLastUsed(ctx context.Context, id string, usedAt time.Time) error
	Delete(ctx context.Context, id string) (bool, error)
}
