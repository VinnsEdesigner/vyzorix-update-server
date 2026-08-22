package channel

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// MembershipServiceAdapter adapts the member service to the channel package's
// MemberLookup.
type MembershipServiceAdapter struct {
	service MembershipService
}

// MembershipService matches the member service contract.
type MembershipService interface {
	GetMembership(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error)
}

// NewMembershipLookup wraps the member service for channel authorization.
func NewMembershipLookup(service MembershipService) MemberLookup {
	return &MembershipServiceAdapter{service: service}
}

// FindByOperatorAndOrg returns the operator's membership.
func (a *MembershipServiceAdapter) FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
	if a.service == nil {
		return nil, ErrForbidden
	}
	return a.service.GetMembership(ctx, operatorID, orgID)
}
