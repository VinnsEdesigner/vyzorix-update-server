package auth

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/go-ldap/ldap/v3"
)

// LDAPConfig holds LDAP server configuration.
type LDAPConfig struct {
	Server     string
	BaseDN     string
	BindDN     string
	BindPass   string
	UserFilter string
	Port       int
	UseTLS     bool
	SkipVerify bool
}

// LDAPUser represents a user retrieved from LDAP.
type LDAPUser struct {
	DN       string
	UID      string
	Email    string
	Name     string
	MemberOf []string
}

// AuthenticateLDAP authenticates a user against LDAP/AD.
func (s *AuthService) AuthenticateLDAP(ctx context.Context, cfg *LDAPConfig, username, password string) (*LDAPUser, error) {
	if cfg == nil || cfg.Server == "" {
		return nil, fmt.Errorf("LDAP not configured")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	proto := "ldap"
	if cfg.UseTLS {
		proto = "ldaps"
	}
	ldapURL := fmt.Sprintf("%s://%s", proto, addr)

	conn, err := ldap.DialURL(ldapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if cfg.UseTLS {
		err = conn.StartTLS(nil)
		if err != nil && !cfg.SkipVerify {
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	if err := conn.Bind(cfg.BindDN, cfg.BindPass); err != nil {
		return nil, fmt.Errorf("failed to bind with service account: %w", err)
	}

	filter := cfg.UserFilter
	if filter == "" {
		filter = "(uid=%s)"
	}
	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(filter, ldap.EscapeFilter(username)),
		[]string{"dn", "uid", "mail", "cn", "memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found or multiple matches")
	}

	entry := sr.Entries[0]
	userDN := entry.DN

	userConn, err := ldap.DialURL(ldapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect for user bind: %w", err)
	}
	defer func() { _ = userConn.Close() }()

	if cfg.UseTLS {
		_ = userConn.StartTLS(nil)
	}

	if err := userConn.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	memberOf := make([]string, 0, len(entry.Attributes))
	for _, attr := range entry.Attributes {
		if attr.Name == "memberOf" {
			memberOf = append(memberOf, attr.Values...)
		}
	}

	return &LDAPUser{
		DN:       userDN,
		UID:      entry.GetAttributeValue("uid"),
		Email:    entry.GetAttributeValue("mail"),
		Name:     entry.GetAttributeValue("cn"),
		MemberOf: memberOf,
	}, nil
}

// SyncOperatorFromLDAP syncs LDAP user to local operator record.
func (s *AuthService) SyncOperatorFromLDAP(ctx context.Context, ldapUser *LDAPUser) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByEmail(ctx, ldapUser.Email)
	if err != nil {
		op = &operator.Operator{
			ID:    shared.GenerateID(),
			Email: ldapUser.Email,
			Name:  ldapUser.Name,
		}
		if createErr := s.operatorRepo.Create(ctx, op); createErr != nil {
			return nil, createErr
		}
		return op, nil
	}

	op.Name = ldapUser.Name
	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// IsLDAPEnabled returns true if LDAP is configured.
func (s *AuthService) IsLDAPEnabled() bool {
	return s.ldapConfig != nil && s.ldapConfig.Server != ""
}
