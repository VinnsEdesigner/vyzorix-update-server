
// Package confirmation provides the business logic for issuing and consuming
// single-use confirmation tokens that gate risky device commands.
package confirmation

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
)

// DefaultConfirmationTTL is used when the command's risk profile does not
// specify a TTL. It mirrors the catalog's DefaultConfirmationTTL.
const DefaultConfirmationTTL = 5 * time.Minute

// Service issues and consumes confirmation tokens. It is the only writer to
// the confirmation store; the command execution handler consumes tokens via
// ConsumeForCommand.
type Service struct {
	repo domainconfirmation.Repository
}

// NewService creates a confirmation Service backed by the given repository.
func NewService(repo domainconfirmation.Repository) *Service {
	return &Service{repo: repo}
}

// RequestConfirmation issues a single-use token authorizing the given
// operator to execute the named command (optionally on a specific device).
// The TTL is taken from the command's risk profile, falling back to
// DefaultConfirmationTTL. The returned token must be presented back within
// the TTL via ConsumeForCommand.
func (s *Service) RequestConfirmation(ctx context.Context, operatorID, orgID, commandName, deviceID string) (*domainconfirmation.PendingConfirmation, error) {
	if operatorID == "" || commandName == "" {
		return nil, errors.New("operatorID and commandName are required")
	}

	profile := command.LookupRiskProfile(commandName)
	ttl := profile.ConfirmationTTL
	if ttl <= 0 {
		ttl = DefaultConfirmationTTL
	}

	now := time.Now()
	c := &domainconfirmation.PendingConfirmation{
		Token:      shared.GenerateID(),
		OperatorID: operatorID,
		OrgID:      orgID,
		Command:    commandName,
		DeviceID:   deviceID,
		RiskTier:   string(profile.Tier),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// ConsumeForCommand validates and consumes a confirmation token for a specific
// operator/command/device execution. It enforces ownership, command match,
// device scope, expiry, and single-use semantics, marking the token consumed
// on success. Returns the consumed confirmation so the caller can include its
// risk tier / timing in audit records.
func (s *Service) ConsumeForCommand(ctx context.Context, token, operatorID, commandName, deviceID string) (*domainconfirmation.PendingConfirmation, error) {
	c, err := s.repo.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if c.IsExpired() {
		return c, domainconfirmation.ErrExpired
	}
	if !c.Matches(operatorID, commandName, deviceID) {
		return c, domainconfirmation.ErrMismatch
	}
	if c.IsConsumed() {
		return c, domainconfirmation.ErrAlreadyConsumed
	}
	return s.repo.Consume(ctx, token, time.Now())
}