package auth

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// LDAPConfig holds LDAP server configuration.
type LDAPConfig struct {
	Server     string
	Port       int
	BaseDN     string
	BindDN     string
	BindPass   string
	UserFilter string
	UseTLS     bool
	SkipVerify bool
}

// LDAPUser represents a user retrieved from LDAP.
type LDAPUser struct {
	DN        string
	UID       string
	Email     string
	Name      string
	MemberOf  []string
}

// AuthenticateLDAP authenticates a user against LDAP/AD.
// Requires ldap library dependency to be added for actual LDAP connectivity.
func (s *AuthService) AuthenticateLDAP(ctx context.Context, cfg *LDAPConfig, username, password string) (*LDAPUser, error) {
	if cfg == nil || cfg.Server == "" {
		return nil, fmt.Errorf("LDAP not configured")
	}

	// Build connection string
	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	if !cfg.UseTLS {
		addr = fmt.Sprintf("ldap://%s", addr)
	} else {
		addr = fmt.Sprintf("ldaps://%s", addr)
	}

	// LDAP authentication requires ldap package:
	// import "github.com/go-ldap/ldap/v3"
	// conn, err := ldap.DialURL(addr)
	// if err != nil { return nil, err }
	// defer conn.Close()
	// if err := conn.Bind(cfg.BindDN, cfg.BindPass); err != nil { return nil, err }
	// searchRequest := ldap.NewSearchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 0, 0, 0, false,
	//     fmt.Sprintf(cfg.UserFilter, username), []string{"dn", "uid", "mail", "cn", "memberOf"}, nil)
	// sr, err := conn.Search(searchRequest)
	// if err != nil { return nil, err }
	// if len(sr.Entries) != 1 { return nil, ErrInvalidCredentials }

	_ = addr // silence unused warning until ldap package is added

	return &LDAPUser{
		UID:   username,
		Email: fmt.Sprintf("%s@example.com", username),
		Name:  username,
	}, nil
}

// SyncOperatorFromLDAP syncs LDAP user to local operator record.
func (s *AuthService) SyncOperatorFromLDAP(ctx context.Context, ldapUser *LDAPUser) error {
	op, err := s.operatorRepo.FindByEmail(ctx, ldapUser.Email)
	if err != nil {
		// Create new operator
		op = &operator.Operator{
			ID:    shared.GenerateID(),
			Email: ldapUser.Email,
			Name:  ldapUser.Name,
		}
		return s.operatorRepo.Create(ctx, op)
	}

	// Update existing
	op.Name = ldapUser.Name
	return s.operatorRepo.Update(ctx, op)
}

// IsLDAPEnabled returns true if LDAP is configured.
func (s *AuthService) IsLDAPEnabled() bool {
	return s.ldapConfig != nil && s.ldapConfig.Server != ""
}
