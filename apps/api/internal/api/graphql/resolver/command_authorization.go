package resolver

import (
	"context"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	domainoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// authorizeCommand enforces the same authorization the REST command path does,
// in three layers:
//
//  1. Role gate: a viewer-tier member may not execute commands at all.
//  2. Per-device ownership: a non-admin member may only command devices they
//     own (device.OperatorID); admins and super_admins may command any device
//     in the org.
//  3. Risk gate: the shared Authorizer applies MFA (critical tier) and
//     confirmation-token requirements identical to REST.
//
// Returning the shared AuthorizeOutcome keeps the GraphQL and REST paths
// behaviorally identical instead of drifting apart.
func (r *Resolver) authorizeCommand(ctx context.Context, op *domainoperator.Operator, orgID string, dev *devicedomain.Device, cmdName, confirmationToken string) cmdapp.AuthorizeOutcome {
	m := op.GetMembership(orgID)
	if m == nil || !m.IsActive() {
		return cmdapp.AuthorizeOutcome{
			Reason: "not a member of this organization",
			Tier:   domaincommand.LookupRiskProfile(cmdName).Tier,
		}
	}

	// Role gate: viewers are read-only.
	if m.Role.Level() < orgdomain.RoleOperator.Level() {
		return cmdapp.AuthorizeOutcome{
			Reason: "insufficient permissions to send commands",
			Tier:   domaincommand.LookupRiskProfile(cmdName).Tier,
		}
	}

	// Per-device ownership: non-admins may only command devices they own.
	// Devices without a recorded owner remain org-shared.
	if m.Role.Level() < orgdomain.RoleAdmin.Level() && dev.OperatorID != "" && dev.OperatorID != op.ID {
		return cmdapp.AuthorizeOutcome{
			Reason: "you do not have access to this device",
			Tier:   domaincommand.LookupRiskProfile(cmdName).Tier,
		}
	}

	// Risk gate (MFA + confirmation), shared with the REST path.
	actor := domaincommand.ActorContext{
		OperatorID:   op.ID,
		OrgID:        orgID,
		IsSuperAdmin: op.IsSuperAdmin(),
		MFAVerified:  gqlcontext.GetMFAVerified(ctx),
	}
	return r.Authorizer.Authorize(ctx, actor, cmdName, dev.ID, confirmationToken)
}
