package wire

// sessionRepoAdapter adapts the domain-backed *storage.SessionRepository to the
// infrastructure/security/session.Repository interface expected by the session
// Manager. The two packages define separate Session structs (security/session
// vs domain/session) with identical fields, so we convert at the boundary.

import (
	"context"
	"time"

	domainsession "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

type sessionRepoAdapter struct {
	repo *storage.SessionRepository
}

func newSessionRepoAdapter(repo *storage.SessionRepository) session.Repository {
	return &sessionRepoAdapter{repo: repo}
}

func toSecuritySession(s *domainsession.Session) *session.Session {
	if s == nil {
		return nil
	}

	return &session.Session{
		ID:                      s.ID,
		OperatorID:              s.OperatorID,
		SelectedOrganizationID:  s.SelectedOrganizationID,
		ExpiresAt:               s.ExpiresAt,
		CreatedAt:               s.CreatedAt,
		IPAddress:               s.IPAddress,
		UserAgent:               s.UserAgent,
	}
}

func toSecuritySessions(s []*domainsession.Session) []*session.Session {
	out := make([]*session.Session, 0, len(s))
	for _, sess := range s {
		out = append(out, toSecuritySession(sess))
	}

	return out
}

func toDomainSession(s *session.Session) *domainsession.Session {
	if s == nil {
		return nil
	}

	return &domainsession.Session{
		ID:                      s.ID,
		OperatorID:              s.OperatorID,
		ExpiresAt:               s.ExpiresAt,
		CreatedAt:               s.CreatedAt,
		SelectedOrganizationID:  s.SelectedOrganizationID,
		IPAddress:               s.IPAddress,
		UserAgent:               s.UserAgent,
	}
}

func (a *sessionRepoAdapter) FindByID(ctx context.Context, id string) (*session.Session, error) {
	s, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toSecuritySession(s), nil
}

func (a *sessionRepoAdapter) FindByOperatorID(ctx context.Context, operatorID string) ([]*session.Session, error) {
	sessions, err := a.repo.FindByOperatorID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return toSecuritySessions(sessions), nil
}

func (a *sessionRepoAdapter) Create(ctx context.Context, s *session.Session) error {
	return a.repo.Create(ctx, toDomainSession(s))
}

func (a *sessionRepoAdapter) Delete(ctx context.Context, id string) error {
	return a.repo.Delete(ctx, id)
}

func (a *sessionRepoAdapter) DeleteByOperatorID(ctx context.Context, operatorID string) error {
	return a.repo.DeleteByOperatorID(ctx, operatorID)
}

func (a *sessionRepoAdapter) DeleteExpired(ctx context.Context) (int, error) {
	return a.repo.DeleteExpired(ctx)
}

func (a *sessionRepoAdapter) Extend(ctx context.Context, id string, newExpiry time.Time) error {
	return a.repo.Extend(ctx, id, newExpiry)
}

func (a *sessionRepoAdapter) RevokeAllOperatorSessions(ctx context.Context, operatorID string) error {
	return a.repo.RevokeAllOperatorSessions(ctx, operatorID)
}
