package resolver

import (
	"context"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	domainoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

// authorizeCommand enforces the same authorization the REST command path does,
// in three layers:
//
//  1. Role gate: scoped permission evaluation — the member's role defaults
//     unioned with any custom per-resource grants must grant command.execute on
//     the devices scope. Viewers lack that grant and are blocked.
//  2. Per-device ownership: a non-admin member may only command devices they
//     own (device.OperatorID); admins and super_admins may command any device
//     in the org. Custom grants can widen this via the evaluator.
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

	eval := r.scopeEvaluator(ctx, op, orgID, string(m.Role))

	// Role gate via scoped permissions: only members granted command.execute on
	// the devices scope may dispatch. Viewers lack that grant.
	if !eval.Grants(permission.ActionCommand, permission.WildcardScope(permission.ScopeDevices)) {
		return cmdapp.AuthorizeOutcome{
			Reason: "insufficient permissions to send commands",
			Tier:   domaincommand.LookupRiskProfile(cmdName).Tier,
		}
	}

	// Per-device ownership: non-admins may only command devices they own, or
	// devices assigned to a group they belong to. Devices without a recorded
	// owner remain org-shared.
	if m.Role.Level() < orgdomain.RoleAdmin.Level() && dev.OperatorID != "" && dev.OperatorID != op.ID {
		if !r.sharesGroupWithDevice(ctx, op, dev.ID) {
			return cmdapp.AuthorizeOutcome{
				Reason: "you do not have access to this device",
				Tier:   domaincommand.LookupRiskProfile(cmdName).Tier,
			}
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

// scopeEvaluator builds the scoped permission evaluator for an operator in an
// org: role defaults unioned with custom grants. Nil grant repo → defaults only.
func (r *Resolver) scopeEvaluator(ctx context.Context, op *domainoperator.Operator, orgID, role string) *permission.Evaluator {
	if r.GrantRepo == nil {
		return permission.NewEvaluator(role, nil)
	}
	grants, err := r.GrantRepo.ListEffective(ctx, op.ID, orgID)
	if err != nil {
		return permission.NewEvaluator(role, nil)
	}
	return permission.NewEvaluator(role, grants)
}

// sharesGroupWithDevice reports whether the device is assigned to a group the
// operator belongs to (team/group-based device scoping, Issue 5).
func (r *Resolver) sharesGroupWithDevice(ctx context.Context, op *domainoperator.Operator, deviceID string) bool {
	if r.GroupRepo == nil {
		return false
	}
	groupIDs, err := r.GroupRepo.GroupIDsForDevice(ctx, deviceID)
	if err != nil {
		return false
	}
	for _, gid := range groupIDs {
		member, err := r.GroupRepo.IsMember(ctx, gid, op.ID)
		if err == nil && member {
			return true
		}
	}
	return false
}
